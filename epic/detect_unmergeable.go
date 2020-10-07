package epic

import (
	"context"
	"log"
	"sync"

	"github.com/google/go-github/v28/github"

	"github.com/voyagegroup/popuko/operation"
)

func DetectUnmergeablePR(ctx context.Context, client *github.Client, ev *github.PushEvent) {
	owner := *ev.Repo.Owner.Name
	log.Printf("debug: repository owner is %v\n", owner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)

	fullRepositoryName := owner + "/" + repo
	defaultBranch := ev.Repo.GetDefaultBranch()
	if defaultBranch == "" {
		// I seem we may not require this fallback path. But GitHub API example is not reliable.
		log.Printf("debug: could not get default branch name from pushed event for %v\n", fullRepositoryName)
		repoInfo, _, err := client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			log.Printf("warn: could not fetch the repository infor for %v by %v\n", fullRepositoryName, err)
			return
		}

		defaultBranch = repoInfo.GetDefaultBranch()
	}
	log.Printf("info: the default branch name is `%v` for %v\n", defaultBranch, fullRepositoryName)

	// At this moment, we only care a pull request which are looking master branch.
	if *ev.Ref != "refs/heads/"+defaultBranch {
		log.Printf("info: pushed branch is not related to me: %v\n", *ev.Ref)
		return
	}

	prSvc := client.PullRequests

	prList, _, err := prSvc.List(ctx, owner, repo, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		log.Println("warn: could not fetch opened pull requests")
		return
	}

	compare := *ev.Compare
	comment := ":umbrella: The latest upstream change (presumably [these](" + compare + ")) made this pull request unmergeable. Please resolve the merge conflicts."

	// Restrict the number of Goroutine which checks unmergeables
	// to avoid the API limits at a moment.
	const maxConcurrency int = 8
	semaphore := make(chan int, maxConcurrency)

	wg := &sync.WaitGroup{}
	for _, item := range prList {
		wg.Add(1)

		go markUnmergeable(ctx, wg, &markUnmergeableInfo{
			issueSvc:          client.Issues,
			prSvc:             prSvc,
			RepoOwner:         owner,
			Repo:              repo,
			DefaultBranchName: defaultBranch,
			Number:            *item.Number,
			Comment:           comment,
			semaphore:         semaphore,
		})
	}
	wg.Wait()
}

type markUnmergeableInfo struct {
	issueSvc          *github.IssuesService
	prSvc             *github.PullRequestsService
	RepoOwner         string
	Repo              string
	DefaultBranchName string
	Number            int
	Comment           string
	semaphore         chan int
}

func markUnmergeable(ctx context.Context, wg *sync.WaitGroup, info *markUnmergeableInfo) {
	info.semaphore <- 0 // wait until the internal buffer takes a space.

	var err error
	defer wg.Done()
	defer func() {
		<-info.semaphore // release the space of the internal buffer

		if err != nil {
			log.Printf("error: %v\n", err)
		}
	}()

	issueSvc := info.issueSvc

	repoOwner := info.RepoOwner
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := info.Repo
	log.Printf("debug: repository name is %v\n", repo)
	number := info.Number
	log.Printf("debug: pull request number is %v\n", number)
	defaultBranchName := info.DefaultBranchName
	log.Printf("debug: the default branch name is %v\n", defaultBranchName)

	pr, _, err := info.prSvc.Get(ctx, repoOwner, repo, number)
	if err != nil || pr == nil {
		log.Println("info: could not get the info for pull request")
		log.Printf("debug: %v\n", err)
		return
	}

	if !operation.IsRelatedToDefaultBranch(pr, repoOwner, defaultBranchName) {
		log.Printf("info: #%v is not related to `%v` branch", number, defaultBranchName)
		return
	}

	currentLabels := operation.GetLabelsByIssue(ctx, issueSvc, repoOwner, repo, number)
	if currentLabels == nil {
		return
	}

	// We don't have to warn to a pull request which have been marked as unmergeable.
	if operation.HasLabelInList(currentLabels, operation.LABEL_NEEDS_REBASE) {
		log.Printf("info: #%v has marked as 'should rebase on the latest master'.\n", number)
		return
	}

	ok, mergeable := operation.IsMergeable(ctx, info.prSvc, repoOwner, repo, number, pr)
	if !ok {
		log.Printf("info: We treat #%v as 'mergeable' to avoid miss detection because we could not fetch the pr info,\n", number)
		return
	}

	if mergeable {
		log.Printf("info: do not have to mark %v as 'unmergeable'\n", number)
		return
	}

	if ok := operation.AddComment(ctx, issueSvc, repoOwner, repo, number, info.Comment); !ok {
		log.Printf("info: could not create the comment about unmergeables to #%v\n", number)
		return
	}

	labels := operation.AddNeedRebaseLabel(currentLabels)
	log.Printf("debug: the changed labels: %v of #%v\n", labels, number)
	_, _, err = issueSvc.ReplaceLabelsForIssue(ctx, repoOwner, repo, number, labels)
	if err != nil {
		log.Printf("could not change labels of #%v\n", number)
		return
	}
}
