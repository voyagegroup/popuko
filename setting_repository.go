package main

type repositoryInfo struct {
	reviewers           *ReviewerSet
	regardAllAsReviewer bool

	ShouldMergeAutomatically bool
	ShouldDeleteMerged       bool
}

func (r *repositoryInfo) isReviewer(name string) bool {
	if r.regardAllAsReviewer {
		return true
	}

	return r.reviewers.Has(name)
}

func (r *repositoryInfo) Reviewers() *ReviewerSet {
	return r.reviewers
}
