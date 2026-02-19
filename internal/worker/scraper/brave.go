package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"github.com/zulfikawr/gosp/pkg/stealth"
)

// BraveScraper implements the Engine interface for Brave Search.
type BraveScraper struct {
	client *http.Client
}

// NewBraveScraper initializes a new Brave scraper using a stealth client.
func NewBraveScraper() *BraveScraper {
	return &BraveScraper{
		// Use the stealth client to spoof TLS fingerprinting (JA3)
		client: stealth.NewStealthClient(10 * time.Second),
	}
}

func (s *BraveScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_BRAVE
}

// Search performs a raw HTTP scrape of Brave search results.
func (s *BraveScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	searchURL := fmt.Sprintf("https://search.brave.com/search?q=%s&offset=%d", url.QueryEscape(query), offset)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// 1. Set headers that match the TLS fingerprint (Chrome 120+)
	req.Header.Set("User-Agent", stealth.GetRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Sec-Ch-Ua", "\"Not_A Brand\";v=\"8\", \"Chromium\";v=\"120\", \"Google Chrome\";v=\"120\"")
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("brave returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []*protocol.ResultItem

	// Try a more generic selector for the web results
	doc.Find("div.snippet, div.search-result").Each(func(i int, sel *goquery.Selection) {
		title := sel.Find("div.title, h2, h3").First().Text()
		link, exists := sel.Find("a").First().Attr("href")
		snippet := sel.Find("div.snippet-description, p").First().Text()

		if exists && title != "" {
			results = append(results, &protocol.ResultItem{
				Title:       title,
				Url:         link,
				Description: snippet,
			})
		}
	})

	logger.Debug("brave scrape complete", "count", len(results), "query", query)
	return results, nil
}
