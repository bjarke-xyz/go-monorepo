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
	now := time.Now()
	price, err := priser.GetPrice(now)
	if err != nil {
		log.Printf("Could not get price: %v\n", err)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Could not get price")
		return
	}

	priceStr := fmt.Sprintf("%f", price.Price)
	priceStrParts := strings.Split(priceStr, ".")
	if len(priceStrParts) != 2 {
		log.Printf("Price str parts was wrong: %v\n", priceStrParts)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Could not get price")
		return
	}

	kroner, kronerErr := strconv.Atoi(priceStrParts[0])
	orer, orerErr := strconv.Atoi(priceStrParts[1][0:2])
	if kronerErr != nil || orerErr != nil {
		log.Printf("Could not convert kroner/orer strings to ints: %v\n", priceStrParts)
		rw.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(rw, "Could not get price")
		return
	}

	response := PriceTodayResponse{
		Price: PriceDetails{
			FullPrice: price.Price,
			Kroner:    kroner,
			Orer:      orer,
		},
		Date: price.Date.Time,
	}
	urlParams := r.URL.Query()
	lang := urlParams.Get("lang")
	if lang == "" {
		lang = "en"
	}
	switch lang {
	case "da":
		response.Message = fmt.Sprintf("Prisen for Blyfri oktan 95 er i dag %v kroner og %v ører", response.Price.Kroner, response.Price.Orer)
	case "en":
		response.Message = fmt.Sprintf("Unleaded 95 costs %v kroner today", response.Price.FullPrice)
	}

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

type PriceTodayResponse struct {
	Price   PriceDetails `json:"price"`
	Date    time.Time    `json:"date"`
	Message string       `json:"message"`
}

type PriceDetails struct {
	FullPrice float32 `json:"fullPrice"`
	Kroner    int     `json:"kroner"`
	Orer      int     `json:"orer"`
}
