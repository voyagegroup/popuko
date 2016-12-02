package main

import (
	"log"
	"sync"

	"github.com/google/go-github/github"
)

func (srv *AppServer) detectUnmergeablePR(ev *github.PushEvent) {
	if *ev.Ref != "refs/heads/master" {
		log.Println(*ev.Ref)
		return
	}

	repoOwner := *ev.Repo.Owner.Name
	repo := *ev.Repo.Name

	client := srv.githubClient
	prSvc := client.PullRequests

	prList, _, err := prSvc.List(repoOwner, repo, &github.PullRequestListOptions{
		State: "open",
	})
	if err != nil {
		log.Println("could not fetch opened pull requests")
		return
	}

	compare := *ev.Compare
	comment := ":umbrella: The latest upstream [changeset](" + compare + ") made this pull request unmergeable. Please resolve the merge conflicts."
	wg := &sync.WaitGroup{}
	for _, item := range prList {
		wg.Add(1)

		go markUnmergeable(wg, client.Issues, &markUnmergeableInfo{
			repoOwner,
			repo,
			*item.Number,
			comment,
		})
	}
	wg.Wait()
}

type markUnmergeableInfo struct {
	RepoOwner string
	Repo      string
	Number    int
	Comment   string
}

func markUnmergeable(wg *sync.WaitGroup, issueSvc *github.IssuesService, info *markUnmergeableInfo) {
	var err error
	defer wg.Done()
	defer func() {
		if err != nil {
			log.Printf("error: %v\n", err)
		}
	}()

	repoOwner := info.RepoOwner
	repo := info.Repo
	number := info.Number

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repo, number, nil)
	if err != nil {
		log.Println("could not get labels by the issue")
		return
	}

	// We don't have to warn to a pull request which have been marked as unmergeable.
	if hasStatusLabel(currentLabels, LABEL_NEEDS_REBASE) {
		return
	}

	_, _, err = issueSvc.CreateComment(repoOwner, repo, number, &github.IssueComment{
		Body: &info.Comment,
	})
	if err != nil {
		log.Println("could not create the comment to unmergeables")
		return
	}

	labels := addNeedRebaseLabel(currentLabels)
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repo, number, labels)
	if err != nil {
		log.Println("could not change labels of the issue")
		return
	}
}
