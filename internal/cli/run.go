package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/pid"
)

var (
	runDaemon bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run both Master and a Worker node in a single process (Dev Mode)",
	Long: `This command launches both the GOSP Master and a local Worker node 
concurrently. Useful for local development and testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		if runDaemon {
			startDaemon()
			return
		}
		startGospAll()
	},
}

func init() {
	runCmd.Flags().BoolVarP(&runDaemon, "daemon", "d", false, "Run GOSP in the background")
	rootCmd.AddCommand(runCmd)
}

func startDaemon() {
	// Re-run the same command but without the --daemon flag and in the background
	args := []string{"run"}
	// Pass other relevant flags if needed
	
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error: Failed to start daemon: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("GOSP started in background (PID: %d)\n", cmd.Process.Pid)
	fmt.Println("Use 'gosp stop' to terminate the process.")
	os.Exit(0)
}

func startGospAll() {
	// 1. Write PID file so 'gosp stop' can find us
	if err := pid.WritePID(defaultPidFile); err != nil {
		logger.Warn("failed to write PID file", "error", err)
	}
	defer pid.RemovePID(defaultPidFile)

	logger.Info("Starting GOSP in unified mode (Master + Worker)...")

	// 2. Setup global context for cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 3. Start Master in a goroutine
	go func() {
		logger.Info("Initializing Master node...")
		runMasterInternal(ctx)
	}()

	// 4. Wait for Master to initialize
	time.Sleep(2 * time.Second)

	// 5. Start Worker in a goroutine
	go func() {
		logger.Info("Initializing local Worker node...")
		runWorkerInternal(ctx)
	}()

	// 6. Block until signal
	<-ctx.Done()
	logger.Info("Shutting down GOSP unified cluster...")
	
	// Allow some time for graceful shutdown
	time.Sleep(1 * time.Second)
	logger.Info("GOSP stopped.")
}
