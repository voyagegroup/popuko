package epic

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/google/go-github/v28/github"
	"github.com/voyagegroup/popuko/setting"
)

func GetRepositoryInfo(ctx context.Context, repoSvc *github.RepositoriesService, owner, name, defaultBranchName string) *setting.RepositoryInfo {
	var repoinfo *setting.RepositoryInfo
	log.Println("info: Use `OWNERS` file.")
	ok, owners := fetchOwnersFile(ctx, repoSvc, owner, name, defaultBranchName)
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

func fetchOwnersFile(ctx context.Context, svc *github.RepositoriesService, owner string, reponame string, defaultBranchName string) (bool, *setting.OwnersFile) {
	fullRepositoryName := owner + "/" + reponame
	if defaultBranchName == "" {
		log.Printf("debug: could not get default branch name from the event for %v\n", fullRepositoryName)
		repoInfo, _, err := svc.Get(ctx, owner, reponame)
		if err != nil {
			log.Printf("warn: could not fetch the repository infor for %v by %v\n", fullRepositoryName, err)
			return false, nil
		}

		defaultBranchName = repoInfo.GetDefaultBranch()
	}
	log.Printf("info: the default branch name is `%v` for %v\n", defaultBranchName, fullRepositoryName)

	file, err := svc.DownloadContents(ctx, owner, reponame, "OWNERS.json", &github.RepositoryContentGetOptions{
		// We always use the file in master which we regard as accepted to the project.
		Ref: defaultBranchName,
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
