package main

type repositoryInfo struct {
	reviewers           *ReviewerSet
	regardAllAsReviewer bool

	EnableAutoMerge      bool
	DeleteAfterAutoMerge bool

	experimentalTryOnAutoBranch bool
}

func (r *repositoryInfo) isReviewer(name string) bool {
	if r.regardAllAsReviewer {
		return true
	}

	return r.reviewers.Has(name)
}

func (r *repositoryInfo) ExperimentalTryOnAutoBranch() bool {
	return r.experimentalTryOnAutoBranch && r.EnableAutoMerge
}

func (r *repositoryInfo) Reviewers() *ReviewerSet {
	return r.reviewers
}

type ReviewerSet struct {
	set map[string]*interface{}
}

func (s *ReviewerSet) Has(person string) bool {
	_, ok := s.set[person]
	return ok
}

func (s *ReviewerSet) Entries() []string {
	list := make([]string, 0)
	for k := range s.set {
		list = append(list, k)
	}
	return list
}

func newReviewerSet(list []string) *ReviewerSet {
	s := make(map[string]*interface{})
	for _, name := range list {
		s[name] = nil
	}

	return &ReviewerSet{
		s,
	}
}
