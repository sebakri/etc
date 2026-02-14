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
	defer func() {
		_ = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				_ = os.Chmod(path, 0777)
			}
			return nil
		})
		_ = os.RemoveAll(tmpDir)
	}()

	m := New(tmpDir, nil)

	// Test case 1: Go tool with version (with 'v' prefix)
	tool := config.Tool{
		Type:    "go",
		Source:  config.Source{"github.com/go-task/task/v3/cmd/task"},
		Version: "v3.40.0",
	}

	err = m.Install(tool)
	if err != nil {
		t.Fatalf("Failed to install versioned go tool (with 'v'): %v", err)
	}

	binPath := filepath.Join(tmpDir, ".box", "bin", "task")
	if _, err := os.Stat(binPath); err != nil {
		t.Errorf("Expected binary at %s, but it does not exist: %v", binPath, err)
	}

	// Test case 2: Go tool with version (WITHOUT 'v' prefix - should fail with a hint)
	tool2 := config.Tool{
		Type:    "go",
		Source:  config.Source{"github.com/go-task/task/v3/cmd/task"},
		Version: "3.41.0",
	}

	err = m.Install(tool2)
	if err == nil {
		t.Errorf("Expected failure for version without 'v' prefix, but it succeeded")
	} else {
		t.Logf("Caught expected failure for version without 'v' prefix: %v", err)
	}
}
