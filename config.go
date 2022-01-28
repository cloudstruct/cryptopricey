package main

import (
	"os"
	"log"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"errors"
)

type DataFile struct {
        Tickers string  `yaml:"tickers"`
        Cron    string  `yaml:"cron"`
        Currency       string  `yaml:"currency"`
}

func readYAML() map[string]*DataFile {
	data := make(map[string]*DataFile)
	dataDir := os.Getenv("DATA_DIR")
	configFile := dataDir + "/conf.yaml"
	log.Printf("********** Loading file: " + configFile)

	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("*********** YAML config does not exist, continuing.")
		} else {
			log.Printf("yamlFile.Get err   #%v ", err)
			panic(err)
		}
	}

	if err = yaml.Unmarshal(yamlFile, &data); err != nil {
		log.Fatal(err)
	}

	return data

}

// handleCryptopriceyConfig will take care of /cryptoprice submissions
func handleCryptopriceyConfig(command slack.SlashCommand, client *slack.Client) error {
	if len(command.Text) < 1 {
		log.Printf("*********** Empty Ticker Symbol")
		return nil
	}

	data := readYAML()

	log.Printf("********** Old Tickers: %+v", data[command.ChannelID].Tickers)
	modalReturn := generateModalRequest()
	log.Printf("%+v", modalReturn)

//	data[command.ChannelID].Tickers = command.Text
//	log.Printf("********** Updated Tickers: %+v", data[command.ChannelID].Tickers)

	dataOut, err := yaml.Marshal(&data)
	if err != nil {
		log.Printf("%+v", dataOut)
		log.Fatal(err)
	}

//	if err = ioutil.WriteFile(configFile, dataOut, 0600); err != nil {
//	     log.Fatal(err)
//	}

//	attachment := slack.Attachment{}
//	attachment.Text = fmt.Sprintf("Ticker list has been updated to [%s].", command.Text)
//	attachment.Color = "#4af030"

	// Send the message to the channel
	// The Channel is available in the command.ChannelID
//	_, _, err3 := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
//	if err3 != nil {
//		return fmt.Errorf("********* failed to post message: %w", err)
//	}
	return nil
}

func generateModalRequest() slack.ModalViewRequest {
	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "CryptoPricey Configuration", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Save", false, false)

	headerText := slack.NewTextBlockObject("mrkdwn", "Please enter the below information to configure CryptoPricey", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	currencyText := slack.NewTextBlockObject("plain_text", "Base Currency", false, false)
	currencyPlaceholder := slack.NewTextBlockObject("plain_text", "USD", false, false)
	currencyElement := slack.NewPlainTextInputBlockElement(currencyPlaceholder, "currency")
	// Notice that blockID is a unique identifier for a block
	currency := slack.NewInputBlock("Currency", currencyText, currencyElement)

	tickersText := slack.NewTextBlockObject("plain_text", "Tickers", false, false)
	tickersPlaceholder := slack.NewTextBlockObject("plain_text", "BTC,ETH,ADA", false, false)
	tickersElement := slack.NewPlainTextInputBlockElement(tickersPlaceholder, "tickers")
	tickers := slack.NewInputBlock("Tickers", tickersText, tickersElement)

	cronText := slack.NewTextBlockObject("plain_text", "Cron Schedule", false, false)
	cronPlaceholder := slack.NewTextBlockObject("plain_text", "* */6 * * *", false, false)
	cronElement := slack.NewPlainTextInputBlockElement(cronPlaceholder, "cron")
	cron := slack.NewInputBlock("Cron", cronText, cronElement)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			currency,
			tickers,
			cron,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	return modalRequest
}
