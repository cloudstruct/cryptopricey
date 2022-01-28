package main

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"io"
	"log"
	"net/http"
	"time"
)

type responseData struct {
	Data Data `json:"data,omitempty"`
}

type Data struct {
	Base     string `json:"base,omitempty"`
	Currency string `json:"currency,omitempty"`
	Amount   string `json:"amount,omitempty"`
}

// handleCryptopriceyCommand will take care of /cryptoprice submissions
func handleCryptopriceyCommand(command slack.SlashCommand, client *slack.Client) error {
	data := readYAML()
	log.Printf("********** Currency: %+v", data[command.ChannelID].Currency)

	// The Input is found in the text field so
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}

	price := getCryptoPrice(command.Text, data[command.ChannelID].Currency)

	// Greet the user
	attachment.Text = fmt.Sprintf("The spot price of %s is '%s'\n", command.Text, price)
	attachment.Color = "#4af030"

	// Send the message to the channel
	// The Channel is available in the command.ChannelID
	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("********* failed to post message: %w", err)
	}
	return nil
}

func getCryptoPrice(ticker string, currency string) string {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(fmt.Sprintf("https://api.coinbase.com/v2/prices/%s-%s/spot", ticker, currency))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var r responseData
	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}

	return r.Data.Amount
}
