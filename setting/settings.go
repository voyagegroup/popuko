package setting

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"log"

	"github.com/BurntSushi/toml"
)

type Settings struct {
	Version int           `toml:"config_version"`
	BotName string        `toml:"botname"`
	Port    int           `toml:"port"`
	Github  GithubSetting `toml:"github"`
}

func (s *Settings) PortStr() string {
	return ":" + strconv.FormatInt(int64(s.Port), 10)
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

func (s *Settings) AcceptRepo(owner, name string) bool {
	return s.Github.accept(owner, name)
}

const RootConfigFile = "/config.toml"

func LoadSettings(dir string) *Settings {
	path, err := filepath.Abs(dir + "/" + RootConfigFile)
	if err != nil {
		log.Printf("error: cannot get the path to %v\n", err)
		return nil
	}

	s := decodeFile(path)
	if s == nil {
		return nil
	}

	initGithubSetting(&s.Github)
	return s
}

func decodeFile(path string) *Settings {
	// Check whether the file exists.
	if _, err := os.Stat(path); err != nil {
		log.Printf("error: %v is not found.\n", path)
		return nil
	}

	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("error: on read %v: %v\n", path, err)
		return nil
	}

	data := string(b)

	var s Settings
	if _, err := toml.Decode(data, &s); err != nil {
		log.Printf("error: on toml decoding: %v\n", err)
		return nil
	}

	return &s
}
