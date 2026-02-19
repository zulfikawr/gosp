package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"github.com/zulfikawr/gosp/pkg/stealth"
)

type BraveScraper struct {
	client *http.Client
}

func NewBraveScraper() *BraveScraper {
	return &BraveScraper{
		client: stealth.NewStealthClient(15 * time.Second),
	}
}

func (s *BraveScraper) ID() protocol.Engine {
	return protocol.Engine_ENGINE_BRAVE
}

func (s *BraveScraper) Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error) {
	// Strategy: Brave Search via the 'API-like' web endpoint
	searchURL := fmt.Sprintf("https://search.brave.com/search?q=%s&offset=%d&source=web", url.QueryEscape(query), offset)
	
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil { return nil, err }

	req.Header.Set("User-Agent", stealth.GetRandomUserAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")

	resp, err := s.client.Do(req)
	if err != nil { return nil, err }
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("brave returned status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil { return nil, err }

	var results []*protocol.ResultItem

	// Brave's template is very consistent. Let's use more robust selectors.
	doc.Find(".snippet").Each(func(i int, sel *goquery.Selection) {
		if int32(len(results)) >= count { return }

		title := strings.TrimSpace(sel.Find(".snippet-title, .snippet-header").Text())
		link, _ := sel.Find("a").First().Attr("href")
		snippet := strings.TrimSpace(sel.Find(".snippet-description, .snippet-content").Text())

		if title != "" && link != "" && !strings.HasPrefix(link, "/") {
			results = append(results, &protocol.ResultItem{
				Title:       title,
				Url:         link,
				Description: snippet,
			})
		}
	})

	return results, nil
}
