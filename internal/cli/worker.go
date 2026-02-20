package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/internal/worker"
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/pid"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"github.com/zulfikawr/gosp/pkg/version"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	workerNoDaemon bool
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Manage Worker nodes (The Scrapers)",
}

func init() {
	// Worker List
	workerCmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List all Worker profiles",
		Run: func(cmd *cobra.Command, args []string) {
			list, _ := config.ListWorkers()
			fmt.Println("GOSP WORKER PROFILES")
			fmt.Println("--------------------")
			for _, id := range list {
				fmt.Println("-", id)
			}
		},
	})

	// Worker Create
	workerCmd.AddCommand(&cobra.Command{
		Use:   "create",
		Short: "Create a new Worker profile",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := &config.WorkerConfig{}
			survey.AskOne(&survey.Input{Message: "Worker ID:", Default: "local-01"}, &cfg.ID)
			survey.AskOne(&survey.Input{Message: "Master gRPC URL:", Default: "localhost:19004"}, &cfg.MasterURL)
			survey.AskOne(&survey.Input{Message: "Region:", Default: "US-Cloud"}, &cfg.Region)
			survey.AskOne(&survey.Input{Message: "Join Token (Required):"}, &cfg.JoinToken)

			if cfg.JoinToken == "" {
				fmt.Println("❌ Error: A Join Token is required to connect to a Master.")
				os.Exit(1)
			}

			config.SaveWorker(cfg)
			fmt.Println("✅ Worker profile created.")
		},
	})

	// Worker Run
	runCmd := &cobra.Command{
		Use:   "run [profile]",
		Short: "Start a Worker node",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := "local-01"
			if len(args) > 0 {
				id = args[0]
			}
			if !workerNoDaemon {
				startWorkerDaemon(id)
				return
			}
			runWorkerService(id)
		},
	}
	runCmd.Flags().BoolVar(&workerNoDaemon, "no-daemon", false, "Run in foreground")
	workerCmd.AddCommand(runCmd)

	// Worker Stop
	workerCmd.AddCommand(&cobra.Command{
		Use:   "stop [profile]",
		Short: "Stop a running Worker node",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := "local-01"
			if len(args) > 0 {
				id = args[0]
			}
			stopService("worker", id)
		},
	})

	// Worker Delete
	workerCmd.AddCommand(&cobra.Command{
		Use:   "delete [profile]",
		Short: "Delete a Worker profile",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			config.DeleteWorker(args[0])
			fmt.Printf("🗑 Worker profile '%s' deleted.\n", args[0])
		},
	})

	rootCmd.AddCommand(workerCmd)
}

func startWorkerDaemon(id string) {
	logDir := filepath.Join(config.GetBaseDir(), "logs")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, "worker_"+id+".log")
	f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)

	cmd := exec.Command(os.Args[0], "worker", "run", id, "--no-daemon")
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Start()
	fmt.Printf("🚀 Worker '%s' connected in background (PID: %d)\n", id, cmd.Process.Pid)
	fmt.Printf("📝 Logs: %s\n", logFile)
	os.Exit(0)
}

func runWorkerService(id string) {
	cfg, err := config.LoadWorker(id)
	if err != nil {
		fmt.Printf("❌ Error: Worker profile '%s' not found.\n", id)

		createNow := false
		survey.AskOne(&survey.Confirm{Message: fmt.Sprintf("Would you like to create profile '%s' now?", id), Default: true}, &createNow)
		if createNow {
			cfg = &config.WorkerConfig{ID: id}
			survey.AskOne(&survey.Input{Message: "Master gRPC URL:", Default: "localhost:19004"}, &cfg.MasterURL)
			survey.AskOne(&survey.Input{Message: "Region:", Default: "US-Cloud"}, &cfg.Region)
			survey.AskOne(&survey.Input{Message: "Join Token (Required):"}, &cfg.JoinToken)
			config.SaveWorker(cfg)
			fmt.Println("✅ Profile created. Connecting...")
		} else {
			os.Exit(1)
		}
	}

	pidPath := config.GetPIDPath("worker", id)
	pid.WritePID(pidPath)
	defer pid.RemovePID(pidPath)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	engines := []protocol.Engine{
		protocol.Engine_ENGINE_GOOGLE,
		protocol.Engine_ENGINE_BRAVE,
		protocol.Engine_ENGINE_BING,
		protocol.Engine_ENGINE_DUCKDUCKGO,
	}

	// NEW: Worker now uses the token from config
	client := worker.NewClient(cfg.ID, version.AppVersion, cfg.MasterURL, engines, insecure.NewCredentials(), cfg.JoinToken)

	go func() {
		for {
			err := client.Run(ctx)
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					logger.Error("worker failed, reconnecting...", "error", err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}()

	<-ctx.Done()
	logger.Info("Worker node stopped.")
}
