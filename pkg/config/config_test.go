package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigFullFlow(t *testing.T) {
	// Setup temporary test directory
	tempDir, err := os.MkdirTemp("", "gosp-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Set the base directory override for the test
	SetBaseDir(tempDir)
	defer SetBaseDir("") // Reset after test

	// 1. Test EnsureDirs
	err = EnsureDirs()
	if err != nil {
		t.Fatalf("EnsureDirs failed: %v", err)
	}

	if _, err := os.Stat(GetMasterDir()); os.IsNotExist(err) {
		t.Error("Master directory was not created")
	}

	// 2. Test Master Lifecycle
	mCfg := &MasterConfig{
		Name:      "prod-master",
		HTTPPort:  "19000",
		GRPCPort:  "19004",
		JoinToken: "token-123",
	}

	err = SaveMaster(mCfg)
	if err != nil {
		t.Fatalf("SaveMaster failed: %v", err)
	}

	loadedMaster, err := LoadMaster("prod-master")
	if err != nil {
		t.Fatalf("LoadMaster failed: %v", err)
	}
	if loadedMaster.JoinToken != "token-123" {
		t.Errorf("Expected token-123, got %s", loadedMaster.JoinToken)
	}

	masters, _ := ListMasters()
	if len(masters) != 1 || masters[0] != "prod-master" {
		t.Errorf("Expected [prod-master], got %v", masters)
	}

	// 3. Test Worker Lifecycle
	wCfg := &WorkerConfig{
		ID:        "worker-42",
		MasterURL: "localhost:19004",
		Region:    "EU",
	}

	err = SaveWorker(wCfg)
	if err != nil {
		t.Fatalf("SaveWorker failed: %v", err)
	}

	loadedWorker, err := LoadWorker("worker-42")
	if err != nil {
		t.Fatalf("LoadWorker failed: %v", err)
	}
	if loadedWorker.Region != "EU" {
		t.Errorf("Expected EU, got %s", loadedWorker.Region)
	}

	workers, _ := ListWorkers()
	if len(workers) != 1 || workers[0] != "worker-42" {
		t.Errorf("Expected [worker-42], got %v", workers)
	}

	// 4. Test Deletion
	err = DeleteMaster("prod-master")
	if err != nil {
		t.Errorf("DeleteMaster failed: %v", err)
	}
	err = DeleteWorker("worker-42")
	if err != nil {
		t.Errorf("DeleteWorker failed: %v", err)
	}

	masters, _ = ListMasters()
	if len(masters) != 0 {
		t.Errorf("Expected 0 masters after deletion, got %d", len(masters))
	}
}

func TestGetPIDPath(t *testing.T) {
	path := GetPIDPath("master", "main")
	if !filepath.IsAbs(path) {
		t.Errorf("Expected absolute path, got %s", path)
	}
}
