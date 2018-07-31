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

	labelsToEnsure := getLabelsToEnsureFromMessage(body)
	prNumber, err := getIDFromUrl(notification.Subject.GetURL())
	if err != nil {
		return fmt.Errorf("failed to extract pr number from url: %v", err)
	}
	if err := c.ensureIssueLabelsExist(ctx, *notification.Repository, prNumber, labelsToEnsure); err != nil {
		return fmt.Errorf("failed to ensure labels: %v", err)
	}
	return nil
}

// For the labels part we have to treat it as an issue, because PRs do not have Labels in this lib
func (c *Controller) ensureIssueLabelsExist(ctx context.Context, repo github.Repository, id int, labels []string) error {
	currentLabels, _, err := c.client.Issues.ListLabelsByIssue(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, nil)
	if err != nil {
		fmt.Errorf("failed to fetch labels for issue %s: %v", repo.GetURL(), err)
	}

	for _, desiredLabel := range labels {
		if !labelSliceContains(currentLabels, desiredLabel) {
			// AddLabelsToIssue(ctx context.Context, owner string, repo string, number int, labels []string) ([]*Label, *Response, error)
			newLabels, _, err := c.client.Issues.AddLabelsToIssue(ctx, repo.GetOwner().GetLogin(), repo.GetName(), id, []string{desiredLabel})
			if err != nil {
				return fmt.Errorf("failed to add label %s on issue %s#%v: %v", desiredLabel, repo.GetURL(), id, err)
			}
			log.Printf("Successfully added label %s on issue %s#%v", desiredLabel, repo.GetURL(), id)
			currentLabels = newLabels
		}
	}
	return nil
}

func labelSliceContains(slice []*github.Label, s string) bool {
	for _, element := range slice {
		if element.GetName() == s {
			return true
		}
	}
	return false
}

func getLabelsToEnsureFromMessage(message string) []string {
	var labels []string
	if cherryPickCommandTarget := getCommandTarget(message, cherryPickCommandName); cherryPickCommandTarget != nil {
		labels = append(labels, fmt.Sprintf("cherry-pick/%s", *cherryPickCommandTarget))
	}

	return labels
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
