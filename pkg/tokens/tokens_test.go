package tokens

import (
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	token, err := Generate()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	parts := strings.Split(token, "-")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts in token, got %d", len(parts))
	}

	for _, part := range parts {
		found := false
		for _, word := range words {
			if word == part {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Token part '%s' not found in words list", part)
		}
	}
}

func TestGenerateUniqueness(t *testing.T) {
	// Generating 100 tokens to check for duplicates (low chance)
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, _ := Generate()
		if seen[token] {
			t.Errorf("Collision detected: %s was generated twice in 100 runs", token)
		}
		seen[token] = true
	}
}
