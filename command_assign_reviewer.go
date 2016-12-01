package main

import (
	"fmt"

	"github.com/google/go-github/github"
)

func (srv *AppServer) commandAssignReviewer(ev *github.IssueCommentEvent, target string) (bool, error) {
	fmt.Printf("Start: assign the reviewer by %v\n", *ev.Comment.ID)
	defer fmt.Printf("End: assign the reviewer by %v\n", *ev.Comment.ID)

	client := srv.githubClient
	issueSvc := client.Issues

	repoOwner := *ev.Repo.Owner.Login
	repo := *ev.Repo.Name
	issue := *ev.Issue.Number

	assignees := []string{target}

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repo, issue, nil)
	if err != nil {
		return false, err
	}

	_, _, err = issueSvc.AddAssignees(repoOwner, repo, issue, assignees)
	if err != nil {
		return false, err
	}

	labels := addAwaitingReviewLabel(currentLabels)
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repo, issue, labels)
	if err != nil {
		return false, err
	}

	return true, nil
}
