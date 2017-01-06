package epic

import (
	"log"

	"github.com/google/go-github/github"

	"github.com/karen-irc/popuko/input"
	"github.com/karen-irc/popuko/operation"
	"github.com/karen-irc/popuko/queue"
	"github.com/karen-irc/popuko/setting"
)

type AcceptCommand struct {
	Owner string
	Name  string

	Client  *github.Client
	BotName string
	Cmd     input.AcceptChangesetCommand
	Info    *setting.RepositoryInfo

	AutoMergeRepo *queue.AutoMergeQRepo
}

func (c *AcceptCommand) AcceptChangesetByReviewer(ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("info: Start: merge the pull request by %v\n", *ev.Comment.ID)
	defer log.Printf("info: End: merge the pull request by %v\n", *ev.Comment.ID)

	if c.BotName != c.Cmd.BotName() {
		log.Printf("info: this command works only if target user is actual our bot.")
		return false, nil
	}

	sender := *ev.Sender.Login
	log.Printf("debug: command is sent from %v\n", sender)

	if !c.Info.IsReviewer(sender) {
		log.Printf("info: %v is not an reviewer registred to this bot.\n", sender)
		return false, nil
	}

	client := c.Client
	issueSvc := client.Issues

	repoOwner := c.Owner
	repoName := c.Name
	issue := *ev.Issue.Number
	log.Printf("debug: issue number is %v\n", issue)

	currentLabels := operation.GetLabelsByIssue(issueSvc, repoOwner, repoName, issue)
	if currentLabels == nil {
		return false, nil
	}

	labels := operation.AddAwaitingMergeLabel(currentLabels)

	// https://github.com/nekoya/popuko/blob/master/web.py
	_, _, err := issueSvc.ReplaceLabelsForIssue(repoOwner, repoName, issue, labels)
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

	if c.Info.EnableAutoMerge {
		qHandle := c.AutoMergeRepo.Get(repoOwner, repoName)
		qHandle.Lock()
		defer qHandle.Unlock()

		q := qHandle.Load()

		item := &queue.AutoMergeQueueItem{
			PullRequest: issue,
			PrHead:      headSha,
		}
		q.Push(item)
		q.Save()

		if q.HasActive() {
			log.Printf("info: pull request (%v) has been queued but other is active.\n", issue)
			{
				comment := ":postbox: This pull request is queued. Please await the time."
				if ok := operation.AddComment(issueSvc, repoOwner, repoName, issue, comment); !ok {
					log.Println("info: could not create the comment to declare to merge this.")
				}
			}
			return true, nil
		}

		ok, next := q.GetNext()
		if !ok || next == nil {
			log.Println("error: this queue should not be empty because `q` is queued just now.")
			return false, nil
		}

		if next != item {
			log.Println("error: `next` should be equal to `q` because there should be only `q` in queue.")
			return false, nil
		}

		ok, commit := operation.TryWithMaster(client, repoOwner, repoName, pr)
		if !ok {
			log.Printf("info: we cannot try #%v with the latest `master`.", issue)
			return false, nil
		}

		item.AutoBranchHead = commit.SHA
		q.SetActive(item)
		log.Printf("info: pin #%v as the active item to queue\n", issue)
		q.Save()
	}

	log.Printf("info: complete merge the pull request %v\n", issue)
	return true, nil
}

func (c *AcceptCommand) AcceptChangesetByOtherReviewer(ev *github.IssueCommentEvent, reviewer string) (bool, error) {
	log.Printf("info: Start: merge the pull request from other reviewer by %v\n", ev.Comment.ID)
	defer log.Printf("info: End:merge the pull request from other reviewer by %v\n", ev.Comment.ID)

	if !c.Info.IsReviewer(reviewer) {
		log.Println("info: could not find the actual reviewer in reviewer list")
		log.Printf("debug: specified actial reviewer %v\n", reviewer)
		return false, nil
	}

	return c.AcceptChangesetByReviewer(ev)
}
