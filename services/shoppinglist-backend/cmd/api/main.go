package main

import (
	"ShoppingList-Backend/cmd/api/router"
	"ShoppingList-Backend/internal/pkg/migration"
	"ShoppingList-Backend/pkg/application"
	"ShoppingList-Backend/pkg/config"
	"ShoppingList-Backend/pkg/logger"
	"ShoppingList-Backend/pkg/server"
	"log"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

// @title ShoppingList V4 Backend API
// @version 1.0
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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
		zap.S().Fatalf("Application setup error: %v", err)
	}

	if cfg.MigrateOnStartup {
		migration.DoMigration("up", cfg.GetDBConnStr())
	}

	r := mux.NewRouter().StrictSlash(true)

	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: []string{"*"}, // TODO: Maybe consider not allowing all origins. For now it's fine
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		Debug:          false,
	})
	n := negroni.New(corsMiddleware, negroni.NewRecovery(), negroni.NewLogger())

	n.UseHandler(r)

	router.SwaggerRoute(app, r)
	router.PrivateRoutes(app, r)
	router.SocketIoRoutes(app, r)

	go app.SocketIo.Serve()
	server.Start(app.Cfg, n)
}
