package scraper

import (
	"github.com/zulfikawr/go-search/pkg/protocol"
)

// Result represents a single search result from any engine.
type Result struct {
	Title       string
	URL         string
	Description string
}

// Engine defines the interface that all search engine scrapers must implement.
type Engine interface {
	// Search performs the actual scrape and returns a list of protocol-compatible results.
	Search(query string, count int32, offset int32) ([]*protocol.ResultItem, error)
	// ID returns the protocol-specific engine identifier.
	ID() protocol.Engine
}
