package epic

import (
	"errors"
	"log"

	"github.com/google/go-github/github"

	"github.com/karen-irc/popuko/input"
	"github.com/karen-irc/popuko/operation"
	"github.com/karen-irc/popuko/queue"
	"github.com/karen-irc/popuko/setting"
)

type CancelApprovedCommand struct {
	BotName       string
	Client        *github.Client
	Owner         string
	Name          string
	Number        int
	Cmd           *input.CancelApprovedByReviewerCommand
	Info          *setting.RepositoryInfo
	AutoMergeRepo *queue.AutoMergeQRepo
}

func (c *CancelApprovedCommand) CancelApprovedChangeSet(ev *github.IssueCommentEvent) (ok bool, err error) {
	id := *ev.Comment.ID
	log.Printf("info: Start: merge the pull request by %v\n", id)
	defer log.Printf("info: End: merge the pull request by %v\n", id)

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

	owner := c.Owner
	name := c.Name
	number := c.Number
	log.Printf("debug: issue number is %v\n", number)

	currentLabels := operation.GetLabelsByIssue(c.Client.Issues, owner, name, number)
	if currentLabels != nil {
		labels := operation.AddAwaitingReviewLabel(currentLabels)

		// https://github.com/nekoya/popuko/blob/master/web.py
		_, _, err = c.Client.Issues.ReplaceLabelsForIssue(owner, name, number, labels)
		if err != nil {
			log.Printf("info: could not change labels by the issue: %v\n", err)
		}
	}

	{
		comment := ":outbox_tray: This has been cancelled from the approved queue by `" + sender + "`"
		if ok := operation.AddComment(c.Client.Issues, owner, name, number, comment); !ok {
			log.Println("info: could not create the comment about what this pull request rejected.")
		}
	}

	if c.Info.EnableAutoMerge {
		qHandle := c.AutoMergeRepo.Get(owner, name)
		if qHandle == nil {
			return false, errors.New("error: cannot get the queue handle")
		}

		qHandle.Lock()
		defer qHandle.Unlock()

		q := qHandle.Load()
		if found := q.RemoveAwaiting(number); found {
			q.Save()
		}
	}

	log.Printf("info: complete to reject the pull request %v\n", number)
	return true, nil
}
