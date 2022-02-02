package main

import (
	"fmt"
	"github.com/robfig/cron/v3"
	"log"
	"errors"
//	"time"
)

func emptyCron(cronObject *cron.Cron) error {
	for _, entry := range cronObject.Entries() {
		err := cronObject.Remove(entry.ID)
		if err != nil {
			log.Printf("******** ERROR: Could not clear entryID '%s' for cronObject: %+v", entry.ID, cronObject)
			return errors.New("ERROR: Could not clear entryID from cronObject.")
		}
	}
	return nil
}

func rebuildCron(cronObject *cron.Cron, command *slack.SlashCommand, httpClient *http.Client)) *cron.Cron {
	err := emptyCron(cronObject)
	if err != nil {
		log.Printf("Error calling emptyCron on cronObject: %+v", cronObject)
	}

	data := readYAML()
	for k, v := range data {
		fmt.Printf("%+v", v)
		if v.Cron != "" {
			if v.Tickers == "" {
				v.Tickers = "BTC"
			}
			if v.Currency == "" {
				v.Currency = "USD"
			}
			cronObject.addFunc(v.Cron, asyncGetCryptoPrice(v.Tickers, v.Currency, httpClient) 
			log.Printf("********* Added cronObject entry for channel ID '%s'.", k)
		}
	}

	return cronObject

}

func main() {
	myCron := cron.New()
	myCron.AddFunc("*/1 * * * *", func() { fmt.Println("hi dummy2") })
	fmt.Printf("%+v", myCron.Entries())
	myCron.Start()
	emptyCron(myCron)
	fmt.Printf("%+v", myCron.Entries())
//	time.Sleep(80 * time.Second)
}
