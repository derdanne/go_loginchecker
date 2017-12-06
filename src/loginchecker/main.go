package main

import (
	"bytes"
	"log"
	"log/syslog"
	"os"
	"time"
	"path/filepath"
)

func main() {
	var (
		config          Config
		messageBuffer   bytes.Buffer
		messageBufferSL bytes.Buffer
	)

	logWriter, logError := syslog.New(syslog.LOG_DAEMON, "loginchecker")
	if logError == nil {
		log.SetOutput(logWriter)
	}

	pwd, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	logWriter.Info("Starting up loginchecker in " + pwd)

	host := detectHostname()
	enableMail := true
	enableSlack := true
	users := createUsersSlice()

	configFile := "./config.yml"
	if len(os.Args) > 1 {
		arg := os.Args[1]
		configFile = arg
	}

	logWriter.Info("Loading config from " + configFile)
	config.getConfig(configFile)

	if config.Mail.Recipients == nil ||
		config.Mail.From == "" ||
		config.Mail.FromName == "" ||
		config.Mail.Subject == "" {

		logWriter.Warning("Alerting via email is turned off. If you want to enable email alerting, make sure you set the needed config params!")
		enableMail = false
	}

	if config.Slack.WebHookUrl == "" ||
		config.Slack.Channel == "" ||
		config.Slack.Author == "" ||
		config.Slack.Message == "" ||
		config.Slack.Username == "" ||
		config.Slack.IconEmoji == "" {

		logWriter.Warning("Alerting via slack is turned off. If you want to enable slack alerting, make sure you set the needed config params!")
		enableSlack = false
	}

	for {
		users = detectUser(users)
		notify := false
		newUser := false
		oldUser := false

		messageBuffer.WriteString("User login(s):\n")

		for _, user := range users {
			if isNotAllowedUser(user.Username, config.AllowedUsers) || isNotAllowedHost(user.Hostname, config.AllowedAddresses) {
				if user.TimeDetected == user.TimeNotified {
					newUser = true
				} else if time.Now().Sub(user.TimeNotified) > time.Duration(config.GraceTime)*time.Second {
					oldUser = true
				}

				if oldUser || newUser {
					messageBuffer.WriteString("\n")
					messageBuffer.WriteString("User ")
					messageBuffer.WriteString(user.Username)
					if newUser {
						messageBuffer.WriteString(" logged in from ")
					} else if oldUser {
						messageBuffer.WriteString(" still logged in from ")
					}
					messageBuffer.WriteString(user.Hostname)

					messageBufferSL.WriteString("User ")
					messageBufferSL.WriteString(user.Username)
					if newUser {
						messageBufferSL.WriteString(" logged in from ")
					} else if oldUser {
						messageBufferSL.WriteString(" still logged in from ")
					}
					messageBufferSL.WriteString(user.Hostname)
					logWriter.Warning(messageBufferSL.String())
					messageBufferSL.Reset()

					user.updateTimeNotified(time.Now())
					notify = true
				}

			}
		}

		if notify {
			message := messageBuffer.String()

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
