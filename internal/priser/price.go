package priser

import (
	"benzinpriser/internal/cache"
	"context"
	"fmt"
	"strings"
	"time"
)

type FuelType int

const (
	FuelTypeUnleaded95 FuelType = iota + 1
	FuelTypeOktane100
	FuelTypeDiesel
)

type PriceGetter interface {
	GetPrice(ctx context.Context, date time.Time, fuelType FuelType) (*Price, error)
}

type PriceService struct {
	Cache *cache.Cache
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

func (t PriceTime) MarshalJSON() ([]byte, error) {
	layout := "2006-01-02T15:04:05"
	formattedTime := t.Time.Format(layout)
	formattedTimeWithQuotes := fmt.Sprintf("\"%v\"", formattedTime)
	return []byte(formattedTimeWithQuotes), nil
}
