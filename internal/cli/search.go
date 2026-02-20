package cli

import (
	"bytes"
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
		fmt.Printf("❌ Error: Could not reach the GOSP Master on port %s.\n", cfg.HTTPPort)
		fmt.Println("👉 Solution: Is the brain running? Try starting it with: 'gosp master run'")
		os.Exit(1)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		var errData map[string]string
		json.Unmarshal(body, &errData)

		fmt.Printf("❌ Error: The Master is online, but the search failed: %s\n", errData["error"])
		if resp.StatusCode == 503 {
			fmt.Println("👉 Solution: A brain needs hands! Ensure a worker is connected with: 'gosp worker run'")
		}
		os.Exit(1)
	}

	if searchFormat == "json" {
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "", "  "); err != nil {
			fmt.Println(string(body))
		} else {
			fmt.Println(prettyJSON.String())
		}
		return
	}

	var searchResp models.SearchResponse
	json.Unmarshal(body, &searchResp)

	fmt.Printf("GOSP Results for: %s\n", searchQuery)
	separator := "----------------------------------------------------------------------------------------------------"
	fmt.Println(separator)
	fmt.Printf("%-3s | %-50s | %s\n", "#", "Title", "URL")
	fmt.Println(separator)
	for i, res := range searchResp.Web.Results {
		title := truncate(res.Title, 50)
		fmt.Printf("%-3d | %-50s | %s\n", i+1, title, res.URL)
	}
	fmt.Println(separator)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-3]) + "..."
}
