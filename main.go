package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {

	config, err := NewConfig()
	if err != nil {
		log.Printf("failed to load env: %v", err)
	}

	err = Migrate("up", config.GetDbConnectionString())
	if err != nil {
		log.Printf("failed to migrate: %v", err)
	}

	appContext := NewAppContext(config)

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		now := time.Now()
		prices, err := appContext.PriceRepository.GetPricesInRange(0, now.AddDate(0, 0, -1), now.AddDate(0, 0, 1))
		if err != nil {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
			"prices":  prices,
		})
	})
	r.GET("/job", func(ctx *gin.Context) {
		job := NewFetchOkDataJob(appContext)
		go job.ExecuteFetchJob(FuelTypeUnleaded95)
	})
	r.Run()
}
