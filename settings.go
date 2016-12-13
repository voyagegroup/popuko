package main

import "strconv"

type Settings struct {
	botName string
	port    uint64
	github  GithubSetting
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
		return github
	} else {
		return s.botName
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
