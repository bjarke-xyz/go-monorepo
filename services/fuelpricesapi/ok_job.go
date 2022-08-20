package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const JobIdentifierOk = "OK_DATA_JOB"

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
	tmpTime, err := time.Parse(okTimeLayout, s)
	if err != nil {
		return err
	}
	okt.Time = tmpTime
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

type OkJobOptions struct {
	FetchFromSource bool
}

func (f *FetchOkDataJob) FetchAndStoreOKPrices(fuelType FuelType, jobOptions OkJobOptions) error {
	var jsonBytes []byte
	var err error
	if jobOptions.FetchFromSource {
		jsonBytes, err = f.fetchOkJsonFromSource(fuelType)
		if err != nil {
			return fmt.Errorf("failed to fetch ok json from source: %v", err)
		}
	} else {
		jsonBytes, err = f.fetchOkJsonFromS3(fuelType)
		if err != nil {
			if errors.Is(err, ErrNoSuchKey) {
				jsonBytes, err = f.fetchOkJsonFromSource(fuelType)
				if err != nil {
					return fmt.Errorf("attempted to fetch from source because it was not found in S3, but it failed: %v", err)
				}
			} else {
				return fmt.Errorf("failed to fetch ok json from s3: %v", err)
			}
		}
	}

	if !jobOptions.FetchFromSource {
		err = f.storeOkJson(jsonBytes, fuelType)
		if err != nil {
			return fmt.Errorf("failed to store ok json to s3: %v", err)
		}
	}

	prices, err := f.processOkJson(fuelType, jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to process ok json: %v", err)
	}

	err = f.storeProcessedOkPrices(fuelType, prices)
	if err != nil {
		return fmt.Errorf("failed to store processed ok prices: %v", err)
	}

	return nil
}

func (f *FetchOkDataJob) ExecuteFetchJob(jobOptions OkJobOptions) error {
	fuelTypes := []FuelType{FuelTypeUnleaded95, FuelTypeOctane100, FuelTypeDiesel}
	for _, fuelType := range fuelTypes {
		err := f.FetchAndStoreOKPrices(fuelType, jobOptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *FetchOkDataJob) processOkJson(fuelType FuelType, jsonBytes []byte) ([]Price, error) {
	okPriceResp := &okPriceHistoryResponse{}
	err := json.Unmarshal(jsonBytes, okPriceResp)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json for type %v: %w", fuelType, err)
	}

	currentPrices, err := f.appContext.PriceRepository.GetPrices(fuelType)
	if err != nil {
		return nil, fmt.Errorf("error getting current prices: %w", err)
	}
	currentPricesByTime := make(map[int64]Price)
	for _, price := range currentPrices {
		currentPricesByTime[price.Date.Unix()] = price
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
					DetectionTimestamp: time.Now().UTC(),
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
				Date:       okPrice.Date.Time,
				Price:      okPrice.Price,
				PrevPrices: prevPrices,
			}
			prices = append(prices, price)
		}
	}

	return prices, nil
}

func (f *FetchOkDataJob) storeProcessedOkPrices(fuelType FuelType, prices []Price) error {
	log.Printf("OK data job: Found %v new prices for %v", len(prices), fuelType.String())

	err := f.appContext.PriceRepository.UpsertPrices(prices)
	if err != nil {
		return fmt.Errorf("failed to upsert prices: %w", err)
	}
	return nil
}

func (f *FetchOkDataJob) fetchOkJsonFromSource(fuelType FuelType) ([]byte, error) {
	url := "https://www.ok.dk/privat/produkter/ok-kort/prisudvikling/getProduktHistorik"
	requestMap := map[string]string{
		"varenr":    strconv.Itoa(fuelType.FuelTypeToOkItemNumber()),
		"pumpepris": "true",
	}
	requestJson, err := json.Marshal(requestMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	response, err := http.Post(url, "application/json", bytes.NewBuffer(requestJson))
	if err != nil {
		return nil, fmt.Errorf("error getting ok prices: %w", err)
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return bodyBytes, nil
}

func (f *FetchOkDataJob) storeOkJson(okJson []byte, fuelType FuelType) error {
	err := f.appContext.Storage.Put(s3Bucket, fuelType.GetStorageKey(), okJson)
	if err != nil {
		return fmt.Errorf("failed to write to storage bucket: %w", err)
	}
	return nil
}

func (f *FetchOkDataJob) fetchOkJsonFromS3(fuelType FuelType) ([]byte, error) {
	bytes, err := f.appContext.Storage.Get(s3Bucket, fuelType.GetStorageKey())
	if err != nil {
		return nil, fmt.Errorf("failed to get ok json: %w", err)
	}
	return bytes, nil
}
