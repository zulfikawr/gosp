# GOSP (Go OpenSearchProtocol) Makefile

# Variables
BINARY_DIR=bin
GOSP_BINARY=$(BINARY_DIR)/gosp
PROTO_DIR=proto
PROTO_GEN_DIR=pkg/protocol
LOG_DIR=logs

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod

.PHONY: all build proto clean test run-master run-worker demo help

all: proto build

## build: Build the unified GOSP binary
build:
	@echo "Building unified GOSP binary..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(GOSP_BINARY) cmd/gosp/main.go
	@echo "Done. Binary is at $(GOSP_BINARY)"

## proto: Generate Go code from Protobuf definitions
proto:
	@echo "Generating Protobuf code..."
	@./scripts/gen-proto.sh

## clean: Remove binaries and logs
clean:
	@echo "Cleaning up..."
	@rm -rf $(BINARY_DIR)
	@rm -rf $(LOG_DIR)
	$(GOCLEAN)
	@echo "Cleanup complete."

## test: Run all Go unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

## run-master: Run the Master node locally
run-master: build
	@echo "Starting GOSP Master..."
	./$(GOSP_BINARY) master --http :19000 --grpc :19004

## run-worker: Run a Worker node locally
run-worker: build
	@echo "Starting GOSP Worker..."
	./$(GOSP_BINARY) worker --master localhost:19004 --id "local-worker-01"

## demo: Run the integrated demo cluster
demo: build
	@echo "Launching GOSP Live Demo..."
	@./scripts/run-demo.sh

## help: Show this help message
help:
	@echo "GOSP Unified CLI"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //'
