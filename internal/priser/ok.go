package priser

import (
	"benzinpriser/internal/cache"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func fuelTypeToOkVarenr(fuelType FuelType) int {
	switch fuelType {
	case FuelTypeUnleaded95:
		return 536
	case FuelTypeOktane100:
		return 533
	case FuelTypeDiesel:
		return 231
	default:
		return 536
	}
}

func fetchOkPrices(varenr int) (*priceResponse, error) {

	postData := map[string]interface{}{
		"varenr":    varenr,
		"pumpepris": true,
	}

	postBody, err := json.Marshal(postData)
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

	prices := &priceResponse{}
	if err := json.Unmarshal(body, prices); err != nil {
		return nil, fmt.Errorf("could not unmarshal json: %w", err)
	}
	return prices, nil
}

func (p *PriceService) getPrices(ctx context.Context, fuelType FuelType, allowCached bool) (*priceResponse, error) {

	okVareNr := fuelTypeToOkVarenr(fuelType)
	key := fmt.Sprintf("pumpepris:ok:%v", okVareNr)
	prices := &priceResponse{}
	cacheDuration := time.Hour * 25
	if allowCached {
		err := p.Cache.Get(ctx, key, &prices)
		if err != nil {
			log.Printf("Error getting prices from cache using key '%v': %v", key, err)
		}
		if err != nil && errors.Is(err, cache.ErrNil) {
			prices, err := fetchOkPrices(okVareNr)
			if err != nil {
				return nil, fmt.Errorf("could not fetch ok prices: %w", err)
			}
			p.Cache.Set(ctx, key, &prices, cacheDuration)
		}
	} else {
		uncachedPrices, err := fetchOkPrices(okVareNr)
		if err != nil {
			return nil, fmt.Errorf("could not fetch ok prices: %w", err)
		}
		p.Cache.Set(ctx, key, uncachedPrices, cacheDuration)
		prices = uncachedPrices
	}
	return prices, nil

}

func findPrice(prices *priceResponse, date time.Time) *Price {
	for _, price := range prices.History {
		if price.Date.Year() == date.Year() && price.Date.Month() == date.Month() && price.Date.Day() == date.Day() {
			return &price
		}
	}

	return nil
}

func (p *PriceService) GetPrice(ctx context.Context, date time.Time, fuelType FuelType) (*Price, error) {
	prices, err := p.getPrices(ctx, fuelType, true)
	if err != nil {
		return nil, fmt.Errorf("could not get all prices: %w", err)
	}
	price := findPrice(prices, date)
	if price != nil {
		return price, nil
	} else {
		prices, err := p.getPrices(ctx, fuelType, false)
		if err != nil {
			return nil, fmt.Errorf("could not get all prices: %w", err)
		}
		price := findPrice(prices, date)
		return price, nil
	}
}

type priceResponse struct {
	ShowPricesFor1000Liter bool    `json:"visPriserFor1000Liter"`
	History                []Price `json:"historik"`
}
