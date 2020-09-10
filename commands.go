package main

import (
	"context"
	"fmt"
	"github.com/google/go-github/v32/github"
	"io/ioutil"
	"net/http"
)

type command interface {
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
	command
	ctx    context.Context
	client *github.Client
}

func (requestCommand) String() string { return "request" }

func (requestCommand) Handle(event *github.IssueCommentEvent) error {
	return nil
}

func NewUnrequestCommand(ctx context.Context, client *github.Client) unrequestCommand {
	return unrequestCommand{
		ctx:    ctx,
		client: client,
	}
}

type unrequestCommand struct {
	command
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
	command
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
