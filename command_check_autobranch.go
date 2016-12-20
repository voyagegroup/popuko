package main

import (
	"log"

	"github.com/google/go-github/github"
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

		_, _, err = issueSvc.CreateComment(repoOwner, repoName, prNum, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
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

		_, _, err = issueSvc.CreateComment(repoOwner, repoName, prNum, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			log.Println("error: could not write the comment about the result of auto branch.")
		}
	}

	{
		comment := ":hourglass: Try to merge " + *prInfo.Head.SHA + " into `master`"
		_, _, err = issueSvc.CreateComment(repoOwner, repoName, prNum, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			log.Println("info: could not create the comment to declare to merge this.")
			return
		}
	}

	{
		// XXX: By the behavior, github uses defautlt merge message
		// if we specify `""` to `commitMessage`.
		_, _, err := prSvc.Merge(repoOwner, repoName, prNum, "", nil)
		if err != nil {
			log.Println("info: could not merge pull request")
			comment := "Could not merge this pull request by:\n```\n" + err.Error() + "\n```"
			_, _, err = issueSvc.CreateComment(repoOwner, repoName, prNum, &github.IssueComment{
				Body: &comment,
			})
			return
		}
	}

	if repoInfo.DeleteAfterAutoMerge {
		branchOwner := *prInfo.Head.Repo.Owner.Login
		log.Printf("debug: branch owner: %v\n", branchOwner)
		branchOwnerRepo := *prInfo.Head.Repo.Name
		log.Printf("debug: repo: %v\n", branchOwnerRepo)
		branchName := *prInfo.Head.Ref
		log.Printf("debug: head ref: %v\n", branchName)

		_, err = client.Git.DeleteRef(branchOwner, branchOwnerRepo, "heads/"+branchName)
		if err != nil {
			log.Println("info: could not delete the merged branch.")
		}
	}

	queue.RemoveActive()
	log.Printf("info: complete merging #%v into master\n", prNum)

	ok, next := queue.GetNext()
	if !ok {
		log.Println("error: this queue should not be empty because `q` is queued just now.")
		return
	}

	if next == nil {
		log.Printf("info: there is no awating item in the queue of %v\n", repoOwner+repoName)
		return
	}

	log.Println("info: the next item has fetched from queue.")

	nextNum := next.PullRequest

	nextInfo, _, err := prSvc.Get(repoOwner, repoName, nextNum)
	if err != nil {
		log.Println("info: could not fetch the pull request information.")
		return
	}

	if *nextInfo.State != "open" {
		log.Printf("info: the pull request #%v has been resolved the state\n", nextNum)
		return
	}

	currentLabels, _, err := issueSvc.ListLabelsByIssue(repoOwner, repoName, nextNum, nil)
	if err != nil {
		log.Println("info: could not get labels by the issue")
		return
	}
	if !hasStatusLabel(currentLabels, LABEL_AWAITING_MERGE) {
		log.Printf("warn: #%v does not have %v\n", nextNum, LABEL_AWAITING_MERGE)
		return
	}

	log.Printf("info: the pullrequest #%v has %v\n", nextNum, LABEL_AWAITING_MERGE)

	ok, _ = createBranchFromMaster(client.Git, repoOwner, repoName, "auto")
	if !ok {
		log.Println("info: cannot create the auto branch")
		return
	}
	log.Println("info: create the auto branch")

	ok, commit := mergeIntoAutoBranch(client.Repositories, repoOwner, repoName, nextInfo.Head)
	if !ok {
		log.Println("info: cannot merge into the auto branch")
		return
	}

	log.Println("info: merge the auto branch")

	next.SHA = commit.SHA
	queue.SetActive(next)
	log.Printf("info: pin #%v as the active item to queue\n", nextNum)

	{
		comment := ":hourglass: " + *nextInfo.Head.SHA + " has been merged into the auto branch " + *commit.HTMLURL
		_, _, err := issueSvc.CreateComment(repoOwner, repoName, nextNum, &github.IssueComment{
			Body: &comment,
		})
		if err != nil {
			log.Println("info: could not create the comment to declare to merge this.")
			return
		}
	}

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
