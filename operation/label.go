package operation

import (
	"log"

	"github.com/google/go-github/github"
)

func HasStatusLabel(issueSvc *github.IssuesService, owner string, name string, issue int, label string) bool {
	current, _, err := issueSvc.ListLabelsByIssue(owner, name, issue, nil)
	if err != nil {
		log.Println("warn: could not get labels by the issue")
		log.Printf("debug: %v\n", err)
		return false
	}

	has := hasStatusLabel(current, label)
	if !has {
		log.Printf("debug: #%v does not have %v\n", issue, label)
	}

	return has
}

func hasStatusLabel(list []*github.Label, target string) bool {
	for _, item := range list {
		label := *item.Name
		if label == target {
			return true
		}
	}
	return false
}
