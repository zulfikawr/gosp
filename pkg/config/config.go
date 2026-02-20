// Package config provides configuration management for GOSP.
// It handles master and worker profile storage, loading, and directory management.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MasterConfig holds the configuration for a master node profile.
type MasterConfig struct {
	Name      string `json:"name"`
	HTTPPort  string `json:"http_port"`
	GRPCPort  string `json:"grpc_port"`
	JoinToken string `json:"join_token"`
}

// WorkerConfig holds the configuration for a worker node profile.
type WorkerConfig struct {
	ID        string `json:"id"`
	MasterURL string `json:"master_url"`
	Region    string `json:"region"`
	JoinToken string `json:"join_token,omitempty"`
}

var baseDirOverride string

// SetBaseDir sets an override for the configuration directory (primarily for testing).
func SetBaseDir(dir string) {
	baseDirOverride = dir
}

// GetBaseDir returns the base directory for GOSP configuration files.
func GetBaseDir() string {
	if baseDirOverride != "" {
		return baseDirOverride
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosp")
}

// GetMasterDir returns the directory for master configuration files.
func GetMasterDir() string {
	return filepath.Join(GetBaseDir(), "masters")
}

// GetWorkerDir returns the directory for worker configuration files.
func GetWorkerDir() string {
	return filepath.Join(GetBaseDir(), "workers")
}

// EnsureDirs creates the necessary config directories.
func EnsureDirs() error {
	dirs := []string{GetMasterDir(), GetWorkerDir()}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

// Master Profile Management

// SaveMaster persists a master configuration to disk.
func SaveMaster(cfg *MasterConfig) error {
	path := filepath.Join(GetMasterDir(), cfg.Name+".json")
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(path, data, 0644)
}

// LoadMaster loads a master configuration from disk by name.
func LoadMaster(name string) (*MasterConfig, error) {
	path := filepath.Join(GetMasterDir(), name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg MasterConfig
	json.Unmarshal(data, &cfg)
	return &cfg, nil
}

// ListMasters returns a list of all saved master profile names.
func ListMasters() ([]string, error) {
	entries, err := os.ReadDir(GetMasterDir())
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			names = append(names, e.Name()[:len(e.Name())-5])
		}
	}
	return names, nil
}

// Worker Profile Management

// SaveWorker persists a worker configuration to disk.
func SaveWorker(cfg *WorkerConfig) error {
	path := filepath.Join(GetWorkerDir(), cfg.ID+".json")
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(path, data, 0644)
}

// LoadWorker loads a worker configuration from disk by ID.
func LoadWorker(id string) (*WorkerConfig, error) {
	path := filepath.Join(GetWorkerDir(), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg WorkerConfig
	json.Unmarshal(data, &cfg)
	return &cfg, nil
}

// ListWorkers returns a list of all saved worker profile IDs.
func ListWorkers() ([]string, error) {
	entries, err := os.ReadDir(GetWorkerDir())
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			ids = append(ids, e.Name()[:len(e.Name())-5])
		}
	}
	return ids, nil
}

// DeleteMaster removes a master profile from disk.
func DeleteMaster(name string) error {
	return os.Remove(filepath.Join(GetMasterDir(), name+".json"))
}

// DeleteWorker removes a worker profile from disk.
func DeleteWorker(id string) error {
	return os.Remove(filepath.Join(GetWorkerDir(), id+".json"))
}

// PID Management per profile

// GetPIDPath returns the path to the PID file for a given role and name.
func GetPIDPath(role, name string) string {
	return filepath.Join(GetBaseDir(), fmt.Sprintf("%s_%s.pid", role, name))
}
