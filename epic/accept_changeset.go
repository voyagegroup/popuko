package epic

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/v28/github"

	"errors"

	"fmt"

	"github.com/voyagegroup/popuko/input"
	"github.com/voyagegroup/popuko/operation"
	"github.com/voyagegroup/popuko/queue"
	"github.com/voyagegroup/popuko/setting"
)

type AcceptCommand struct {
	Owner string
	Name  string

	Client  *github.Client
	BotName string
	Info    *setting.RepositoryInfo

	AutoMergeRepo *queue.AutoMergeQRepo
}

func (c *AcceptCommand) AcceptChangesetByOthers(ctx context.Context, ev *github.IssueCommentEvent, cmd *input.AcceptChangeByOthersCommand) (bool, error) {
	log.Printf("info: Start: merge the pull request by %v\n", *ev.Comment.ID)
	defer log.Printf("info: End: merge the pull request by %v\n", *ev.Comment.ID)

	if c.BotName != cmd.BotName() {
		log.Printf("info: this command works only if target user is actual our bot.")
		return false, nil
	}

	sender := *ev.Sender.Login
	log.Printf("debug: command is sent from %v\n", sender)

	if c.Info.IsReviewer(sender) {
		log.Printf("info: this bot try to merge #%v by the reviewer (`%v`)\n", ev.Issue.GetID(), sender)
		return c.acceptChangeset(ctx, ev, cmd)
	}

	if c.Info.IsInMergeableUserList(sender) {
		opener := ev.Issue.User.GetName()
		if isMergeableByMergeableUser(sender, opener, cmd.Reviewer) {
			log.Printf("info: this bot try to merge #%v (opened by `%v`) by the mergeable user (`%v`) with reviewer (%v)\n", ev.Issue.GetID(), opener, sender, cmd.Reviewer)
			return c.acceptChangeset(ctx, ev, cmd)
		}
	}

	log.Printf("info: %v cannnot merge the pull request #%v\n", sender, ev.Issue.GetID())
	return false, nil
}

func isMergeableByMergeableUser(commander, opener string, reviewer []string) bool {
	if commander != opener {
		log.Printf("info: commander `(%v)` is diffetent from the opener (`%v`) for this pull request\n", commander, opener)
		return false
	}

	for _, r := range reviewer {
		if r == commander {
			log.Printf("info: commander `(%v)` could not review this pull request by self (reviewer: %v)\n", commander, reviewer)
			return false
		}
	}

	return true
}

func (c *AcceptCommand) AcceptChangesetByReviewer(ctx context.Context, ev *github.IssueCommentEvent, cmd *input.AcceptChangeByReviewerCommand) (bool, error) {
	log.Printf("info: Start: merge the pull request by %v\n", *ev.Comment.ID)
	defer log.Printf("info: End: merge the pull request by %v\n", *ev.Comment.ID)

	if c.BotName != cmd.BotName() {
		log.Printf("info: this command works only if target user is actual our bot.")
		return false, nil
	}

	sender := *ev.Sender.Login
	log.Printf("debug: command is sent from %v\n", sender)

	if !c.Info.IsReviewer(sender) {
		log.Printf("info: %v is not an reviewer registred to this bot.\n", sender)
		return false, nil
	}

	return c.acceptChangeset(ctx, ev, cmd)
}

func (c *AcceptCommand) acceptChangeset(ctx context.Context, ev *github.IssueCommentEvent, cmd input.AcceptChangesetCommand) (bool, error) {
	sender := *ev.Sender.Login

	client := c.Client
	issueSvc := client.Issues

	repoOwner := c.Owner
	repoName := c.Name
	issue := *ev.Issue.Number
	log.Printf("debug: issue number is %v\n", issue)

	currentLabels := operation.GetLabelsByIssue(ctx, issueSvc, repoOwner, repoName, issue)
	if currentLabels == nil {
		return false, nil
	}

	labels := operation.AddAwaitingMergeLabel(currentLabels)

	// https://github.com/nekoya/popuko/blob/master/web.py
	_, _, err := issueSvc.ReplaceLabelsForIssue(ctx, repoOwner, repoName, issue, labels)
	if err != nil {
		log.Println("info: could not change labels by the issue")
		return false, err
	}

	prSvc := client.PullRequests
	pr, _, err := prSvc.Get(ctx, repoOwner, repoName, issue)
	if err != nil {
		log.Println("info: could not fetch the pull request information.")
		return false, err
	}

	headSha := *pr.Head.SHA
	if ok := commentApprovedSha(ctx, cmd, issueSvc, repoOwner, repoName, issue, headSha, sender); !ok {
		log.Println("info: could not create the comment to declare the head is approved.")
		return false, err
	}

	if c.Info.EnableAutoMerge {
		qHandle := c.AutoMergeRepo.Get(repoOwner, repoName)
		if qHandle == nil {
			log.Println("error: cannot get the queue handle")
			return false, errors.New("error: cannot get the queue handle")
		}

		qHandle.Lock()
		defer qHandle.Unlock()

		q := qHandle.Load()

		item := &queue.AutoMergeQueueItem{
			PullRequest: issue,
			PrHead:      headSha,
		}
		ok, mutated := queuePullReq(q, item)
		if !ok {
			return false, errors.New("error: we cannot recover the error")
		}

		if mutated {
			q.Save()
		}

		if q.HasActive() {
			commentAsPostponed(ctx, issueSvc, repoOwner, repoName, issue)
			return true, nil
		}

		if next := q.Front(); next != item {
			commentAsPostponed(ctx, issueSvc, repoOwner, repoName, issue)
		}

		tryNextItem(ctx, client, repoOwner, repoName, q, c.Info.AutoBranchName)
	}

	log.Printf("info: complete merge the pull request %v\n", issue)
	return true, nil
}

func commentApprovedSha(
	ctx context.Context,
	cmd input.AcceptChangesetCommand,
	issues *github.IssuesService,
	owner,
	name string,
	number int,
	sha string,
	sender string) bool {

	var reviewers string
	switch cmd := cmd.(type) {
	case *input.AcceptChangeByOthersCommand:
		{
			list := make([]string, 0, len(cmd.Reviewer))
			for _, name := range cmd.Reviewer {
				list = append(list, fmt.Sprintf("`%v`", name))
			}
			reviewers = strings.Join(list, ", ")
		}
	case *input.AcceptChangeByReviewerCommand:
		reviewers = fmt.Sprintf("`%v`", sender)
	default:
		log.Printf("error: %+v is not handled.", cmd)
		return false
	}

	comment := fmt.Sprintf(":pushpin: Commit %v has been approved by %v", sha, reviewers)
	if ok := operation.AddComment(ctx, issues, owner, name, number, comment); !ok {
		log.Println("info: could not create the comment to declare the head is approved.")
		return false
	}

	return true
}

func queuePullReq(queue *queue.AutoMergeQueue, item *queue.AutoMergeQueueItem) (ok bool, mutated bool) {
	if queue.HasActive() {
		active := queue.GetActive()
		if active.PullRequest == item.PullRequest {
			if active.PrHead == item.PrHead {
				// noop
				return true, false
			}

			queue.RemoveActive()
			if ok := queue.Push(item); !ok {
				return false, false
			}

			return true, true
		}
	}

	has, awaiting := queue.IsAwaiting(item.PullRequest)
	if has {
		if sameHead := (awaiting.PrHead == item.PrHead); sameHead {
			return true, false
		}

		if ok := queue.RemoveAwaiting(item.PullRequest); !ok {
			log.Println("error: ASSERT!: cannot remove awaiting item")
			log.Printf("error: queue %+v\n", queue)
			log.Printf("error: item %+v\n", item)
			return false, false
		}
	}

	if ok := queue.Push(item); !ok {
		return false, false
	}

	return true, true
}

func commentAsPostponed(ctx context.Context, issueSvc *github.IssuesService, owner, name string, issue int) {
	log.Printf("info: pull request (%v) has been queued but other is active.\n", issue)
	{
		comment := ":postbox: This pull request is queued. Please await the time."
		if ok := operation.AddComment(ctx, issueSvc, owner, name, issue, comment); !ok {
			log.Println("info: could not create the comment to declare to merge this.")
		}
	}
}
