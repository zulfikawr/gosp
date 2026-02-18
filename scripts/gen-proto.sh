#!/bin/bash

# Ensure we're in the project root
cd "$(dirname "$0")/.."

# Check if protoc is installed
if ! command -v protoc &> /dev/null
then
    echo "protoc could not be found. Please install it with: sudo apt install protobuf-compiler"
    exit 1
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null
then
    echo "protoc-gen-go could not be found. Installing..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check if protoc-gen-go-grpc is installed
if ! command -v protoc-gen-go-grpc &> /dev/null
then
    echo "protoc-gen-go-grpc could not be found. Installing..."
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
fi

# Create pkg/protocol if it doesn't exist
mkdir -p pkg/protocol

# Generate Go code
protoc --proto_path=proto \
       --go_out=pkg/protocol --go_opt=paths=source_relative \
       --go-grpc_out=pkg/protocol --go-grpc_opt=paths=source_relative \
       proto/search.proto

echo "Protobuf generation complete. Files are in pkg/protocol."
