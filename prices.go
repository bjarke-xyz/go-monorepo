package main

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type FuelType int8

const (
	FuelTypeUnleaded95 FuelType = iota
	FuelTypeOctane100
	FuelTypeDiesel
)

func (f FuelType) FuelTypeToOkItemNumber() int {
	switch f {
	case FuelTypeOctane100:
		return 533
	case FuelTypeDiesel:
		return 231
	default:
		return 536
	}
}

func (f FuelType) String() string {
	switch f {
	case FuelTypeOctane100:
		return "Octane100"
	case FuelTypeDiesel:
		return "Diesel"
	default:
		return "Unleaded95"
	}
}

func (f FuelType) GetStorageKey() string {
	return "/go/prices/" + f.String() + ".json"
}

type PreviousPriceSlice []PreviousPrice

type PreviousPrice struct {
	DetectionTimestamp time.Time `json:"detectionTimestamp"`
	Price              float32   `json:"price"`
}

func (pp *PreviousPriceSlice) Scan(val interface{}) error {
	switch v := val.(type) {
	case []byte:
		json.Unmarshal(v, &pp)
		return nil
	case string:
		json.Unmarshal([]byte(v), &pp)
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
func (pp PreviousPriceSlice) Value() (driver.Value, error) {
	return json.Marshal(&pp)
}

type Price struct {
	FuelType   FuelType           `json:"fuelType"`
	Timestamp  time.Time          `db:"ts" json:"timestamp"`
	Price      float32            `json:"price"`
	PrevPrices PreviousPriceSlice `db:"prev_prices" json:"prevPrices"`
}

type PriceRepository struct {
	config *Config
}

func NewPriceRepository(config *Config) *PriceRepository {
	return &PriceRepository{
		config: config,
	}
}

func (p *PriceRepository) GetPricesInRange(fuelType FuelType, from time.Time, to time.Time) ([]Price, error) {
	db, err := GetDb(p.config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	prices := []Price{}
	err = db.Select(&prices, "SELECT * FROM fuelprices WHERE fueltype = $1 AND ts between $2 AND $3", fuelType, from, to)
	if err != nil {
		return nil, err
	}
	return prices, nil
}

func (p *PriceRepository) GetPrices(fuelType FuelType) ([]Price, error) {
	db, err := GetDb(p.config)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	prices := []Price{}
	err = db.Select(&prices, "SELECT * FROM fuelprices WHERE fueltype = $1", fuelType)
	if err != nil {
		return nil, err
	}
	return prices, nil
}

func (p *PriceRepository) UpsertPrices(prices []Price) error {
	if len(prices) == 0 {
		return nil
	}
	db, err := GetDb(p.config)
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	// excluded contains the data of the row, where the insert failed
	// So we can use that for the update
	_, err = db.NamedExec(
		"INSERT INTO fuelprices (fueltype, ts, price, prev_prices) "+
			"VALUES (:fueltype, :ts, :price, :prev_prices) "+
			"ON CONFLICT ON CONSTRAINT fuelprices_pkey "+
			"DO UPDATE SET price = excluded.price, prev_prices = excluded.prev_prices", prices)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to do upserts: %w", err)
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil

}
