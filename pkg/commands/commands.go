package commands

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

var (
	CommandRegex = regexp.MustCompile(`^/([a-zA-Z0-9_-]+)`)

	requestReviewers = regexp.MustCompile(`[ ,;:]]`)
)

type Command interface {
	String() string
	Handle(event *github.IssueCommentEvent) error
}

func NewRequestCommand(ctx context.Context, client *github.Client) requestCommand {
	return requestCommand{
		ctx:    ctx,
		client: client,
	}
}

type requestCommand struct {
	Command
	ctx    context.Context
	client *github.Client
}

func (requestCommand) String() string { return "request" }

func (c requestCommand) Handle(event *github.IssueCommentEvent) error {
	split := CommandRegex.Split(event.GetComment().GetBody(), 1)
	if len(split) == 0 {
		return fmt.Errorf("must request at least one reviewer")
	}

	reviewers := requestReviewers.Split(strings.TrimSpace(split[0]), -1)

	if event.Issue.IsPullRequest() {
		_, response, err := c.client.PullRequests.RequestReviewers(c.ctx, event.Repo.GetOrganization().GetLogin(), event.Repo.GetName(), event.Issue.GetNumber(), github.ReviewersRequest{
			Reviewers: reviewers,
			// todo
			//TeamReviewers: nil,
		})
		if err != nil {
			log.Print(err)
			return fmt.Errorf("error adding reviewers: %d", response.StatusCode)
		}
		if response.StatusCode != http.StatusOK {
			defer response.Body.Close()
			rep, err := ioutil.ReadAll(response.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf("%s", string(rep))
		}
	} else {
		return fmt.Errorf("requesting reviewers only works on Pull Requests")
	}
	return nil
}

func NewUnrequestCommand(ctx context.Context, client *github.Client) unrequestCommand {
	return unrequestCommand{
		ctx:    ctx,
		client: client,
	}
}

type unrequestCommand struct {
	Command
	ctx    context.Context
	client *github.Client
}

func (unrequestCommand) String() string { return "unrequest" }

func (unrequestCommand) Handle(event *github.IssueCommentEvent) error {
	return nil
}

func NewHelpCommand(ctx context.Context, client *github.Client) helpCommand {
	return helpCommand{
		ctx:    ctx,
		client: client,
	}
}

type helpCommand struct {
	Command
	ctx    context.Context
	client *github.Client
}

// todo implement
func helpText() string { return "" }

func (helpCommand) String() string { return "help" }

func (c helpCommand) Handle(event *github.IssueCommentEvent) error {
	body := helpText()
	// todo process users and organizations alike
	_, response, err := c.client.Issues.CreateComment(c.ctx, event.Repo.GetOrganization().GetLogin(), event.Repo.GetName(), event.Issue.GetNumber(), &github.IssueComment{
		Body: &body,
	})
	if err != nil {
		return err
	}
	if response.StatusCode != http.StatusOK {
		defer response.Body.Close()
		all, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("error reading body: %w", err)
		}
		return fmt.Errorf("error publishing help: %d:%s", response.StatusCode, string(all))
	}
	return nil
}
