package setting

import (
	"log"
)

type RepositorySetting struct {
	Owner        string
	Name         string
	ReviewerList []string
	Reviewers    ReviewerSet

	EnableAutoMerge      bool
	DeleteAfterAutoMerge bool
	RegardAllAsReviewer  bool

	// Use `OWNERS` file as reviewer list in the repository's root.
	UseOwnersFile bool
}

func (s *RepositorySetting) Init() {
	set := newReviewerSet(s.ReviewerList)
	s.ReviewerList = nil
	s.Reviewers = *set
}

func (s *RepositorySetting) Fullname() string {
	return s.Owner + "/" + s.Name
}

func (r *RepositorySetting) ToRepoInfo() (bool, *RepositoryInfo) {
	info := RepositoryInfo{
		reviewers:            &r.Reviewers,
		regardAllAsReviewer:  r.RegardAllAsReviewer,
		EnableAutoMerge:      r.EnableAutoMerge,
		DeleteAfterAutoMerge: r.DeleteAfterAutoMerge,
	}
	return true, &info
}

func (r *RepositorySetting) Log() {
	log.Println("--------------------------------")
	log.Printf("%v\n", r.Fullname())

	if r.UseOwnersFile {
		log.Println("See /OWNERS.json to confirm all configurations for each repos.")
	} else {
		log.Printf("* Enable auto-merging by this bot: %v\n", r.EnableAutoMerge)
		log.Printf("* Try to delete a branch after auto merging: %v\n", r.DeleteAfterAutoMerge)
		if r.RegardAllAsReviewer {
			log.Println("* Privide the reviewer privilege for all user who can comment to the repo.")
		} else {
			log.Println("* reviewers:")
			for _, name := range r.Reviewers.Entries() {
				log.Printf("  - %v\n", name)
			}
		}
	}

	log.Println("")
}
