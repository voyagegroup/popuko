package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/karen-irc/popuko/queue"
	"github.com/karen-irc/popuko/setting"
)

// AppServer is just an this application.
type AppServer struct {
	githubClient  *github.Client
	autoMergeRepo *queue.AutoMergeQRepo
}

func (srv *AppServer) handleGithubHook(rw http.ResponseWriter, req *http.Request) {
	log.Println("info: Start: handle GitHub WebHook")
	log.Printf("info: Path is %v\n", req.URL.Path)
	defer log.Println("info End: handle GitHub WebHook")

	if req.Method != "POST" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	payload, err := github.ValidatePayload(req, config.WebHookSecret())
	if err != nil {
		rw.WriteHeader(http.StatusPreconditionFailed)
		io.WriteString(rw, err.Error())
		return
	}

	event, err := github.ParseWebHook(github.WebHookType(req), payload)
	if err != nil {
		rw.WriteHeader(http.StatusPreconditionFailed)
		io.WriteString(rw, err.Error())
		return
	}

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		ok, err := srv.processIssueCommentEvent(event)
		rw.WriteHeader(http.StatusOK)
		if ok {
			io.WriteString(rw, "result: \n")
		}

		if err != nil {
			log.Printf("info: %v\n", err)
			io.WriteString(rw, err.Error())
		}
		return
	case *github.PushEvent:
		srv.processPushEvent(event)
		rw.WriteHeader(http.StatusOK)
		return
	case *github.StatusEvent:
		srv.processStatusEvent(event)
		rw.WriteHeader(http.StatusOK)
		return
	default:
		rw.WriteHeader(http.StatusOK)
		log.Println("warn: Unsupported type events")
		log.Println(reflect.TypeOf(event))
		io.WriteString(rw, "This event type is not supported: "+github.WebHookType(req))
		return
	}
}

func (srv *AppServer) processIssueCommentEvent(ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("Start: processCommitCommentEvent by %v\n", *ev.Comment.ID)
	defer log.Printf("End: processCommitCommentEvent by %v\n", *ev.Comment.ID)

	body := *ev.Comment.Body
	ok, cmd := parseCommand(body)
	if !ok {
		return false, fmt.Errorf("No operations which this bot should handle.")
	}

	if cmd == nil {
		return false, fmt.Errorf("error: unexpected result of parsing comment body")
	}

	repoOwner := *ev.Repo.Owner.Login
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)

	repoInfo := createRepositoryInfo(srv.githubClient.Repositories, repoOwner, repo)
	if repoInfo == nil {
		return false, fmt.Errorf("debug: cannot get repositoryInfo")
	}

	var queue *queue.AutoMergeQueue
	if repoInfo.ExperimentalTryOnAutoBranch() {
		srv.autoMergeRepo.Lock()
		queue = srv.autoMergeRepo.Get(repoOwner, repo)
		srv.autoMergeRepo.Unlock()
	}

	switch cmd := cmd.(type) {
	case *AssignReviewerCommand:
		return srv.commandAssignReviewer(ev, cmd.Reviewer)
	case *AcceptChangeByReviewerCommand:
		commander := AcceptCommand{
			repoOwner,
			repo,
			srv.githubClient,
			config.BotNameForGithub(),
			cmd,
			repoInfo,
			queue,
		}
		return commander.commandAcceptChangesetByReviewer(ev)
	case *AcceptChangeByOthersCommand:
		commander := AcceptCommand{
			repoOwner,
			repo,
			srv.githubClient,
			config.BotNameForGithub(),
			cmd,
			repoInfo,
			queue,
		}
		return commander.commandAcceptChangesetByOtherReviewer(ev, cmd.Reviewer[0])
	default:
		return false, fmt.Errorf("error: unreachable")
	}
}

func (srv *AppServer) processPushEvent(ev *github.PushEvent) {
	log.Println("info: Start: processPushEvent by push id")
	defer log.Println("info: End: processPushEvent by push id")
	srv.detectUnmergeablePR(ev)
}

func (srv *AppServer) processStatusEvent(ev *github.StatusEvent) {
	log.Println("info: Start: processStatusEvent")
	defer log.Println("info: End: processStatusEvent")
	srv.checkAutoBranch(ev)
}

func createGithubClient(config *setting.Settings) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: config.GithubToken(),
		},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	return client
}

func createRepositoryInfo(repoSvc *github.RepositoriesService, owner, name string) *setting.RepositoryInfo {
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
