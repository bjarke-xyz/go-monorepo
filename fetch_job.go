package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type FetchOkDataJob struct {
	appContext *AppContext
}

func NewFetchOkDataJob(appContext *AppContext) *FetchOkDataJob {
	return &FetchOkDataJob{
		appContext: appContext,
	}
}

type okPriceHistoryResponse struct {
	ShowPricesFor1000Liter bool                 `json:"visPriserFor1000Liter"`
	History                []okPriseHistoryItem `json:"historik"`
}

type okTime struct {
	time.Time
}

const okTimeLayout = "2006-01-02T15:04:05"

var nilTime = (time.Time{}).UnixNano()

func (okt *okTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		okt.Time = time.Time{}
		return
	}
	okt.Time, err = time.Parse(okTimeLayout, s)
	return
}
func (okt *okTime) MarshalJSON() ([]byte, error) {
	if okt.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", okt.Time.Format(okTimeLayout))), nil
}
func (okt *okTime) IsSet() bool {
	return okt.UnixNano() != nilTime
}

type okPriseHistoryItem struct {
	Date   okTime  `json:"dato"`
	ItemNo int     `json:"varenr"`
	Price  float32 `json:"pris"`
}

const s3Bucket = "fuelprices"

func (f *FetchOkDataJob) ExecuteFetchJob(fuelType FuelType) {
	start := time.Now()

	err := f.fetchOkJson(fuelType)
	if err != nil {
		log.Printf("failed to fetch ok json: %v", err)
		return
	}
	err = f.processOkJson(fuelType)
	if err != nil {
		log.Printf("failed to process ok json: %v", err)
		return
	}
	duration := time.Since(start)
	log.Printf("OK Fetch Data job completed successfully in %v ms", duration.Milliseconds())
}

func (f *FetchOkDataJob) processOkJson(fuelType FuelType) error {
	jsonBytes, err := f.appContext.Storage.Get(s3Bucket, fuelType.GetStorageKey())
	if err != nil {
		return fmt.Errorf("error getting json for type %v: %w", fuelType, err)
	}
	okPriceResp := &okPriceHistoryResponse{}
	err = json.Unmarshal(jsonBytes, okPriceResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal json for type %v: %w", fuelType, err)
	}

	currentPrices, err := f.appContext.PriceRepository.GetPrices(fuelType)
	if err != nil {
		return fmt.Errorf("error getting current prices: %w", err)
	}
	currentPricesByTime := make(map[int64]Price)
	for _, price := range currentPrices {
		currentPricesByTime[price.Timestamp.Unix()] = price
	}
	prices := make([]Price, 0)
	for _, okPrice := range okPriceResp.History {
		currentPrice, ok := currentPricesByTime[okPrice.Date.Time.Unix()]
		prevPrices := make([]PreviousPrice, 0)
		// includePrice is used to check if we should update/insert this price at all
		includePrice := false
		if ok {
			// We found a currentPrice, so carry its prevPrices along
			for _, prevPrice := range currentPrice.PrevPrices {
				prevPrices = append(prevPrices, prevPrice)
			}
			if currentPrice.Price != okPrice.Price {
				// The price for a already known date has changed, so its an updated price
				includePrice = true
				prevPrices = append(prevPrices, PreviousPrice{
					DetectionTimestamp: time.Now(),
					Price:              currentPrice.Price,
				})
			}
		} else {
			// We did not find a price for this timestamp in the db, so its a new price
			includePrice = true
		}
		if includePrice {
			price := Price{
				FuelType:   fuelType,
				Timestamp:  okPrice.Date.Time,
				Price:      okPrice.Price,
				PrevPrices: prevPrices,
			}
			prices = append(prices, price)
		}
	}

	log.Printf("Fetch job: %v prices", len(prices))

	err = f.appContext.PriceRepository.UpsertPrices(prices)
	if err != nil {
		return fmt.Errorf("failed to upsert prices: %w", err)
	}

	return nil
}

func (f *FetchOkDataJob) fetchOkJson(fuelType FuelType) error {
	url := "https://www.ok.dk/privat/produkter/ok-kort/prisudvikling/getProduktHistorik"
	requestMap := map[string]string{
		"varenr":    strconv.Itoa(fuelType.FuelTypeToOkItemNumber()),
		"pumpepris": "true",
	}
	requestJson, err := json.Marshal(requestMap)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestJson))
	if err != nil {
		return fmt.Errorf("error getting ok prices: %w", err)
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %w", err)
	}

	err = f.appContext.Storage.Put(s3Bucket, fuelType.GetStorageKey(), bodyBytes)
	if err != nil {
		return fmt.Errorf("failed to write to storage bucket: %w", err)
	}

	return nil
}
