package epic

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"github.com/voyagegroup/popuko/operation"
)

func AssignReviewer(ctx context.Context, client *github.Client, ev *github.IssueCommentEvent, assignees []string) (bool, error) {
	log.Printf("info: Start: assign the reviewer by %v\n", *ev.Comment.ID)
	defer log.Printf("info: End: assign the reviewer by %v\n", *ev.Comment.ID)

	issueSvc := client.Issues

	repoOwner := *ev.Repo.Owner.Login
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)
	issue := *ev.Issue.Number
	log.Printf("debug: issue number is %v\n", issue)

	currentLabels := operation.GetLabelsByIssue(ctx, issueSvc, repoOwner, repo, issue)
	if currentLabels == nil {
		return false, nil
	}

	log.Printf("debug: assignees is %v\n", assignees)

	_, _, err := issueSvc.AddAssignees(ctx, repoOwner, repo, issue, assignees)
	if err != nil {
		log.Println("info: could not change assignees.")
		return false, err
	}

	labels := operation.AddAwaitingReviewLabel(currentLabels)
	_, _, err = issueSvc.ReplaceLabelsForIssue(ctx, repoOwner, repo, issue, labels)
	if err != nil {
		log.Println("info: could not change labels.")
		return false, err
	}

	log.Println("info: Complete assign the reviewer with no errors.")

	return true, nil
}
