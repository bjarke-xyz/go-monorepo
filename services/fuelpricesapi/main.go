package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	config, err := NewConfig()
	if err != nil {
		log.Printf("failed to load env: %v", err)
	}

	err = Migrate("up", config.ConnectionString())
	if err != nil {
		log.Printf("failed to migrate: %v", err)
	}

	appContext := NewAppContext(config)

	defer appContext.JobManager.Stop()
	appContext.JobManager.Cron("*/1 * * * *", JobIdentifierOk, func() error {
		job := NewFetchOkDataJob(appContext)
		return job.ExecuteFetchJob(OkJobOptions{
			FetchFromSource: false,
		})
	}, config.AppEnv == AppEnvProduction)
	go appContext.JobManager.Start()

	httpHandler := NewHttpHandler(appContext)
	if config.AppEnv == AppEnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	r.GET("/prices", httpHandler.GetPrices)
	r.GET("/prices/all", httpHandler.GetAllPrices)
	r.POST("/job", httpHandler.RunJob(config.JobKey))
	r.Run()
}
