package main

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

type AcceptCommand struct {
	client *github.Client
	repo   *RepositorySetting
}

func (c *AcceptCommand) commandAcceptChangesetByReviewer(ev *github.IssueCommentEvent) (bool, error) {
	fmt.Printf("Start: assign the reviewer by %v\n", ev.Comment.ID)
	defer fmt.Printf("End: assign the reviewer by %v\n", ev.Comment.ID)

	sender := *ev.Sender.Login
	if !c.repo.Reviewers().Has(sender) {
		fmt.Printf("%v is not an reviewer registred to this bot.\n", sender)
		return false, nil
	}

	client := c.client
	issueSvc := client.Issues

	repoOwner := c.repo.Owner()
	repoName := c.repo.Name()
	issue := *ev.Issue.Number

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repoName, issue, nil)
	if err != nil {
		fmt.Println("could not get labels by the issue")
		return false, err
	}
	labels := addAwaitingMergeLabel(currentLabels)

	// https://github.com/nekoya/popuko/blob/master/web.py
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repoName, issue, labels)
	if err != nil {
		fmt.Println("could not change labels by the issue")
		return false, err
	}

	{
		comment := "Try to merge this pull request which has been approved by `" + sender + "`"
		_, _, err := issueSvc.CreateComment(repoOwner, repoName, issue, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			fmt.Println("could not create the comment to declare to merge this.")
			return false, err
		}
	}

	// XXX: the commit comment should be default?
	prSvc := client.PullRequests
	_, _, err = prSvc.Merge(repoOwner, repoName, issue, "", nil)
	if err != nil {
		fmt.Println("could not merge pull request")
		comment := "Could not merge this pull request by:\n```\n" + err.Error() + "\n```"
		_, _, err := issueSvc.CreateComment(repoOwner, repoName, issue, &github.IssueComment{
			Body: &comment,
		})
		return false, err
	}

	// delete branch
	/*
		pr, _, err := prSvc.Get(repoOwner, repoName, issue)
		if err != nil {
			fmt.Println("could not fetch the pull request information.")
			return false, err
		}

		fmt.Printf("sender: %v\n", sender)
		fmt.Printf("repo: %v\n", repoName)
		fmt.Printf("head ref: %v\n", *pr.Head.Ref)

		_, err = client.Git.DeleteRef(repoOwner, repoName, *pr.Head.Ref)
		if err != nil {
			fmt.Println("could not delete the merged branch.")
			return false, err
		}
	*/

	return true, nil
}

func (c *AcceptCommand) commandAcceptChangesetByOtherReviewer(ev *github.IssueCommentEvent, command string) (bool, error) {
	tmp := strings.Split(command, "=")
	if len(tmp) < 2 {
		fmt.Println("No specify who is the actual reviewer.")
		return false, nil
	}

	actual := tmp[1]
	if !c.repo.Reviewers().Has(actual) {
		fmt.Println("could not find the actual reviewer in reviewer list")
		return false, nil
	}

	return c.commandAcceptChangesetByReviewer(ev)
}
