package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

func (srv *AppServer) commandAcceptChangesetByReviewer(ev *github.IssueCommentEvent) (bool, error) {
	fmt.Printf("Start: assign the reviewer by %v\n", ev.Comment.ID)
	defer fmt.Printf("End: assign the reviewer by %v\n", ev.Comment.ID)

	sender := *ev.Sender.Login
	if !config.Reviewers().Has(sender) {
		fmt.Printf("%v is not an reviewer registred to this bot.\n", sender)
		return false, nil
	}

	client := srv.githubClient
	issueSvc := client.Issues

	repoOwner := *ev.Repo.Owner.Login
	repo := *ev.Repo.Name
	issue := *ev.Issue.Number

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repo, issue, nil)
	if err != nil {
		fmt.Println("could not get labels by the issue")
		return false, err
	}
	labels := addAwaitingMergeLabel(currentLabels)

	// https://github.com/nekoya/popuko/blob/master/web.py
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repo, issue, labels)
	if err != nil {
		fmt.Println("could not change labels by the issue")
		return false, err
	}

	{
		comment := "Try to merge this pull request which has been approved by `" + sender + "`"
		_, _, err := issueSvc.CreateComment(repoOwner, repo, issue, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			fmt.Println("could not create the comment to declare to merge this.")
			return false, err
		}
	}

	// XXX: the commit comment should be default?
	prSvc := client.PullRequests
	_, _, err = prSvc.Merge(repoOwner, repo, issue, "", nil)
	if err != nil {
		fmt.Println("could not merge pull request")
		comment := "Could not merge this pull request by:\n```\n" + err.Error() + "\n```"
		_, _, err := issueSvc.CreateComment(repoOwner, repo, issue, &github.IssueComment{
			Body: &comment,
		})
		return false, err
	}

	// delete branch
	/*
		pr, _, err := prSvc.Get(repoOwner, repo, issue)
		if err != nil {
			fmt.Println("could not fetch the pull request information.")
			return false, err
		}

		fmt.Printf("sender: %v\n", sender)
		fmt.Printf("repo: %v\n", repo)
		fmt.Printf("head ref: %v\n", *pr.Head.Ref)

		_, err = client.Git.DeleteRef(repoOwner, repo, *pr.Head.Ref)
		if err != nil {
			fmt.Println("could not delete the merged branch.")
			return false, err
		}
	*/

	return true, nil
}

func (srv *AppServer) commandAcceptChangesetByOtherReviewer(ev *github.IssueCommentEvent, command string) (bool, error) {
	tmp := strings.Split(command, "=")
	if len(tmp) < 2 {
		fmt.Println("No specify who is the actual reviewer.")
		return false, nil
	}

	actual := tmp[1]
	if !config.Reviewers().Has(actual) {
		fmt.Println("could not find the actual reviewer in reviewer list")
		return false, nil
	}

	return srv.commandAcceptChangesetByReviewer(ev)
}
