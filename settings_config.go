package main

import "log"

type RepositorySetting struct {
	owner        string
	name         string
	reviewerList []string
	reviewers    ReviewerSet

	shouldMergeAutomatically bool
	shouldDeleteMerged       bool
	regardAllAsReviewer      bool

	// Use `OWNERS` file as reviewer list in the repository's root.
	useOwnersFile bool
}

func (s *RepositorySetting) Init() {
	set := newReviewerSet(s.reviewerList)
	s.reviewerList = nil
	s.reviewers = *set
}

func (s *RepositorySetting) Owner() string {
	return s.owner
}

func (s *RepositorySetting) Name() string {
	return s.name
}

func (s *RepositorySetting) Fullname() string {
	return s.owner + "/" + s.name
}

func (r *RepositorySetting) UseOwnersFile() bool {
	return r.useOwnersFile
}

func (r *RepositorySetting) ToRepoInfo() (bool, *repositoryInfo) {
	info := repositoryInfo{
		reviewers:                &r.reviewers,
		regardAllAsReviewer:      r.regardAllAsReviewer,
		ShouldMergeAutomatically: r.shouldMergeAutomatically,
		DeleteAfterAutoMerge:     r.shouldDeleteMerged,
	}
	return true, &info
}

func (r *RepositorySetting) log() {
	log.Println("--------------------------------")
	log.Printf("%v\n", r.Fullname())

	if r.UseOwnersFile() {
		log.Println("See /OWNERS.json to confirm all configurations for each repos.")
	} else {
		log.Printf("* Enable auto-merging by this bot: %v\n", r.shouldMergeAutomatically)
		log.Printf("* Try to delete a branch after auto merging: %v\n", r.shouldDeleteMerged)
		if r.regardAllAsReviewer {
			log.Println("* Privide the reviewer privilege for all user who can comment to the repo.")
		} else {
			log.Println("* reviewers:")
			for _, name := range r.reviewers.Entries() {
				log.Printf("  - %v\n", name)
			}
		}
	}

	log.Println("")
}
