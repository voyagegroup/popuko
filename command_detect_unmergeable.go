package main

import (
	"log"
	"sync"
	"time"

	"github.com/google/go-github/github"

	"github.com/karen-irc/popuko/operation"
)

func (srv *AppServer) detectUnmergeablePR(ev *github.PushEvent) {
	// At this moment, we only care a pull request which are looking master branch.
	if *ev.Ref != "refs/heads/master" {
		log.Printf("info: pushed branch is not related to me: %v\n", *ev.Ref)
		return
	}

	repoOwner := *ev.Repo.Owner.Name
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)

	client := srv.githubClient
	prSvc := client.PullRequests

	prList, _, err := prSvc.List(repoOwner, repo, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		log.Println("warn: could not fetch opened pull requests")
		return
	}

	compare := *ev.Compare
	comment := ":umbrella: The latest upstream change (presumably [these](" + compare + ")) made this pull request unmergeable. Please resolve the merge conflicts."
	wg := &sync.WaitGroup{}
	for _, item := range prList {
		wg.Add(1)

		go markUnmergeable(wg, &markUnmergeableInfo{
			client.Issues,
			prSvc,
			repoOwner,
			repo,
			*item.Number,
			comment,
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
}

func markUnmergeable(wg *sync.WaitGroup, info *markUnmergeableInfo) {
	var err error
	defer wg.Done()
	defer func() {
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

	currentLabels := operation.GetLabelsByIssue(issueSvc, repoOwner, repo, number)
	if currentLabels == nil {
		return
	}

	// We don't have to warn to a pull request which have been marked as unmergeable.
	if operation.HasLabelInList(currentLabels, operation.LABEL_NEEDS_REBASE) {
		log.Println("info: The pull request has marked as 'shold rebase on the latest master'")
		return
	}

	ok, mergeable := isMergeable(info.prSvc, repoOwner, repo, number)
	if !ok {
		log.Println("info: We treat it as 'mergeable' to avoid miss detection if we could not fetch the pr info,")
		return
	}

	if mergeable {
		log.Println("info: do not have to mark as 'unmergeable'")
		return
	}

	if ok := operation.AddComment(issueSvc, repoOwner, repo, number, info.Comment); !ok {
		log.Println("info: could not create the comment to unmergeables")
		return
	}

	labels := operation.AddNeedRebaseLabel(currentLabels)
	log.Printf("debug: the changed labels: %v\n", labels)
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repo, number, labels)
	if err != nil {
		log.Println("could not change labels of the issue")
		return
	}
}

func isMergeable(prSvc *github.PullRequestsService, owner string, name string, issue int) (bool, bool) {
	ok, pr := getPrInfo(prSvc, owner, name, issue)
	if !ok || pr == nil {
		return false, false
	}

	mergeable := pr.Mergeable
	if mergeable == nil {
		// By the document https://developer.github.com/v3/pulls/#get-a-single-pull-request
		// this state is still in checking.

		// sleep same time: https://github.com/barosl/homu/blob/2104e4b154d2fba15d515b478a5bd6105c1522f6/homu/main.py#L722
		time.Sleep(5 * time.Second)
		ok, pr := getPrInfo(prSvc, owner, name, issue)
		if !ok || pr == nil {
			// Conclude it is not mergeable heurístically
			return true, true
		}

		mergeable = pr.Mergeable
		if mergeable == nil {
			return true, false
		}
	}

	return true, *mergeable
}

func getPrInfo(prSvc *github.PullRequestsService, owner string, name string, issue int) (ok bool, info *github.PullRequest) {
	pr, _, err := prSvc.Get(owner, name, issue)
	if err != nil {
		log.Println("info: could not get the info for pull request")
		log.Printf("debug: %v\n", err)
		return false, nil
	}

	return true, pr
}
