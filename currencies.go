package main

import (
	"encoding/json"
	"errors"
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

	resp, err := client.Get("https://api.coinbase.com/v2/currencies")
	if err != nil {
		log.Println("*********** Client.Get Error")
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("*********** io.ReadAll Error")
		panic(err)
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println("*********** json.Unmarshal Error")
		panic(err)
	}

	return r["data"]
}

func validateCurrency(currenciesList []currencyData, currency string) error {
	pass := false
	for _, data := range currenciesList {
		if data.Id == currency {
			pass = true
		}
	}

	if !pass {
		return errors.New("Currency input is not a valid selection.")
	}

	return nil
}
