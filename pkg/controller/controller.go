package controller

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

const (
	cherryPickCommandName = "/cherry-pick"
	cherryPickLabelPrefix = "cherry-pick"
)

type Controller struct {
	client *github.Client
}

func New(client *github.Client) *Controller {
	return &Controller{client: client}
}

func (c *Controller) HandleNotification(notification *github.Notification) error {
	if *notification.Subject.Type != "PullRequest" {
		log.Println("Dropping message because its not a PR...")
		return nil
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelFunc()

	body, err := c.getComment(ctx, *notification.Repository, notification.Subject.GetLatestCommentURL())
	if err != nil {
		return err
	}
	log.Printf("Got message body: %s", body)

	prNumber, err := getIDFromUrl(notification.Subject.GetURL())
	if err != nil {
		return fmt.Errorf("failed to extract pr number from url: %v", err)
	}
	if err := c.syncLabels(ctx, *notification.Repository, prNumber, body); err != nil {
		return fmt.Errorf("failed to sync labels: %v", err)
	}
	return nil
}

func getCommandTarget(message, command string) *string {
	words := strings.Fields(message)
	for idx, word := range words {
		if word == cherryPickCommandName && len(message) >= idx {
			return &words[idx+1]
		}
	}
	return nil
}

func (c *Controller) getComment(ctx context.Context, repo github.Repository, commentUrl string) (string, error) {
	commentID, err := getIDFromUrl(commentUrl)
	if err != nil {
		return "", err
	}

	repoComment, _, err := c.client.Issues.GetComment(ctx, repo.GetOwner().GetLogin(), repo.GetName(), commentID)
	if err != nil {
		return "", err
	}
	return repoComment.GetBody(), nil
}

func getIDFromUrl(commentUrl string) (int, error) {
	commentURLSplitted := strings.Split(commentUrl, "/")
	commentIDAsString := commentURLSplitted[len(commentURLSplitted)-1]
	return strconv.Atoi(commentIDAsString)
}

func (c *Controller) writeMessageToIssue(ctx context.Context, repo github.Repository, id int, message string) error {
	issueComment := &github.IssueComment{Body: &message}
	_, _, err := c.client.Issues.CreateComment(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, issueComment)
	return err
}
