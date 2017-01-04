package main

import (
	"log"
	"net/http"

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

	log.Println("===== popuko =====")
	log.Printf("version (git revision): %s\n", revision)
	log.Printf("builddate: %s\n", builddate)
	log.Printf("listen http on port: %v\n", config.PortStr())
	log.Printf("botname for GitHub: %v\n", "@"+config.BotNameForGithub())
	log.Println("==================")

	github := createGithubClient(config)
	if github == nil {
		panic("Cannot create the github client")
	}

	server := AppServer{github, queue.NewAutoMergeQRepo()}

	http.HandleFunc("/github", server.handleGithubHook)
	http.ListenAndServe(config.PortStr(), nil)
}
