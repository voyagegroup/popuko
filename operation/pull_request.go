package operation

import (
	"log"
	"time"

	"github.com/google/go-github/github"
)

func IsMergeable(prSvc *github.PullRequestsService, owner, name string, issue int, pr *github.PullRequest) (bool, bool) {
	return isMergeable(prSvc, owner, name, issue, pr, 0)
}

func isMergeable(prSvc *github.PullRequestsService, owner, name string, issue int, pr *github.PullRequest, nest uint) (bool, bool) {
	mergeable := pr.Mergeable
	if mergeable == nil {
		// By the document https://developer.github.com/v3/pulls/#get-a-single-pull-request
		// this state is still in checking if pr.Mergeable == nil.
		if nest > 1 {
			// We tried once.
			// Conclude it is not mergeable heurístically
			return true, false
		}

		// sleep same time: https://github.com/barosl/homu/blob/2104e4b154d2fba15d515b478a5bd6105c1522f6/homu/main.py#L722
		time.Sleep(5 * time.Second)
		pr, _, err := prSvc.Get(owner, name, issue)
		if err != nil || pr == nil {
			log.Println("info: could not get the info for pull request")
			log.Printf("debug: %v\n", err)
			return false, false
		}
		return isMergeable(prSvc, owner, name, issue, pr, nest+1)
	}

	return true, *mergeable
}
