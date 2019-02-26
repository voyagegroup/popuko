package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/voyagegroup/popuko/epic"
	"github.com/voyagegroup/popuko/input"
	"github.com/voyagegroup/popuko/queue"
	"github.com/voyagegroup/popuko/setting"
)

// AppServer is just an this application.
type AppServer struct {
	githubClient  *github.Client
	autoMergeRepo *queue.AutoMergeQRepo
	setting       *setting.Settings
}

const prefixWebHookPath = "/github"

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

	ctx := req.Context()

	switch event := event.(type) {
	case *github.IssueCommentEvent:
		ok, err := srv.processIssueCommentEvent(ctx, event)
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
		srv.processPushEvent(ctx, event)
		rw.WriteHeader(http.StatusOK)
		return
	case *github.StatusEvent:
		srv.processStatusEvent(ctx, event)
		rw.WriteHeader(http.StatusOK)
		return
	case *github.CheckSuiteEvent:
		srv.processCheckSuiteEvent(ctx, event)
		rw.WriteHeader(http.StatusOK)
		return
	case *github.PullRequestEvent:
		srv.processPullRequestEvent(ctx, event)
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

func (srv *AppServer) processIssueCommentEvent(ctx context.Context, ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("Start: processCommitCommentEvent by %v\n", *ev.Comment.ID)
	defer log.Printf("End: processCommitCommentEvent by %v\n", *ev.Comment.ID)

	if action := ev.Action; (action == nil) || (*action != "created") {
		return false, fmt.Errorf("info: accept `action === \"created\"` only")
	}

	repoOwner := *ev.Repo.Owner.Login
	repo := *ev.Repo.Name
	if !srv.setting.AcceptRepo(repoOwner, repo) {
		n := repoOwner + "/" + repo
		log.Printf("======= error: =======\n This event is from an unaccepted repository: %v\n==============", n)
		return false, fmt.Errorf("%v is not accepted", n)
	}

	body := *ev.Comment.Body
	ok, cmd := input.ParseCommand(body)
	if !ok {
		return false, fmt.Errorf("No operations which this bot should handle")
	}

	if cmd == nil {
		return false, fmt.Errorf("error: unexpected result of parsing comment body")
	}

	repoInfo := epic.GetRepositoryInfo(ctx, srv.githubClient.Repositories, repoOwner, repo)
	if repoInfo == nil {
		return false, fmt.Errorf("debug: cannot get repositoryInfo")
	}

	switch cmd := cmd.(type) {
	case *input.AssignReviewerCommand:
		return epic.AssignReviewer(ctx, srv.githubClient, ev, cmd.Reviewer)
	case *input.AcceptChangeByReviewerCommand:
		commander := epic.AcceptCommand{
			Owner:         repoOwner,
			Name:          repo,
			Client:        srv.githubClient,
			BotName:       config.BotNameForGithub(),
			Info:          repoInfo,
			AutoMergeRepo: srv.autoMergeRepo,
		}
		return commander.AcceptChangesetByReviewer(ctx, ev, cmd)
	case *input.AcceptChangeByOthersCommand:
		commander := epic.AcceptCommand{
			Owner:         repoOwner,
			Name:          repo,
			Client:        srv.githubClient,
			BotName:       config.BotNameForGithub(),
			Info:          repoInfo,
			AutoMergeRepo: srv.autoMergeRepo,
		}
		return commander.AcceptChangesetByOthers(ctx, ev, cmd)
	case *input.CancelApprovedByReviewerCommand:
		commander := epic.CancelApprovedCommand{
			BotName:       config.BotNameForGithub(),
			Client:        srv.githubClient,
			Owner:         repoOwner,
			Name:          repo,
			Number:        *ev.Issue.Number,
			Cmd:           cmd,
			Info:          repoInfo,
			AutoMergeRepo: srv.autoMergeRepo,
		}
		return commander.CancelApprovedChangeSet(ctx, ev)
	default:
		return false, fmt.Errorf("error: unreachable")
	}
}

func (srv *AppServer) processPushEvent(ctx context.Context, ev *github.PushEvent) {
	log.Println("info: Start: processPushEvent by push id")
	defer log.Println("info: End: processPushEvent by push id")

	repoOwner := *ev.Repo.Owner.Name
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)
	if !srv.setting.AcceptRepo(repoOwner, repo) {
		n := repoOwner + "/" + repo
		log.Printf("======= error: =======\n This event is from an unaccepted repository: %v\n==============", n)
		return
	}

	epic.DetectUnmergeablePR(ctx, srv.githubClient, ev)
}

func (srv *AppServer) processStatusEvent(ctx context.Context, ev *github.StatusEvent) {
	log.Println("info: Start: processStatusEvent")
	defer log.Println("info: End: processStatusEvent")

	repoOwner := *ev.Repo.Owner.Login
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)
	if !srv.setting.AcceptRepo(repoOwner, repo) {
		n := repoOwner + "/" + repo
		log.Printf("======= error: =======\n This event is from an unaccepted repository: %v\n==============", n)
		return
	}

	epic.CheckAutoBranchWithStatusEvent(ctx, srv.githubClient, srv.autoMergeRepo, ev)
}

func (srv *AppServer) processCheckSuiteEvent(ctx context.Context, ev *github.CheckSuiteEvent) {
	log.Println("info: Start: processCheckSuiteEvent")
	defer log.Println("info: End: processCheckSuiteEvent")

	repoOwner := *ev.Repo.Owner.Login
	log.Printf("debug: repository owner is %v\n", repoOwner)
	repo := *ev.Repo.Name
	log.Printf("debug: repository name is %v\n", repo)
	if !srv.setting.AcceptRepo(repoOwner, repo) {
		n := repoOwner + "/" + repo
		log.Printf("======= error: =======\n This event is from an unaccepted repository: %v\n==============", n)
		return
	}

	epic.CheckAutoBranchWithCheckSuiteEvent(ctx, srv.githubClient, srv.autoMergeRepo, ev)
}

func (srv *AppServer) processPullRequestEvent(ctx context.Context, ev *github.PullRequestEvent) {
	log.Println("info: Start: processPullRequestEvent")
	defer log.Println("info: End: processPullRequestEvent")

	if action := *ev.Action; action != "closed" {
		log.Printf("info: action type is `%v` which is not handled by this bot\n", action)
		return
	}

	repo := ev.Repo
	if repo == nil {
		log.Println("warn: ev.Repo is nil")
		return
	}

	pr := ev.PullRequest
	if pr == nil {
		log.Println("warn: ev.PullRequest is nil")
		return
	}

	epic.RemoveAllStatusLabel(ctx, srv.githubClient, repo, pr)
}

func createGithubClient(config *setting.Settings) (*github.Client, error) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: config.GithubToken(),
		},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	if config.Github.BaseURL == "" {
		client := github.NewClient(tc)
		return client, nil
	} else {
		if config.Github.UploadURL == "" {
			return nil, errors.New("upload_url is blank. If use enterprise, set config base_url and upload_url.")
		}
		return github.NewEnterpriseClient(config.Github.BaseURL, config.Github.UploadURL, tc)
	}
}

const prefixRestAPI = "/api/v0"
const prefixQueueInfoAPI = "/queue/"

func (srv *AppServer) handleRESTApiRequest(rw http.ResponseWriter, req *http.Request) {
	p := strings.TrimPrefix(req.URL.Path, prefixRestAPI)
	if strings.HasPrefix(p, prefixQueueInfoAPI) {
		repo := strings.TrimPrefix(p, prefixQueueInfoAPI)
		srv.getQueueInfoForRepository(rw, req, repo)
		return
	}

	rw.WriteHeader(http.StatusNotFound)
}

func (srv *AppServer) getQueueInfoForRepository(rw http.ResponseWriter, req *http.Request, repo string) {
	if req.Method != "GET" {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var owner string
	var name string
	{
		tmp := strings.Split(repo, "/")
		if !(len(tmp) == 2) && !(len(tmp) == 3) { // accept `/bar/foo/` style.
			rw.WriteHeader(http.StatusNotFound)
			m := "info: the repo name is invalid"
			log.Printf(m+"%+v\n", tmp)
			io.WriteString(rw, m)
			return
		}

		owner = tmp[0]
		name = tmp[1]
	}

	qhandle := srv.autoMergeRepo.Get(owner, name)
	if qhandle == nil {
		rw.WriteHeader(http.StatusNotFound)
		m := fmt.Sprintf("error: cannot get the queue handle for `%v/%v`", owner, name)
		log.Println(m)
		io.WriteString(rw, m)
		return
	}

	qhandle.Lock()
	defer qhandle.Unlock()

	b := qhandle.LoadAsRawByte()
	if b == nil {
		rw.WriteHeader(http.StatusInternalServerError)
		m := fmt.Sprintf("error: cannot get the queue information for `%v/%v`", owner, name)
		log.Println(m)
		io.WriteString(rw, m)
		return
	}

	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	rw.WriteHeader(http.StatusOK)
	rw.Write(b)
}
