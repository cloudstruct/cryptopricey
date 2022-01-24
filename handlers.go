package main

import (
	"os"
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"gopkg.in/yaml.v2"
	"io/ioutil"
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
		return handleCryptopriceyConfig(command, client)
	case "/cryptoprice-tickers":
		return nil, handleCryptopriceyTickers(command, client)
	}
	return nil, nil
}

type DataFile struct {
        Tickers string  `yaml:"tickers"`
        Time    string  `yaml:"time"`
        Frequency       string  `yaml:"frequency"`
}

// handleCryptopriceyTickers will take care of /cryptoprice submissions
func handleCryptopriceyTickers(command slack.SlashCommand, client *slack.Client) error {
	data := make(map[string]*DataFile)

	dataDir := os.Getenv("DATA_DIR")
	configFile := dataDir + "/conf.yaml"
	log.Printf("********** Loading file: " + configFile)


	if len(command.Text) < 1 {
		log.Printf("*********** Empty Ticker Symbol")
		return nil
	}

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("*********** YAML config does not exist, continuing.")
		} else {
			log.Printf("yamlFile.Get err   #%v ", err)
			panic(err)
		}
	}

	if err := yaml.Unmarshal(yamlFile, &data); err != nil {
		log.Fatal(err)
	}

	log.Printf("********** Old Tickers: %+v", data[command.ChannelID].Tickers)
	data[command.ChannelID].Tickers = command.Text
	log.Printf("********** Updated Tickers: %+v", data[command.ChannelID].Tickers)

	dataOut, err2 := yaml.Marshal(&data)
	if err2 != nil {
		log.Fatal(err)
	}

	if err3 := ioutil.WriteFile(configFile, dataOut, 0600); err3 != nil {
	     log.Fatal(err3)
	}

	attachment := slack.Attachment{}
	attachment.Text = fmt.Sprintf("Ticker list has been updated to [%s].", command.Text)
	attachment.Color = "#4af030"

	// Send the message to the channel
	// The Channel is available in the command.ChannelID
	_, _, err3 := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err3 != nil {
		return fmt.Errorf("********* failed to post message: %w", err)
	}
	return nil
}

// handleCryptopriceyConfig will allow the user to configure Cryptopricey via Slack UI
func handleCryptopriceyConfig(command slack.SlashCommand, client *slack.Client) (interface{}, error) {
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	// Allow user to select the time for announcements
	timePicker := slack.NewTimePickerBlockElement("timeselection")

	// Create the Accessories that will be included in the Block
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
		return fmt.Errorf("********* failed to post message: %w", err)
	}
	return nil
}

func handleInteractionEvent(interaction slack.InteractionCallback, client *slack.Client) error {
	// This is where we would handle the interaction
	// Switch depending on the Type
	log.Printf("********** The response was of type: %s\n", interaction.Type)
	switch interaction.Type {
	case slack.InteractionTypeBlockActions:
		// This is a block action, so we need to handle it
		for _, action := range interaction.ActionCallback.BlockActions {
			if action.ActionID == "timeselection" {
				log.Printf("********* ActionID: %s", action.ActionID)
				log.Printf("********* Selected Time: %s", action.SelectedTime)
				log.Printf("********* ChannelID: %s", interaction.Container.ChannelID)
				//				log.Printf("********* Selected option: %s", action.TimePickerElement.InitialTime)
			}
			if action.ActionID == "frequency" {
				log.Printf("********* ActionID: %s", action.ActionID)
				log.Printf("********* Selected option: %s", action.SelectedOption.Value)
			}
		}

	default:

	}

	return nil
}
