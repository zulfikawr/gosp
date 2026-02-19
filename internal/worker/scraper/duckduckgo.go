package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/protocol"
)

// DuckDuckGoScraper implements the Engine interface for DuckDuckGo (HTML version).
type DuckDuckGoScraper struct {
	client *http.Client
}

// NewDuckDuckGoScraper initializes a new DuckDuckGo scraper.
func NewDuckDuckGoScraper() *DuckDuckGoScraper {
	return &DuckDuckGoScraper{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *DuckDuckGoScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_DUCKDUCKGO
}

func (s *DuckDuckGoScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	// Using the standard DuckDuckGo HTML endpoint which is less aggressive with bot detection
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))
	
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// High-reputation headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("duckduckgo returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []*protocol.ResultItem

	// DuckDuckGo HTML specific selectors
	doc.Find(".result").Each(func(i int, sel *goquery.Selection) {
		if int32(len(results)) >= count {
			return
		}

		titleNode := sel.Find(".result__a")
		title := titleNode.Text()
		link, exists := titleNode.Attr("href")
		snippet := sel.Find(".result__snippet").Text()

		// Clean DuckDuckGo redirect URLs to get the final destination
		if exists && strings.Contains(link, "uddg=") {
			if u, err := url.Parse(link); err == nil {
				actualURL := u.Query().Get("uddg")
				if actualURL != "" {
					link = actualURL
				}
			}
		}

		// Filter out ads or empty results
		if exists && title != "" && !sel.HasClass("result--ad") {
			results = append(results, &protocol.ResultItem{
				Title:       title,
				Url:         link,
				Description: snippet,
			})
		}
	})

	logger.Debug("duckduckgo scrape complete", "count", len(results), "query", query)
	return results, nil
}
