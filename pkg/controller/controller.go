package controller

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"
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
		glog.V(6).Infoln("Dropping message because its not a PR...")
		return nil
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelFunc()

	prNumber, err := getIDFromUrl(notification.Subject.GetURL())
	if err != nil {
		return fmt.Errorf("failed to extract pr number from url: %v", err)
	}

	// We have to fetch the repository because the repository struct in notification.Repository
	// does not contain any clone url
	repo, _, err := c.client.Repositories.Get(ctx,
		notification.Repository.GetOwner().GetLogin(),
		notification.Repository.GetName())
	if err != nil {
		return fmt.Errorf("failed to fetch repo: %v", err)
	}
	latestCommentURL := notification.Subject.GetLatestCommentURL()
	// Notification caused by a comment by someone else are only distinguishible from
	// other kinds of notification by their latestCommentURL pointing to a comment and not to the PR/Issue
	if strings.Contains(latestCommentURL, "issues/comments") {
		body, err := c.getComment(ctx, *repo, notification.Subject.GetLatestCommentURL())
		if err != nil {
			return err
		}
		glog.V(7).Infof("Got message body: %s", body)

		if err := c.syncLabels(ctx, *repo, prNumber, body); err != nil {
			return fmt.Errorf("failed to sync labels: %v", err)
		}
	}

	if err := c.syncCherryPicks(ctx, *repo, prNumber); err != nil {
		return fmt.Errorf("failed to sync cherry picks: %v", err)
	}

	glog.V(6).Infoln("Successfully finished processing notifications")
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
