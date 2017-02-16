package epic

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/karen-irc/popuko/operation"
)

func RemoveAllStatusLabel(client *github.Client, repo *github.Repository, pr *github.PullRequest) {
	owner := *repo.Owner.Login
	name := *repo.Name
	number := *pr.Number

	if pr.Merged == nil {
		// This value should be boolean, not be nil.
		log.Printf("error: we cannot get the merge status of #%v. We abort the process for safetyness & to save the API limit.\n", number)
		return
	}

	currentLabels := operation.GetLabelsByIssue(client.Issues, owner, name, number)
	if currentLabels == nil {
		log.Printf("warn: could not get all labels of #%v\n", number)
		return
	}

	labels := operation.RemoveStatusLabelFromList(currentLabels)
	_, _, err := client.Issues.ReplaceLabelsForIssue(owner, name, number, labels)
	if err != nil {
		log.Printf("warn: could not remove all `S-***` labels from #%v\n", number)
		return
	}

	log.Printf("info: finish to remove all status labels from #%v\n", number)
}
