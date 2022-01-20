package main

import (
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
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
	_, _, err = client.PostMessage(event.Channel, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}

// handleSlashCommand will take a slash command and route to the appropriate function
func handleSlashCommand(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// We need to switch depending on the command
	switch command.Command {
	case "/cryptoprice":
		// This was a hello command, so pass it along to the proper function
		return nil, handleCryptopriceyCommand(command, client)
	case "/cryptoprice-config":
		return handleCryptopriceyConfig(command, client)
	}

	return nil, nil
}

// handleCryptopriceyConfig will allow the user to configure Cryptopricey via Slack UI
func handleCryptopriceyConfig(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	// Allow user to select the time for announcements
	timePicker := slack.NewTimePickerBlockElement("timeselection")

	// Create the Accessory that will be included in the Block and add the radiobox to it
	timeAccessory := slack.NewAccessory(timePicker)

	headerText := slack.NewTextBlockObject(slack.MarkdownType, "### CryptoPricey Config ###", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	timeText := slack.NewTextBlockObject(slack.MarkdownType, "Please select an announcement time:", false, false)
	timeSection := slack.NewSectionBlock(timeText, nil, timeAccessory)

	selectMenuTitle := slack.NewTextBlockObject("plain_text", "Please select an announcement frequency:", false, false)
	selectMenuText := slack.NewTextBlockObject("plain_text", "Frequencies:", false, false)

	selectMenuElement := slack.NewOptionsSelectBlockElement(
		"static_select",
		selectMenuText,
		"frequency",
		&slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: "plain_text", Text: "3 Hours"}, Value: "3h"},
		&slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: "plain_text", Text: "6 Hours"}, Value: "6h"},
		&slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: "plain_text", Text: "12 Hours"}, Value: "12h"},
		&slack.OptionBlockObject{Text: &slack.TextBlockObject{Type: "plain_text", Text: "24 Hours"}, Value: "24h"},
	)

	selectMenuBlock := slack.NewInputBlock("announcement_frequency", selectMenuTitle, selectMenuElement)

	// Add Blocks to the attachment
	attachment.Blocks = slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			timeSection,
			selectMenuBlock,
		},
	}

	attachment.Color = "#4af030"
	return attachment, nil

}

// handleCryptopriceyCommand will take care of /cryptoprice submissions
func handleCryptopriceyCommand(command slack.SlashCommand, client *slack.Client) error {
	// The Input is found in the text field so
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	adaPrice := getCryptoPrice(command.Text)

	// Greet the user
	attachment.Text = fmt.Sprintf("The spot price of %s is '%s'\n", command.Text, adaPrice)
	attachment.Color = "#4af030"

	// Send the message to the channel
	// The Channel is available in the command.ChannelID
	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("failed to post message: %w", err)
	}
	return nil
}
