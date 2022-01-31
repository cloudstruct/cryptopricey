package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type currencyData struct {
	Id      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	MinSize string `json:"min_size,omitempty"`
}

func getCurrencies() []currencyData {
	var r map[string][]currencyData
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(fmt.Sprintf("https://api.coinbase.com/v2/currencies"))
	if err != nil {
		log.Printf("*********** Client.Get Error")
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("*********** io.ReadAll Error")
		panic(err)
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Printf("*********** json.Unmarshal Error")
		panic(err)
	}

	return r["data"]
}

func validateCurrency(currenciesList []currencyData, currency string) error {
	pass := false
	log.Printf("********* Evaluating Currency '%s'.", currency)
	for _, data := range currenciesList {
		if data.Id == currency {
			pass = true
		}
	}

	if pass == false {
		return errors.New("Currency input is not a valid selection.")
	} else {
	}

	return nil
}
