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

	for true {
		unreadNotifications, err := getUnreadNotifications(client, ctx)
		if err != nil {
			die(fmt.Errorf("error while getting unread notifications: %v", err))
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
		time.Sleep(time.Duration(conf.SleepTime.Nanoseconds()))
	}
}

func getClient(githubToken string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: conf.AccessToken},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	return github.NewClient(tc)
}
