package main

import (
	"benzinpriser/handlers"
	"benzinpriser/middleware"
	"benzinpriser/priser"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}
	r := mux.NewRouter()
	router(r)
	log.Printf("Listening on %v\n", port)
	http.ListenAndServe(port, r)
}

func router(r *mux.Router) {
	r.Use(middleware.LoggingMiddleware)

	handlerCtx := handlers.NewHandlerCtx(&priser.PriceService{})

	r.HandleFunc("/prices/today", handlers.HandlerPricesToday(handlerCtx))
}
