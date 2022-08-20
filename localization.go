package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Language string

const (
	LangDa Language = "da"
	LangEn Language = "en"
)

func (l Language) GetErrorText() string {
	switch l {
	case LangDa:
		return "Der blev ikke fundet priser for den dato"
	default:
		return "No prices were found for that date"
	}
}

func (l Language) GetText(prices *DayPrices, fuelType FuelType) string {
	switch l {
	case LangDa:
		return getTextDanish(prices, fuelType)
	default:
		return getTextEnglish(prices, fuelType)
	}
}

func getTextEnglish(prices *DayPrices, fuelType FuelType) string {
	lang := LangEn
	text := fmt.Sprintf("Today, the price of %v is %.2f kroner.", lang.fuelTypeString(fuelType), prices.Today.Price)
	if prices.Yesterday != nil && prices.Yesterday.Price > 0 {
		diffText := lang.getDiffText(prices.Today, prices.Yesterday)
		text = fmt.Sprintf("%v Yesterday the price was %v: %.2f kroner.", text, diffText, prices.Yesterday.Price)
	}
	if prices.Tomorrow != nil && prices.Tomorrow.Price > 0 {
		diffText := lang.getDiffText(prices.Today, prices.Tomorrow)
		text = fmt.Sprintf("%v Tomorrow the price will be %v: %.2f kroner.", text, diffText, prices.Tomorrow.Price)
	}
	return text
}

func getTextDanish(prices *DayPrices, fuelType FuelType) string {
	lang := LangDa
	kroner, orer, err := priceToKronerAndOrer(prices.Today)
	if err != nil {
		log.Printf("failed to convert today price to kroner and orer: %v", err)
		return lang.GetErrorText()
	}

	text := fmt.Sprintf("%v koster %v kroner og %v ører i dag.", lang.fuelTypeString(fuelType), kroner, orer)
	if prices.Yesterday.Price > 0 {
		kroner, orer, err = priceToKronerAndOrer(prices.Yesterday)
		if err != nil {
			log.Printf("failed to convert yesterday price to kroner and orer: %v", err)
			return lang.GetErrorText()
		}
		diffText := lang.getDiffText(prices.Today, prices.Yesterday)
		text = fmt.Sprintf("%v I går var prisen %v: %v kroner og %v ører.", text, diffText, kroner, orer)
	}
	if prices.Tomorrow.Price > 0 {
		kroner, orer, err = priceToKronerAndOrer(prices.Tomorrow)
		if err != nil {
			log.Printf("failed to convert tomorrow price to kroner and orer: %v", err)
			return lang.GetErrorText()
		}
		diffText := lang.getDiffText(prices.Today, prices.Tomorrow)
		text = fmt.Sprintf("%v I morgen vil prisen være %v: %v kroner og %v ører.", text, diffText, kroner, orer)
	}
	return text
}

func (l Language) getDiffText(today *Price, otherDay *Price) string {
	switch l {
	case LangDa:
		if otherDay.Price > today.Price {
			return "højere"
		} else if otherDay.Price < today.Price {
			return "lavere"
		} else {
			return "den samme"
		}
	default:
		if otherDay.Price > today.Price {
			return "higher"
		} else if otherDay.Price < today.Price {
			return "lower"
		} else {
			return "the same"
		}
	}
}

func priceToKronerAndOrer(price *Price) (kroner string, orer string, err error) {
	parts := strings.Split(fmt.Sprintf("%f", price.Price), ".")
	if len(parts) != 2 {
		return "0", "0", fmt.Errorf("failed to parse price")
	}
	kronerInt, err := strconv.Atoi(parts[0])
	if err != nil {
		return "0", "0", fmt.Errorf("failed to parse kroner: %w", err)
	}
	kroner = strconv.Itoa(kronerInt)
	orer = strings.TrimPrefix(parts[1][0:2], "0")
	return kroner, orer, nil
}

func (l Language) fuelTypeString(fuelType FuelType) string {
	switch l {
	case LangDa:
		switch fuelType {
		case FuelTypeOctane100:
			return "Oktan 100"
		case FuelTypeDiesel:
			return "Diesel"
		default:
			return "Blyfri oktan 95"
		}
	default:
		switch fuelType {
		case FuelTypeOctane100:
			return "Oktan 100"
		case FuelTypeDiesel:
			return "Diesel"
		default:
			return "Unleaded octane 95"
		}
	}
}
