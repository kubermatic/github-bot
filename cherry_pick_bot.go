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

		logger.Info("Sleeping ...")
		time.Sleep(time.Duration(conf.SleepTime.Nanoseconds()))
	}
}
