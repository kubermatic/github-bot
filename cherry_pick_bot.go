package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"github.com/adtac/cherry-pick-bot/pkg/controller"

	"github.com/op/go-logging"
)

var configPath = flag.String("config", "config.toml", "Path for the config file")

var logger = logging.MustGetLogger("cherry-pick-bot")

func main() {
	err := loadConfig(*configPath)
	if err != nil {
		die(fmt.Errorf("error while reading configuration file: %v", err))
	}

	loadEnvironment()
	ctx, client := authenticate()

	logger.Notice("Ready for action!")

	controller := controller.New(client)

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

		for _, notification := range unreadNotifications {
			login, project, prId, err := extractNotification(notification)
			if err != nil {
				die(fmt.Errorf("error while extracting notification data: %v", err))
			}

			changeRepo(login, project)

			if *notification.Reason == "mention" {
				// check if email is public
				lastUser, err := getLastUserMentioned(client, ctx, login, project, prId)

				if err != nil {
					logger.Infof("error while getting mentioner: %v", err)
				}

				userName := *lastUser.Login

				logger.Infof("Got a call from %s on %s/%s #%d", userName, login, project, prId)

				if lastUser.Email == nil {
					logger.Infof("%s email isn't public.. Skipping...", userName)
					comment(client, ctx, login, project, prId, invalidEmail)
					continue
				}

				// spoof the cherry-pick committer to make it look like the person commenting
				// did it; also clear any ongoing rebases or cherry-picks
				spoofUser(lastUser)
				clear()

				logger.Infof("Performing cherry pick for %s/%s #%d ...", login, project, prId)
				err = performCherryPick(client, ctx, login, project, prId)
				if err != nil {
					logger.Error(err.Error())
					continue
				}

				logger.Info("Creating pull request ...", login, project, prId)
				err = createCherryPR(client, ctx, login, project, prId)
				if err != nil {
					logger.Error(err.Error())
					continue
				}
			}
		}

		logger.Info("Sleeping ...")
		time.Sleep(time.Duration(conf.SleepTime.Nanoseconds()))
	}
}
