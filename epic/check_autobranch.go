package epic

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"github.com/voyagegroup/popuko/operation"
	"github.com/voyagegroup/popuko/queue"
	"github.com/voyagegroup/popuko/setting"
)

func CheckAutoBranch(ctx context.Context, client *github.Client, autoMergeRepo *queue.AutoMergeQRepo, ev *github.CheckRunEvent) {
	log.Println("info: Start: checkAutoBranch")
	defer log.Println("info: End: checkAutoBranch")

	if *ev.CheckRun.Status != "completed" {
		log.Println("info: Not handle queued or in_progress check-run event")
		return
	}
	log.Printf("info: Start to handle check-run event: %v\n", *ev.CheckRun.Status)

	repoOwner := *ev.Repo.Owner.Login
	repoName := *ev.Repo.Name
	log.Printf("info: Target repository is %v/%v\n", repoOwner, repoName)

	repoInfo := GetRepositoryInfo(ctx, client.Repositories, repoOwner, repoName)
	if repoInfo == nil {
		log.Println("debug: cannot get repositoryInfo")
		return
	}

	log.Println("info: success to load the configure.")

	if !repoInfo.EnableAutoMerge {
		log.Println("info: this repository does not enable merging into master automatically.")
		return
	}
	log.Println("info: Start to handle auto merging the branch.")

	qHandle := autoMergeRepo.Get(repoOwner, repoName)
	if qHandle == nil {
		log.Println("error: cannot get the queue handle")
		return
	}

	qHandle.Lock()
	defer qHandle.Unlock()

	q := qHandle.Load()

	if !q.HasActive() {
		log.Println("info: there is no testing item")
		return
	}

	active := q.GetActive()
	if active == nil {
		log.Println("error: `active` should not be null")
		return
	}
	log.Println("info: got the active item.")

	if !isRelatedToAutoBranch(active, ev, repoInfo.AutoBranchName) {
		log.Printf("info: The event's tip sha does not equal to the one which is tesing actively in %v/%v\n", repoOwner, repoName)
		return
	}
	log.Println("info: the status event is related to auto branch.")

	mergeSucceedItem(ctx, client, repoOwner, repoName, repoInfo, q, ev)

	q.RemoveActive()
	q.Save()

	tryNextItem(ctx, client, repoOwner, repoName, q, repoInfo.AutoBranchName)

	log.Println("info: complete to start the next trying")
}

func isRelatedToAutoBranch(active *queue.AutoMergeQueueItem, ev *github.CheckRunEvent, autoBranch string) bool {
	if *ev.CheckRun.CheckSuite.HeadBranch != autoBranch {
		log.Printf("warn: this check-run event (%v) is not the auto branch\n", *ev.CheckRun.ID)
		return false
	}

	if ok := checkCommitHashOnTrying(active, ev); !ok {
		return false
	}

	log.Println("info: the tip of auto branch is same as `active.SHA`")
	return true
}

func checkCommitHashOnTrying(active *queue.AutoMergeQueueItem, ev *github.CheckRunEvent) bool {
	autoTipSha := active.AutoBranchHead
	if autoTipSha == nil {
		return false
	}

	if *autoTipSha != *ev.CheckRun.HeadSHA {
		log.Printf("debug: The commit hash which contained by the check-run event: %v\n", *ev.CheckRun.HeadSHA)
		log.Printf("debug: The commit hash is pinned to the status queue as the tip of auto branch: %v\n", autoTipSha)
		return false
	}

	return true
}

func mergeSucceedItem(
	ctx context.Context,
	client *github.Client,
	owner string,
	name string,
	repoInfo *setting.RepositoryInfo,
	q *queue.AutoMergeQueue,
	ev *github.CheckRunEvent) bool {

	active := q.GetActive()

	prNum := active.PullRequest

	prInfo, _, err := client.PullRequests.Get(ctx, owner, name, prNum)
	if err != nil {
		log.Println("info: could not fetch the pull request information.")
		return false
	}

	if *prInfo.State != "open" {
		log.Printf("info: the pull request #%v has been resolved the state\n", prNum)
		return true
	}

	if *ev.CheckRun.Conclusion != "success" {
		log.Println("info: could not merge pull request")

		comment := ":collision: The result of what tried to merge this pull request is `" + *ev.CheckRun.Conclusion + "`."
		commentStatus(ctx, client, owner, name, prNum, comment, repoInfo.AutoBranchName)

		currentLabels := operation.GetLabelsByIssue(ctx, client.Issues, owner, name, prNum)
		if currentLabels == nil {
			return false
		}

		labels := operation.AddFailsTestsWithUpsreamLabel(currentLabels)
		_, _, err = client.Issues.ReplaceLabelsForIssue(ctx, owner, name, prNum, labels)
		if err != nil {
			log.Println("warn: could not change labels of the issue")
		}

		return false
	}

	comment := ":tada: The result of what tried to merge this pull request is `" + *ev.CheckRun.Conclusion + "`."
	commentStatus(ctx, client, owner, name, prNum, comment, repoInfo.AutoBranchName)

	if ok := operation.MergePullRequest(ctx, client, owner, name, prInfo, active.PrHead); !ok {
		log.Printf("info: cannot merge pull request #%v\n", prNum)
		return false
	}

	if repoInfo.DeleteAfterAutoMerge {
		operation.DeleteBranchByPullRequest(ctx, client.Git, prInfo)
	}

	log.Printf("info: complete merging #%v into master\n", prNum)
	return true
}

func commentStatus(ctx context.Context, client *github.Client, owner, name string, prNum int, comment string, autoBranch string) {
	status, _, err := client.Repositories.GetCombinedStatus(ctx, owner, name, autoBranch, nil)
	if err != nil {
		log.Println("error: could not get the status about the auto branch.")
	}

	if status != nil {
		comment += "\n\n"

		for _, s := range status.Statuses {
			if s.TargetURL == nil {
				continue
			}

			var item string
			if s.Description == nil || *s.Description == "" {
				item = fmt.Sprintf("* %v\n", *s.TargetURL)
			} else {
				item = fmt.Sprintf("* [%v](%v)\n", *s.Description, *s.TargetURL)
			}

			comment += item
		}
	}

	if ok := operation.AddComment(ctx, client.Issues, owner, name, prNum, comment); !ok {
		log.Println("error: could not write the comment about the result of auto branch.")
	}
}

func tryNextItem(ctx context.Context, client *github.Client, owner, name string, q *queue.AutoMergeQueue, autoBranch string) (ok, hasNext bool) {
	defer q.Save()

	next, nextInfo := getNextAvailableItem(ctx, client, owner, name, q)
	if next == nil {
		log.Printf("info: there is no awating item in the queue of %v/%v\n", owner, name)
		return true, false
	}

	nextNum := next.PullRequest

	ok, commit := operation.TryWithMaster(ctx, client, owner, name, nextInfo, autoBranch)
	if !ok {
		log.Printf("info: we cannot try #%v with the latest `master`.", nextNum)
		return tryNextItem(ctx, client, owner, name, q, autoBranch)
	}

	next.AutoBranchHead = &commit
	q.SetActive(next)
	log.Printf("info: pin #%v as the active item to queue\n", nextNum)

	return true, true
}

func getNextAvailableItem(
	ctx context.Context,
	client *github.Client,
	owner string,
	name string,
	queue *queue.AutoMergeQueue) (*queue.AutoMergeQueueItem, *github.PullRequest) {

	issueSvc := client.Issues
	prSvc := client.PullRequests

	log.Println("Start to find the next item")
	defer log.Println("End to find the next item")

	for {
		ok, next := queue.TakeNext()
		if !ok || next == nil {
			log.Printf("debug: there is no awating item in the queue of %v/%v\n", owner, name)
			return nil, nil
		}

		log.Println("debug: the next item has fetched from queue.")
		prNum := next.PullRequest

		nextInfo, _, err := prSvc.Get(ctx, owner, name, prNum)
		if err != nil {
			log.Println("debug: could not fetch the pull request information.")
			continue
		}

		if next.PrHead != *nextInfo.Head.SHA {
			operation.CommentHeadIsDifferentFromAccepted(ctx, issueSvc, owner, name, prNum)
			continue
		}

		if state := *nextInfo.State; state != "open" {
			log.Printf("debug: the pull request #%v has been resolved the state as `%v`\n", prNum, state)
			continue
		}

		ok, mergeable := operation.IsMergeable(ctx, prSvc, owner, name, prNum, nextInfo)
		if !ok {
			log.Println("info: We treat it as 'mergeable' to avoid miss detection because we could not fetch the pr info,")
			continue
		}

		if !mergeable {
			comment := ":lock: Merge conflict"
			if ok := operation.AddComment(ctx, issueSvc, owner, name, prNum, comment); !ok {
				log.Println("error: could not write the comment about the result of auto branch.")
			}

			currentLabels := operation.GetLabelsByIssue(ctx, issueSvc, owner, name, prNum)
			if currentLabels == nil {
				continue
			}

			labels := operation.AddNeedRebaseLabel(currentLabels)
			log.Printf("debug: the changed labels: %v\n", labels)
			_, _, err = issueSvc.ReplaceLabelsForIssue(ctx, owner, name, prNum, labels)
			if err != nil {
				log.Println("warn: could not change labels of the issue")
			}

			continue
		} else {
			label := operation.GetLabelsByIssue(ctx, issueSvc, owner, name, prNum)
			if label == nil {
				continue
			}

			if !operation.HasLabelInList(label, operation.LABEL_AWAITING_MERGE) {
				continue
			}
		}

		return next, nextInfo
	}
}
