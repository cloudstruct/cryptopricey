package main

import (
	"errors"
	"fmt"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"log"
	"net/http"
	"strings"
)

func emptyCron(cronObject *cron.Cron) error {
	var preCronSize int
	var postCronSize int

	preCronSize = len(cronObject.Entries())
	for _, entry := range cronObject.Entries() {
		cronObject.Remove(entry.ID)
		postCronSize = len(cronObject.Entries())
		if postCronSize >= preCronSize {
			log.Printf("******** ERROR: Could not clear entryID '%+v' for cronObject: %+v", entry.ID, cronObject)
			return errors.New("ERROR: Could not clear entryID from cronObject.")
		}
	}
	return nil
}

func rebuildCron(cronObject *cron.Cron, client *slack.Client, httpClient *http.Client) (*cron.Cron, error) {
	err := emptyCron(cronObject)
	if err != nil {
		log.Printf("Error calling emptyCron on cronObject: %+v", cronObject)
		return cronObject, err
	}

	data := readYAML()
	for channel_id, v := range data {
		channelConfig := v
		channelId := channel_id
		if channelConfig.Cron != "" {
			if channelConfig.Tickers == "" {
				channelConfig.Tickers = "BTC"
			}
			if channelConfig.Currency == "" {
				channelConfig.Currency = "USD"
			}
			_, err = cronObject.AddFunc(channelConfig.Cron, func() {
				err := announceCron(channelId, channelConfig.Tickers, channelConfig.Currency, client, httpClient)
				if err != nil {
					panic(err)
				}
			})
			if err != nil {
				log.Printf("********** ERROR: cronObject.AddFunc failed: '%+v'", err)
				return cronObject, err
			} else {
				log.Printf("********* Added cronObject '%s', '%s/%s' on channel ID '%s'.", channelConfig.Cron, channelConfig.Tickers, channelConfig.Currency, channelId)
				log.Printf("********* %+v", cronObject.Entries())
			}
		}
	}

	return cronObject, nil

}

func announceCron(channelid string, tickers string, currency string, client *slack.Client, httpClient *http.Client) error {
	var responseTextList []string

	prices := asyncGetCryptoPrice(tickers, currency, httpClient)
	if prices == nil {
		responseTextList = append(responseTextList, fmt.Sprintf("Tickerlist '%s' contains more than 5 tickers.", tickers))
		log.Printf("********** Tickerlist '%s' contains more than 5 tickers", tickers)
	} else {
		for _, price := range prices {
			responseTextList = append(responseTextList, fmt.Sprintf("The spot price of '%s-%s' is '%s'.", price.Data.Base, currency, price.Data.Amount))
		}
	}

	attachment := slack.Attachment{}
	attachment.Color = "#4af030"
	attachment.Text = strings.Join(responseTextList, "\n")

	// Send the message to the channel
	_, _, err := client.PostMessage(channelid, slack.MsgOptionAttachments(attachment))
	if err != nil {
		return fmt.Errorf("********* failed to post message: %w", err)
	}

	return nil
}
