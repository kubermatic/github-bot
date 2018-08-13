package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"

	"github.com/kubermatic/github-bot/pkg/controller"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

const (
	sleepTime = 15 * time.Second
)

func main() {
	flag.Parse()
	githubToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	if githubToken == "" {
		glog.Fatalln("Environment variable 'GITHUB_ACCESSS_TOKEN' must not be emtpy!")
	}
	ctx := context.Background()
	client := getClient(ctx, githubToken)

	controller := controller.New(client)

	for {
		unreadNotifications, err := getUnreadNotifications(ctx, client)
		if err != nil {
			glog.Errorf("Error getting unread notifications: %v", err)
			continue
		}

		glog.V(6).Infof("Got %d notifications!", len(unreadNotifications))
		for _, unreadNotification := range unreadNotifications {
			if serializedMsg, err := json.Marshal(unreadNotification); err == nil {
				glog.V(6).Infof("Serialized notification:\n---\n%s\n---\n", string(serializedMsg))
			} else {
				glog.Errorf("Failed to serialize message: %v", err)
			}
			if err := controller.HandleNotification(unreadNotification); err != nil {
				glog.Errorf("Failed to handle notification: %v", err)
			}
		}

		//TODO: only mark processed notifications as read
		client.Activity.MarkNotificationsRead(ctx, time.Now())

		glog.V(7).Info("sleeping...")
		time.Sleep(sleepTime)
	}
}

func getClient(ctx context.Context, githubToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	return github.NewClient(oauth2.NewClient(ctx, ts))
}

func getUnreadNotifications(ctx context.Context, client *github.Client) ([]*github.Notification, error) {
	notifications, resp, err := client.Activity.ListNotifications(
		ctx, &github.NotificationListOptions{All: true})

	if err != nil {
		return nil, err
	} else if s := resp.Response.StatusCode; s != http.StatusOK {
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
