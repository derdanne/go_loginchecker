package main

import (
	"bytes"
	"os"
	"time"
)

func main() {
	var (
		config        Config
		messageBuffer bytes.Buffer
	)

	host := detectHostname()
	enableMail := true
	enableSlack := true
	users := createUsersSlice()

	configFile := "./config.yml"
	if len(os.Args) > 1 {
		arg := os.Args[1]
		configFile = arg
	}
	config.getConfig(configFile)

	if config.Mail.Recipients == nil ||
		config.Mail.From == "" ||
		config.Mail.FromName == "" ||
		config.Mail.Subject == "" {

		println("Alerting via email is turned off. If you want to enable email alerting, make sure you set the needed config params!")
		enableMail = false
	}

	if config.Slack.WebHookUrl == "" ||
		config.Slack.Channel == "" ||
		config.Slack.Author == "" ||
		config.Slack.Message == "" ||
		config.Slack.Username == "" ||
		config.Slack.IconEmoji == "" {

		println("Alerting via slack is turned off. If you want to enable slack alerting, make sure you set the needed config params!")
		enableSlack = false
	}

	for {
		users = detectUser(users)
		notify := false
		messageBuffer.WriteString("Unauthorized user login(s):\n")

		for _, user := range users {
			if isNotAllowedUser(user.Username, config.AllowedUsers) || isNotAllowedHost(user.Hostname, config.AllowedAddresses) {
				if user.TimeDetected == user.TimeNotified || time.Now().Sub(user.TimeNotified) > time.Duration(config.GraceTime)*time.Second {
					messageBuffer.WriteString("\n")
					messageBuffer.WriteString("User ")
					messageBuffer.WriteString(user.Username)
					messageBuffer.WriteString(" is not allowed to access this host from ")
					messageBuffer.WriteString(user.Hostname)
					messageBuffer.WriteString("!")
					user.updateTimeNotified(time.Now())
					notify = true
				}
			}
		}

		if notify {
			message := messageBuffer.String()
			println(message)
			if enableMail {
				for _, alertMailTo := range config.Mail.Recipients {
					sendMail(
						alertMailTo,
						config.Mail.From,
						config.Mail.FromName,
						config.Mail.Subject+" on host "+host.Hostname,
						message,
					)
				}
			}
			if enableSlack {
				sendSlack(
					config.Slack.WebHookUrl,
					config.Slack.Channel,
					config.Slack.Author,
					config.Slack.Message+" on host "+host.Hostname,
					config.Slack.Username,
					config.Slack.IconEmoji,
					message,
				)
			}
		}
		messageBuffer.Reset()
		time.Sleep(time.Duration(config.RecheckTime) * time.Second)
	}
}
