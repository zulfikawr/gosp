package pid

import (
	"os"
	"testing"
)

func TestPIDFlow(t *testing.T) {
	tempFile, err := os.CreateTemp("", "gosp-pid-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())
	tempFile.Close()

	// Test Write
	err = WritePID(tempFile.Name())
	if err != nil {
		t.Errorf("Failed to write PID: %v", err)
	}

	// Test Read
	pid, err := ReadPID(tempFile.Name())
	if err != nil {
		t.Errorf("Failed to read PID: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), pid)
	}

	// Test IsRunning
	if !IsRunning(pid) {
		t.Errorf("Expected PID %d to be running", pid)
	}

	// Test Remove
	err = RemovePID(tempFile.Name())
	if err != nil {
		t.Errorf("Failed to remove PID: %v", err)
	}
}

func TestIsRunningInvalid(t *testing.T) {
	// PID 0 is generally not a valid process for this check
	if IsRunning(0) {
		t.Error("Expected PID 0 to not be running")
	}
}
