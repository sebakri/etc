package installer

import (
	"os"
	"path/filepath"
	"testing"

	"box/internal/config"
)

func TestInstallGoWithVersion(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "box-install-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	m := New(tmpDir, nil)

	// Test case 1: Go tool with version (with 'v' prefix)
	tool := config.Tool{
		Name:    "task-v",
		Type:    "go",
		Source:  "github.com/go-task/task/v3/cmd/task",
		Version: "v3.40.0",
	}

	err = m.Install(tool)
	if err != nil {
		t.Fatalf("Failed to install versioned go tool (with 'v'): %v", err)
	}

	binPath := filepath.Join(tmpDir, ".box", "bin", "task-v")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("Expected binary at %s, but it does not exist: %v", binPath, err)
	}

	// Test case 2: Go tool with version (WITHOUT 'v' prefix - should be fixed to work)
	tool2 := config.Tool{
		Name:    "task-no-v",
		Type:    "go",
		Source:  "github.com/go-task/task/v3/cmd/task",
		Version: "3.41.0",
	}

	err = m.Install(tool2)
	if err != nil {
		t.Errorf("Failed to install versioned go tool (without 'v' prefix): %v", err)
	}

	binPath2 := filepath.Join(tmpDir, ".box", "bin", "task-no-v")
	if _, err := os.Stat(binPath2); err != nil {
		t.Errorf("Expected binary at %s, but it does not exist: %v", binPath2, err)
	}
}
