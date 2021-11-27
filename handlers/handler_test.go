package handlers_test

import (
	"benzinpriser/handlers"
	"benzinpriser/priser"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandlerPricesToday(t *testing.T) {
	t.Run("Same prices", func(tt *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prices/today", nil)
		res := httptest.NewRecorder()

		handlerCtx := handlers.NewHandlerCtx(&SamePrices{})
		handlers.HandlerPricesToday(handlerCtx)(res, req)

		expected := "the price will not change"
		if !strings.Contains(res.Body.String(), expected) {
			t.Errorf("expected body of %q to contain %q", res.Body.String(), expected)
		}
	})

	t.Run("Cheaper prices", func(tt *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prices/today?now=2021-11-27", nil)
		res := httptest.NewRecorder()

		handlerCtx := handlers.NewHandlerCtx(&CheaperPrices{})
		handlers.HandlerPricesToday(handlerCtx)(res, req)

		expected := "it will be cheaper"
		if !strings.Contains(res.Body.String(), expected) {
			t.Errorf("expected body of %q to contain %q", res.Body.String(), expected)
		}
	})

	t.Run("More expensive prices", func(tt *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prices/today?now=2021-11-27", nil)
		res := httptest.NewRecorder()

		handlerCtx := handlers.NewHandlerCtx(&ExpensivePrices{})
		handlers.HandlerPricesToday(handlerCtx)(res, req)

		expected := "it will be more expensive"
		if !strings.Contains(res.Body.String(), expected) {
			t.Errorf("expected body of %q to contain %q", res.Body.String(), expected)
		}
	})

	t.Run("Date has no price", func(tt *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prices/today?now=2099-11-27", nil)
		res := httptest.NewRecorder()

		handlerCtx := handlers.NewHandlerCtx(&NoPrice{})
		handlers.HandlerPricesToday(handlerCtx)(res, req)

		expected := "No data for date"
		if !strings.Contains(res.Body.String(), expected) {
			t.Errorf("expected body of %q to contain %q", res.Body.String(), expected)
		}
	})

	t.Run("Invalid `now` format", func(tt *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prices/today?now=hest-kanin-365", nil)
		res := httptest.NewRecorder()

		handlerCtx := handlers.NewHandlerCtx(&NoPrice{})
		handlers.HandlerPricesToday(handlerCtx)(res, req)

		expected := "Invalid `now` format"
		if !strings.Contains(res.Body.String(), expected) {
			t.Errorf("expected body of %q to contain %q", res.Body.String(), expected)
		}
	})
}

type SamePrices struct {
	priser.PriceGetter
}

func (p *SamePrices) GetPrice(date time.Time) (*priser.Price, error) {
	return &priser.Price{
		Date: &priser.PriceTime{
			Time: date,
		},
		Price: 6.00,
	}, nil
}

type CheaperPrices struct {
	priser.PriceGetter
}

func (p *CheaperPrices) GetPrice(date time.Time) (*priser.Price, error) {
	if date.Day() == 27 {
		return &priser.Price{
			Date: &priser.PriceTime{
				Time: date,
			},
			Price: 6.00,
		}, nil
	}
	return &priser.Price{
		Date: &priser.PriceTime{
			Time: date,
		},
		Price: 5.79,
	}, nil
}

type ExpensivePrices struct {
	priser.PriceGetter
}

func (p *ExpensivePrices) GetPrice(date time.Time) (*priser.Price, error) {
	if date.Day() == 27 {
		return &priser.Price{
			Date: &priser.PriceTime{
				Time: date,
			},
			Price: 6.00,
		}, nil
	}
	return &priser.Price{
		Date: &priser.PriceTime{
			Time: date,
		},
		Price: 6.79,
	}, nil
}

type NoPrice struct {
	priser.PriceGetter
}

func (p *NoPrice) GetPrice(date time.Time) (*priser.Price, error) {
	return nil, nil
}
