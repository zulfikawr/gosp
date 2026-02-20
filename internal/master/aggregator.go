// Package master provides the master node functionality for GOSP.
// It handles worker coordination, task dispatching, and HTTP API serving.
package master

import (
	"sort"
	"strings"

	"github.com/zulfikawr/gosp/pkg/protocol"
)

// ResultAggregator handles merging and deduplicating results from multiple workers/engines.
type ResultAggregator struct{}

// NewResultAggregator initializes a new result aggregator.
func NewResultAggregator() *ResultAggregator {
	return &ResultAggregator{}
}

// Aggregate takes responses from multiple workers and merges them into a single Brave-compatible response.
func (a *ResultAggregator) Aggregate(responses ...*protocol.SearchResponse) *protocol.SearchResponse {
	if len(responses) == 0 {
		return &protocol.SearchResponse{
			ErrorCode: protocol.ErrorCode_ERROR_CODE_PROVIDER_DOWN,
		}
	}

	uniqueResults := make(map[string]*protocol.ResultItem)
	var finalResults []*protocol.ResultItem

	// 1. Deduplication by normalized URL
	for _, resp := range responses {
		for _, item := range resp.Results {
			normURL := normalizeURL(item.Url)
			if _, exists := uniqueResults[normURL]; !exists {
				uniqueResults[normURL] = item
				finalResults = append(finalResults, item)
			}
		}
	}

	// 2. Basic Ranking Logic (Heuristic-based)
	// For now, we keep the original order but we could rank based on engine priority or title match.
	sort.Slice(finalResults, func(i, j int) bool {
		// Heuristic: Prefer results with descriptions over those without
		if finalResults[i].Description != "" && finalResults[j].Description == "" {
			return true
		}
		return false
	})

	return &protocol.SearchResponse{
		TaskId:    responses[0].TaskId,
		Results:   finalResults,
		ErrorCode: protocol.ErrorCode_ERROR_CODE_SUCCESS,
	}
}

// normalizeURL strips common noise from URLs to prevent duplicate results (e.g., trailing slashes, tracking params).
func normalizeURL(u string) string {
	u = strings.TrimSpace(u)
	u = strings.ToLower(u)
	u = strings.TrimSuffix(u, "/")

	// Remove common tracking parameters (utm_*, etc.)
	if idx := strings.Index(u, "?"); idx != -1 {
		params := strings.Split(u[idx+1:], "&")
		var cleanParams []string
		for _, p := range params {
			if !strings.HasPrefix(p, "utm_") && !strings.HasPrefix(p, "fbclid") {
				cleanParams = append(cleanParams, p)
			}
		}
		if len(cleanParams) > 0 {
			u = u[:idx+1] + strings.Join(cleanParams, "&")
		} else {
			u = u[:idx]
		}
	}

	return u
}
