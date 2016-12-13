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

type ReviewerSet struct {
	regardAllAsReviewer bool
	set                 map[string]*interface{}
}

func (s *ReviewerSet) Has(person string) bool {
	if s.regardAllAsReviewer {
		return true
	}

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

func newReviewerSet(list []string, regardAllAsReviewer bool) *ReviewerSet {
	if regardAllAsReviewer {
		return &ReviewerSet{
			true,
			nil,
		}
	}

	s := make(map[string]*interface{})
	for _, name := range list {
		s[name] = nil
	}

	return &ReviewerSet{
		false,
		s,
	}
}
