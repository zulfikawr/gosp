package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MasterConfig struct {
	Name     string `json:"name"`
	HTTPPort string `json:"http_port"`
	GRPCPort string `json:"grpc_port"`
}

type WorkerConfig struct {
	ID        string `json:"id"`
	MasterURL string `json:"master_url"`
	Region    string `json:"region"`
	JoinToken string `json:"join_token,omitempty"`
}

func GetBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosp")
}

func GetMasterDir() string {
	return filepath.Join(GetBaseDir(), "masters")
}

func GetWorkerDir() string {
	return filepath.Join(GetBaseDir(), "workers")
}

// EnsureDirs creates the necessary config directories
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
func SaveMaster(cfg *MasterConfig) error {
	path := filepath.Join(GetMasterDir(), cfg.Name+".json")
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(path, data, 0644)
}

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
func SaveWorker(cfg *WorkerConfig) error {
	path := filepath.Join(GetWorkerDir(), cfg.ID+".json")
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(path, data, 0644)
}

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

func DeleteMaster(name string) error {
	return os.Remove(filepath.Join(GetMasterDir(), name+".json"))
}

func DeleteWorker(id string) error {
	return os.Remove(filepath.Join(GetWorkerDir(), id+".json"))
}

// PID Management per profile
func GetPIDPath(role, name string) string {
	return filepath.Join(GetBaseDir(), fmt.Sprintf("%s_%s.pid", role, name))
}
