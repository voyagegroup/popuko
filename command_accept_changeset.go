package main

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/karen-irc/popuko/operation"
)

type AcceptCommand struct {
	owner string
	name  string

	client  *github.Client
	botName string
	cmd     AcceptChangesetCommand
	info    *repositoryInfo

	queue *autoMergeQueue
}

func (c *AcceptCommand) commandAcceptChangesetByReviewer(ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("info: Start: merge the pull request by %v\n", *ev.Comment.ID)
	defer log.Printf("info: End: merge the pull request by %v\n", *ev.Comment.ID)

	if c.botName != c.cmd.BotName() {
		log.Printf("info: this command works only if target user is actual our bot.")
		return false, nil
	}

	sender := *ev.Sender.Login
	log.Printf("debug: command is sent from %v\n", sender)

	if !c.info.isReviewer(sender) {
		log.Printf("info: %v is not an reviewer registred to this bot.\n", sender)
		return false, nil
	}

	client := c.client
	issueSvc := client.Issues

	repoOwner := c.owner
	repoName := c.name
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

	prSvc := client.PullRequests
	pr, _, err := prSvc.Get(repoOwner, repoName, issue)
	if err != nil {
		log.Println("info: could not fetch the pull request information.")
		return false, err
	}

	headSha := *pr.Head.SHA
	{
		comment := ":pushpin: Commit " + headSha + " has been approved by `" + sender + "`"
		if ok := operation.AddComment(issueSvc, repoOwner, repoName, issue, comment); !ok {
			log.Println("info: could not create the comment to declare the head is approved.")
			return false, nil
		}
	}

	if c.info.EnableAutoMerge {
		if c.info.ExperimentalTryOnAutoBranch() {
			c.queue.Lock()
			defer c.queue.Unlock()

			q := &autoMergeQueueItem{
				PullRequest: issue,
				SHA:         nil,
			}
			c.queue.Push(q)

			if c.queue.HasActive() {
				log.Printf("info: pull request (%v) has been queued but other is active.\n", issue)
				{
					comment := ":postbox: This pull request is queued. Please await the time."
					if ok := operation.AddComment(issueSvc, repoOwner, repoName, issue, comment); !ok {
						log.Println("info: could not create the comment to declare to merge this.")
					}
				}
				return true, nil
			}

			ok, next := c.queue.GetNext()
			if !ok || next == nil {
				log.Println("error: this queue should not be empty because `q` is queued just now.")
				return false, nil
			}

			if next != q {
				log.Println("error: `next` should be equal to `q` because there should be only `q` in queue.")
				return false, nil
			}

			ok, _ = createBranchFromMaster(client.Git, repoOwner, repoName, "auto")
			if !ok {
				log.Println("info: cannot create the auto branch")
				return false, nil
			}
			log.Println("info: create the auto branch")

			ok, commit := mergeIntoAutoBranch(client.Repositories, repoOwner, repoName, pr.Head)
			if !ok {
				log.Println("info: cannot merge into the auto branch")
				return false, nil
			}

			log.Println("info: merge the auto branch")

			q.SHA = commit.SHA
			c.queue.SetActive(q)
			log.Printf("info: pin #%v as the active item to queue\n", issue)

			{
				comment := ":hourglass: " + headSha + " has been merged into the auto branch " + *commit.HTMLURL
				if ok := operation.AddComment(issueSvc, repoOwner, repoName, issue, comment); !ok {
					log.Println("info: could not create the comment to declare to merge this.")
					return false, nil
				}
			}
		} else {
			{
				comment := ":hourglass: Try to merge " + headSha
				if ok := operation.AddComment(issueSvc, repoOwner, repoName, issue, comment); !ok {
					log.Println("info: could not create the comment to declare to merge this.")
					return false, nil
				}
			}

			// XXX: By the behavior, github uses defautlt merge message
			// if we specify `""` to `commitMessage`.
			_, _, err = prSvc.Merge(repoOwner, repoName, issue, "", nil)
			if err != nil {
				log.Println("info: could not merge pull request")
				comment := "Could not merge this pull request by:\n```\n" + err.Error() + "\n```"
				if ok := operation.AddComment(issueSvc, repoOwner, repoName, issue, comment); !ok {
					log.Println("info: could not create the comment to express no merging the pull request")
				}
			}

			if c.info.DeleteAfterAutoMerge {
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
	}

	log.Printf("info: complete merge the pull request %v\n", issue)
	return true, nil
}

func (c *AcceptCommand) commandAcceptChangesetByOtherReviewer(ev *github.IssueCommentEvent, reviewer string) (bool, error) {
	log.Printf("info: Start: merge the pull request from other reviewer by %v\n", ev.Comment.ID)
	defer log.Printf("info: End:merge the pull request from other reviewer by %v\n", ev.Comment.ID)

	if !c.info.isReviewer(reviewer) {
		log.Println("info: could not find the actual reviewer in reviewer list")
		log.Printf("debug: specified actial reviewer %v\n", reviewer)
		return false, nil
	}

	return c.commandAcceptChangesetByReviewer(ev)
}

func createBranchFromMaster(svc *github.GitService, owner string, repo string, branchName string) (ok bool, ref *github.Reference) {
	refName := "refs/heads/" + branchName

	log.Printf("info: clean up %v by deleting it\n", refName)
	if _, err := svc.DeleteRef(owner, repo, refName); err != nil {
		log.Printf("info: could not clean up %v by %v, but we continue to create %v optimistically\n", refName, err, refName)
	}

	const base string = "refs/heads/master"
	ref, _, err := svc.GetRef(owner, repo, base)
	if err != nil {
		log.Printf("warn: cannot get reference about %v\n", base)
		return
	}

	branchRef := github.Reference{
		Ref:    &refName,
		URL:    nil, // XXX: This field is unused on creating ref.
		Object: ref.Object,
	}

	ref, _, err = svc.CreateRef(owner, repo, &branchRef)
	if err != nil {
		log.Printf("warn: cannot create a new ref %v\n", refName)
		return
	}

	return true, ref
}

func mergeIntoAutoBranch(svc *github.RepositoriesService, owner string, repo string, head *github.PullRequestBranch) (ok bool, commit *github.RepositoryCommit) {
	base := "auto"
	message := "Auto merging"
	req := github.RepositoryMergeRequest{
		Base:          &base,
		Head:          head.SHA,
		CommitMessage: &message,
	}

	commit, _, err := svc.Merge(owner, repo, &req)
	if err != nil {
		log.Printf("warn: could not merge '%v' branch into master\n", base)
		return
	}

	return true, commit
}
