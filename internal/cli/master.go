// Package cli provides the command-line interface for GOSP (Go OpenSearchProtocol).
// It defines all available commands including master, worker, search, and status operations.
package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/internal/master"
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/pid"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"github.com/zulfikawr/gosp/pkg/tokens"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var (
	masterNoDaemon bool
)

// masterCmd is the base command for all master-related operations.
var masterCmd = &cobra.Command{
	Use:   "master",
	Short: "Manage Master nodes (The Brain)",
}

func init() {
	masterCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all Master profiles",
		Run: func(cmd *cobra.Command, args []string) {
			list, _ := config.ListMasters()
			fmt.Println("GOSP MASTER PROFILES")
			fmt.Println("--------------------")
			for _, name := range list {
				fmt.Println("-", name)
			}
		},
	})

	masterCmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new Master profile",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := &config.MasterConfig{}
			survey.AskOne(&survey.Input{Message: "Master Name:", Default: "main"}, &cfg.Name)
			survey.AskOne(&survey.Input{Message: "HTTP Port:", Default: "19000"}, &cfg.HTTPPort)
			survey.AskOne(&survey.Input{Message: "gRPC Port:", Default: "19004"}, &cfg.GRPCPort)

			// Generate persistent join token
			token, _ := tokens.Generate()
			cfg.JoinToken = token

			config.SaveMaster(cfg)
			fmt.Println("✅ Master profile created.")
			fmt.Printf("🔑 Join Token: %s\n", cfg.JoinToken)
			fmt.Println("   (Workers will need this token to join the cluster)")
		},
	})

	runCmd := &cobra.Command{
		Use:   "run [profile]",
		Short: "Start a Master node",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "main"
			if len(args) > 0 {
				name = args[0]
			}
			if !masterNoDaemon {
				startMasterDaemon(name)
				return
			}
			runMasterService(name)
		},
	}
	runCmd.Flags().BoolVar(&masterNoDaemon, "no-daemon", false, "Run in foreground")
	masterCmd.AddCommand(runCmd)

	masterCmd.AddCommand(&cobra.Command{
		Use:   "stop [profile]",
		Short: "Stop a running Master node",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "main"
			if len(args) > 0 {
				name = args[0]
			}
			stopService("master", name)
		},
	})

	masterCmd.AddCommand(&cobra.Command{
		Use:   "delete [profile]",
		Short: "Delete a Master profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config.DeleteMaster(args[0])
			fmt.Printf("🗑 Master profile '%s' deleted.\n", args[0])
		},
	})

	rootCmd.AddCommand(masterCmd)
}

// GRPCServer implements the gRPC SearchService for master-worker communication.
type GRPCServer struct {
	protocol.UnimplementedSearchServiceServer
	registry   *master.Registry
	dispatcher *master.Dispatcher
	token      string
}

// validateToken verifies the authorization token from incoming gRPC requests.
func (s *GRPCServer) validateToken(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.Unauthenticated, "metadata is not provided")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "authorization token is not provided")
	}

	if values[0] != s.token {
		return status.Errorf(codes.Unauthenticated, "invalid authorization token")
	}

	return nil
}

// Register handles worker registration requests via gRPC.
func (s *GRPCServer) Register(ctx context.Context, req *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	if err := s.validateToken(ctx); err != nil {
		return nil, err
	}
	p, ok := peer.FromContext(ctx)
	remoteAddr := "unknown"
	if ok {
		remoteAddr = p.Addr.String()
	}
	s.registry.Register(req, remoteAddr)
	return &protocol.RegisterResponse{Success: true, Message: "Registered"}, nil
}

// Connect establishes a bidirectional gRPC stream for worker communication.
func (s *GRPCServer) Connect(stream protocol.SearchService_ConnectServer) error {
	if err := s.validateToken(stream.Context()); err != nil {
		return err
	}
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

// startMasterDaemon launches the master as a background process with output redirected to a log file.
func startMasterDaemon(name string) {
	logDir := filepath.Join(config.GetBaseDir(), "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("❌ Error: Failed to create log directory: %v\n", err)
		os.Exit(1)
	}
	logFile := filepath.Join(logDir, "master_"+name+".log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Printf("❌ Error: Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	// Get the absolute path to the current executable
	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("❌ Error: Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(execPath, "master", "run", name, "--no-daemon")
	cmd.Stdout = f
	cmd.Stderr = f
	if err := cmd.Start(); err != nil {
		fmt.Printf("❌ Error: Failed to start master: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("🚀 Master '%s' started in background (PID: %d)\n", name, cmd.Process.Pid)
	fmt.Printf("📝 Logs: %s\n", logFile)
	os.Exit(0)
}

// runMasterService runs the master node in the foreground, initializing gRPC and HTTP servers.
func runMasterService(name string) {
	cfg, err := config.LoadMaster(name)
	if err != nil {
		fmt.Printf("❌ Error: Master profile '%s' not found.\n", name)
		createNow := false
		survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("Would you like to create profile '%s' now?", name), Default: true}, &createNow)
		if createNow {
			cfg = &config.MasterConfig{Name: name}
			survey.AskOne(&survey.Input{Message: "HTTP Port:", Default: "19000"}, &cfg.HTTPPort)
			survey.AskOne(&survey.Input{Message: "gRPC Port:", Default: "19004"}, &cfg.GRPCPort)
			token, _ := tokens.Generate()
			cfg.JoinToken = token
			config.SaveMaster(cfg)
			fmt.Println("✅ Profile created. Starting...")
		} else {
			os.Exit(1)
		}
	}

	pidPath := config.GetPIDPath("master", name)
	pid.WritePID(pidPath)
	defer pid.RemovePID(pidPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	reg := master.NewRegistry(60 * time.Second)
	sched := master.NewRoundRobinScheduler(reg)
	disp := master.NewDispatcher(sched, reg)
	aggr := master.NewResultAggregator()

	lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
	if err != nil {
		fmt.Printf("❌ Error: Failed to listen on gRPC port %s: %v\n", cfg.GRPCPort, err)
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	protocol.RegisterSearchServiceServer(grpcServer, &GRPCServer{
		registry:   reg,
		dispatcher: disp,
		token:      cfg.JoinToken,
	})

	go func() {
		logger.Info("gRPC Master listening", "port", cfg.GRPCPort, "profile", name)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server failed", "error", err)
		}
	}()

	httpServer := master.NewHTTPServer(disp, aggr)
	go func() {
		// Check for certs in current directory
		if _, err := os.Stat("cert.pem"); err == nil {
			if _, err := os.Stat("key.pem"); err == nil {
				httpServer.ListenTLS(":"+cfg.HTTPPort, "cert.pem", "key.pem")
				return
			}
		}
		httpServer.Listen(":" + cfg.HTTPPort)
	}()

	<-ctx.Done()
	logger.Info("Stopping Master...")
	grpcServer.GracefulStop()
}

// stopService sends a SIGTERM signal to stop a running service by its PID file.
func stopService(role, name string) {
	pidPath := config.GetPIDPath(role, name)
	p, err := pid.ReadPID(pidPath)
	if err != nil {
		fmt.Printf("Error: %s '%s' is not running.\n", role, name)
		return
	}
	process, _ := os.FindProcess(p)
	process.Signal(syscall.SIGTERM)
	fmt.Printf("Stopping %s '%s' (PID: %d)...\n", role, name, p)
}
