package operation

import (
	"log"

	"github.com/google/go-github/github"
)

func AddComment(issueSvc *github.IssuesService, owner string, name string, issue int, body string) bool {
	_, _, err := issueSvc.CreateComment(owner, name, issue, &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		log.Printf("info: could not create the comment to %v/%v#%v\n", owner, name, issue)
		log.Printf("debug: error is:%v\n", err)
		return false
	}

	return true
}
