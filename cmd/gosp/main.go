// Package main is the entry point for the GOSP (Go OpenSearchProtocol) application.
// It initializes and executes the CLI command hierarchy.
package main

import (
	"github.com/zulfikawr/gosp/internal/cli"
)

func main() {
	cli.Execute()
}
