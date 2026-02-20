package master

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zulfikawr/gosp/pkg/models"
	"github.com/zulfikawr/gosp/pkg/protocol"
)

func TestSearchAPI_EndToEnd(t *testing.T) {
	// 1. Setup Master Components
	reg := NewRegistry(10 * time.Second)
	sched := NewRoundRobinScheduler(reg)
	disp := NewDispatcher(sched, reg)
	aggr := NewResultAggregator()
	server := NewHTTPServer(disp, aggr)

	// 2. Register a "Fake" Worker
	workerID := "test-worker-01"
	reg.Register(&protocol.RegisterRequest{
		WorkerId:         workerID,
		SupportedEngines: []protocol.Engine{protocol.Engine_ENGINE_GOOGLE, protocol.Engine_ENGINE_DUCKDUCKGO},
		Version:          "v0.1.0",
		Region:           "TEST-LOC",
	}, "127.0.0.1")

	worker := reg.GetWorker(workerID)
	if worker == nil {
		t.Fatal("Worker registration failed")
	}

	// 3. Mock Worker Task Processing Loop
	go func() {
		// Listen for task assignment
		cmd := <-worker.CommandChan
		task := cmd.GetTask()
		if task == nil {
			return
		}

		// Simulate scraping and send response back to dispatcher
		disp.HandleResponse(&protocol.SearchResponse{
			TaskId: task.TaskId,
			Results: []*protocol.ResultItem{
				{
					Title:        "GOSP Test Result",
					Url:          "https://gosp.test",
					Description:  "This is a simulated result for integration testing.",
					SourceEngine: protocol.Engine_ENGINE_DUCKDUCKGO,
					WorkerId:     workerID,
				},
			},
			ScrapeLatencyMs: 100,
		})
	}()

	// 4. Perform HTTP Request to Search API
	// We use /res/v1/web/search to test the Brave compatibility route
	req := httptest.NewRequest("GET", "/res/v1/web/search?q=golang&engine=duckduckgo", nil)
	resp, err := server.app.Test(req, 5000) // 5s timeout
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// 5. Verify JSON Response Schema (Brave Parity)
	var searchResp models.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		t.Fatalf("Failed to decode JSON response: %v", err)
	}

	if searchResp.Type != "search" {
		t.Errorf("Expected type 'search', got %s", searchResp.Type)
	}

	if len(searchResp.Web.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(searchResp.Web.Results))
	}

	result := searchResp.Web.Results[0]
	if result.Title != "GOSP Test Result" {
		t.Errorf("Expected title 'GOSP Test Result', got %s", result.Title)
	}

	if result.URL != "https://gosp.test" {
		t.Errorf("Expected URL https://gosp.test, got %s", result.URL)
	}
}

func TestClusterStatusAPI(t *testing.T) {
	// Create registry with a 1-hour timeout for testing so workers don't get pruned
	reg := NewRegistry(1 * time.Hour)
	sched := NewRoundRobinScheduler(reg)
	disp := NewDispatcher(sched, reg)
	aggr := NewResultAggregator()
	server := NewHTTPServer(disp, aggr)

	// Register a worker
	reg.Register(&protocol.RegisterRequest{
		WorkerId: "status-worker",
	}, "127.0.0.1")

	req := httptest.NewRequest("GET", "/cluster/status", nil)
	resp, err := server.app.Test(req)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != 200 {
		// Read body to see error message
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		t.Errorf("Expected status 200, got %d. Body: %v", resp.StatusCode, body)
	}

	var status map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&status)

	if val, ok := status["active_workers"].(float64); !ok || val != 1 {
		t.Errorf("Expected 1 active worker, got %v", status["active_workers"])
	}
}
