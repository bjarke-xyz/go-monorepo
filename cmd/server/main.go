package main

import (
	"benzinpriser/internal/cache"
	"benzinpriser/internal/configuration"
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

	redisAddr := configuration.GetSwarmSecret("prod_redis_addr", os.Getenv("REDIS_ADDR"))
	redisUser := configuration.GetSwarmSecret("prod_redis_user", os.Getenv("REDIS_USERNAME"))
	redisPass := configuration.GetSwarmSecret("prod_redis_pass", os.Getenv("REDIS_PASSWORD"))

	cache := cache.NewRedisCache(redisAddr, redisUser, redisPass)
	handlerCtx := handlers.NewHandlerCtx(&priser.PriceService{
		Cache: cache,
	})

	r.HandleFunc("/prices/today", handlers.HandlerPricesToday(handlerCtx))
}
