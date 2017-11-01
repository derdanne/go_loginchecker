package main

import (
	"github.com/monochromegane/slack-incoming-webhooks"
	"log"
)

func sendSlack(webhookUrl string, channel string, author string, message string, username string, iconEmoji string, attachementMessage string) {
	client := &slack_incoming_webhooks.Client{
		WebhookURL: webhookUrl,
	}

	attachment := &slack_incoming_webhooks.Attachment{
		AuthorName: author,
		Text:       attachementMessage,
		Color:      "danger",
	}

	payload := &slack_incoming_webhooks.Payload{
		Text:      message,
		Channel:   channel,
		Username:  username,
		IconEmoji: iconEmoji,
	}
	payload.AddAttachment(attachment)

	execErr := client.Post(payload)
	if execErr != nil {
		log.Print(execErr)
	}
}
