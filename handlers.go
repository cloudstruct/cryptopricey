package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"log"
	"net/http"
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
func handleSlashCommand(command slack.SlashCommand, client *slack.Client, httpClient *http.Client) (interface{}, error) {
	// We need to switch depending on the command
	switch command.Command {
	case "/cryptoprice":
		return nil, handleCryptopriceyCommand(command, client, httpClient)
	case "/cryptoprice-config":
		return nil, handleCryptopriceyConfig(command, client)
	}
	return nil, nil
}

func handleInteractionEvent(mainCron *cron.Cron, interaction slack.InteractionCallback, client *slack.Client, httpClient *http.Client) error {
	data := readYAML()

	currencyAttachment := slack.Attachment{}
	tickersAttachment := slack.Attachment{}
	cronAttachment := slack.Attachment{}

	currencyAttachment.Color = "#4af030"
	tickersAttachment.Color = "#5af035"
	cronAttachment.Color = "#6af039"

	yamlModified := false

	// This is where we would handle the interaction
	// Switch depending on the Type
	switch interaction.Type {
	case slack.InteractionTypeViewSubmission:
		if interaction.View.State.Values["Currency"]["currency"].Value != "" {
			currencyValue := interaction.View.State.Values["Currency"]["currency"].Value
			err := validateCurrency(getCurrencies(), currencyValue)
			if err == nil {
				// set new currency in YAML struct
				data[interaction.View.PrivateMetadata].Currency = currencyValue
				currencyAttachment.Text = fmt.Sprintf("Base Currency has been updated to `%s`.", data[interaction.View.PrivateMetadata].Currency)
				yamlModified = true
			} else {
				// Report invalid currency
				log.Printf("********** Currency '%s' NOT validated successfully.", currencyValue)
				currencyAttachment.Text = fmt.Sprintf("Currency *not* updated.  Invalid currency provided: ` %s `", currencyValue)
			}
		}
		if interaction.View.State.Values["Tickers"]["tickers"].Value != "" {
			// Set new tickers in YAML
			data[interaction.View.PrivateMetadata].Tickers = interaction.View.State.Values["Tickers"]["tickers"].Value
			tickersAttachment.Text = fmt.Sprintf("Ticker list has been updated to `%s`.", data[interaction.View.PrivateMetadata].Tickers)
			yamlModified = true
		}
		if interaction.View.State.Values["Cron"]["cron"].Value != "" {
			// Validate cron if possible, set new cron in yaml
			data[interaction.View.PrivateMetadata].Cron = interaction.View.State.Values["Cron"]["cron"].Value
			cronAttachment.Text = fmt.Sprintf("Cron has been updated to `%s`.", data[interaction.View.PrivateMetadata].Cron)
			yamlModified = true
		}
	default:

	}

	if yamlModified == true {
		err := writeYAML(data)
		if err != nil {
			return fmt.Errorf("********* Error writing to YAML config: %w", err)
		}

		// Rebuild the cron list
		_, err = rebuildCron(mainCron, client, httpClient)
		if err != nil {
			return fmt.Errorf("********* Error rebuilding Cron: %w", err)
		}

		// Send the message to the channel
		_, _, err = client.PostMessage(interaction.View.PrivateMetadata, slack.MsgOptionAttachments(currencyAttachment, tickersAttachment, cronAttachment))
		if err != nil {
			return fmt.Errorf("********* failed to post message: %w", err)
		}
	}

	return nil
}
