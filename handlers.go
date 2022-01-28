package main

import (
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log"
	"strings"
	"time"
)

// handleEventMessage will take an event and handle it properly based on the type of event
func handleEventMessage(event slackevents.EventsAPIEvent, client *slack.Client) error {
	switch event.Type {
	// First we check if this is an CallbackEvent
	case slackevents.CallbackEvent:

		innerEvent := event.InnerEvent
		// Yet Another Type switch on the actual Data to see if its an AppMentionEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.AppMentionEvent:
			// The application has been mentioned since this Event is a Mention event
			err := handleAppMentionEvent(ev, client)
			if err != nil {
				return err
			}
		}
	default:
		return errors.New("unsupported event type")
	}
	return nil
}

// handleAppMentionEvent is used to take care of the AppMentionEvent when the bot is mentioned
func handleAppMentionEvent(event *slackevents.AppMentionEvent, client *slack.Client) error {

	// Grab the user name based on the ID of the one who mentioned the bot
	user, err := client.GetUserInfo(event.User)
	if err != nil {
		return err
	}

	// Capture what the user said to the bot and standarize
	text := strings.ToLower(event.Text)

	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	// Add Some default context like user who mentioned the bot
	attachment.Fields = []slack.AttachmentField{
		{
			Title: "Date",
			Value: time.Now().String(),
		}, {
			Title: "Initializer",
			Value: user.Name,
		},
	}

	if strings.Contains(text, "hello") {
		// Greet the user
		attachment.Text = fmt.Sprintf("Hello %s", user.Name)
		attachment.Pretext = "Greetings"
		attachment.Color = "#4af030"
	} else {
		// Send a message to the user
		attachment.Text = fmt.Sprintf("How can I help you %s?", user.Name)
		attachment.Pretext = "How can I be of service"
		attachment.Color = "#3d3d3d"
	}

	// Send the message to the channel
	// The Channel is available in the event message
	_, _, err2 := client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err2 != nil {
		return fmt.Errorf("failed to post message: %w", err2)
	}
	return nil
}

// handleSlashCommand will take a slash command and route to the appropriate function
func handleSlashCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// We need to switch depending on the command
	switch command.Command {
	case "/cryptoprice":
		return nil, handleCryptopriceyCommand(command, client)
	case "/cryptoprice-config":
		return nil, handleCryptopriceyConfig(command, client)
	}
	return nil, nil
}

func handleInteractionEvent(interaction slack.InteractionCallback, client *slack.Client) error {
	// This is where we would handle the interaction
	// Switch depending on the Type
	log.Printf("********** The response was of type: %s\n", interaction.Type)
	switch interaction.Type {
	case slack.InteractionTypeViewSubmission:
		if interaction.View.State.Values["Currency"]["currency"].Value != nil {
			err := validateCurrency(getCurrencies(), interaction.View.State.Values["Currency"]["currency"].Value)
			if err == nil {
				// set new currency in YAML
			}
		}
		if interaction.View.State.Values["Tickers"]["tickers"].Value != nil {
			// Set new tickers in YAML
		}
		if interaction.View.State.Values["Cron"]["cron"].Value != nil {
			// Validate cron if possible, set new cron in yaml
		}
	default:

	}

	return nil
}
