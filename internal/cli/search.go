package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/zulfikawr/gosp/pkg/models"
)

var (
	searchQuery    string
	searchEngine   string
	searchCount    int
	searchMetadata bool
	searchFormat   string
	searchApiURL   string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Perform a search via a GOSP Master node",
	Run: func(cmd *cobra.Command, args []string) {
		runSearch()
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Search query")
	searchCmd.Flags().StringVarP(&searchEngine, "engine", "e", "duckduckgo", "Search engine (google, brave, duckduckgo)")
	searchCmd.Flags().IntVarP(&searchCount, "count", "c", 10, "Number of results")
	searchCmd.Flags().BoolVarP(&searchMetadata, "metadata", "m", false, "Show OSP metadata")
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, json)")
	searchCmd.Flags().StringVarP(&searchApiURL, "url", "u", "http://localhost:19000", "Master API URL")
	
	searchCmd.MarkFlagRequired("query")
	rootCmd.AddCommand(searchCmd)
}

func runSearch() {
	params := url.Values{}
	params.Add("q", searchQuery)
	params.Add("engine", searchEngine)
	params.Add("count", strconv.Itoa(searchCount))
	if searchMetadata {
		params.Add("metadata", "true")
	}

	fullURL := fmt.Sprintf("%s/web/search?%s", searchApiURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error: Failed to connect to Master: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error: Master returned status %d\nBody: %s\n", resp.StatusCode, string(body))
		os.Exit(1)
	}

	if searchFormat == "json" {
		fmt.Println(string(body))
		return
	}

	var searchResp models.SearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		fmt.Printf("Error: Failed to parse JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nGOSP Search Results for: %s\n", searchQuery)
	fmt.Printf("Latency: %dms | Total: %d\n", searchResp.Meta.LatencyMs, searchResp.Meta.Total)
	fmt.Println("--------------------------------------------------------------------------------")
	fmt.Printf("%-3s | %-50s | %s\n", "#", "Title", "URL")
	fmt.Println("--------------------------------------------------------------------------------")

	for i, res := range searchResp.Web.Results {
		fmt.Printf("%-3d | %-50s | %s\n", i+1, truncate(res.Title, 50), res.URL)
	}
	fmt.Println("--------------------------------------------------------------------------------")

	if searchMetadata && searchResp.Performance != nil {
		fmt.Println("\n--- OSP Metadata ---")
		fmt.Printf("Worker Scrape: %dms\n", searchResp.Performance.WorkerScrapeMs)
		fmt.Printf("Master Agg:    %dms\n", searchResp.Performance.MasterAggMs)
		fmt.Printf("Nodes Queried: %d\n", searchResp.Cluster.NodesQueried)
		fmt.Println()
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
