package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

var (
	clusterApiURL string
)

var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Monitor the GOSP cluster health",
}

var clusterStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of workers and Master",
	Run: func(cmd *cobra.Command, args []string) {
		runClusterStatus()
	},
}

func init() {
	clusterCmd.PersistentFlags().StringVarP(&clusterApiURL, "url", "u", "http://localhost:19000", "Master API URL")
	clusterCmd.AddCommand(clusterStatusCmd)
	rootCmd.AddCommand(clusterCmd)
}

func runClusterStatus() {
	resp, err := http.Get(clusterApiURL + "/cluster/status")
	if err != nil {
		fmt.Printf("Error: Failed to connect to Master: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var status map[string]interface{}
	json.Unmarshal(body, &status)

	fmt.Println("\nGOSP CLUSTER STATUS")
	fmt.Println("-------------------")
	fmt.Printf("Active Workers: %v\n", status["active_workers"])
	fmt.Printf("Master Version: %v\n\n", status["version"])

	workers, ok := status["workers"].([]interface{})
	if ok && len(workers) > 0 {
		fmt.Printf("%-20s | %-15s | %-10s | %-15s | %s\n", "Worker ID", "Region", "CPU", "Memory", "Addr")
		fmt.Println("---------------------------------------------------------------------------------------------")
		for _, w := range workers {
			node := w.(map[string]interface{})
			fmt.Printf("%-20v | %-15v | %-10.1f%% | %-15.1f MB | %v\n", 
				node["ID"], 
				node["Region"], 
				node["CPUUsage"], 
				node["MemoryUsage"], 
				node["RemoteAddr"])
		}
	} else {
		fmt.Println("No workers currently connected.")
	}
	fmt.Println()
}
