package setting

import "strconv"

type Settings struct {
	BotName string
	Port    uint64
	Github  GithubSetting
}

func (s *Settings) PortStr() string {
	return ":" + strconv.FormatUint(s.Port, 10)
}

func (s *Settings) BotNameForGithub() string {
	github := s.Github.BotName
	if github != "" {
		return github
	} else {
		return s.BotName
	}
}

func (s *Settings) GithubToken() string {
	return s.Github.Token
}

func (s *Settings) WebHookSecret() []byte {
	return []byte(s.Github.HookSecret)
}
