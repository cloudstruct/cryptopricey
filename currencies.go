package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type currencyData struct {
	Id	string `json:"id,omitempty"`
	Name	string `json:"name,omitempty"`
	MinSize	string `json:"min_size,omitempty"`
}


func getCurrencies() ([]currencyData) {
	var r map[string][]currencyData
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(fmt.Sprintf("https://api.coinbase.com/v2/currencies"))
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

	fmt.Println(r["data"])
	return r["data"]
}