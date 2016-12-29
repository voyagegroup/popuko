package main

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/karen-irc/popuko/operation"
)

func (srv *AppServer) checkAutoBranch(ev *github.StatusEvent) {
	log.Println("info: Start: checkAutoBranch")
	defer log.Println("info: End: checkAutoBranch")

	if *ev.State == "pending" {
		log.Println("info: Not handle pending status event")
		return
	}
	log.Printf("info: Start to handle status event: %v\n", *ev.State)

	repoOwner := *ev.Repo.Owner.Login
	repoName := *ev.Repo.Name
	log.Printf("info: Target repository is %v/%v\n", repoOwner, repoName)

	repoConfig := config.Repositories().Get(repoOwner, repoName)
	if repoConfig == nil {
		log.Println("info: Not found registred repo config.")
		return
	}

	repoInfo := createRepositoryInfo(repoConfig, srv.githubClient.Repositories)
	if repoInfo == nil {
		log.Println("debug: cannot get repositoryInfo")
		return
	}

	log.Println("info: success to load the configure.")

	if !(repoInfo.EnableAutoMerge && repoInfo.ExperimentalTryOnAutoBranch()) {
		log.Println("info: this repository does not enable merging into master automatically.")
		return
	}

	log.Println("info: Start to handle auto merging the branch.")

	srv.autoMergeRepo.Lock()
	queue := srv.autoMergeRepo.Get(repoOwner, repoName)
	srv.autoMergeRepo.Unlock()

	queue.Lock()
	defer queue.Unlock()

	if !queue.HasActive() {
		log.Println("info: there is no testing item")
		return
	}

	active := queue.GetActive()
	if active == nil {
		log.Println("error: `active` should not be null")
		return
	}

	log.Println("info: got the active item.")

	if !isIncludeAutoBranch(ev.Branches) {
		log.Printf("warn: this status event (%v) does not include the auto branch\n", *ev.ID)
		return
	}

	log.Println("info: the status event is related to auto branch.")

	if active.SHA == nil {
		log.Println("error: ASSERT! `active.SHA` should not be null")
		return
	}

	autoTipSha := *active.SHA
	if autoTipSha != *ev.SHA {
		log.Printf("debug: The commit hash which contained by the status event: %v\n", *ev.SHA)
		log.Printf("debug: The commit hash is pinned to the status queue as the tip of auto branch: %v\n", autoTipSha)
		log.Printf("info: The event's tip sha does not equal to the one which is tesing actively in %v/%v\n", repoOwner, repoName)
		return
	}
	log.Println("info: the tip of auto branch is same as `active.SHA`")

	client := srv.githubClient
	issueSvc := client.Issues
	prSvc := client.PullRequests
	prNum := active.PullRequest

	prInfo, _, err := prSvc.Get(repoOwner, repoName, prNum)
	if err != nil {
		log.Println("info: could not fetch the pull request information.")
		return
	}

	if *prInfo.State != "open" {
		log.Printf("info: the pull request #%v has been resolved the state\n", prNum)
		return
	}

	if *ev.State != "success" {
		log.Println("info: could not merge pull request")

		client := srv.githubClient
		issueSvc := client.Issues
		repoSvc := client.Repositories

		prNum := active.PullRequest

		status, _, err := repoSvc.GetCombinedStatus(repoOwner, repoName, "auto", nil)
		if err != nil {
			log.Println("error: could not get the status about the auto branch.")
		}

		comment := ":collision: " + *ev.State + ": The branch testing to merge this pull request into master has been troubled."
		if status != nil {
			comment += "\n\n"

			for _, s := range status.Statuses {
				if s.Description == nil || s.TargetURL == nil {
					continue
				}

				item := "* [" + *s.Description + "](" + *s.TargetURL + ")\n"
				comment += item
			}
		}

		if ok := operation.AddComment(issueSvc, repoOwner, repoName, prNum, comment); !ok {
			log.Println("error: could not write the comment about the result of auto branch.")
		}

		return
	}

	{
		repoSvc := client.Repositories
		status, _, err := repoSvc.GetCombinedStatus(repoOwner, repoName, "auto", nil)
		if err != nil {
			log.Println("error: could not get the status about the auto branch.")
		}

		comment := ":tada: " + *ev.State + ": The branch testing to merge this pull request into master has been succeed."
		if status != nil {
			comment += "\n\n"

			for _, s := range status.Statuses {
				if s.Description == nil || s.TargetURL == nil {
					continue
				}

				item := "* [" + *s.Description + "](" + *s.TargetURL + ")\n"
				comment += item
			}
		}

		if ok := operation.AddComment(issueSvc, repoOwner, repoName, prNum, comment); !ok {
			log.Println("error: could not write the comment about the result of auto branch.")
		}
	}

	if ok := operation.MergePullRequest(client, repoOwner, repoName, prInfo); !ok {
		log.Printf("info: cannot merge pull request #%v\n", prNum)
		return
	}

	if repoInfo.DeleteAfterAutoMerge {
		operation.DeleteBranchByPullRequest(client.Git, prInfo)
	}

	queue.RemoveActive()
	log.Printf("info: complete merging #%v into master\n", prNum)

	next, nextInfo := getNextAvailableItem(queue, issueSvc, prSvc, repoOwner, repoName)
	if next == nil {
		log.Printf("info: there is no awating item in the queue of %v\n", repoOwner+repoName)
		return
	}

	nextNum := next.PullRequest

	ok, commit := operation.TryWithMaster(client, repoOwner, repoName, nextInfo)
	if !ok {
		log.Printf("info: we cannot try #%v with the latest `master`.", nextNum)
		// FIXME: We should to try the next pull req in the queue.
		return
	}

	next.SHA = commit.SHA
	queue.SetActive(next)
	log.Printf("info: pin #%v as the active item to queue\n", nextNum)

	log.Println("info: complete to start the next trying")
}

func isIncludeAutoBranch(branches []*github.Branch) bool {
	for _, b := range branches {
		if b == nil {
			continue
		}

		if b.Name == nil {
			continue
		}

		if *b.Name == "auto" {
			return true
		}
	}

	return false
}

func getNextAvailableItem(queue *autoMergeQueue,
	issueSvc *github.IssuesService,
	prSvc *github.PullRequestsService,
	owner string,
	name string) (item *autoMergeQueueItem, info *github.PullRequest) {

	log.Println("Start to find the next item")
	defer log.Println("End to find the next item")

	for {
		ok, next := queue.GetNext()
		if !ok || next == nil {
			log.Printf("debug: there is no awating item in the queue of %v/%v\n", owner, name)
			return
		}

		log.Println("debug: the next item has fetched from queue.")

		nextInfo, _, err := prSvc.Get(owner, name, next.PullRequest)
		if err != nil {
			log.Println("debug: could not fetch the pull request information.")
			continue
		}

		if *nextInfo.State != "open" {
			log.Printf("debug: the pull request #%v has been resolved the state as `%v`\n", next.PullRequest, *nextInfo.State)
			continue
		}

		if !operation.HasStatusLabel(issueSvc, owner, name, next.PullRequest, LABEL_AWAITING_MERGE) {
			continue
		}

		// XXX: We trust the result of detectUnmergeablePR instead of checking mergeable by myself.

		return next, nextInfo
	}
}
