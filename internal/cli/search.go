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
	"github.com/zulfikawr/gosp/pkg/config"
	"github.com/zulfikawr/gosp/pkg/models"
)

var (
	searchQuery    string
	searchEngine   string
	searchCount    int
	searchMetadata bool
	searchFormat   string
	searchProfile  string
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Perform a search query",
	Run: func(cmd *cobra.Command, args []string) {
		runSearch()
	},
}

func init() {
	searchCmd.Flags().StringVarP(&searchQuery, "query", "q", "", "Search query")
	searchCmd.Flags().StringVarP(&searchEngine, "engine", "e", "duckduckgo", "Search engine")
	searchCmd.Flags().IntVarP(&searchCount, "count", "c", 10, "Number of results")
	searchCmd.Flags().BoolVarP(&searchMetadata, "metadata", "m", false, "Show OSP metadata")
	searchCmd.Flags().StringVarP(&searchFormat, "format", "f", "table", "Output format (table, json)")
	searchCmd.Flags().StringVarP(&searchProfile, "profile", "p", "main", "Master profile to use")
	
	searchCmd.MarkFlagRequired("query")
	rootCmd.AddCommand(searchCmd)
}

func runSearch() {
	cfg, err := config.LoadMaster(searchProfile)
	if err != nil {
		fmt.Printf("Error: Master profile '%s' not found.\n", searchProfile)
		os.Exit(1)
	}

	params := url.Values{}
	params.Add("q", searchQuery)
	params.Add("engine", searchEngine)
	params.Add("count", strconv.Itoa(searchCount))
	if searchMetadata {
		params.Add("metadata", "true")
	}

	fullURL := fmt.Sprintf("http://localhost:%s/web/search?%s", cfg.HTTPPort, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		fmt.Printf("Error: Failed to connect to Master on port %s. Is it running?\n", cfg.HTTPPort)
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
	json.Unmarshal(body, &searchResp)

	fmt.Printf("\nGOSP Results for: %s\n", searchQuery)
	fmt.Println("--------------------------------------------------------------------------------")
	for i, res := range searchResp.Web.Results {
		fmt.Printf("%-2d | %-50s | %s\n", i+1, truncate(res.Title, 50), res.URL)
	}
	fmt.Println("--------------------------------------------------------------------------------\n")
}

func truncate(s string, n int) string {
	if len(s) <= n { return s }
	return s[:n] + "..."
}
