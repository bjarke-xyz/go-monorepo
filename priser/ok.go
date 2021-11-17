package priser

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func getPrices(date time.Time) (*priceResponse, error) {
	filenameSuffix := date.Format("2006-01-02")

	filename := fmt.Sprintf("pumpepris-%v.json", filenameSuffix)
	tmpDir := os.TempDir()
	filepath := path.Join(tmpDir, filename)
	data, err := os.ReadFile(filepath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("error opening file %v: %w", filepath, err)
		}
		postBody, err := json.Marshal(map[string]interface{}{
			"varenr":    536,
			"pumpepris": true,
		})
		if err != nil {
			return nil, fmt.Errorf("could not marshal postBody: %w", err)
		}
		postBodyBuffer := bytes.NewBuffer(postBody)
		resp, err := http.Post("https://www.ok.dk/privat/produkter/benzinkort/prisudvikling/getProduktHistorik", "application/json", postBodyBuffer)
		if err != nil {
			return nil, fmt.Errorf("error making request to ok.dk: %w", err)
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading body: %w", err)
		}
		err = os.WriteFile(filepath, body, 0644)
		if err != nil {
			return nil, fmt.Errorf("error writing cache file %v: %w", filepath, err)
		}

		data, err = os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("error reading recently written to file %v: %w", filepath, err)
		}
	}

	var prices priceResponse
	if err := json.Unmarshal(data, &prices); err != nil {
		return nil, fmt.Errorf("could not unmarshal json: %w", err)
	}
	return &prices, nil
}

func GetPrice(date time.Time) (*Price, error) {
	prices, err := getPrices(date)
	if err != nil {
		return nil, fmt.Errorf("could not get all prices: %w", err)
	}
	for _, price := range prices.History {
		if price.Date.Year() == date.Year() && price.Date.Month() == date.Month() && price.Date.Day() == date.Day() {
			return &price, nil
		}
	}
	return nil, nil
}

type Price struct {
	Date  *PriceTime `json:"dato"`
	Price float32    `json:"pris"`
}

type PriceTime struct {
	time.Time
}

func (t *PriceTime) UnmarshalJSON(b []byte) error {
	str := string(b)
	str = strings.Trim(str, "\"")
	layout := "2006-01-02T15:04:05"
	parsedTime, err := time.Parse(layout, str)
	if err != nil {
		return err
	}
	*t = PriceTime{parsedTime}
	return nil
}

type priceResponse struct {
	ShowPricesFor1000Liter bool    `json:"visPriserFor1000Liter"`
	History                []Price `json:"historik"`
}
