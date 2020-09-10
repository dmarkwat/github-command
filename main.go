package main

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
)

var (
	ctx    = context.Background()
	client = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_ACCESS_TOKEN")},
	)))

	help = NewHelpCommand(ctx, client)

	commandMap = map[string]command{
		"request":   NewRequestCommand(ctx, client),
		"unrequest": NewUnrequestCommand(ctx, client),
		"help":      help,
	}

	commandRegex = regexp.MustCompile(`^/([a-zA-Z0-9_-]+)`)
)

func handleIssueComment(event *github.IssueCommentEvent) error {
	switch *event.Action {
	case "created":
		body := event.Comment.Body
		if strings.HasPrefix(*body, "/") {
			matches := commandRegex.FindStringSubmatch(*body)
			if matches == nil || len(matches) < 2 {
				log.Printf("no matches found for request: %s", *event.Issue.URL)
				return nil
			}
			// not first entry per FindStringSubmatch doc
			cmdStr := matches[1]
			cmd, ok := commandMap[cmdStr]
			if !ok {
				log.Printf("/%s command isn't available for: %s", cmdStr, *event.Issue.URL)
				cmd = help
			} else {
				log.Printf("handling /%s command for: %s", cmd.String(), *event.Issue.URL)
			}
			return cmd.Handle(event)
		} else {
			// todo debug only; expose as metric
			log.Printf("comment did not start with /")
		}
	default:
		// todo debug only; expose as metric
		log.Printf("handler only responds to creating issues: %s", *event.Action)
	}
	return nil
}

func handleWebhook(resp http.ResponseWriter, req *http.Request) {
	var event github.Event

	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &event)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := event.ParsePayload()
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	switch v := payload.(type) {
	case github.IssueCommentEvent:
		err = handleIssueComment(&v)
		if err != nil {
			resp.WriteHeader(http.StatusInternalServerError)
		}
	default:
		resp.WriteHeader(http.StatusBadRequest)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/command", handleWebhook)

	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Fatal(server.ListenAndServe())
}
