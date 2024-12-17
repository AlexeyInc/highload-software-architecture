package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	UAH_USD_API_URL  = "https://bank.gov.ua/NBUStatService/v1/statdirectory/exchangenew?json"
	GA4_ENDPOINT     = "https://www.google-analytics.com/mp/collect"
	MEASUREMENT_ID   = "G-BJW0D1KLQV"
	API_SECRET       = "6d0z1kXCTtGkOrvH4cv61w"
	CLIENT_ID        = "alex-proj-client-420"
	GA4_CUSTOM_EVENT = "uah_usd_exchange_rate"
)

type ExchangeRate struct {
	CurrencyCode string  `json:"cc"`
	Rate         float64 `json:"rate"`
}

func main() {
	restyClient := resty.New().SetTimeout(10 * time.Second)

	for {
		fmt.Println("Fetching UAH/USD exchange rate...")
		rate, err := fetchUAHtoUSDRate(restyClient)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			time.Sleep(time.Minute)
			continue
		}
		fmt.Printf("Current UAH/USD rate: %.2f\n", rate)

		fmt.Println("Sending rate to GA4...")
		if err := sendToGA4(restyClient, rate); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		time.Sleep(10 * time.Minute)
	}
}

func fetchUAHtoUSDRate(client *resty.Client) (float64, error) {
	result := []ExchangeRate{}
	resp, err := client.R().
		SetResult(&result).
		Get(UAH_USD_API_URL)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch exchange rate: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		fmt.Printf("Not OK status: %d\n", resp.StatusCode())
	}

	for _, rate := range result {
		if rate.CurrencyCode == "USD" {
			return rate.Rate, nil
		}
	}
	return 0, fmt.Errorf("UAH/USD rate not found")
}

func sendToGA4(client *resty.Client, rate float64) error {
	payload := map[string]interface{}{
		"client_id": CLIENT_ID,
		"events": []map[string]interface{}{
			{
				"name": GA4_CUSTOM_EVENT,
				"params": map[string]interface{}{
					"uah_usd_rate": rate,
				},
			},
		},
	}

	url := fmt.Sprintf("%s?measurement_id=%s&api_secret=%s", GA4_ENDPOINT, MEASUREMENT_ID, API_SECRET)

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(payload).
		Post(url)
	if err != nil {
		return fmt.Errorf("failed to send request to GA4: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("GA4 status: %s", resp.Status())
	}

	fmt.Println("Data sent successfully to GA4")
	return nil
}
