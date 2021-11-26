package main

import (
	"benzinpriser/priser"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/prices/today", PricesTodayHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	log.Printf("Listening on %v\n", port)
	http.ListenAndServe(port, r)
}

func PricesTodayHandler(rw http.ResponseWriter, r *http.Request) {
	log.Printf("GET /prices/today\n")
	now := time.Now()
	todayPrice, err := priser.GetPrice(now)
	if err != nil {
		log.Printf("Could not get price: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Could not get price")
		return
	}

	tomorrowTime := now.AddDate(0, 0, 1)
	tomorrowPrice, err := priser.GetPrice(tomorrowTime)
	if err != nil {
		log.Printf("Could not get tomorrow price for date %v: %v\n", tomorrowTime, err)
	}

	response, err := ToResponse([]*priser.Price{todayPrice, tomorrowPrice})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(rw, err.Error())
		return
	}

	urlParams := r.URL.Query()
	lang := urlParams.Get("lang")

	SetMessage(response, lang)

	payload, err := json.Marshal(response)
	if err != nil {
		log.Println("Could not marshal json")
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Could not marshal json")
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(payload)
}

func SetMessage(response *PriceTodayResponse, lang string) {
	if lang == "" {
		lang = "en"
	}

	todayPrice := response.Prices[0]
	switch lang {
	case "da":
		response.Message = fmt.Sprintf("Dagens dato er %v. Blyfri oktan 95 koster %v kroner og %v ører.", GetDateString(response.Date, lang), todayPrice.Kroner, todayPrice.Orer)
	case "en":
		response.Message = fmt.Sprintf("Today is %v. The price of Unleaded octane 95 is %v kroner.", GetDateString(response.Date, lang), todayPrice.FullPrice)
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

func ToResponse(prices []*priser.Price) (*PriceTodayResponse, error) {
	if len(prices) == 0 {
		return nil, fmt.Errorf("no prices")
	}
	response := &PriceTodayResponse{
		Prices: make([]PriceDetails, 0),
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
		response.Prices = append(response.Prices, PriceDetails{
			FullPrice: price.Price,
			Kroner:    kroner,
			Orer:      orer,
		})
	}
	return response, nil
}

func GetDateString(date time.Time, lang string) string {
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

type PriceTodayResponse struct {
	Prices  []PriceDetails `json:"prices"`
	Date    time.Time      `json:"date"`
	Message string         `json:"message"`
}

type PriceDetails struct {
	FullPrice float32 `json:"fullPrice"`
	Kroner    int     `json:"kroner"`
	Orer      int     `json:"orer"`
}
