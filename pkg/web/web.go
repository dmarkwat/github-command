package web

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dmarkwat/github-command/pkg/commands"
	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	ctx    = context.Background()
	client = github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GH_ACCESS_TOKEN")},
	)))

	help = commands.NewHelpCommand(ctx, client)

	commandMap = map[string]commands.Command{
		"request":   commands.NewRequestCommand(ctx, client),
		"unrequest": commands.NewUnrequestCommand(ctx, client),
		"help":      help,
	}

	lineStartRegex = regexp.MustCompile(`^`)
)

func replyErrorComment(ctx context.Context, client *github.Client, event *github.IssueCommentEvent, reply error) {
	replyStr := fmt.Sprintf("Whoops! Error encountered: %s", reply.Error())
	_, response, err := client.Issues.CreateComment(ctx, event.Repo.GetOrganization().GetLogin(), event.Repo.GetName(), event.Issue.GetNumber(), &github.IssueComment{
		Body: &replyStr,
	})
	if err != nil {
		log.Print(err)
		return
	}
	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		rep, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Print(err)
			return
		}
		log.Printf("error replying to commenter: %s", string(rep))
	}
}

func makeBlockQuote(text string) string {
	split := lineStartRegex.Split(text, -1)
	for idx, part := range split {
		split[idx] = "> " + part
	}
	return strings.Join(split, "")
}

func handleIssueComment(event *github.IssueCommentEvent) error {
	switch *event.Action {
	case "created":
		body := event.Comment.Body
		if strings.HasPrefix(*body, "/") {
			matches := commands.CommandRegex.FindStringSubmatch(*body)
			if matches == nil || len(matches) < 2 {
				replyErrorComment(ctx, client, event, fmt.Errorf("no commands matched the request: %s", *body))
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
			if err := cmd.Handle(event); err != nil {
				// sink the error here: we consider this a valid code flow--not an error needing a failed status code
				replyErrorComment(ctx, client, event, err)
			}
			return nil
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

func DigestsMatch(body []byte, webhookKey, signature string) error {
	h := hmac.New(sha1.New, []byte(webhookKey))
	_, err := h.Write(body)
	if err != nil {
		return err
	}
	generated := fmt.Sprintf("sha1=%s", hex.EncodeToString(h.Sum(nil)))
	if subtle.ConstantTimeCompare([]byte(generated), []byte(signature)) != 1 {
		return fmt.Errorf("signatures did not match: %s != %s", generated, signature)
	}
	return nil
}

func HandleWebhook(resp http.ResponseWriter, req *http.Request, webhookKey string) {
	var event github.Event

	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	signature := req.Header.Get("X-Hub-Signature")
	if signature == "" {
		log.Printf("no signature provided in request header")
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	err = DigestsMatch(body, webhookKey, signature)
	if err != nil {
		log.Print(err)
		resp.WriteHeader(http.StatusForbidden)
		return
	}

	err = json.Unmarshal(body, &event)
	if err != nil {
		log.Print(err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	payload, err := event.ParsePayload()
	if err != nil {
		log.Print(err)
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	switch v := payload.(type) {
	case github.IssueCommentEvent:
		err = handleIssueComment(&v)
		if err != nil {
			log.Print(err)
			resp.WriteHeader(http.StatusInternalServerError)
		}
	default:
		log.Printf("Unrecognized payload type: %v", v)
		resp.WriteHeader(http.StatusBadRequest)
	}
}
