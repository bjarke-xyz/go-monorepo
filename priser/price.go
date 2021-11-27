package priser

import (
	"strings"
	"time"
)

type PriceGetter interface {
	GetPrice(date time.Time) (*Price, error)
}

type PriceService struct {
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
