package master

import (
	"context"
	"fmt"
	"sync"

	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/metrics"
	"github.com/zulfikawr/go-search/pkg/protocol"
)

// Dispatcher coordinates search task assignment and result collection.
type Dispatcher struct {
	scheduler Scheduler
	registry  *Registry
	
	mu           sync.RWMutex
	pendingTasks map[string]chan *protocol.SearchResponse
}

// NewDispatcher initializes a new task dispatcher.
func NewDispatcher(s Scheduler, r *Registry) *Dispatcher {
	return &Dispatcher{
		scheduler:    s,
		registry:     r,
		pendingTasks: make(map[string]chan *protocol.SearchResponse),
	}
}

// Dispatch sends a search request to a suitable worker and waits for the result.
func (d *Dispatcher) Dispatch(ctx context.Context, req *protocol.SearchRequest) (*protocol.SearchResponse, error) {
	// 1. Select a worker
	worker, err := d.scheduler.NextWorker(req.Engine)
	if err != nil {
		return nil, fmt.Errorf("dispatch: %w", err)
	}

	// 2. Register a pending task channel to receive the result later.
	resChan := make(chan *protocol.SearchResponse, 1)
	d.mu.Lock()
	d.pendingTasks[req.TaskId] = resChan
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		delete(d.pendingTasks, req.TaskId)
		d.mu.Unlock()
	}()

	// 3. Send the command to the worker's command channel.
	// The Master's gRPC stream handler will pick this up and send it to the worker.
	command := &protocol.MasterCommand{
		Command: &protocol.MasterCommand_Task{
			Task: req,
		},
	}

	select {
	case worker.CommandChan <- command:
		logger.Debug("task dispatched to worker", "task_id", req.TaskId, "worker_id", worker.ID)
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return nil, fmt.Errorf("worker %s command channel is full", worker.ID)
	}

	// 4. Wait for the response from the worker (pushed via HandleResponse).
	select {
	case res := <-resChan:
		metrics.SearchLatency.WithLabelValues(req.Engine.String()).Observe(float64(res.LatencyMs) / 1000.0)
		return res, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("task timed out: %w", ctx.Err())
	}
}

// HandleResponse correlates a SearchResponse from a worker with a pending task.
func (d *Dispatcher) HandleResponse(res *protocol.SearchResponse) {
	d.mu.RLock()
	resChan, exists := d.pendingTasks[res.TaskId]
	d.mu.RUnlock()

	if exists {
		resChan <- res
	} else {
		logger.Warn("received response for unknown task", "task_id", res.TaskId)
	}
}
