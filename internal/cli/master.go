package cli

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/internal/master"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"context"
)

var (
	masterHttpAddr string
	masterGrpcAddr string
)

var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Start the GOSP Master node",
	Run: func(cmd *cobra.Command, args []string) {
		runMaster()
	},
}

func init() {
	masterCmd.Flags().StringVar(&masterHttpAddr, "http", ":19000", "HTTP API listen address")
	masterCmd.Flags().StringVar(&masterGrpcAddr, "grpc", ":19004", "gRPC listen address")
	rootCmd.AddCommand(masterCmd)
}

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

	errChan := make(chan error, 2)

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

func runMaster() {
	reg := master.NewRegistry(60 * time.Second)
	sched := master.NewRoundRobinScheduler(reg)
	disp := master.NewDispatcher(sched, reg)
	aggr := master.NewResultAggregator()

	lis, err := net.Listen("tcp", masterGrpcAddr)
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
		logger.Info("gRPC Master listening", "addr", masterGrpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	httpServer := master.NewHTTPServer(disp, aggr)
	go func() {
		if err := httpServer.Listen(masterHttpAddr); err != nil {
			logger.Error("HTTP server failed", "error", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down Master node...")
	grpcServer.GracefulStop()
	logger.Info("Master node stopped.")
}
