package main

import (
	"fmt"
	"net/http"
)

func main() {
	config.Init()

	fmt.Println("===== popuko =====")
	fmt.Printf("listen http on port: %v\n", config.PortStr())
	fmt.Printf("botname for GitHub: %v\n", config.BotNameForGithub())
	fmt.Println("==================")

	github := createGithubClient(config)
	if github == nil {
		panic("Cannot create the github client")
	}

	server := AppServer{github}

	http.HandleFunc("/github", server.handleGithubHook)
	http.ListenAndServe(config.PortStr(), nil)
}
