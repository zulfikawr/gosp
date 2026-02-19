package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/config"
)

var rootCmd = &cobra.Command{
	Use:   "gosp",
	Short: "GOSP (Go OpenSearchProtocol) - Distributed Search Orchestrator",
}

func Execute() {
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
