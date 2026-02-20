// Package pid provides process ID file management for GOSP services.
// It handles writing, reading, and validating PID files for daemon processes.
package pid

import (
	"os"
	"strconv"
	"syscall"
)

// WritePID saves the current process ID to the specified file.
func WritePID(path string) error {
	pid := os.Getpid()
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

// ReadPID reads the process ID from the specified file.
func ReadPID(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// RemovePID deletes the PID file.
func RemovePID(path string) error {
	return os.Remove(path)
}

// IsRunning checks if a process with the given PID is actually running.
func IsRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds. Send signal 0 to check if process exists.
	err = process.Signal(syscall.Signal(0))
	return err == nil
}
