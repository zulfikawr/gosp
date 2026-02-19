package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/logger"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run both Master and a Worker node in a single process (Dev Mode)",
	Long: `This command launches both the GOSP Master and a local Worker node 
concurrently. Useful for local development and testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		startGospAll()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func startGospAll() {
	logger.Info("Starting GOSP in unified mode (Master + Worker)...")

	// 1. Setup global context for cancellation
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Start Master in a goroutine
	// Note: We use the default port settings defined in master.go
	go func() {
		logger.Info("Initializing Master node...")
		// We can't call runMaster() directly because it has its own signal handler.
		// For the 'run' command, we'll implement a simplified version or refactor.
		runMasterInternal(ctx)
	}()

	// 3. Wait for Master to initialize
	time.Sleep(2 * time.Second)

	// 4. Start Worker in a goroutine
	go func() {
		logger.Info("Initializing local Worker node...")
		runWorkerInternal(ctx)
	}()

	// 5. Block until signal
	<-ctx.Done()
	logger.Info("Shutting down GOSP unified cluster...")
	
	// Allow some time for graceful shutdown
	time.Sleep(1 * time.Second)
	logger.Info("GOSP stopped.")
}
