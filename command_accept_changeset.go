package main

import (
	"log"

	"github.com/google/go-github/github"
)

type AcceptCommand struct {
	client    *github.Client
	repo      *RepositorySetting
	reviewers *ReviewerSet
}

func (c *AcceptCommand) commandAcceptChangesetByReviewer(ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("info: Start: merge the pull request by %v\n", *ev.Comment.ID)
	defer log.Printf("info: End: merge the pull request by %v\n", *ev.Comment.ID)

	sender := *ev.Sender.Login
	log.Printf("debug: command is sent from %v\n", sender)

	if !c.reviewers.Has(sender) {
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

	if c.repo.ShouldMergeAutomatically() {
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

		// XXX: By the behavior, github uses defautlt merge message
		// if we specify `""` to `commitMessage`.
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
		if c.repo.ShouldDeleteMerged() {
			pr, _, err := prSvc.Get(repoOwner, repoName, issue)
			if err != nil {
				log.Println("info: could not fetch the pull request information.")
				return false, err
			}

			branchOwner := *pr.Head.Repo.Owner.Login
			log.Printf("debug: branch owner: %v\n", branchOwner)
			branchOwnerRepo := *pr.Head.Repo.Name
			log.Printf("debug: repo: %v\n", branchOwnerRepo)
			branchName := *pr.Head.Ref
			log.Printf("debug: head ref: %v\n", branchName)

			_, err = client.Git.DeleteRef(branchOwner, branchOwnerRepo, "heads/"+branchName)
			if err != nil {
				log.Println("info: could not delete the merged branch.")
				return false, err
			}
		}
	}

	log.Printf("info: complete merge the pull request %v\n", issue)
	return true, nil
}

func (c *AcceptCommand) commandAcceptChangesetByOtherReviewer(ev *github.IssueCommentEvent, reviewer string) (bool, error) {
	log.Printf("info: Start: merge the pull request from other reviewer by %v\n", ev.Comment.ID)
	defer log.Printf("info: End:merge the pull request from other reviewer by %v\n", ev.Comment.ID)

	if !c.reviewers.Has(reviewer) {
		log.Println("info: could not find the actual reviewer in reviewer list")
		log.Printf("debug: specified actial reviewer %v\n", reviewer)
		return false, nil
	}

	return c.commandAcceptChangesetByReviewer(ev)
}
