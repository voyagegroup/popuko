package operation

import (
	"log"
	"strings"

	"github.com/google/go-github/github"
)

const (
	STATUS_LABEL_PREFIX   string = "S-"
	LABEL_AWAITING_REVIEW string = "S-awaiting-review"
	LABEL_AWAITING_MERGE  string = "S-awaiting-merge"
	LABEL_NEEDS_REBASE    string = "S-needs-rebase"
)

func AddAwaitingReviewLabel(list []*github.Label) []string {
	return changeStatusLabel(list, LABEL_AWAITING_REVIEW)
}

func AddAwaitingMergeLabel(list []*github.Label) []string {
	return changeStatusLabel(list, LABEL_AWAITING_MERGE)
}

func AddNeedRebaseLabel(list []*github.Label) []string {
	return changeStatusLabel(list, LABEL_NEEDS_REBASE)
}

func changeStatusLabel(list []*github.Label, new string) []string {
	result := make([]string, 0, 0)
	for _, item := range list {
		label := *item.Name
		if strings.Index(label, STATUS_LABEL_PREFIX) == 0 {
			continue
		} else {
			result = append(result, label)
		}
	}
	result = append(result, new)
	return result
}

func GetLabelsByIssue(issueSvc *github.IssuesService, owner string, name string, issue int) []*github.Label {
	currentLabels, _, err := issueSvc.ListLabelsByIssue(owner, name, issue, nil)
	if err != nil {
		log.Println("info: could not get labels by the issue")
		log.Printf("debug: %v\n", err)
		return nil
	}
	log.Printf("debug: the current labels: %v\n", currentLabels)
	return currentLabels
}

func HasLabelInList(list []*github.Label, target string) bool {
	for _, item := range list {
		label := *item.Name
		if label == target {
			return true
		}
	}
	return false
}