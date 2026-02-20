package version

import "testing"

func TestVersion(t *testing.T) {
	if AppVersion == "" {
		t.Error("AppVersion should not be empty")
	}
}
