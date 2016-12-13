package main

import (
	"log"
	"net/http"
)

var config *Settings

var (
	revision  string
	builddate string
)

func main() {
	config = createSettings()
	config.Init()

	log.Println("===== popuko =====")
	log.Printf("version (git revision): %s\n", revision)
	log.Printf("builddate: %s\n", builddate)
	log.Printf("listen http on port: %v\n", config.PortStr())
	log.Printf("botname for GitHub: %v\n", "@"+config.BotNameForGithub())
	{
		log.Println("---- popuko handling repositories -------")
		repomap := config.Repositories()
		for _, v := range repomap.Entries() {
			dumpRepositorySetting(&v)
			log.Println("")
		}
	}
	log.Println("==================")

	github := createGithubClient(config)
	if github == nil {
		panic("Cannot create the github client")
	}

	server := AppServer{github}

	http.HandleFunc("/github", server.handleGithubHook)
	http.ListenAndServe(config.PortStr(), nil)
}

func dumpRepositorySetting(v *RepositorySetting) {
	log.Printf("%v\n", v.Fullname())

	_, info := v.ToRepoInfo()

	log.Printf("  Try to delete a merged branch: %v\n", info.ShouldDeleteMerged)
	log.Printf("  Use OWNERS.json: %v\n", v.UseOwnersFile())
	if v.UseOwnersFile() {
		log.Println("  reviewers: see OWNERS.json in the repository")
	} else {
		log.Println("  reviewers:")
		reviewer := info.Reviewers()
		for _, name := range reviewer.Entries() {
			log.Printf("    - %v\n", name)
		}
	}
}
