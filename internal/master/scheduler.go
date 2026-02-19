package master

import (
	"sync"

	"github.com/zulfikawr/gosp/pkg/protocol"
)

// Scheduler defines the interface for worker selection strategies.
type Scheduler interface {
	NextWorker(engine protocol.Engine) (*WorkerNode, error)
}

// RoundRobinScheduler implements a simple round-robin worker selection.
type RoundRobinScheduler struct {
	registry *Registry
	mu       sync.Mutex
	lastIdx  map[protocol.Engine]int
}

// NewRoundRobinScheduler initializes a new round-robin scheduler.
func NewRoundRobinScheduler(r *Registry) *RoundRobinScheduler {
	return &RoundRobinScheduler{
		registry: r,
		lastIdx:  make(map[protocol.Engine]int),
	}
}

// NextWorker picks the next healthy worker that supports the requested engine.
func (s *RoundRobinScheduler) NextWorker(engine protocol.Engine) (*WorkerNode, error) {
	workers := s.registry.GetHealthyWorkers()
	
	// Filter workers by engine support
	var eligible []*WorkerNode
	for _, w := range workers {
		for _, e := range w.SupportedEngines {
			if e == engine {
				eligible = append(eligible, w)
				break
			}
		}
	}

	if len(eligible) == 0 {
		return nil, ToProtocolError(protocol.ErrorCode_ERROR_CODE_PROVIDER_DOWN, "no healthy workers support this engine")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	idx := (s.lastIdx[engine] + 1) % len(eligible)
	s.lastIdx[engine] = idx

	return eligible[idx], nil
}

// ToProtocolError is a helper to convert ErrorCode to a standard error.
func ToProtocolError(e protocol.ErrorCode, msg string) error {
	return &protocolError{code: e, msg: msg}
}

type protocolError struct {
	code protocol.ErrorCode
	msg  string
}

func (p *protocolError) Error() string { return p.msg }
