package setting

import (
	"log"
	"os"
	"path/filepath"
	"testing"
)

const kConfigFile = "example.config.toml"

func TestLoadConfigToml(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("cannot get the current dir: %v\n", err)
		return
	}

	path, err := filepath.Abs(dir + "/../" + kConfigFile)
	if err != nil {
		t.Errorf("cannot get the abs path: %v\n", err)
		return
	}
	log.Printf("the file path: %v\n", path)

	result := decodeFile(path)
	if result == nil {
		t.Errorf("cannot decode the file: %v\n", path)
		return
	}

	if actual := result.Version; actual != 0 {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.Port; actual != 3000 {
		t.Errorf("%v\n", actual)
		return
	}

	if actual := result.Github.BotName; actual != "popuko" {
		t.Errorf("%v\n", actual)
		return
	}

	if actual := result.Github.Token; actual != "api_token" {
		t.Errorf("%v\n", actual)
		return
	}

	if actual := result.Github.HookSecret; actual != "webhook_secret" {
		t.Errorf("%v\n", actual)
		return
	}

	if actual := result.Github.acceptedRepos; actual != nil {
		t.Fatalf("%v\n", actual)
	}
}
