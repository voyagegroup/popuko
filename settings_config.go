package main

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
		ShouldDeleteMerged:       r.shouldDeleteMerged,
	}
	return true, &info
}
