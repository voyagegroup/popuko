package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/karen-irc/popuko/epic"
	"github.com/karen-irc/popuko/input"
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
	ok, cmd := input.ParseCommand(body)
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

	repoInfo := epic.GetRepositoryInfo(srv.githubClient.Repositories, repoOwner, repo)
	if repoInfo == nil {
		return false, fmt.Errorf("debug: cannot get repositoryInfo")
	}

	switch cmd := cmd.(type) {
	case *input.AssignReviewerCommand:
		return epic.AssignReviewer(srv.githubClient, ev, cmd.Reviewer)
	case *input.AcceptChangeByReviewerCommand:
		commander := epic.AcceptCommand{
			repoOwner,
			repo,
			srv.githubClient,
			config.BotNameForGithub(),
			cmd,
			repoInfo,
			srv.autoMergeRepo,
		}
		return commander.AcceptChangesetByReviewer(ev)
	case *input.AcceptChangeByOthersCommand:
		commander := epic.AcceptCommand{
			repoOwner,
			repo,
			srv.githubClient,
			config.BotNameForGithub(),
			cmd,
			repoInfo,
			srv.autoMergeRepo,
		}
		return commander.AcceptChangesetByOtherReviewer(ev, cmd.Reviewer[0])
	case *input.RejectChangeByReviewerCommand:
		commander := epic.RejectChangeSetCommand{
			BotName:       config.BotNameForGithub(),
			Client:        srv.githubClient,
			Owner:         repoOwner,
			Name:          repo,
			Number:        *ev.Issue.Number,
			Cmd:           cmd,
			Info:          repoInfo,
			AutoMergeRepo: srv.autoMergeRepo,
		}
		return commander.RejectChangeset(ev)
	default:
		return false, fmt.Errorf("error: unreachable")
	}
}

func (srv *AppServer) processPushEvent(ev *github.PushEvent) {
	log.Println("info: Start: processPushEvent by push id")
	defer log.Println("info: End: processPushEvent by push id")
	epic.DetectUnmergeablePR(srv.githubClient, ev)
}

func (srv *AppServer) processStatusEvent(ev *github.StatusEvent) {
	log.Println("info: Start: processStatusEvent")
	defer log.Println("info: End: processStatusEvent")
	epic.CheckAutoBranch(srv.githubClient, srv.autoMergeRepo, ev)
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
