package operation

import (
	"log"

	"github.com/google/go-github/github"
)

func CreateBranchFromMaster(svc *github.GitService, owner string, repo string, branchName string) (ok bool, ref *github.Reference) {
	refName := "refs/heads/" + branchName

	log.Printf("info: clean up %v by deleting it\n", refName)
	if _, err := svc.DeleteRef(owner, repo, refName); err != nil {
		log.Printf("info: could not clean up %v by %v, but we continue to create %v optimistically\n", refName, err, refName)
	}

	const base string = "refs/heads/master"
	ref, _, err := svc.GetRef(owner, repo, base)
	if err != nil {
		log.Printf("warn: cannot get reference about %v\n", base)
		return
	}

	branchRef := github.Reference{
		Ref:    &refName,
		URL:    nil, // XXX: This field is unused on creating ref.
		Object: ref.Object,
	}

	ref, _, err = svc.CreateRef(owner, repo, &branchRef)
	if err != nil {
		log.Printf("warn: cannot create a new ref %v\n", refName)
		return
	}

	return true, ref
}

func MergeIntoAutoBranch(svc *github.RepositoriesService, owner string, repo string, head *github.PullRequestBranch) (ok bool, commit *github.RepositoryCommit) {
	base := "auto"
	message := "Auto merging"
	req := github.RepositoryMergeRequest{
		Base:          &base,
		Head:          head.SHA,
		CommitMessage: &message,
	}

	commit, _, err := svc.Merge(owner, repo, &req)
	if err != nil {
		log.Printf("warn: could not merge '%v' branch into master\n", base)
		return
	}

	return true, commit
}
