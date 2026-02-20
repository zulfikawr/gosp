package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zulfikawr/gosp/pkg/protocol"
)

// GoogleScraper implements the Engine interface for Google Search.
type GoogleScraper struct {
	client *http.Client
}

// NewGoogleScraper initializes a new Google scraper.
func NewGoogleScraper() *GoogleScraper {
	return &GoogleScraper{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *GoogleScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_GOOGLE
}

// Search performs a scrape of Google results.
// Note: This implementation is optimized for Residential/Home IPs.
// Standard VPS IPs are frequently blocked by Google's automated systems.
func (s *GoogleScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	// Using the mobile search URL which provides a simpler HTML structure for parsing.
	searchURL := fmt.Sprintf("https://www.google.com/search?q=%s&num=%d&start=%d&gbv=1&hl=en", url.QueryEscape(query), count, offset)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	// Use a high-quality Mobile Safari User-Agent
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("google returned status code %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var results []*protocol.ResultItem

	// Parse results using standard mobile selectors
	doc.Find("div.kCrYT").Each(func(i int, sel *goquery.Selection) {
		if int32(len(results)) >= count { return }
		
		titleNode := sel.Find("h3")
		if titleNode.Length() == 0 { return }
		
		title := strings.TrimSpace(titleNode.Text())
		linkNode := sel.Find("a").First()
		link, exists := linkNode.Attr("href")

		// Extract snippet from the next container if possible
		snippet := sel.Next().Find(".BNeawe.s3707d.AP7Wnd").First().Text()
		if snippet == "" {
			snippet = sel.Find(".BNeawe").Last().Text()
		}

		if exists && title != "" && strings.HasPrefix(link, "/url?q=") {
			results = append(results, &protocol.ResultItem{
				Title:       title, 
				Url:         cleanGoogleURL(link), 
				Description: strings.TrimSpace(snippet),
			})
		}
	})

	if len(results) > 0 {
		return results, nil
	}

	return nil, fmt.Errorf("zero results found (likely bot-detection page)")
}

func cleanGoogleURL(link string) string {
	if strings.HasPrefix(link, "/url?q=") {
		u, err := url.Parse("https://www.google.com" + link)
		if err == nil {
			return u.Query().Get("q")
		}
	}
	return link
}
