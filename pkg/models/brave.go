// Package models defines data structures for GOSP API responses.
// It includes Brave Search API-compatible schemas and OSP-specific extensions.
package models

// SearchResponse matches the Brave Search API result schema.
type SearchResponse struct {
	Type  string        `json:"type"`
	Query QueryMetadata `json:"query"`
	Web   WebResults    `json:"web"`
	News  *NewsResults  `json:"news,omitempty"`
	Video *VideoResults `json:"videos,omitempty"`
	Image *ImageResults `json:"images,omitempty"`
	Meta  ResponseMeta  `json:"meta"`

	// OSP Extended Metadata (Opt-in via ?metadata=true)
	Performance *OSPPerformance `json:"osp_performance,omitempty"`
	Cluster     *OSPCluster     `json:"osp_cluster,omitempty"`
	Diagnostics *OSPDiagnostics `json:"osp_diagnostics,omitempty"`
}

// QueryMetadata contains information about the search query.
type QueryMetadata struct {
	Original  string `json:"original"`
	Altered   string `json:"altered,omitempty"`
	Canonical string `json:"canonical,omitempty"`
}

// WebResults contains a collection of web search results.
type WebResults struct {
	Type    string      `json:"type"`
	Results []WebResult `json:"results"`
}

// WebResult represents a single web search result.
type WebResult struct {
	Title       string       `json:"title"`
	URL         string       `json:"url"`
	Description string       `json:"description"`
	Age         string       `json:"age,omitempty"`
	Language    string       `json:"language,omitempty"`
	Thumbnail   *Thumbnail   `json:"thumbnail,omitempty"`
	Profile     *ProfileMeta `json:"profile,omitempty"`

	// OSP Specific Result Signals
	Signals *OSPSignals `json:"osp_signals,omitempty"`
}

// OSPSignals contains OSP-specific metadata for a search result.
type OSPSignals struct {
	Source    string  `json:"source"`
	WorkerID  string  `json:"worker_id"`
	Region    string  `json:"region"`
	Relevance float32 `json:"relevance,omitempty"`
}

// OSPPerformance contains timing information for the search operation.
type OSPPerformance struct {
	WorkerScrapeMs uint32 `json:"worker_scrape_ms"`
	MasterAggMs    uint32 `json:"master_agg_ms"`
	TotalLatencyMs uint32 `json:"total_latency_ms"`
}

// OSPCluster contains information about the cluster that processed the request.
type OSPCluster struct {
	NodesQueried    int    `json:"nodes_queried"`
	CacheStatus     string `json:"cache_status"`
	ProtocolVersion string `json:"protocol_version"`
}

// OSPDiagnostics contains diagnostic information for debugging failed requests.
type OSPDiagnostics struct {
	WorkerID     string `json:"worker_id,omitempty"`
	TargetEngine string `json:"target_engine,omitempty"`
	RawError     string `json:"raw_error,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
}

// Thumbnail represents a thumbnail image for a search result.
type Thumbnail struct {
	Src string `json:"src"`
}

// ProfileMeta contains profile information for a search result source.
type ProfileMeta struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Img  string `json:"img,omitempty"`
}

// NewsResults contains news search results.
type NewsResults struct {
	Results []interface{} `json:"results"`
}

// VideoResults contains video search results.
type VideoResults struct {
	Results []interface{} `json:"results"`
}

// ImageResults contains image search results.
type ImageResults struct {
	Results []interface{} `json:"results"`
}

// ResponseMeta contains metadata about the search response.
type ResponseMeta struct {
	LatencyMs uint32 `json:"latency_ms"`
	Total     int    `json:"total"`
}
