package scraper

import (
	"fmt"
	"math/rand"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:122.0) Gecko/20100101 Firefox/122.0",
}

// GetRandomUserAgent returns a random modern User-Agent string to help avoid detection.
func GetRandomUserAgent() string {
	return userAgents[rand.Intn(len(userAgents))]
}

// FormatGoogleURL prepares the Google search query URL.
func FormatGoogleURL(query string, count int32, offset int32) string {
	return fmt.Sprintf("https://www.google.com/search?q=%s&num=%d&start=%d", query, count, offset)
}

// FormatBraveURL prepares the Brave search query URL.
func FormatBraveURL(query string, count int32, offset int32) string {
	return fmt.Sprintf("https://search.brave.com/search?q=%s&offset=%d", query, offset)
}
