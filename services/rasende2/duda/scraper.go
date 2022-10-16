package duda

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var (
	chromeUserAgent       = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36"
	disabledSitesCacheKey = "disabled-sites"
)

type Link struct {
	Url   string
	Title string
}

type disabledSite struct {
	Reason string
}

type Scraper struct {
	cache *Cache
}

func NewScraper(cache *Cache) *Scraper {
	return &Scraper{
		cache: cache,
	}
}

func (s *Scraper) getDisabledSites() (map[string]disabledSite, error) {
	disabledSites := make(map[string]disabledSite)
	disabledSitesJson, ok := s.cache.Get(disabledSitesCacheKey)
	if ok {
		err := json.Unmarshal([]byte(disabledSitesJson), &disabledSites)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal disabled sites: %w", err)
		}
	}
	return disabledSites, nil
}

func (s *Scraper) saveDisabledSites(disabledSites map[string]disabledSite) error {
	disabledSitesBytes, err := json.Marshal(disabledSites)
	if err != nil {
		return fmt.Errorf("error marshaling disabled sites: %w", err)
	}
	disabledSitesJson := string(disabledSitesBytes)
	s.cache.Put(disabledSitesCacheKey, disabledSitesJson)
	return nil
}

func (s *Scraper) GetContent(link Link) (string, error) {
	cacheKey := "links/" + strings.ReplaceAll(link.Url, "/", "_")
	content, ok := s.cache.Get(cacheKey)
	if ok {
		return content, nil
	}
	req, err := http.NewRequest("GET", link.Url, nil)
	if err != nil {
		log.Printf("error creating request: %v", err)
	}
	req.Header.Set("User-Agent", chromeUserAgent)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error getting %v: %w", link.Url, err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("%v returned non-200: %v", link.Url, resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading body: %w", err)
	}
	bodyStr := string(body)
	s.cache.Put(cacheKey, bodyStr)
	return bodyStr, nil
}

func (s *Scraper) DownloadContents(links []Link) ([]Link, error) {

	disabledSites, err := s.getDisabledSites()
	if err != nil {
		return nil, err
	}

	workingLinks := make([]Link, 0)
	for _, link := range links {
		if _, ok := disabledSites[link.Url]; ok {
			continue
		}
		_, err := s.GetContent(link)
		if err != nil {
			log.Printf("error getting %v: %v\n", link.Url, err)
			disabledSites[link.Url] = disabledSite{Reason: fmt.Sprintf("error getting: %v", err)}
			continue
		}
		workingLinks = append(workingLinks, link)
	}
	err = s.saveDisabledSites(disabledSites)
	if err != nil {
		return nil, err
	}
	return workingLinks, nil
}

func (s *Scraper) GetMediaUrls() ([]Link, error) {

	duda := "https://duda.dk/aviser/"

	cacheKey := "dudahtml"
	cachedHtml, ok := s.cache.Get(cacheKey)
	if !ok {
		log.Println("duda not found in cache, getting")
		client := &http.Client{}
		req, err := http.NewRequest("GET", duda, nil)
		if err != nil {
			return nil, fmt.Errorf("could not create request: %w", err)
		}
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
		req.Header.Set("User-Agent", chromeUserAgent)
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("could not do request: %w", err)
		}
		if res.StatusCode != 200 {
			return nil, fmt.Errorf("request returned %v", res.StatusCode)
		}

		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("could not read body: %w", err)
		}
		cachedHtml = string(body)
		s.cache.Put(cacheKey, cachedHtml)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(cachedHtml))
	if err != nil {
		return nil, fmt.Errorf("could create goquery document: %w", err)
	}

	links := make([]Link, 0)
	doc.Find(".single-page-content.single-content.entry.wpex-clr").Each(func(i int, s *goquery.Selection) {
		s.Find("a").Each(func(j int, a *goquery.Selection) {
			title := strings.TrimSpace(a.Text())
			url, ok := a.Attr("href")
			if ok {
				links = append(links, Link{
					Url:   url,
					Title: title,
				})
			}
		})
	})
	return links, nil

}
