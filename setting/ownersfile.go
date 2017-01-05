package setting

import (
	"log"
)

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

	// Enable to merge branch automatically by this bot after you command `r+`.
	// If you merge by hand and this bot should change the status label,
	// disable this option.
	EnableAutoMerge bool `json:"auto_merge.enabled,omitempty"`

	// Delete the branch by this bot after this bot had merged it
	// if you enable this option.
	// The operation may not delete contributor's branch by API
	// restriction. This only clean up only the upstream repository
	// managed by this bot.
	DeleteAfterAutoMerge bool `json:"auto_merge.delete_branch,omitempty"`
}

func (o *OwnersFile) reviewers() (ok bool, set *ReviewerSet) {
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

	set = newReviewerSet(list)
	return true, set
}

func (o *OwnersFile) ToRepoInfo() (bool, *RepositoryInfo) {
	ok, r := o.reviewers()
	if !ok {
		return false, nil
	}

	info := RepositoryInfo{
		reviewers:            r,
		regardAllAsReviewer:  o.RegardAllAsReviewer,
		EnableAutoMerge:      o.EnableAutoMerge,
		DeleteAfterAutoMerge: o.DeleteAfterAutoMerge,
	}
	return true, &info
}
