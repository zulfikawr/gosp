package scraper

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/protocol"
)

// BraveScraper implements the Engine interface for Brave Search.
type BraveScraper struct {
	client *http.Client
}

// NewBraveScraper initializes a new Brave scraper.
func NewBraveScraper() *BraveScraper {
	return &BraveScraper{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (s *BraveScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_BRAVE
}

// Search performs a raw HTTP scrape of Brave search results.
func (s *BraveScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	url := fmt.Sprintf("https://search.brave.com/search?q=%s&offset=%d", query, offset)
	
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
		return nil, fmt.Errorf("brave returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []*protocol.ResultItem

	// Brave's search result selector (can change frequently)
	doc.Find("div.snippet").Each(func(i int, sel *goquery.Selection) {
		title := sel.Find("div.title").First().Text()
		link, exists := sel.Find("a.result-header").First().Attr("href")
		snippet := sel.Find("div.snippet-description").First().Text()

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
