package worker

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"time"

	"github.com/zulfikawr/go-search/internal/worker/scraper"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/protocol"
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
	
	// conn is the active gRPC connection
	conn *grpc.ClientConn

	// scrapers is a pool of initialized search engines.
	scrapers map[protocol.Engine]scraper.Engine
}

// NewClient initializes a new OSP Worker client.
func NewClient(id, version, masterAddr string, engines []protocol.Engine, creds credentials.TransportCredentials) *Client {
	c := &Client{
		id:               id,
		version:          version,
		masterAddr:       masterAddr,
		supportedEngines: engines,
		creds:            creds,
		scrapers:         make(map[protocol.Engine]scraper.Engine),
	}

	// Initialize supported scrapers
	for _, e := range engines {
		switch e {
		case protocol.Engine_ENGINE_GOOGLE:
			c.scrapers[e] = scraper.NewGoogleScraper()
		case protocol.Engine_ENGINE_BRAVE:
			c.scrapers[e] = scraper.NewBraveScraper()
		}
	}

	return c
}

// Run starts the worker client lifecycle: Connect -> Register -> Stream.
func (c *Client) Run(ctx context.Context) error {
	var err error
	
	// 1. Establish gRPC Connection
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(c.creds),
	}
	
	c.conn, err = grpc.DialContext(ctx, c.masterAddr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial master: %w", err)
	}
	defer c.conn.Close()

	client := protocol.NewSearchServiceClient(c.conn)

	// 2. Register with Master
	regReq := &protocol.RegisterRequest{
		WorkerId:         c.id,
		Version:          c.version,
		SupportedEngines: c.supportedEngines,
	}
	
	_, err = client.Register(ctx, regReq)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	logger.Info("worker registered successfully", "worker_id", c.id)

	// 3. Open Bidirectional Stream
	stream, err := client.Connect(ctx)
	if err != nil {
		return fmt.Errorf("failed to open stream: %w", err)
	}

	// 4. Start Heartbeat & Task Loop
	return c.handleStream(ctx, stream)
}

func (c *Client) handleStream(ctx context.Context, stream protocol.SearchService_ConnectClient) error {
	errChan := make(chan error, 2)

	// Go-routine: Send Heartbeats and Status
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
					CpuUsage:    0.0, 
					MemoryUsage: float32(m.Alloc) / 1024 / 1024, // MB
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

	// Go-routine: Receive Tasks from Master
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
					if err != nil {
						logger.Error("scrape failed", "error", err, "task_id", task.TaskId)
						resp = &protocol.SearchResponse{
							TaskId:        task.TaskId,
							ErrorCode:     protocol.ErrorCode_ERROR_CODE_INTERNAL_ERROR,
							ErrorMessage:  err.Error(),
						}
					} else {
						resp = &protocol.SearchResponse{
							TaskId:    task.TaskId,
							ErrorCode: protocol.ErrorCode_ERROR_CODE_SUCCESS,
							Results:   results,
							LatencyMs: uint32(time.Since(start).Milliseconds()),
						}
					}
				} else {
					resp = &protocol.SearchResponse{
						TaskId:       task.TaskId,
						ErrorCode:    protocol.ErrorCode_ERROR_CODE_PROVIDER_DOWN,
						ErrorMessage: "engine not supported by this worker",
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
