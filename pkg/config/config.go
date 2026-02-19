package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Role       string `json:"role"`        // master, worker, or both
	NodeID     string `json:"node_id"`     
	Region     string `json:"region"`      
	MasterURL  string `json:"master_url"`  // For workers
	HTTPPort   string `json:"http_port"`   // For master
	GRPCPort   string `json:"grpc_port"`   // For master
	JoinToken  string `json:"join_token,omitempty"`
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gosp", "config.json")
}

func Load() (*Config, error) {
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path := GetConfigPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
