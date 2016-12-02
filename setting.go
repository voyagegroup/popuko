package main

import (
	"strconv"
)

type Settings struct {
	botName string
	port    uint64
	github  GithubSetting
}

type GithubSetting struct {
	botName      string
	token        string
	hookSecret   string
	reviewerList *[]string
	reviewers    *ReviewerSet
}

func (s *Settings) PortStr() string {
	return ":" + strconv.FormatUint(s.port, 10)
}

func (s *Settings) Init() {
	set := newReviewerSet(*s.github.reviewerList)
	s.github.reviewerList = nil
	s.github.reviewers = set
}

func (s *Settings) BotNameForGithub() string {
	github := s.github.botName
	if github != "" {
		return "@" + github
	} else {
		return "@" + s.botName
	}
}

func (s *Settings) GithubToken() string {
	return s.github.token
}

func (s *Settings) WebHookSecret() []byte {
	return []byte(s.github.hookSecret)
}

func (s *Settings) Reviewers() *ReviewerSet {
	return s.github.reviewers
}

type ReviewerSet struct {
	set map[string]*interface{}
}

func (s *ReviewerSet) Has(person string) bool {
	_, ok := s.set[person]
	return ok
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
