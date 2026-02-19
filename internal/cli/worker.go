package cli

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/internal/worker"
	"github.com/zulfikawr/gosp/pkg/logger"
	"github.com/zulfikawr/gosp/pkg/protocol"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	workerMasterAddr string
	workerID         string
	workerInsecure   bool
	workerRegion     string
)

var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Start a GOSP Worker node",
	Run: func(cmd *cobra.Command, args []string) {
		runWorker()
	},
}

func init() {
	workerCmd.Flags().StringVar(&workerMasterAddr, "master", "localhost:19004", "OSP Master gRPC address")
	workerCmd.Flags().StringVar(&workerID, "id", "", "Unique worker ID")
	workerCmd.Flags().BoolVar(&workerInsecure, "insecure", true, "Use insecure connection")
	workerCmd.Flags().StringVar(&workerRegion, "region", "US-Cloud", "Geographic region of this worker")
	rootCmd.AddCommand(workerCmd)
}

func runWorker() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	runWorkerInternal(ctx)
}

func runWorkerInternal(ctx context.Context) {
	id := workerID
	if id == "" {
		id = "worker-" + uuid.New().String()[:8]
	}

	var creds credentials.TransportCredentials
	if workerInsecure {
		creds = insecure.NewCredentials()
	} else {
		logger.Warn("mTLS not yet configured in CLI, falling back to insecure")
		creds = insecure.NewCredentials()
	}

	engines := []protocol.Engine{
		protocol.Engine_ENGINE_GOOGLE,
		protocol.Engine_ENGINE_BRAVE,
		protocol.Engine_ENGINE_BING,
		protocol.Engine_ENGINE_DUCKDUCKGO,
	}

	client := worker.NewClient(id, "v0.1.0", workerMasterAddr, engines, creds)
	
	logger.Info("starting GOSP worker node", "worker_id", id, "master", workerMasterAddr, "region", workerRegion)

	// Reconnection Loop
	for {
		err := client.Run(ctx)
		if err != nil {
			// Check if context was canceled
			select {
			case <-ctx.Done():
				logger.Info("GOSP worker node stopped.")
				return
			default:
				logger.Error("worker client failed", "error", err)
				time.Sleep(5 * time.Second)
				logger.Info("reconnecting to master...")
				continue
			}
		}
		break
	}
}
