package epic

import (
	"context"
	"log"

	"github.com/google/go-github/v28/github"
	"github.com/voyagegroup/popuko/operation"
)

func RemoveAllStatusLabel(ctx context.Context, client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	owner := *repo.Owner.Login
	name := *repo.Name
	number := *pr.Number

	if pr.Merged == nil {
		// This value should be boolean, not be nil.
		log.Printf("error: we cannot get the merge status of #%v. We abort the process for safetyness & to save the API limit.\n", number)
		return
	}

	currentLabels := operation.GetLabelsByIssue(ctx, client.Issues, owner, name, number)
	if currentLabels == nil {
		log.Printf("warn: could not get all labels of #%v\n", number)
		return
	}

	labels := operation.RemoveStatusLabelFromList(currentLabels)
	_, _, err := client.Issues.ReplaceLabelsForIssue(ctx, owner, name, number, labels)
	if err != nil {
		log.Printf("warn: could not remove all `S-***` labels from #%v\n", number)
		return
	}

	log.Printf("info: finish to remove all status labels from #%v\n", number)
}
