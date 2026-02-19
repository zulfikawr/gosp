package scraper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/protocol"
)

// GoogleScraper implements the Engine interface for Google Search.
type GoogleScraper struct {
	client *http.Client
}

// NewGoogleScraper initializes a new Google scraper.
func NewGoogleScraper() *GoogleScraper {
	return &GoogleScraper{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *GoogleScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_GOOGLE
}

// Search performs a raw HTTP scrape of Google search results.
func (s *GoogleScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	url := fmt.Sprintf("https://www.google.com/search?q=%s&num=%d&start=%d", query, count, offset)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", GetRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

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
