package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zulfikawr/go-search/internal/master"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	httpAddr = flag.String("http", ":18789", "HTTP API listen address")
	grpcAddr = flag.String("grpc", ":50051", "gRPC listen address")
)

// GRPCServer implements the OSP Master-Worker gRPC protocol.
type GRPCServer struct {
	protocol.UnimplementedSearchServiceServer
	registry   *master.Registry
	dispatcher *master.Dispatcher
}

func (s *GRPCServer) Register(ctx context.Context, req *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	p, ok := peer.FromContext(ctx)
	remoteAddr := "unknown"
	if ok {
		remoteAddr = p.Addr.String()
	}
	s.registry.Register(req, remoteAddr)
	return &protocol.RegisterResponse{Success: true, Message: "Registered"}, nil
}

func (s *GRPCServer) Connect(stream protocol.SearchService_ConnectServer) error {
	// 1. Initial status for ID identification
	status, err := stream.Recv()
	if err != nil {
		return err
	}
	workerID := status.WorkerId
	worker := s.registry.GetWorker(workerID)
	if worker == nil {
		return fmt.Errorf("worker %s not registered", workerID)
	}

	logger.Info("worker connected via gRPC stream", "worker_id", workerID)
	defer s.registry.Deregister(workerID)

	// 2. Bidirectional Loop: Dispatch tasks and receive results
	errChan := make(chan error, 2)

	// Go-routine: Receive results/heartbeats from worker
	go func() {
		for {
			status, err := stream.Recv()
			if err != nil {
				errChan <- err
				return
			}
			s.registry.UpdateStatus(workerID, status)
			if status.CompletedTask != nil {
				s.dispatcher.HandleResponse(status.CompletedTask)
			}
		}
	}()

	// Go-routine: Send commands/tasks to worker
	go func() {
		for {
			select {
			case cmd := <-worker.CommandChan:
				if err := stream.Send(cmd); err != nil {
					errChan <- err
					return
				}
			case <-stream.Context().Done():
				errChan <- stream.Context().Err()
				return
			}
		}
	}()

	return <-errChan
}

func main() {
	flag.Parse()

	// 1. Initialize Registry and Scheduler
	reg := master.NewRegistry(60 * time.Second)
	sched := master.NewRoundRobinScheduler(reg)
	disp := master.NewDispatcher(sched, reg)
	aggr := master.NewResultAggregator()

	// 2. Start gRPC Server
	lis, err := net.Listen("tcp", *grpcAddr)
	if err != nil {
		logger.Error("failed to listen gRPC", "error", err)
		os.Exit(1)
	}
	
	grpcServer := grpc.NewServer()
	protocol.RegisterSearchServiceServer(grpcServer, &GRPCServer{
		registry:   reg,
		dispatcher: disp,
	})
	
	go func() {
		logger.Info("gRPC Master listening", "addr", *grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	// 3. Start HTTP API Server
	httpServer := master.NewHTTPServer(disp, aggr)
	go func() {
		if err := httpServer.Listen(*httpAddr); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	// 4. Graceful Shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Master node...")
	grpcServer.GracefulStop()
	logger.Info("Master node stopped.")
}
