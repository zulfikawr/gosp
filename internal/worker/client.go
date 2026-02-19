package worker

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/zulfikawr/gosp/internal/worker/scraper"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// Client handles the worker's connection and communication with the OSP Master.
type Client struct {
	id               string
	version          string
	masterAddr       string
	supportedEngines []protocol.Engine
	creds            credentials.TransportCredentials
	region           string
	
	conn *grpc.ClientConn
	scrapers map[protocol.Engine]scraper.Engine
}

func NewClient(id, version, masterAddr string, engines []protocol.Engine, creds credentials.TransportCredentials) *Client {
	c := &Client{
		id:               id,
		version:          version,
		masterAddr:       masterAddr,
		supportedEngines: engines,
		creds:            creds,
		region:           "US-Cloud", // Default region
		scrapers:         make(map[protocol.Engine]scraper.Engine),
	}

	for _, e := range engines {
		switch e {
		case protocol.Engine_ENGINE_GOOGLE:
			c.scrapers[e] = scraper.NewGoogleScraper()
		case protocol.Engine_ENGINE_BRAVE:
			c.scrapers[e] = scraper.NewBraveScraper()
		case protocol.Engine_ENGINE_DUCKDUCKGO:
			c.scrapers[e] = scraper.NewDuckDuckGoScraper()
		}
	}

	return c
}

func (c *Client) Run(ctx context.Context) error {
	var err error
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(c.creds),
	}
	
	c.conn, err = grpc.DialContext(ctx, c.masterAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial master: %w", err)
	}
	defer c.conn.Close()

	client := protocol.NewSearchServiceClient(c.conn)

	regReq := &protocol.RegisterRequest{
		WorkerId:         c.id,
		Version:          c.version,
		SupportedEngines: c.supportedEngines,
		Region:           c.region,
	}
	
	_, err = client.Register(ctx, regReq)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	logger.Info("worker registered successfully", "worker_id", c.id, "region", c.region)

	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	initialStatus := &protocol.WorkerStatus{
		WorkerId: c.id,
	}
	if err := stream.Send(initialStatus); err != nil {
		return fmt.Errorf("failed to send initial status: %w", err)
	}

	return c.handleStream(ctx, stream)
}

func (c *Client) handleStream(ctx context.Context, stream protocol.SearchService_ConnectClient) error {
	errChan := make(chan error, 2)

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				var m runtime.MemStats
				runtime.ReadMemStats(&m)
				
				status := &protocol.WorkerStatus{
					WorkerId:    c.id,
					MemoryUsage: float32(m.Alloc) / 1024 / 1024,
					ActiveTasks: 0,
				}
				
				if err := stream.Send(status); err != nil {
					errChan <- err
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		for {
			cmd, err := stream.Recv()
			if err == io.EOF {
				errChan <- nil
				return
			}
			if err != nil {
				errChan <- err
				return
			}

			if task := cmd.GetTask(); task != nil {
				start := time.Now()
				logger.Info("received search task", "task_id", task.TaskId, "query", task.Query, "engine", task.Engine)
				
				var resp *protocol.SearchResponse
				if s, exists := c.scrapers[task.Engine]; exists {
					results, err := s.Search(task.Query, task.Count, task.Offset)
					
					// Enrich results with OSP metadata
					for _, r := range results {
						r.SourceEngine = task.Engine
						r.WorkerId = c.id
						r.WorkerRegion = c.region
					}

					if err != nil {
						logger.Error("scrape failed", "error", err, "task_id", task.TaskId)
						resp = &protocol.SearchResponse{
							TaskId:        task.TaskId,
							ErrorCode:     protocol.ErrorCode_ERROR_CODE_INTERNAL_ERROR,
							ErrorMessage:  err.Error(),
							WorkerId:      c.id,
							WorkerRegion:  c.region,
						}
					} else {
						resp = &protocol.SearchResponse{
							TaskId:           task.TaskId,
							ErrorCode:        protocol.ErrorCode_ERROR_CODE_SUCCESS,
							Results:          results,
							ScrapeLatencyMs:  uint32(time.Since(start).Milliseconds()),
							WorkerId:         c.id,
							WorkerRegion:     c.region,
						}
					}
				} else {
					resp = &protocol.SearchResponse{
						TaskId:       task.TaskId,
						ErrorCode:    protocol.ErrorCode_ERROR_CODE_PROVIDER_DOWN,
						ErrorMessage: "engine not supported by this worker",
						WorkerId:     c.id,
						WorkerRegion: c.region,
					}
				}
				
				status := &protocol.WorkerStatus{
					WorkerId:      c.id,
					CompletedTask: resp,
				}
				stream.Send(status)
			}
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
