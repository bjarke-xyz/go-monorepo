package main

import (
	"log"
	"strings"

	"github.com/bjarke-xyz/rasende2/duda"
	"github.com/bjarke-xyz/rasende2/pkg"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	context := &pkg.AppContext{
		Cache: pkg.NewCache("./cache"),
	}
	dudaScraper := duda.NewScraper(context)
	links, err := dudaScraper.GetMediaUrls()
	if err != nil {
		log.Fatal(err)
	}
	workingLinks, err := dudaScraper.DownloadContents(links)
	if err != nil {
		log.Fatal(err)
	}

	for i, link := range workingLinks {
		content, err := dudaScraper.GetContent(link)
		if err != nil {
			log.Printf("error getting content for %v: %v", link.Url, err)
		}
		if strings.Contains(content, "rss") {
			log.Printf("(%v) %v: %v", i, link.Url, len(content))
		}
	}

}
