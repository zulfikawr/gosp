package cli

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/pid"
)

const defaultPidFile = "/tmp/gosp.pid"

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background GOSP process",
	Run: func(cmd *cobra.Command, args []string) {
		runStop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop() {
	p, err := pid.ReadPID(defaultPidFile)
	if err != nil {
		fmt.Printf("Error: GOSP does not appear to be running (no PID file at %s)\n", defaultPidFile)
		return
	}

	if !pid.IsRunning(p) {
		fmt.Printf("Warning: PID file exists but process %d is not running. Cleaning up...\n", p)
		pid.RemovePID(defaultPidFile)
		return
	}

	fmt.Printf("Stopping GOSP (PID %d) gracefully...\n", p)
	
	// Send SIGTERM to trigger graceful shutdown
	process, _ := os.FindProcess(p)
	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		fmt.Printf("Error: Failed to signal process: %v\n", err)
		return
	}

	// Wait for the process to exit and cleanup the PID file
	for i := 0; i < 10; i++ {
		if !pid.IsRunning(p) {
			pid.RemovePID(defaultPidFile)
			fmt.Println("GOSP stopped successfully.")
			return
		}
		time.Sleep(1 * time.Second)
	}

	fmt.Println("GOSP is taking too long to stop. You may need to kill it manually.")
}
