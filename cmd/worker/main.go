package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/zulfikawr/go-search/internal/worker"
	"github.com/zulfikawr/go-search/pkg/logger"
	"github.com/zulfikawr/go-search/pkg/protocol"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	masterAddr   = flag.String("master", "localhost:50051", "OSP Master gRPC address")
	workerID     = flag.String("id", "", "Unique worker ID (optional, will generate if empty)")
	useInsecure  = flag.Bool("insecure", true, "Use insecure gRPC connection (disable for production mTLS)")
)

func main() {
	flag.Parse()

	// 1. Generate ID if not provided
	id := *workerID
	if id == "" {
		id = "worker-" + uuid.New().String()[:8]
	}

	// 2. Setup Credentials
	var creds credentials.TransportCredentials
	if *useInsecure {
		creds = insecure.NewCredentials()
	} else {
		// Future: load mTLS certs from pkg/protocol
		logger.Warn("mTLS not yet configured in CLI, falling back to insecure")
		creds = insecure.NewCredentials()
	}

	// 3. Define Supported Engines
	engines := []protocol.Engine{
		protocol.Engine_ENGINE_GOOGLE,
		protocol.Engine_ENGINE_BRAVE,
		protocol.Engine_ENGINE_BING,
	}

	// 4. Initialize and Run Worker Client
	client := worker.NewClient(id, "v0.1.0", *masterAddr, engines, creds)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting OSP worker node", "worker_id", id, "master", *masterAddr)

	// Reconnection Loop
	for {
		err := client.Run(ctx)
		if err == context.Canceled {
			break
		}
		if err != nil {
			logger.Error("worker client failed", "error", err)
			time.Sleep(5 * time.Second) // Backoff before reconnect
			logger.Info("reconnecting to master...")
			continue
		}
		break
	}

	logger.Info("OSP worker node stopped.")
}
