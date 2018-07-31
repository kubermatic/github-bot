package main

import (
	"log"

	"github.com/google/go-github/github"
)

func handleNotification(notification *github.Notification) error {
	if *notification.Subject.Type != "PullRequest" {
		log.Println("Dropping message because its not a PR...")
		return nil
	}
	return nil
}

//func (s *RepositoriesService) GetComment(ctx context.Context, owner, repo string, id int64) (*RepositoryComment, *Response, error) {
//
//func getComment
