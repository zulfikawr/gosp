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

// GoogleScraper implements the Engine interface for Google Search.
type GoogleScraper struct {
	client *http.Client
}

// NewGoogleScraper initializes a new Google scraper using a stealth client.
func NewGoogleScraper() *GoogleScraper {
	return &GoogleScraper{
		// Use the stealth client to spoof TLS fingerprinting (JA3)
		client: stealth.NewStealthClient(10 * time.Second),
	}
}

func (s *GoogleScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_GOOGLE
}

// Search performs a raw HTTP scrape of Google search results.
func (s *GoogleScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&num=%d&start=%d", url.QueryEscape(query), count, offset)
	
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// 1. Set headers that match the TLS fingerprint (Chrome 120+)
	req.Header.Set("User-Agent", stealth.GetRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
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
		if resp.StatusCode == 429 {
			return nil, fmt.Errorf("google rate-limited (429)")
		}
		return nil, fmt.Errorf("google returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []*protocol.ResultItem

	// Google's search result selector (can change frequently)
	doc.Find("div.g").Each(func(i int, sel *goquery.Selection) {
		title := sel.Find("h3").First().Text()
		link, exists := sel.Find("a").First().Attr("href")
		snippet := sel.Find("div.VwiC3b").First().Text()

		if exists && title != "" {
			results = append(results, &protocol.ResultItem{
				Title:       title,
				Url:         link,
				Description: snippet,
			})
		}
	})

	logger.Debug("google scrape complete", "count", len(results), "query", query)
	return results, nil
}
