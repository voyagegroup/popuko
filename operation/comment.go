package operation

import (
	"context"
	"log"

	"github.com/google/go-github/v28/github"
)

func AddComment(ctx context.Context, issueSvc *github.IssuesService, owner string, name string, issue int, body string) bool {
	_, _, err := issueSvc.CreateComment(ctx, owner, name, issue, &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		log.Printf("info: could not create the comment to %v/%v#%v\n", owner, name, issue)
		log.Printf("debug: error is:%v\n", err)
		return false
	}

	return true
}

func CommentHeadIsDifferentFromAccepted(ctx context.Context, issueSvc *github.IssuesService, owner string, name string, prNum int) {
	log.Printf("info: the head of #%v is changed from r+.\n", prNum)

	comment := ":no_entry_sign: The current head is changed from when this had been accepted. Please review again. :no_entry_sign:"
	if ok := AddComment(ctx, issueSvc, owner, name, prNum, comment); !ok {
		log.Println("error: could not write the comment about the result of auto branch.")
	}

	currentLabels := GetLabelsByIssue(ctx, issueSvc, owner, name, prNum)
	if currentLabels == nil {
		return
	}

	labels := AddAwaitingReviewLabel(currentLabels)
	_, _, err := issueSvc.ReplaceLabelsForIssue(ctx, owner, name, prNum, labels)
	if err != nil {
		log.Println("warn: could not change labels of the issue")
	}
}
