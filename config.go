package main

import (
	"errors"
	"fmt"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"os"
)

type DataFile struct {
	Tickers  string `yaml:"tickers"`
	Cron     string `yaml:"cron"`
	Currency string `yaml:"currency"`
}

func readYAML() map[string]*DataFile {
	data := make(map[string]*DataFile)
	configFile := os.Getenv("DATA_DIR") + "/conf.yaml"

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

func writeYAML(data map[string]*DataFile) error {
	dataOut, err := yaml.Marshal(&data)
	configFile := os.Getenv("DATA_DIR") + "/conf.yaml"

	if err != nil {
		log.Printf("%+v", dataOut)
		log.Fatal(err)
	}

	if err = ioutil.WriteFile(configFile, dataOut, 0600); err != nil {
		log.Fatal(err)
	}

	return nil
}

// handleCryptopriceyConfig will take care of /cryptoprice-config submissions
func handleCryptopriceyConfig(command slack.SlashCommand, client *slack.Client) error {
	data := readYAML()
	modalRequest := generateModalRequest(command, data, command.ChannelID)

	_, err := client.OpenView(command.TriggerID, modalRequest)
	if err != nil {
		fmt.Printf("Error opening view: %s", err)
	}

	return nil
}

func generateModalRequest(command slack.SlashCommand, data map[string]*DataFile, channelid string) slack.ModalViewRequest {
	currencyPlaceholderText := "USD"
	currencyOptional := false
	tickersPlaceholderText := "BTC,ETH,ADA"
	tickersOptional := false
	cronPlaceholderText := "* * */6 * *"
	cronOptional := false

	if _, ok := data[command.ChannelID]; ok {
		if data[command.ChannelID].Currency != "" {
			currencyPlaceholderText = data[command.ChannelID].Currency
			currencyOptional = true
		}

		if data[command.ChannelID].Tickers != "" {
			tickersPlaceholderText = data[command.ChannelID].Tickers
			tickersOptional = true
		}

		if data[command.ChannelID].Cron != "" {
			cronPlaceholderText = data[command.ChannelID].Cron
			cronOptional = true
		}
	}

	// Create a ModalViewRequest with a header and two inputs
	titleText := slack.NewTextBlockObject("plain_text", "CryptoPricey Config", false, false)
	closeText := slack.NewTextBlockObject("plain_text", "Close", false, false)
	submitText := slack.NewTextBlockObject("plain_text", "Submit", false, false)

	// Create a ModalViewRequest with a header and two inputs

	headerText := slack.NewTextBlockObject("mrkdwn", "Configuration fields are optional once intially set.\nPlaceholder text is set to current values in that case.", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	currencyText := slack.NewTextBlockObject("plain_text", "Base Currency", false, false)
	currencyPlaceholder := slack.NewTextBlockObject("plain_text", currencyPlaceholderText, false, false)
	currencyElement := slack.NewPlainTextInputBlockElement(currencyPlaceholder, "currency")
	// Notice that blockID is a unique identifier for a block
	currency := slack.NewInputBlock("Currency", currencyText, currencyElement)
	currency.Optional = currencyOptional

	tickersText := slack.NewTextBlockObject("plain_text", "Tickers", false, false)
	tickersPlaceholder := slack.NewTextBlockObject("plain_text", tickersPlaceholderText, false, false)
	tickersElement := slack.NewPlainTextInputBlockElement(tickersPlaceholder, "tickers")
	tickers := slack.NewInputBlock("Tickers", tickersText, tickersElement)
	tickers.Optional = tickersOptional

	cronText := slack.NewTextBlockObject("plain_text", "Cron Schedule (UTC)", false, false)
	cronPlaceholder := slack.NewTextBlockObject("plain_text", cronPlaceholderText, false, false)
	cronElement := slack.NewPlainTextInputBlockElement(cronPlaceholder, "cron")
	cron := slack.NewInputBlock("Cron", cronText, cronElement)
	cron.Optional = cronOptional

	// Remove config section
	removeBtnTxt := slack.NewTextBlockObject("plain_text", "DELETE", false, false)
	removeBtn := slack.NewButtonBlockElement("delete", "delete", removeBtnTxt)
	removeBtn.Style = "danger"
	removeAccessory := slack.NewAccessory(removeBtn)
	removeText := slack.NewTextBlockObject("mrkdwn", "Remove the existing config for this channel.\nCaution: This will remove your channel config permanently!", false, false)
	removeSection := slack.NewSectionBlock(removeText, nil, removeAccessory)

	blocks := slack.Blocks{
		BlockSet: []slack.Block{
			headerSection,
			currency,
			tickers,
			cron,
			removeSection,
		},
	}

	var modalRequest slack.ModalViewRequest
	modalRequest.Type = slack.ViewType("modal")
	modalRequest.Title = titleText
	modalRequest.Close = closeText
	modalRequest.Submit = submitText
	modalRequest.Blocks = blocks
	modalRequest.PrivateMetadata = command.ChannelID
	return modalRequest
}
