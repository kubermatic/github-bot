package controller

import (
	"log"

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
	return nil
}
