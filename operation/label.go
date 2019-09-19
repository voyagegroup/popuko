package operation

import (
	"context"
	"log"
	"strings"

	"github.com/google/go-github/v28/github"
)

const (
	STATUS_LABEL_PREFIX             string = "S-"
	LABEL_AWAITING_REVIEW           string = "S-awaiting-review"
	LABEL_AWAITING_MERGE            string = "S-awaiting-merge"
	LABEL_NEEDS_REBASE              string = "S-needs-rebase"
	LABEL_FAILS_TESTS_WITH_UPSTREAM string = "S-fails-tests-with-upstream"
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

func AddFailsTestsWithUpsreamLabel(list []*github.Label) []string {
	return changeStatusLabel(list, LABEL_FAILS_TESTS_WITH_UPSTREAM)
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

func GetLabelsByIssue(ctx context.Context, issueSvc *github.IssuesService, owner string, name string, issue int) []*github.Label {
	currentLabels, _, err := issueSvc.ListLabelsByIssue(ctx, owner, name, issue, nil)
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

func RemoveStatusLabelFromList(list []*github.Label) []string {
	r := make([]string, 0)
	for _, item := range list {
		label := *item.Name
		if !strings.HasPrefix(label, STATUS_LABEL_PREFIX) {
			r = append(r, label)
		}
	}
	return r
}
