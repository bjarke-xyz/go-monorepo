package rss

import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"time"

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
	limitStr := c.DefaultQuery("limit", "5")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 5
	}
	searchContentStr := c.DefaultQuery("content", "false")
	searchContent, err := strconv.ParseBool(searchContentStr)
	if err != nil {
		searchContent = false
	}
	results, err := h.service.SearchItems(c.Request.Context(), query, searchContent)
	if err != nil {
		log.Printf("failed to get items with query %v: %v", query, err)
		c.JSON(http.StatusInternalServerError, SearchResult{})
		return
	}
	if len(results) > limit {
		results = results[0:limit]
	}
	c.JSON(http.StatusOK, SearchResult{
		HighlightedWords: []string{query},
		Items:            results,
	})
}

type ChartDataset struct {
	Label string `json:"label"`
	Data  []int  `json:"data"`
}

type ChartResult struct {
	Type     string         `json:"type"`
	Title    string         `json:"title"`
	Labels   []string       `json:"labels"`
	Datasets []ChartDataset `json:"datasets"`
}

type ChartsResult struct {
	Charts []ChartResult `json:"charts"`
}

func MakeLineChart(items []RssItemDto, title string, datasetLabel string) ChartResult {
	dateFormat := "01-02"
	today := time.Now()
	sevenDaysAgo := today.Add(-time.Hour * 24 * 6)
	lastWeekItemsGroupedByDate := make(map[string]int)
	for _, item := range items {
		if item.Published.Before(today) && item.Published.After(sevenDaysAgo) {
			key := item.Published.Format(dateFormat)
			_, ok := lastWeekItemsGroupedByDate[key]
			if !ok {
				lastWeekItemsGroupedByDate[key] = 0
			}
			lastWeekItemsGroupedByDate[key] = lastWeekItemsGroupedByDate[key] + 1
		}
	}
	labels := make([]string, 0)
	data := make([]int, 0)
	for d := sevenDaysAgo; !d.After(today); d = d.AddDate(0, 0, 1) {
		labels = append(labels, d.Format(dateFormat))
		datum, ok := lastWeekItemsGroupedByDate[d.Format(dateFormat)]
		if ok {
			data = append(data, datum)
		} else {
			data = append(data, 0)
		}
	}

	return ChartResult{
		Type:   "line",
		Title:  title,
		Labels: labels,
		Datasets: []ChartDataset{
			{
				Label: datasetLabel,
				Data:  data,
			},
		},
	}
}

func MakeDoughnutChart(items []RssItemDto, title string) ChartResult {
	sitesSet := make(map[string][]RssItemDto)
	for _, item := range items {
		_, ok := sitesSet[item.SiteName]
		if !ok {
			sitesSet[item.SiteName] = make([]RssItemDto, 0)
		}
		sitesSet[item.SiteName] = append(sitesSet[item.SiteName], item)

	}

	labels := make([]string, 0)
	data := make([]int, 0)
	for siteName := range sitesSet {
		labels = append(labels, siteName)
	}
	sort.Strings(labels)
	for _, siteName := range labels {
		siteItems, ok := sitesSet[siteName]
		if ok {
			data = append(data, len(siteItems))
		}
	}

	return ChartResult{
		Type:   "pie",
		Title:  title,
		Labels: labels,
		Datasets: []ChartDataset{
			{
				Label: "",
				Data:  data,
			},
		},
	}
}

func (h *HttpHandlers) HandleCharts(c *gin.Context) {
	query := c.Query("q")
	results, err := h.service.SearchItems(c.Request.Context(), query, false)
	if err != nil {
		log.Printf("failed to get items with query %v: %v", query, err)
		c.JSON(http.StatusInternalServerError, nil)
		return
	}
	lineTitle := "Den seneste uges raserier"
	lineDatasetLabel := "Raseriudbrud"
	doughnutTitle := "Raseri i de forskellige medier"
	if query != "rasende" {
		lineTitle = "Den seneste uges brug af '" + query + "'"
		lineDatasetLabel = "Antal '" + query + "'"
		doughnutTitle = "Brug af '" + query + "' i de forskellige medier"
	}
	c.JSON(http.StatusOK, ChartsResult{
		Charts: []ChartResult{
			MakeLineChart(results, lineTitle, lineDatasetLabel),
			MakeDoughnutChart(results, doughnutTitle),
		},
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
