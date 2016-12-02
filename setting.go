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
	botName    string
	token      string
	hookSecret string
	repoList   []RepositorySetting
	repoMap    RepositoryMap
}

func (s *Settings) PortStr() string {
	return ":" + strconv.FormatUint(s.port, 10)
}

func (s *Settings) Init() {
	m := newRepositoryMap(s.github.repoList)
	s.github.repoList = nil
	s.github.repoMap = *m
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

func (s *Settings) Repositories() *RepositoryMap {
	return &s.github.repoMap
}

type RepositoryMap struct {
	inner map[string]RepositorySetting
}

func newRepositoryMap(list []RepositorySetting) *RepositoryMap {
	m := make(map[string]RepositorySetting)
	for _, item := range list {
		item.Init()

		k := item.Fullname()
		m[k] = item
	}
	return &RepositoryMap{
		m,
	}
}

func (m *RepositoryMap) Entries() map[string]RepositorySetting {
	return m.inner
}

func (m *RepositoryMap) Get(owner string, repo string) *RepositorySetting {
	k := owner + "/" + repo
	item, ok := m.inner[k]
	if ok {
		return &item
	}
	return nil
}

type RepositorySetting struct {
	owner        string
	name         string
	reviewerList []string
	reviewers    ReviewerSet
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

func (s *RepositorySetting) Reviewers() *ReviewerSet {
	return &s.reviewers
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
