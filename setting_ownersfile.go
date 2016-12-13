package main

import "log"

type OwnersFile struct {
	Version      float64       `json:"version"`
	RawReviewers []interface{} `json:"reviewers"`

	// Provide a reviewer privilege for all users whoc can write some comment to
	// pull request.
	//
	// This feature is for the internal repository in your company
	// and there is no restrictions for non-reviewer/
	// NOT FOR PUBLIC OPEN SOURCE PROJECT.
	// You must not enable this option for an open source project.
	RegardAllAsReviewer bool `json:"regard_all_as_reviewer,omitempty"`
}

func (o *OwnersFile) Reviewers() (ok bool, set *ReviewerSet) {
	var list []string

	if !o.RegardAllAsReviewer {
		for _, v := range o.RawReviewers {
			n, ok := v.(string)
			if !ok {
				log.Printf("debug: %v\n", o.RawReviewers)
				return false, nil
			}

			list = append(list, n)
		}
	} else {
		log.Println("debug: This `OwnersFile` provides reviewer privilege for all users who can comment to this repo.")
	}

	set = newReviewerSet(list, o.RegardAllAsReviewer)
	return true, set
}

func (o *OwnersFile) ToRepoInfo() (bool, *repositoryInfo) {
	ok, r := o.Reviewers()
	if !ok {
		return false, nil
	}

	info := repositoryInfo{
		reviewers:                r,
		regardAllAsReviewer:      o.RegardAllAsReviewer,
		ShouldMergeAutomatically: false, // TODO
		ShouldDeleteMerged:       false, // TODO
	}
	return true, &info
}
