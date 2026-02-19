package cli

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/pid"
)

func init() {
	masterCmd.AddCommand(&cobra.Command{
		Use:   "status [profile]",
		Short: "Show Master and cluster status",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			name := "main"
			if len(args) > 0 { name = args[0] }
			showMasterStatus(name)
		},
	})

	workerCmd.AddCommand(&cobra.Command{
		Use:   "status [profile]",
		Short: "Show local Worker status",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id := "local-01"
			if len(args) > 0 { id = args[0] }
			showWorkerStatus(id)
		},
	})
}

func showMasterStatus(name string) {
	cfg, err := config.LoadMaster(name)
	if err != nil {
		fmt.Printf("Error: Master profile '%s' not found.\n", name)
		return
	}

	pidPath := config.GetPIDPath("master", name)
	isRunning := false
	if p, err := pid.ReadPID(pidPath); err == nil {
		isRunning = pid.IsRunning(p)
	}

	fmt.Printf("GOSP MASTER: %s\n", name)
	fmt.Println("--------------------")
	statusStr := "OFFLINE"
	if isRunning { statusStr = "ONLINE" }
	fmt.Printf("Process:    %s\n", statusStr)
	fmt.Printf("HTTP Port:  %s\n", cfg.HTTPPort)
	fmt.Printf("gRPC Port:  %s\n", cfg.GRPCPort)

	if isRunning { 
		fetchClusterStatus(cfg.HTTPPort) 
	}
}

func fetchClusterStatus(port string) {
	resp, err := http.Get("http://localhost:" + port + "/cluster/status")
	if err != nil {
		fmt.Println("Error: Failed to fetch cluster metrics from API.")
		return
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&data)

	fmt.Println("CLUSTER OVERVIEW")
	fmt.Printf("Active Workers: %v\n", data["active_workers"])
	
	workers, ok := data["workers"].([]interface{})
	if ok && len(workers) > 0 {
		fmt.Printf("%-20s | %-15s | %-10s\n", "Worker ID", "Region", "Status")
		fmt.Println("---------------------------------------------------------")
		for _, w := range workers {
			node := w.(map[string]interface{})
			fmt.Printf("%-20v | %-15v | %-10s\n", node["ID"], node["Region"], "ACTIVE")
		}
	}
}

func showWorkerStatus(id string) {
	cfg, err := config.LoadWorker(id)
	if err != nil {
		fmt.Printf("Error: Worker profile '%s' not found.\n", id)
		return
	}

	pidPath := config.GetPIDPath("worker", id)
	isRunning := false
	if p, err := pid.ReadPID(pidPath); err == nil {
		isRunning = pid.IsRunning(p)
	}

	fmt.Printf("GOSP WORKER: %s\n", id)
	fmt.Println("--------------------")
	statusStr := "OFFLINE"
	if isRunning { statusStr = "ONLINE (Connected to "+cfg.MasterURL+")" }
	fmt.Printf("Status: %s\n", statusStr)
	fmt.Printf("Region: %s\n", cfg.Region)
}
