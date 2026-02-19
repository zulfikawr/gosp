package scraper

import (
	"fmt"
	"net/url"
)

// FormatGoogleURL prepares the Google search query URL.
func FormatGoogleURL(query string, count int32, offset int32) string {
	return fmt.Sprintf("https://www.google.com/search?q=%s&num=%d&start=%d", url.QueryEscape(query), count, offset)
}

// FormatBraveURL prepares the Brave search query URL.
func FormatBraveURL(query string, count int32, offset int32) string {
	return fmt.Sprintf("https://search.brave.com/search?q=%s&offset=%d", url.QueryEscape(query), offset)
}
