package main

import (
	"ShoppingList-Backend/internal/pkg/migration"
	"ShoppingList-Backend/pkg/config"
	"ShoppingList-Backend/pkg/logger"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	godotenv.Load()
	godotenv.Load("/run/secrets/env")
	cfg := config.New()
	logger.SetLogs(zap.DebugLevel, cfg.LogFormat)
	log := zap.S()

	direction := cfg.Migrate
	if direction != "down" && direction != "up" {
		log.Fatalf("-migrate accepts [up, down] values only")
	}

	migration.DoMigration(direction, cfg.GetDBConnStr())

}
