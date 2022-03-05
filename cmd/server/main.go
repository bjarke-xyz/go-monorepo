package main

import (
	"benzinpriser/internal/cache"
	"benzinpriser/internal/handlers"
	"benzinpriser/internal/middleware"
	"benzinpriser/internal/priser"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080"
	}

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	r := mux.NewRouter()
	router(r)
	log.Printf("Listening on %v\n", port)
	http.ListenAndServe(port, r)
}

func router(r *mux.Router) {
	r.Use(middleware.LoggingMiddleware)

	cache := cache.NewRedisCache(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_USERNAME"), os.Getenv("REDIS_PASSWORD"))
	handlerCtx := handlers.NewHandlerCtx(&priser.PriceService{
		Cache: cache,
	})

	r.HandleFunc("/prices/today", handlers.HandlerPricesToday(handlerCtx))
}
