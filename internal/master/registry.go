package master

import (
	"sync"
	"time"

	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/metrics"
	"github.com/zulfikawr/gosp/pkg/protocol"
)

// WorkerNode represents a connected worker node in the cluster.
type WorkerNode struct {
	ID               string
	Version          string
	SupportedEngines []protocol.Engine
	LastHeartbeat    time.Time
	CPUUsage         float32
	MemoryUsage      float32
	ActiveTasks      uint32
	RemoteAddr       string
	Region           string
	
	// CommandChan is used by the Dispatcher to send tasks to this specific worker.
	CommandChan chan *protocol.MasterCommand
}

// Registry manages the state of all connected worker nodes.
type Registry struct {
	mu      sync.RWMutex
	workers map[string]*WorkerNode
	
	pruneTimeout time.Duration
}

// NewRegistry initializes a new worker registry.
func NewRegistry(pruneTimeout time.Duration) *Registry {
	r := &Registry{
		workers:      make(map[string]*WorkerNode),
		pruneTimeout: pruneTimeout,
	}

	go r.startPruner()

	return r
}

// Register adds or updates a worker node in the registry.
func (r *Registry) Register(req *protocol.RegisterRequest, remoteAddr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, exists := r.workers[req.WorkerId]
	if !exists {
		node = &WorkerNode{
			ID:          req.WorkerId,
			CommandChan: make(chan *protocol.MasterCommand, 10),
		}
		r.workers[req.WorkerId] = node
		logger.Info("new worker registered", "worker_id", req.WorkerId, "addr", remoteAddr, "engines", req.SupportedEngines)
	}

	node.Version = req.Version
	node.SupportedEngines = req.SupportedEngines
	node.RemoteAddr = remoteAddr
	node.Region = req.Region
	node.LastHeartbeat = time.Now()

	metrics.ActiveWorkerCount.Set(float64(len(r.workers)))
}

// UpdateStatus updates the status of an existing worker.
func (r *Registry) UpdateStatus(workerID string, status *protocol.WorkerStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, exists := r.workers[workerID]; exists {
		node.LastHeartbeat = time.Now()
		node.CPUUsage = status.CpuUsage
		node.MemoryUsage = status.MemoryUsage
		node.ActiveTasks = status.ActiveTasks
	}
}

// Deregister removes a worker from the registry.
func (r *Registry) Deregister(workerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workers[workerID]; exists {
		delete(r.workers, workerID)
		logger.Info("worker deregistered", "worker_id", workerID)
		metrics.ActiveWorkerCount.Set(float64(len(r.workers)))
	}
}

// GetHealthyWorkers returns a list of workers that have sent heartbeats recently.
func (r *Registry) GetHealthyWorkers() []*WorkerNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var healthy []*WorkerNode
	now := time.Now()

	for _, node := range r.workers {
		if now.Sub(node.LastHeartbeat) < r.pruneTimeout {
			healthy = append(healthy, node)
		}
	}

	return healthy
}

func (r *Registry) GetWorker(id string) *WorkerNode {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.workers[id]
}

// startPruner runs a background goroutine to remove inactive workers.
func (r *Registry) startPruner() {
	ticker := time.NewTicker(r.pruneTimeout / 2)
	defer ticker.Stop()

	for range ticker.C {
		r.prune()
	}
}

func (r *Registry) prune() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for id, node := range r.workers {
		if now.Sub(node.LastHeartbeat) >= r.pruneTimeout {
			logger.Warn("pruning inactive worker", "worker_id", id, "last_seen", node.LastHeartbeat)
			delete(r.workers, id)
		}
	}

	metrics.ActiveWorkerCount.Set(float64(len(r.workers)))
}
