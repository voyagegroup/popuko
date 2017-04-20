package epic

import (
	"context"
	"log"
	"sync"

	"github.com/google/go-github/github"

	"github.com/voyagegroup/popuko/operation"
)

const masterBranchName = "master"

func DetectUnmergeablePR(ctx context.Context, client *github.Client, ev *github.PushEvent) {
	// At this moment, we only care a pull request which are looking master branch.
	if *ev.Ref != "refs/heads/"+masterBranchName {
		log.Printf("info: pushed branch is not related to me: %v\n", *ev.Ref)
		return
	}

	repoOwner := *ev.Repo.Owner.Name
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)

	prSvc := client.PullRequests

	prList, _, err := prSvc.List(ctx, repoOwner, repo, &github.PullRequestListOptions{
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
			client.Issues,
			prSvc,
			repoOwner,
			repo,
			*item.Number,
			comment,
			semaphore,
		})
	}
	wg.Wait()
}

type markUnmergeableInfo struct {
	issueSvc  *github.IssuesService
	prSvc     *github.PullRequestsService
	RepoOwner string
	Repo      string
	Number    int
	Comment   string
	semaphore chan int
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

	pr, _, err := info.prSvc.Get(ctx, repoOwner, repo, number)
	if err != nil || pr == nil {
		log.Println("info: could not get the info for pull request")
		log.Printf("debug: %v\n", err)
		return
	}

	if !operation.IsRelatedToMaster(pr, repoOwner, masterBranchName) {
		log.Printf("info: #%v is not related to `%v` branch", number, masterBranchName)
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
