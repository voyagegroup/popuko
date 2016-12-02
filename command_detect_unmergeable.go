package main

import (
	"log"
	"sync"

	"github.com/google/go-github/github"
)

func (srv *AppServer) detectUnmergeablePR(ev *github.PushEvent) {
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
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := info.Repo
	log.Printf("debug: repository name is %v\n", repo)
	number := info.Number
	log.Printf("debug: pull request number is %v\n", number)

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repo, number, nil)
	if err != nil {
		log.Println("info: could not get labels by the issue")
		log.Printf("debug: %v\n", err)
		return
	}
	log.Printf("debug: the current labels: %v\n", currentLabels)

	// We don't have to warn to a pull request which have been marked as unmergeable.
	if hasStatusLabel(currentLabels, LABEL_NEEDS_REBASE) {
		log.Println("info: The pull request has marked as 'shold rebase on the latest master'")
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
	log.Printf("debug: the changed labels: %v\n", labels)
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repo, number, labels)
	if err != nil {
		log.Println("could not change labels of the issue")
		return
	}
}
