package main

type RepositorySetting struct {
	owner        string
	name         string
	reviewerList []string
	reviewers    ReviewerSet

	shouldMergeAutomatically bool
	shouldDeleteMerged       bool
	// Use `OWNERS` file as reviewer list in the repository's root.
	useOwnersFile bool
}

func (s *RepositorySetting) Init() {
	set := newReviewerSet(s.reviewerList, false)
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
		regardAllAsReviewer:      false, // TODO
		ShouldMergeAutomatically: r.shouldMergeAutomatically,
		ShouldDeleteMerged:       r.shouldDeleteMerged,
	}
	return true, &info
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
