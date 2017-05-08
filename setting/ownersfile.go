package setting

import (
	"log"
)

const autoBranchName string = "auto"

type OwnersFile struct {
	Version      float64       `json:"version"`
	RawReviewers []interface{} `json:"reviewers"`

	// Users in this list can merge only a pull request opened by themselves.
	// They only can command `@<botname> r=<reviewer_name>` and `<reviewer_name>`
	// must be different from their names.
	RawMergeableUsers []interface{} `json:"mergeable_users,omitempty"`

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

	// The name of the branch which is used for "Auto-Merging" to test changesets
	// before merging it into upstream. The default value is defined as `autoBranchName`.
	AutoBranchName string `json:"auto_branch.branch_name.auto",omitempty`
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

func (o *OwnersFile) mergeables() (ok bool, set *ReviewerSet) {
	var list []string

	for _, v := range o.RawMergeableUsers {
		n, ok := v.(string)
		if !ok {
			log.Printf("debug: %v\n", o.RawMergeableUsers)
			return false, nil
		}

		list = append(list, n)
	}

	set = newReviewerSet(list)
	return true, set
}

func (o *OwnersFile) ToRepoInfo() (bool, *RepositoryInfo) {
	ok, r := o.reviewers()
	if !ok {
		return false, nil
	}

	ok, mergeables := o.mergeables()
	if !ok {
		return false, nil
	}

	autoName := o.AutoBranchName
	if autoName == "" {
		autoName = autoBranchName
	}

	info := RepositoryInfo{
		reviewers:            r,
		mergeables:           mergeables,
		regardAllAsReviewer:  o.RegardAllAsReviewer,
		EnableAutoMerge:      o.EnableAutoMerge,
		DeleteAfterAutoMerge: o.DeleteAfterAutoMerge,
		AutoBranchName:       autoName,
	}
	return true, &info
}
