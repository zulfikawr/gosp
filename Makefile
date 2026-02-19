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

## run: Run both Master and Worker in a single process (Dev Mode)
run: build
	@echo "Starting GOSP unified cluster..."
	./$(GOSP_BINARY) run

## stop: Stop the background GOSP process
stop: build
	@echo "Stopping GOSP background process..."
	./$(GOSP_BINARY) stop

## demo: Alias for 'run'
demo: run

## help: Show this help message
help:
	@echo "GOSP Unified CLI"
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //'
