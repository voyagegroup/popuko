package main

import (
	"log"
	"strings"

	"github.com/google/go-github/github"
)

type AcceptCommand struct {
	client *github.Client
	repo   *RepositorySetting
}

func (c *AcceptCommand) commandAcceptChangesetByReviewer(ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("info: Start: merge the pull request by %v\n", ev.Comment.ID)
	defer log.Printf("info: End:merge the pull request by %v\n", ev.Comment.ID)

	sender := *ev.Sender.Login
	log.Printf("debug: command is sent from %v\n", sender)

	if !c.repo.Reviewers().Has(sender) {
		log.Printf("info: %v is not an reviewer registred to this bot.\n", sender)
		return false, nil
	}

	client := c.client
	issueSvc := client.Issues

	repoOwner := c.repo.Owner()
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repoName := c.repo.Name()
	log.Printf("debug: repository name is %v\n", repoName)
	issue := *ev.Issue.Number
	log.Printf("debug: issue number is %v\n", issue)

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repoName, issue, nil)
	if err != nil {
		log.Println("info: could not get labels by the issue")
		return false, err
	}
	labels := addAwaitingMergeLabel(currentLabels)

	// https://github.com/nekoya/popuko/blob/master/web.py
	_, _, err = issueSvc.ReplaceLabelsForIssue(repoOwner, repoName, issue, labels)
	if err != nil {
		log.Println("info: could not change labels by the issue")
		return false, err
	}

	{
		comment := "Try to merge this pull request which has been approved by `" + sender + "`"
		_, _, err := issueSvc.CreateComment(repoOwner, repoName, issue, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			log.Println("info: could not create the comment to declare to merge this.")
			return false, err
		}
	}

	// XXX: the commit comment should be default?
	prSvc := client.PullRequests
	_, _, err = prSvc.Merge(repoOwner, repoName, issue, "", nil)
	if err != nil {
		log.Println("info: could not merge pull request")
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
			log.Println("could not fetch the pull request information.")
			return false, err
		}

		log.Printf("sender: %v\n", sender)
		log.Printf("repo: %v\n", repoName)
		log.Printf("head ref: %v\n", *pr.Head.Ref)

		_, err = client.Git.DeleteRef(repoOwner, repoName, *pr.Head.Ref)
		if err != nil {
			log.Println("could not delete the merged branch.")
			return false, err
		}
	*/

	return true, nil
}

func (c *AcceptCommand) commandAcceptChangesetByOtherReviewer(ev *github.IssueCommentEvent, command string) (bool, error) {
	log.Printf("info: Start: merge the pull request from other reviewer by %v\n", ev.Comment.ID)
	defer log.Printf("info: End:merge the pull request from other reviewer by %v\n", ev.Comment.ID)

	tmp := strings.Split(command, "=")
	if len(tmp) < 2 {
		log.Println("info: No specify who is the actual reviewer.")
		return false, nil
	}

	actual := tmp[1]
	if !c.repo.Reviewers().Has(actual) {
		log.Println("info: could not find the actual reviewer in reviewer list")
		log.Printf("debug: specified actial reviewer %v\n", actual)
		return false, nil
	}

	return c.commandAcceptChangesetByReviewer(ev)
}
