package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of GOSP",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("GOSP v%s\n", version.AppVersion)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
