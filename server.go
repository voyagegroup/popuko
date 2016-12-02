package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// AppServer is just an this application.
type AppServer struct {
	githubClient *github.Client
}

func (srv *AppServer) handleGithubHook(rw http.ResponseWriter, req *http.Request) {
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
		if ok {
			io.WriteString(rw, "result: \n")
		}

		if err != nil {
			io.WriteString(rw, err.Error())
		}

		rw.WriteHeader(http.StatusOK)
		return
	case *github.PushEvent:
		srv.processPushEvent(event)
		rw.WriteHeader(http.StatusOK)
		return
	default:
		rw.WriteHeader(http.StatusOK)
		log.Println(reflect.TypeOf(event))
		io.WriteString(rw, "This event type is not supported: "+github.WebHookType(req))
		return
	}
}

func (srv *AppServer) processIssueCommentEvent(ev *github.IssueCommentEvent) (bool, error) {
	log.Printf("Start: processCommitCommentEvent by %v\n", *ev.Comment.ID)
	defer log.Printf("End: processCommitCommentEvent by %v\n", *ev.Comment.ID)

	body := ev.Comment.Body
	tmp := strings.Split(*body, " ")

	// If there are no possibility that the comment body is not formatted
	// `@botname command`, stop to process.
	if len(tmp) < 2 {
		err := fmt.Errorf("The comment body is not expected format: `%v`\n", body)
		return false, err
	}

	trigger := tmp[0]
	command := tmp[1]

	log.Printf("trigger: %v\n", trigger)
	log.Printf("command: %v\n", command)

	var args string
	if len(tmp) > 2 {
		args = tmp[2]
		log.Printf("args: %v\n", args)
	}

	repoOwner := *ev.Repo.Owner.Login
	repo := *ev.Repo.Name
	repoInfo := config.Repositories().Get(repoOwner, repo)
	if repoInfo == nil {
		return false, fmt.Errorf("Not found registred repo config.")
	}

	// `@reviewer r?`
	{
		target := strings.TrimPrefix(trigger, "@")
		if repoInfo.Reviewers().Has(target) && command == "r?" {
			return srv.commandAssignReviewer(ev, target)
		}
	}

	// not for me
	if trigger != config.BotNameForGithub() {
		err := fmt.Errorf("The trigger is not me: `%v`\n", trigger)
		return false, err
	}

	commander := AcceptCommand{
		srv.githubClient,
		repoInfo,
	}
	// `@botname command`
	if command == "r+" {
		return commander.commandAcceptChangesetByReviewer(ev)
	} else if strings.Index(command, "r=") == 0 {
		return commander.commandAcceptChangesetByOtherReviewer(ev, command)
	}

	return false, fmt.Errorf("No operations which this bot should handle.")
}

func (srv *AppServer) processPushEvent(ev *github.PushEvent) {
	log.Printf("Start: processPushEvent by push id: %v\n", ev.PushID)
	defer log.Printf("End: processPushEvent by push id: %v\n", ev.PushID)
	srv.detectUnmergeablePR(ev)
}

func createGithubClient(config *Settings) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: config.GithubToken(),
		},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)
	return client
}
