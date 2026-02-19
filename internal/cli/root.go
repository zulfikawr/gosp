package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gosp",
	Short: "GOSP (Go OpenSearchProtocol) - A distributed search protocol",
	Long: `GOSP is a high-performance, decentralized search protocol that provides 
a free alternative to standard search APIs by leveraging distributed worker nodes.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
}
