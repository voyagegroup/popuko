package operation

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v28/github"
)

func IsMergeable(ctx context.Context, prSvc *github.PullRequestsService, owner, name string, issue int, pr *github.PullRequest) (bool, bool) {
	return isMergeable(ctx, prSvc, owner, name, issue, pr, 0)
}

func isMergeable(ctx context.Context, prSvc *github.PullRequestsService, owner, name string, issue int, pr *github.PullRequest, nest uint) (bool, bool) {
	mergeable := pr.Mergeable
	if mergeable == nil {
		// By the document https://developer.github.com/v3/pulls/#get-a-single-pull-request
		// this state is still in checking if pr.Mergeable == nil.
		if nest > 1 {
			// We tried once.
			// We conclude that the pull request is mergeable.
			// If we conclude it is not mergeable, it is too eager.
			// Even if it's mergeable, we would conclude it's not mergeable. It's mis-detection.
			log.Printf("info: we cannot get the mergeable status of #%v again. We treat it is MERGEABLE heurístically \n", issue)
			return true, true
		}

		// sleep same time: https://github.com/barosl/homu/blob/2104e4b154d2fba15d515b478a5bd6105c1522f6/homu/main.py#L722
		time.Sleep(5 * time.Second)

		pr, _, err := prSvc.Get(ctx, owner, name, issue)
		if err != nil || pr == nil {
			log.Printf("info: could not get the info for #%v\n", issue)
			log.Printf("debug: %v\n", err)
			return false, false
		}
		return isMergeable(ctx, prSvc, owner, name, issue, pr, nest+1)
	}

	return true, *mergeable
}

func IsRelatedToDefaultBranch(pr *github.PullRequest, owner, master string) bool {
	base := pr.Base
	if base == nil {
		log.Printf("info: #%v's Base is `nil`\n", *pr.Number)
		return false
	}

	baseRef := base.Ref
	if baseRef == nil {
		log.Printf("info: #%v's Base.Ref is `nil`\n", *pr.Number)
		return false
	}

	if *baseRef != master {
		log.Printf("info: #%v's Base.Ref is not equals to `%v`\n", *pr.Number, master)
		return false
	}

	baseLabel := base.Label
	if baseLabel == nil {
		log.Printf("info: #%v's Base.Label is `nil`\n", *pr.Number)
		return false
	}

	// Check the pr is from the forked one.
	if strings.Contains(*baseLabel, ":") {
		if !strings.HasPrefix(*baseLabel, owner) {
			log.Printf("info: #%v is come from the forked but `%v` is not related to us\n", *pr.Number, *baseLabel)
			return false
		}

		if !strings.HasSuffix(*baseLabel, master) {
			log.Printf("info: #%v is come from the forked but `%v` is not related to our default branch (%v)\n", *pr.Number, *baseLabel, master)
			return false
		}
	} else {
		if *baseLabel != master {
			log.Printf("info: #%v's base is `%v` but our master is `%v`.", *pr.Number, *baseLabel, master)
			return false
		}
	}

	return true
}
