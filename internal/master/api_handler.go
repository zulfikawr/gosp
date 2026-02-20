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

type HTTPServer struct {
	app        *fiber.App
	dispatcher *Dispatcher
	aggregator *ResultAggregator
}

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
	s.app.Get("/web/search", s.handleSearch)
	s.app.Get("/cluster/status", s.handleClusterStatus)
}

func (s *HTTPServer) handleClusterStatus(c *fiber.Ctx) error {
	workers := s.dispatcher.registry.GetHealthyWorkers()
	return c.JSON(fiber.Map{
		"active_workers": len(workers),
		"workers":        workers,
		"version":        version.AppVersion,
	})
}

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

func (s *HTTPServer) Listen(addr string) error {
	logger.Info("HTTP Master API listening", "addr", addr)
	return s.app.Listen(addr)
}
