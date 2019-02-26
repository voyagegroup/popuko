package setting

type RepositoryInfo struct {
	reviewers           *ReviewerSet
	regardAllAsReviewer bool
	mergeables          *ReviewerSet

	EnableAutoMerge      bool
	DeleteAfterAutoMerge bool
	AutoBranchName       string
}

func (r *RepositoryInfo) IsReviewer(name string) bool {
	if r.regardAllAsReviewer {
		return true
	}

	return r.reviewers.Has(name)
}

func (r *RepositoryInfo) IsInMergeableUserList(name string) bool {
	return r.mergeables.Has(name)
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
