package handlers

import (
	"benzinpriser/priser"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func HandlerPricesToday(handlerCtx *HandlerCtx) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		urlParams := r.URL.Query()
		nowStr := urlParams.Get("now")
		now := time.Now()
		if nowStr != "" {
			var err error
			now, err = time.Parse("2006-01-02", nowStr)
			if err != nil {
				rw.WriteHeader(http.StatusBadRequest)
				fmt.Fprintf(rw, "Invalid `now` format")
				return
			}
		}
		todayPrice, err := handlerCtx.prices.GetPrice(now)
		if err != nil {
			log.Printf("Could not get price: %v\n", err)
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(rw, "Could not get price")
			return
		}
		if todayPrice == nil {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(rw, "No data for date %v", now)
			return
		}

		tomorrowTime := now.AddDate(0, 0, 1)
		tomorrowPrice, err := handlerCtx.prices.GetPrice(tomorrowTime)
		if err != nil {
			log.Printf("Could not get tomorrow price for date %v: %v\n", tomorrowTime, err)
		}

		response, err := toResponse([]*priser.Price{todayPrice, tomorrowPrice})
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(rw, err.Error())
			return
		}

		lang := urlParams.Get("lang")

		setMessage(response, lang)

		err = json.NewEncoder(rw).Encode(response)
		if err != nil {
			log.Println("Could not marshal json")
			rw.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(rw, "Could not marshal json")
			return
		}
	}
}

func setMessage(response *priceTodayResponse, lang string) {
	if lang == "" {
		lang = "en"
	}

	todayPrice := response.Prices[0]
	switch lang {
	case "da":
		response.Message = fmt.Sprintf("Dagens dato er %v. Blyfri oktan 95 koster %v kroner og %v ører.", getDateString(response.Date, lang), todayPrice.Kroner, todayPrice.Orer)
	case "en":
		response.Message = fmt.Sprintf("Today is %v. The price of Unleaded octane 95 is %v kroner.", getDateString(response.Date, lang), todayPrice.FullPrice)
	}

	if len(response.Prices) > 1 {
		tomorrowPrice := response.Prices[1]
		interposedPhrase := make(map[string]string)
		if tomorrowPrice.FullPrice > todayPrice.FullPrice {
			interposedPhrase["da"] = "bliver det dyrer"
			interposedPhrase["en"] = "it will be more expensive"
		} else if tomorrowPrice.FullPrice < todayPrice.FullPrice {
			interposedPhrase["da"] = "bliver det billigere"
			interposedPhrase["en"] = "it will be cheaper"
		} else {
			interposedPhrase["da"] = "er prisen den samme"
			interposedPhrase["en"] = "the price will not change"
		}

		switch lang {
		case "da":
			response.Message = fmt.Sprintf("%v I morgen %v, %v kroner og %v ører.", response.Message, interposedPhrase[lang], tomorrowPrice.Kroner, tomorrowPrice.Orer)
		case "en":
			response.Message = fmt.Sprintf("%v Tomorrow %v, %v kroner.", response.Message, interposedPhrase[lang], tomorrowPrice.FullPrice)
		}
	}
}

func toResponse(prices []*priser.Price) (*priceTodayResponse, error) {
	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices")
	}
	response := &priceTodayResponse{
		Prices: make([]priceDetails, 0),
		Date:   prices[0].Date.Time,
	}

	for _, price := range prices {
		if price == nil {
			continue
		}
		priceStr := fmt.Sprintf("%f", price.Price)
		priceStrParts := strings.Split(priceStr, ".")
		if len(priceStrParts) != 2 {
			log.Printf("Price str parts was wrong: %v\n", priceStrParts)
			continue
		}

		kroner, kronerErr := strconv.Atoi(priceStrParts[0])
		orer, orerErr := strconv.Atoi(priceStrParts[1][0:2])
		if kronerErr != nil || orerErr != nil {
			log.Printf("Could not convert kroner/orer strings to ints: %v\n", priceStrParts)
			continue
		}
		response.Prices = append(response.Prices, priceDetails{
			FullPrice: price.Price,
			Kroner:    kroner,
			Orer:      orer,
		})
	}
	return response, nil
}

func getDateString(date time.Time, lang string) string {
	switch lang {
	case "da":
		days := map[string]string{
			"Monday":    "Mandag",
			"Tuesday":   "Tirsdag",
			"Wednesday": "Onsdag",
			"Thursday":  "Torsdag",
			"Friday":    "Fredag",
			"Saturday":  "Lørdag",
			"Sunday":    "Søndag",
		}
		months := map[string]string{
			"January":   "Januar",
			"February":  "Februar",
			"March":     "Marts",
			"April":     "April",
			"May":       "Maj",
			"June":      "Juni",
			"July":      "Juli",
			"August":    "August",
			"September": "September",
			"October":   "Oktober",
			"November":  "November",
			"December":  "December",
		}
		day := days[date.Weekday().String()]
		month := months[date.Month().String()]
		str := fmt.Sprintf("%v %v. %v", day, date.Day(), month)
		return str
	default:
		return date.Weekday().String()
	}
}

type priceTodayResponse struct {
	Prices  []priceDetails `json:"prices"`
	Date    time.Time      `json:"date"`
	Message string         `json:"message"`
}

type priceDetails struct {
	FullPrice float32 `json:"fullPrice"`
	Kroner    int     `json:"kroner"`
	Orer      int     `json:"orer"`
}

func NewHandlerCtx(priceGetter priser.PriceGetter) *HandlerCtx {
	return &HandlerCtx{
		prices: priceGetter,
	}
}

type HandlerCtx struct {
	prices priser.PriceGetter
}
