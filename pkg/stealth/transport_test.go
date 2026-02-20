package stealth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewStealthClient(t *testing.T) {
	client := NewStealthClient(5 * time.Second)
	if client == nil {
		t.Fatal("Expected client, got nil")
	}

	if client.Timeout != 5*time.Second {
		t.Errorf("Expected timeout 5s, got %v", client.Timeout)
	}
}

func TestStealthClient_Connection(t *testing.T) {
	// Create a local test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	client := NewStealthClient(2 * time.Second)

	// Test standard HTTP connection (DialContext path)
	resp, err := client.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestGetRandomUserAgent(t *testing.T) {
	ua := GetRandomUserAgent()
	if ua == "" {
		t.Error("Expected non-empty User-Agent")
	}
}
