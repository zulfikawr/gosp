package models

// SearchResponse matches the Brave Search API result schema.
type SearchResponse struct {
	Type  string         `json:"type"`
	Query QueryMetadata  `json:"query"`
	Web   WebResults     `json:"web"`
	News  *NewsResults   `json:"news,omitempty"`
	Video *VideoResults  `json:"videos,omitempty"`
	Image *ImageResults  `json:"images,omitempty"`
	Meta  ResponseMeta   `json:"meta"`

	// OSP Extended Metadata (Opt-in via ?metadata=true)
	Performance *OSPPerformance `json:"osp_performance,omitempty"`
	Cluster     *OSPCluster     `json:"osp_cluster,omitempty"`
	Diagnostics *OSPDiagnostics `json:"osp_diagnostics,omitempty"`
}

type QueryMetadata struct {
	Original  string `json:"original"`
	Altered   string `json:"altered,omitempty"`
	Canonical string `json:"canonical,omitempty"`
}

type WebResults struct {
	Type    string       `json:"type"`
	Results []WebResult  `json:"results"`
}

type WebResult struct {
	Title       string        `json:"title"`
	URL         string        `json:"url"`
	Description string        `json:"description"`
	Age         string        `json:"age,omitempty"`
	Language    string        `json:"language,omitempty"`
	Thumbnail   *Thumbnail    `json:"thumbnail,omitempty"`
	Profile     *ProfileMeta  `json:"profile,omitempty"`

	// OSP Specific Result Signals
	Signals *OSPSignals `json:"osp_signals,omitempty"`
}

type OSPSignals struct {
	Source    string  `json:"source"`
	WorkerID  string  `json:"worker_id"`
	Region    string  `json:"region"`
	Relevance float32 `json:"relevance,omitempty"`
}

type OSPPerformance struct {
	WorkerScrapeMs uint32 `json:"worker_scrape_ms"`
	MasterAggMs    uint32 `json:"master_agg_ms"`
	TotalLatencyMs uint32 `json:"total_latency_ms"`
}

type OSPCluster struct {
	NodesQueried    int    `json:"nodes_queried"`
	CacheStatus     string `json:"cache_status"`
	ProtocolVersion string `json:"protocol_version"`
}

type OSPDiagnostics struct {
	WorkerID     string `json:"worker_id,omitempty"`
	TargetEngine string `json:"target_engine,omitempty"`
	RawError     string `json:"raw_error,omitempty"`
	TraceID      string `json:"trace_id,omitempty"`
}

type Thumbnail struct {
	Src string `json:"src"`
}

type ProfileMeta struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Img  string `json:"img,omitempty"`
}

type NewsResults struct {
	Results []interface{} `json:"results"`
}

type VideoResults struct {
	Results []interface{} `json:"results"`
}

type ImageResults struct {
	Results []interface{} `json:"results"`
}

type ResponseMeta struct {
	LatencyMs uint32 `json:"latency_ms"`
	Total     int    `json:"total"`
}
