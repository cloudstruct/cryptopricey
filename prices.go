package main

import (
	"encoding/json"
	"fmt"
	"github.com/slack-go/slack"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
)

type responseData struct {
	Data Data `json:"data,omitempty"`
}

type Data struct {
	Base     string `json:"base,omitempty"`
	Currency string `json:"currency,omitempty"`
	Amount   string `json:"amount,omitempty"`
}

func getCryptoPrice(ticker string, currency string, httpClient *http.Client, ch chan<- responseData, wg *sync.WaitGroup) {
	var r responseData

	resp, err := httpClient.Get(fmt.Sprintf("https://api.coinbase.com/v2/prices/%s-%s/spot", ticker, currency))
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		panic(err)
	}

	if r.Data.Base == "" {
		r.Data.Base = ticker
		r.Data.Amount = "not_supported"
	}

	ch <- r
	wg.Done()
}

func asyncGetCryptoPrice(tickers string, currency string, httpClient *http.Client) []responseData {
	var responses []responseData
	var wg sync.WaitGroup

	tickerList := strings.Split(tickers, ",")

	// Open up channel for Async HTTP
	ch := make(chan responseData)

	if len(tickerList) > 5 {
		log.Printf("********** Tickerlist '%s' contains more than 5 tickers", tickers)
		return nil
	} else {
		for _, ticker := range tickerList {
			wg.Add(1)
			go getCryptoPrice(ticker, currency, httpClient, ch, &wg)
		}

		// Close the channel in the background
		go func() {
			wg.Wait()
			close(ch)
		}()

		for resp := range ch {
			responses = append(responses, resp)
		}

		return responses
	}
}

// handleCryptopriceyCommand will take care of /cryptoprice submissions
func handleCryptopriceyCommand(command slack.SlashCommand, client *slack.Client, httpClient *http.Client) error {
	var responseTextList []string
	var currency string
	data := readYAML()

	if _, found := data[command.ChannelID]; found {
		currency = data[command.ChannelID].Currency
		if currency == "" {
			currency = "USD"
		}
	} else {
		currency = "USD"
	}

	// The Input is found in the text field so
	// Create the attachment and assigned based on the message
	attachment := slack.Attachment{}
	attachment.Color = "#4af030"

	prices := asyncGetCryptoPrice(command.Text, currency, httpClient)
	if prices == nil {
		responseTextList = append(responseTextList, fmt.Sprintf("Tickerlist '%s' contains more than 5 tickers.", command.Text))
		log.Printf("********** Tickerlist '%s' contains more than 5 tickers", command.Text)

	} else {
		for _, price := range prices {
			if price.Data.Amount == "not_supported" {
				responseTextList = append(responseTextList, fmt.Sprintf("The cryptocurrency pair '%s-%s' is not currently supported.", price.Data.Base, currency))
			} else {
				responseTextList = append(responseTextList, fmt.Sprintf("The spot price of '%s-%s' is '%s'.", price.Data.Base, currency, price.Data.Amount))
			}
		}
	}

	attachment.Text = strings.Join(responseTextList, "\n")

	// Send the message to the channel
	_, _, err := client.PostMessage(command.ChannelID, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("********* failed to post message: %w", err)
	}
	return nil
}
