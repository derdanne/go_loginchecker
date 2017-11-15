package main

import (
	"github.com/monochromegane/slack-incoming-webhooks"
	"log"
)

func sendSlack(webHookUrl string, channel string, author string, message string, username string, iconEmoji string, attachmentMessage string) {
	client := &slack_incoming_webhooks.Client{
		WebhookURL: webHookUrl,
	}

	attachment := &slack_incoming_webhooks.Attachment{
		AuthorName: author,
		Text:       attachmentMessage,
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
