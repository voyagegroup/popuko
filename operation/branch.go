package operation

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/v28/github"
)

func createAutoBranch(ctx context.Context, svc *github.GitService, owner string, repo string, number int, branchName string) (ok bool, ref *github.Reference) {
	refName := "refs/heads/" + branchName

	log.Printf("info: clean up %v by deleting it\n", refName)
	if _, err := svc.DeleteRef(ctx, owner, repo, refName); err != nil {
		log.Printf("info: could not clean up %v by %v, but we continue to create %v optimistically\n", refName, err, refName)
	}

	// see:
	// https://github.com/voyagegroup/popuko/issues/93
	// https://help.github.com/articles/checking-out-pull-requests-locally/
	base := fmt.Sprintf("refs/pull/%d/merge", number)
	log.Printf("debug: `ref` is: %v\n", base)
	ref, _, err := svc.GetRef(ctx, owner, repo, base)
	if err != nil {
		log.Printf("warn: cannot get reference about %v\n", base)
		return
	}

	branchRef := github.Reference{
		Ref:    &refName,
		URL:    nil, // XXX: This field is unused on creating ref.
		Object: ref.Object,
	}

	ref, _, err = svc.CreateRef(ctx, owner, repo, &branchRef)
	if err != nil {
		log.Printf("warn: cannot create a new ref %v\n", refName)
		return
	}

	return true, ref
}

func TryWithDefaultBranch(ctx context.Context, client *github.Client, owner string, name string, info *github.PullRequest, autoBranch string) (bool, string) {
	number := *info.Number

	ok, ref := createAutoBranch(ctx, client.Git, owner, name, number, autoBranch)
	if !ok {
		log.Println("info: cannot create the auto branch")
		return false, ""
	}
	log.Println("info: create the auto branch")

	sha := *ref.Object.SHA

	{
		number := *info.Number
		headSha := *info.Head.SHA
		c := ":hourglass: " + headSha + " has been merged into the auto branch " + sha
		if ok := AddComment(ctx, client.Issues, owner, name, number, c); !ok {
			log.Println("info: could not create the comment to declare to merge this.")
		}
	}

	return true, sha
}

func DeleteBranchByPullRequest(ctx context.Context, svc *github.GitService, pr *github.PullRequest) (bool, error) {
	owner := *pr.Head.Repo.Owner.Login
	log.Printf("debug: branch owner: %v\n", owner)
	repo := *pr.Head.Repo.Name
	log.Printf("debug: repo: %v\n", repo)
	branch := *pr.Head.Ref
	log.Printf("debug: head ref: %v\n", branch)

	_, err := svc.DeleteRef(ctx, owner, repo, "heads/"+branch)
	if err != nil {
		log.Println("info: could not delete the merged branch.")
		return false, err
	}

	return true, nil
}

func MergePullRequest(ctx context.Context, client *github.Client, owner string, name string, info *github.PullRequest, acceptedSha string) bool {
	number := *info.Number

	// Even if we checks the head at here, the new commits may be pushed from user
	// before we merge it actually. To prevent to such case, we also set `github.PullRequestOptions.SHA`.
	if acceptedSha != *info.Head.SHA {
		CommentHeadIsDifferentFromAccepted(ctx, client.Issues, owner, name, number)
		return false
	}
	option := &github.PullRequestOptions{
		SHA: acceptedSha, // To ensure that we only accept the accepted changeset.
	}

	// XXX: By the behavior, github uses defautlt merge message
	// if we specify `""` to `commitMessage`.
	_, _, err := client.PullRequests.Merge(ctx, owner, name, number, "", option)
	if err != nil {
		log.Println("warn: could not merge pull request")
		comment := ":skull:　Could not merge this pull request by:\n```\n" + err.Error() + "\n```"
		if ok := AddComment(ctx, client.Issues, owner, name, number, comment); !ok {
			log.Println("warn: could not create the comment to express no merging the pull request")
		}
		return false
	}

	return true
}

func IsIncludeAutoBranch(branches []*github.Branch, auto string) bool {
	for _, b := range branches {
		if b == nil {
			continue
		}

		if b.Name == nil {
			continue
		}

		if *b.Name == auto {
			return true
		}
	}

	return false
}
