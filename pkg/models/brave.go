package models

// SearchResponse matches the Brave Search API result schema.
// This is the model the Master node returns via HTTP JSON.
type SearchResponse struct {
	Type  string         `json:"type"`
	Query QueryMetadata  `json:"query"`
	Web   WebResults     `json:"web"`
	News  *NewsResults   `json:"news,omitempty"`
	Video *VideoResults  `json:"videos,omitempty"`
	Image *ImageResults  `json:"images,omitempty"`
	Meta  ResponseMeta   `json:"meta"`
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
	Cluster   string `json:"cluster,omitempty"`
}
