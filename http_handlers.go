package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type HttpHandler struct {
	appContext *AppContext
}

func NewHttpHandler(appContext *AppContext) *HttpHandler {
	return &HttpHandler{
		appContext: appContext,
	}
}

func (h *HttpHandler) RunJob(jobKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != jobKey {
			c.AbortWithStatus(401)
			return
		}
		go h.appContext.JobManager.RunJob(JobIdentifierOk)
	}
}

func (h *HttpHandler) GetPrices(c *gin.Context) {
	arguments := parseArguments(c)
	prices, err := h.appContext.PriceRepository.GetPricesForDate(arguments.fuelType, arguments.date)
	if err != nil {
		log.Printf("failed to get prices: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": arguments.language.GetErrorText(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": arguments.language.GetText(prices, arguments.fuelType),
		"prices":  prices,
	})
}

func (h *HttpHandler) GetAllPrices(c *gin.Context) {
	from := parseDate(c.Query("from"), time.Now().AddDate(-1, 0, 0).Truncate(24*time.Hour))
	to := parseDate(c.Query("to"), time.Now().Truncate(24*time.Hour))
	fuelType := parseFuelType(c.Query("type"))

	prices, err := h.appContext.PriceRepository.GetPricesBetweenDates(fuelType, from, to)
	if err != nil {
		log.Printf("failed to get all prices: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "could not get prices",
		})
		return
	}
	c.JSON(http.StatusOK, prices)
}

type getPricesArguments struct {
	date     time.Time
	fuelType FuelType
	language Language
	noCache  bool
}

func parseArguments(c *gin.Context) getPricesArguments {
	return getPricesArguments{
		date:     parseDate(c.Query("now"), time.Now().Truncate(24*time.Hour)),
		fuelType: parseFuelType(c.DefaultQuery("type", c.Query("fueltype"))),
		language: parseLanguage(c.DefaultQuery("lang", c.Query("language"))),
		noCache:  parseNoCache(c.DefaultQuery("nocache", c.Query("noCache"))),
	}
}

func parseDate(dateStr string, defaultTime time.Time) time.Time {
	date := defaultTime
	if dateStr != "" {
		layout := "2006-01-02"
		tmpDate, err := time.Parse(layout, dateStr)
		if err != nil {
			log.Printf("failed to parse date: %v", err)
		} else {
			date = tmpDate
		}
	}
	return date
}

func parseLanguage(langStr string) Language {
	switch strings.ToLower(langStr) {
	case "da":
		return LangDa
	case "en":
		return LangEn
	default:
		return LangEn
	}
}

func parseFuelType(fuelTypeStr string) FuelType {
	switch strings.ToLower(fuelTypeStr) {
	case "unleaded95":
		return FuelTypeUnleaded95
	case "octane100":
		return FuelTypeOctane100
	case "diesel":
		return FuelTypeDiesel
	default:
		return FuelTypeUnleaded95
	}
}

func parseNoCache(noCacheStr string) bool {
	boolVal, err := strconv.ParseBool(noCacheStr)
	if err != nil {
		return false
	}
	return boolVal
}
