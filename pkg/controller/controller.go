package controller

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/github"
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
	return nil
}

func (c *Controller) getComment(ctx context.Context, repo github.Repository, commentUrl string) (string, error) {
	commentID, err := getCommentIDFromCommentUrl(commentUrl)
	if err != nil {
		return "", err
	}

	repoComment, _, err := c.client.Issues.GetComment(ctx, repo.GetOwner().GetLogin(), repo.GetName(), commentID)
	if err != nil {
		return "", err
	}
	return repoComment.GetBody(), nil
}

func getCommentIDFromCommentUrl(commentUrl string) (int, error) {
	commentURLSplitted := strings.Split(commentUrl, "/")
	commentIDAsString := commentURLSplitted[len(commentURLSplitted)-1]
	return strconv.Atoi(commentIDAsString)
}
