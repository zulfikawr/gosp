package master

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/models"
	"github.com/zulfikawr/go-search/pkg/protocol"
)

// HTTPServer handles public search requests and maps them to OSP tasks.
type HTTPServer struct {
	app        *fiber.App
	dispatcher *Dispatcher
	aggregator *ResultAggregator
}

// NewHTTPServer initializes a new Fiber-based HTTP server for the Master node.
func NewHTTPServer(d *Dispatcher, a *ResultAggregator) *HTTPServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
	})

	s := &HTTPServer{
		app:        app,
		dispatcher: d,
		aggregator: a,
	}

	s.setupRoutes()
	return s
}

func (s *HTTPServer) setupRoutes() {
	// Brave API Parity Endpoint
	s.app.Get("/web/search", s.handleSearch)
}

// handleSearch implements the /web/search endpoint matching Brave Search API v1.
func (s *HTTPServer) handleSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query parameter 'q' is required",
		})
	}

	// 1. Prepare OSP Search Request
	taskID := uuid.New().String()
	// Default to Google if not specified, or we could round-robin across multiple engines here.
	engine := protocol.Engine_ENGINE_GOOGLE 
	
	req := &protocol.SearchRequest{
		TaskId: taskID,
		Query:  query,
		Engine: engine,
		Count:  int32(c.QueryInt("count", 10)),
		Offset: int32(c.QueryInt("offset", 0)),
		Params: make(map[string]string),
	}

	// 2. Dispatch to Workers with a strict timeout (Brave API targets < 1s)
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := s.dispatcher.Dispatch(ctx, req)
	if err != nil {
		logger.Error("dispatch_failed", "error", err, "query", query)
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("failed to fetch results: %v", err),
		})
	}

	// 3. Aggregate (if we had multiple sources, we'd dispatch them concurrently)
	finalResp := s.aggregator.Aggregate(resp)
	latency := uint32(time.Since(start).Milliseconds())

	// 4. Map to Brave Search API Response Schema
	braveResp := models.SearchResponse{
		Type: "search",
		Query: models.QueryMetadata{
			Original: query,
		},
		Web: models.WebResults{
			Type:    "search_results",
			Results: make([]models.WebResult, 0),
		},
		Meta: models.ResponseMeta{
			LatencyMs: latency,
			Total:     len(finalResp.Results),
		},
	}

	for _, res := range finalResp.Results {
		braveResp.Web.Results = append(braveResp.Web.Results, models.WebResult{
			Title:       res.Title,
			URL:         res.Url,
			Description: res.Description,
			Age:         res.Age,
			Language:    res.Language,
		})
	}

	return c.JSON(braveResp)
}

// Listen starts the HTTP server.
func (s *HTTPServer) Listen(addr string) error {
	logger.Info("HTTP Master API listening", "addr", addr)
	return s.app.Listen(addr)
}
