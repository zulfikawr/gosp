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

	id := *workerID
	if id == "" {
		id = "worker-" + uuid.New().String()[:8]
	}

	var creds credentials.TransportCredentials
	if *useInsecure {
		creds = insecure.NewCredentials()
	} else {
		logger.Warn("mTLS not yet configured in CLI, falling back to insecure")
		creds = insecure.NewCredentials()
	}

	// Supported Engines (RE-INJECT DUCKDUCKGO)
	engines := []protocol.Engine{
		protocol.Engine_ENGINE_GOOGLE,
		protocol.Engine_ENGINE_BRAVE,
		protocol.Engine_ENGINE_BING,
		protocol.Engine_ENGINE_DUCKDUCKGO,
	}

	client := worker.NewClient(id, "v0.1.0", *masterAddr, engines, creds)
	
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting OSP worker node", "worker_id", id, "master", *masterAddr)

	for {
		err := client.Run(ctx)
		if err == context.Canceled {
			break
		}
		if err != nil {
			logger.Error("worker client failed", "error", err)
			time.Sleep(5 * time.Second)
			logger.Info("reconnecting to master...")
			continue
		}
		break
	}

	logger.Info("OSP worker node stopped.")
}
