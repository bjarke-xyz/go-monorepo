package main

import (
	"ShoppingList-Backend/pkg/application"
	"ShoppingList-Backend/pkg/config"
	"ShoppingList-Backend/pkg/logger"
	"ShoppingList-Backend/pkg/worker"
	"log"
	"sync"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}
	err = godotenv.Load("/run/secrets/env")
	if err != nil {
		log.Printf("Error loading /run/secrets/env file: %v", err)
	}

	cfg := config.New()
	logger.SetLogs(zap.DebugLevel, cfg.LogFormat)

	app, err := application.Get(cfg)
	if err != nil {
		zap.S().Fatalf("Database error: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	pool := worker.NewWorkerPool(app)
	pool.PeriodicallyEnqueue("20 40 7 * * *", worker.JobDemoCleanUp)
	go worker.Start(pool, &wg)

	webuiServer := worker.NewWebUI(app)
	go worker.StartWebUI(webuiServer, &wg)

	wg.Wait()

}
