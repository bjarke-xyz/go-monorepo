package rss

import (
	"log"
	"net/http"

	"github.com/bjarke-xyz/rasende2/pkg"
	"github.com/gin-gonic/gin"
)

type HttpHandlers struct {
	context *pkg.AppContext
	service *RssService
}

func NewHttpHandlers(context *pkg.AppContext, service *RssService) *HttpHandlers {
	return &HttpHandlers{
		context: context,
		service: service,
	}
}

type SearchResult struct {
	HighlightedWords []string     `json:"highlightedWords"`
	Items            []RssItemDto `json:"items"`
}

func (h *HttpHandlers) HandleSearch(c *gin.Context) {
	query := c.Query("q")
	results, err := h.service.SearchItems(query)
	if err != nil {
		log.Printf("failed to get items with query %v: %v", query, err)
		c.JSON(http.StatusInternalServerError, SearchResult{})
		return
	}
	c.JSON(http.StatusOK, SearchResult{
		HighlightedWords: []string{query},
		Items:            results,
	})
}

func (h *HttpHandlers) RunJob(key string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != key {
			c.AbortWithStatus(401)
			return
		}
		h.context.JobManager.RunJob(JobIdentifierIngestion)
		c.Status(http.StatusOK)
	}
}
