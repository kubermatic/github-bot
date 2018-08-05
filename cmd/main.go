package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/adtac/cherry-pick-bot/pkg/controller"

	"github.com/google/go-github/github"
	"github.com/op/go-logging"
)

var configPath = flag.String("config", "config.toml", "Path for the config file")

var logger = logging.MustGetLogger("cherry-pick-bot")

func main() {
	githubToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubToken == "" {
		log.Fatalln("Environment variable 'GITHUB_ACCESSS_TOKEN' must not be emtpy!")
	}
	client := getClient(githubToken)

	logger.Notice("Ready for action!")

	controller := controller.New(client)
	ctx := context.Background()

	for {
		unreadNotifications, err := getUnreadNotifications(client, ctx)
		if err != nil {
			log.Printf("Error getting unread notifications: %v", err)
			continue
		}

		logger.Infof("Got %d notifications!", len(unreadNotifications))
		for _, unreadNotification := range unreadNotifications {
			if serializedMsg, err := json.Marshal(unreadNotification); err == nil {
				logger.Infof("Serialized notification:\n---\n%s\n---\n", string(serializedMsg))
			} else {
				logger.Infof("Failed to serialize message: %v", err)
			}
			if err := controller.HandleNotification(unreadNotification); err != nil {
				logger.Infof("Failed to handle notification: %v", err)
			}
		}

		//TODO: only mark processed notifications as read
		client.Activity.MarkNotificationsRead(ctx, time.Now())

		logger.Info("Sleeping ...")
		time.Sleep(15 * time.Second)
	}
}

func getClient(githubToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}

func getUnreadNotifications(client *github.Client, ctx context.Context) ([]*github.Notification, error) {
	notifications, resp, err := client.Activity.ListNotifications(
		ctx, &github.NotificationListOptions{All: true})

	if err != nil {
		return nil, err
	} else if s := resp.Response.StatusCode; s != 200 {
		return nil, fmt.Errorf("response status code is %d", s)
	}

	unreadNotifications := make([]*github.Notification, 0)
	for _, notification := range notifications {
		if notification.GetUnread() {
			unreadNotifications = append(unreadNotifications, notification)
		}
	}
	return unreadNotifications, nil
}
