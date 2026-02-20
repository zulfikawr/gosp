// Package master provides the master node functionality for GOSP.
// It handles worker coordination, task dispatching, and HTTP API serving.
package master

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/models"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"github.com/zulfikawr/gosp/pkg/version"
)

// HTTPServer handles incoming HTTP requests for the master node.
type HTTPServer struct {
	app        *fiber.App
	dispatcher *Dispatcher
	aggregator *ResultAggregator
}

// NewHTTPServer creates a new HTTP server instance with the given dispatcher and aggregator.
func NewHTTPServer(d *Dispatcher, a *ResultAggregator) *HTTPServer {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		ReadTimeout:           10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			logger.Error("fiber error", "error", err)
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		},
	})

	s := &HTTPServer{
		app:        app,
		dispatcher: d,
		aggregator: a,
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures the HTTP routes for the server.
func (s *HTTPServer) setupRoutes() {
	s.app.Get("/web/search", s.handleSearch)
	// Brave API Compatibility Route
	s.app.Get("/res/v1/web/search", s.handleSearch)
	s.app.Get("/cluster/status", s.handleClusterStatus)
}

// handleClusterStatus returns the current cluster status including connected workers.
func (s *HTTPServer) handleClusterStatus(c *fiber.Ctx) error {
	logger.Debug("handling cluster status request")
	workers := s.dispatcher.registry.GetHealthyWorkers()

	// Create a safe view of workers for JSON serialization (avoiding CommandChan)
	type WorkerView struct {
		ID               string            `json:"id"`
		Version          string            `json:"version"`
		SupportedEngines []protocol.Engine `json:"supported_engines"`
		Region           string            `json:"region"`
		RemoteAddr       string            `json:"remote_addr"`
		CPUUsage         float32           `json:"cpu_usage"`
		MemoryUsage      float32           `json:"memory_usage"`
	}

	workerViews := make([]WorkerView, len(workers))
	for i, w := range workers {
		workerViews[i] = WorkerView{
			ID:               w.ID,
			Version:          w.Version,
			SupportedEngines: w.SupportedEngines,
			Region:           w.Region,
			RemoteAddr:       w.RemoteAddr,
			CPUUsage:         w.CPUUsage,
			MemoryUsage:      w.MemoryUsage,
		}
	}

	return c.JSON(fiber.Map{
		"active_workers": len(workers),
		"workers":        workerViews,
		"version":        version.AppVersion,
	})
}

// handleSearch processes incoming search requests and returns aggregated results.
func (s *HTTPServer) handleSearch(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "query parameter 'q' is required",
		})
	}

	showMetadata := c.Query("metadata") == "true"
	taskID := uuid.New().String()

	// Engine selection
	engineStr := c.Query("engine", "duckduckgo")
	engine := protocol.Engine_ENGINE_DUCKDUCKGO
	switch engineStr {
	case "1", "google":
		engine = protocol.Engine_ENGINE_GOOGLE
	case "2", "bing":
		engine = protocol.Engine_ENGINE_BING
	case "3", "brave":
		engine = protocol.Engine_ENGINE_BRAVE
	case "4", "duckduckgo":
		engine = protocol.Engine_ENGINE_DUCKDUCKGO
	}

	req := &protocol.SearchRequest{
		TaskId: taskID,
		Query:  query,
		Engine: engine,
		Count:  int32(c.QueryInt("count", 10)),
		Offset: int32(c.QueryInt("offset", 0)),
		Params: make(map[string]string),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	resp, err := s.dispatcher.Dispatch(ctx, req)
	masterLatency := uint32(time.Since(start).Milliseconds())

	if err != nil {
		logger.Error("dispatch_failed", "error", err, "query", query)

		errResp := fiber.Map{"error": fmt.Sprintf("failed to fetch results: %v", err)}
		if showMetadata {
			errResp["osp_diagnostics"] = models.OSPDiagnostics{
				TargetEngine: engine.String(),
				RawError:     err.Error(),
				TraceID:      taskID,
			}
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(errResp)
	}

	finalResp := s.aggregator.Aggregate(resp)

	// Map to Brave API Schema
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
			LatencyMs: masterLatency,
			Total:     len(finalResp.Results),
		},
	}

	// Inject Results
	for _, res := range finalResp.Results {
		item := models.WebResult{
			Title:       res.Title,
			URL:         res.Url,
			Description: res.Description,
			Age:         res.Age,
			Language:    res.Language,
		}

		// Optional OSP Signals per result
		if showMetadata {
			item.Signals = &models.OSPSignals{
				Source:   res.SourceEngine.String(),
				WorkerID: res.WorkerId,
				Region:   res.WorkerRegion,
			}
		}

		braveResp.Web.Results = append(braveResp.Web.Results, item)
	}

	// Optional OSP Root Metadata
	if showMetadata {
		braveResp.Performance = &models.OSPPerformance{
			WorkerScrapeMs: resp.ScrapeLatencyMs,
			MasterAggMs:    masterLatency - resp.ScrapeLatencyMs,
			TotalLatencyMs: masterLatency,
		}
		braveResp.Cluster = &models.OSPCluster{
			NodesQueried:    1,
			CacheStatus:     "MISS",
			ProtocolVersion: "v0.1.0",
		}
	}

	return c.JSON(braveResp)
}

// Listen starts the HTTP server on the specified address.
func (s *HTTPServer) Listen(addr string) error {
	logger.Info("HTTP Master API listening", "addr", addr)
	return s.app.Listen(addr)
}

// ListenTLS starts the HTTP server with TLS on the specified address.
func (s *HTTPServer) ListenTLS(addr, certFile, keyFile string) error {
	logger.Info("HTTPS Master API listening", "addr", addr)
	return s.app.ListenTLS(addr, certFile, keyFile)
}
