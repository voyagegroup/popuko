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
		t.Fatalf("cannot get the current dir: %v\n", err)
	}

	path, err := filepath.Abs(dir + "/../" + kConfigFile)
	if err != nil {
		t.Fatalf("cannot get the abs path: %v\n", err)
	}
	log.Printf("the file path: %v\n", path)

	result := decodeFile(path)
	if result == nil {
		t.Fatalf("cannot decode the file: %v\n", path)
	}

	if actual := result.Version; actual != 0 {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.BotName; actual != "popuko" {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.Port; actual != 3000 {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.Github.BotName; actual != "" {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.Github.Token; actual != "api_token" {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.Github.HookSecret; actual != "webhook_secret" {
		t.Fatalf("%v\n", actual)
	}

	if actual := result.Github.acceptedRepos; actual != nil {
		t.Fatalf("%v\n", actual)
	}
}
