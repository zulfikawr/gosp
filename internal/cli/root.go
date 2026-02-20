// Package cli provides the command-line interface for GOSP (Go OpenSearchProtocol).
// It defines all available commands including master, worker, search, and status operations.
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/version"
)

// rootCmd is the base command for the GOSP CLI application.
var rootCmd = &cobra.Command{
	Use:     "gosp",
	Short:   "GOSP (Go OpenSearchProtocol) - Distributed Search Orchestrator",
	Version: version.AppVersion,
}

// Execute runs the root command and handles any execution errors.
// This is the main entry point called from main.go.
func Execute() {
	rootCmd.Version = version.AppVersion
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Ensure system directories exist
	config.EnsureDirs()

	// Disable the default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
