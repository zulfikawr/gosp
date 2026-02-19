package cli

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/internal/master"
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/pid"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

var (
	masterNoDaemon bool
)

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
			config.SaveMaster(cfg)
			fmt.Println("✅ Master profile created.")
		},
	})

	runCmd := &cobra.Command{
		Use:   "run [profile]",
		Short: "Start a Master node",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "main"
			if len(args) > 0 { name = args[0] }
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
			if len(args) > 0 { name = args[0] }
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

type GRPCServer struct {
	protocol.UnimplementedSearchServiceServer
	registry   *master.Registry
	dispatcher *master.Dispatcher
}

func (s *GRPCServer) Register(ctx context.Context, req *protocol.RegisterRequest) (*protocol.RegisterResponse, error) {
	p, ok := peer.FromContext(ctx)
	remoteAddr := "unknown"
	if ok { remoteAddr = p.Addr.String() }
	s.registry.Register(req, remoteAddr)
	return &protocol.RegisterResponse{Success: true, Message: "Registered"}, nil
}

func (s *GRPCServer) Connect(stream protocol.SearchService_ConnectServer) error {
	status, err := stream.Recv()
	if err != nil { return err }
	workerID := status.WorkerId
	worker := s.registry.GetWorker(workerID)
	if worker == nil { return fmt.Errorf("worker %s not registered", workerID) }

	logger.Info("worker connected via gRPC stream", "worker_id", workerID)
	defer s.registry.Deregister(workerID)

	errChan := make(chan error, 2)
	go func() {
		for {
			status, err := stream.Recv()
			if err != nil { errChan <- err; return }
			s.registry.UpdateStatus(workerID, status)
			if status.CompletedTask != nil { s.dispatcher.HandleResponse(status.CompletedTask) }
		}
	}()
	go func() {
		for {
			select {
			case cmd := <-worker.CommandChan:
				if err := stream.Send(cmd); err != nil { errChan <- err; return }
			case <-stream.Context().Done():
				errChan <- stream.Context().Err(); return
			}
		}
	}()
	return <-errChan
}

func startMasterDaemon(name string) {
	cmd := exec.Command(os.Args[0], "master", "run", name, "--no-daemon")
	cmd.Start()
	fmt.Printf("🚀 Master '%s' started in background (PID: %d)\n", name, cmd.Process.Pid)
	os.Exit(0)
}

func runMasterService(name string) {
	cfg, err := config.LoadMaster(name)
	if err != nil {
		fmt.Printf("Error: Profile '%s' not found.\n", name)
		os.Exit(1)
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

	lis, _ := net.Listen("tcp", ":"+cfg.GRPCPort)
	grpcServer := grpc.NewServer()
	protocol.RegisterSearchServiceServer(grpcServer, &GRPCServer{registry: reg, dispatcher: disp})
	
	go func() {
		logger.Info("gRPC Master listening", "port", cfg.GRPCPort, "profile", name)
		grpcServer.Serve(lis)
	}()

	httpServer := master.NewHTTPServer(disp, aggr)
	go httpServer.Listen(":" + cfg.HTTPPort)

	<-ctx.Done()
	logger.Info("Stopping Master...")
	grpcServer.GracefulStop()
}

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
