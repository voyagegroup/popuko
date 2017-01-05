package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/karen-irc/popuko/queue"
	"github.com/karen-irc/popuko/setting"
)

var config *setting.Settings

var (
	revision  string
	builddate string
)

func main() {
	config = createSettings()

	var configDir string
	{
		c := "Specify the base dir of config as absolute path. default: $" + setting.XdgConfigHomeEnvKey
		flag.StringVar(&configDir, "config-base-dir", "", c)
	}
	flag.Parse()

	ok, root := setting.HomeDir(configDir)
	if !ok {
		log.Println("info: cannot find the config dir fot this.")
		return
	}

	// Check whether the dir exists. If there is none, create it.
	if _, err := os.Stat(root); err != nil {
		if err := os.MkdirAll(root, os.ModePerm); err != nil {
			log.Println("error: cannot create the config home dir.")
			return
		}
	}

	log.Println("===== popuko =====")
	log.Printf("version (git revision): %s\n", revision)
	log.Printf("builddate: %s\n", builddate)
	log.Printf("listen http on port: %v\n", config.PortStr())
	log.Printf("botname for GitHub: %v\n", "@"+config.BotNameForGithub())
	log.Printf("config dir: %v\n", root)
	log.Println("==================")

	github := createGithubClient(config)
	if github == nil {
		panic("Cannot create the github client")
	}

	q := queue.NewAutoMergeQRepo(root)
	if q == nil {
		panic("Fail to initialize the merge queue")
	}

	server := AppServer{github, q}

	http.HandleFunc("/github", server.handleGithubHook)
	http.ListenAndServe(config.PortStr(), nil)
}
