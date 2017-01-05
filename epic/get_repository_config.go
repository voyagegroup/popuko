package epic

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/google/go-github/github"
	"github.com/karen-irc/popuko/setting"
)

func GetRepositoryInfo(repoSvc *github.RepositoriesService, owner, name string) *setting.RepositoryInfo {
	var repoinfo *setting.RepositoryInfo
	log.Println("info: Use `OWNERS` file.")
	ok, owners := fetchOwnersFile(repoSvc, owner, name)
	if !ok {
		log.Println("error: could not handle OWNERS file.")
		return nil
	}

	ok, repoinfo = owners.ToRepoInfo()
	if !ok {
		log.Println("error: could not get reviewer list")
		return nil
	}

	return repoinfo
}

func fetchOwnersFile(svc *github.RepositoriesService, owner string, reponame string) (bool, *setting.OwnersFile) {
	file, err := svc.DownloadContents(owner, reponame, "OWNERS.json", &github.RepositoryContentGetOptions{
		// We always use the file in master which we regard as accepted to the project.
		Ref: "master",
	})
	if err != nil {
		log.Printf("error: could not fetch `OWNERS.json`: %v\n", err)
		return false, nil
	}

	raw, err := ioutil.ReadAll(file)
	defer file.Close()
	if err != nil {
		log.Printf("error: could not read `OWNERS.json`: %v\n", err)
		return false, nil
	}
	log.Printf("debug: OWNERS.json:\n%v\n", string(raw))

	var decoded setting.OwnersFile
	if err := json.Unmarshal(raw, &decoded); err != nil {
		log.Printf("error: could not decode `OWNERS.json`: %v\n", err.Error())
		return false, nil
	}

	return true, &decoded
}
