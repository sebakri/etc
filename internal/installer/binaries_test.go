package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/sebakri/box/internal/config"
)

func TestBinariesField(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tmpDir, err := os.MkdirTemp("", "box-binaries-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = filepath.Walk(tmpDir, func(path string, _ os.FileInfo, err error) error {
			if err == nil {
				_ = os.Chmod(path, 0700)
			}
			return nil
		})
		_ = os.RemoveAll(tmpDir)
	}()

	m := New(tmpDir, "", nil, nil)

	t.Run("Go with explicit binaries", func(t *testing.T) {
		if _, err := exec.LookPath("go"); err != nil {
			t.Skip("go not found in PATH")
		}

		tool := config.Tool{
			Type:     "go",
			Source:   config.Source{"github.com/go-task/task/v3/cmd/task"},
			Version:  "v3.40.0",
			Binaries: []string{"task"},
		}

		if err := m.Install(tool); err != nil {
			t.Fatalf("Failed to install go tool: %v", err)
		}

		binPath := filepath.Join(tmpDir, ".box", "bin", "task")
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		if _, err := os.Stat(binPath); err != nil {
			t.Errorf("Expected binary at %s, but it does not exist: %v", binPath, err)
		}
	})

	t.Run("Script with explicit binaries", func(t *testing.T) {
		script := "echo 'hello' > $BOX_BIN_DIR/hello-world"
		binName := "hello-world"
		if runtime.GOOS == "windows" {
			script = "echo hello > %BOX_BIN_DIR%\\hello-world"
		}

		tool := config.Tool{
			Type:     "script",
			Source:   config.Source{script},
			Binaries: []string{binName},
		}

		if err := m.Install(tool); err != nil {
			t.Fatalf("Failed to install via script: %v", err)
		}

		binPath := filepath.Join(tmpDir, ".box", "bin", binName)
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		if _, err := os.Stat(binPath); err != nil {
			t.Errorf("Expected binary at %s, but it does not exist: %v", binPath, err)
		}
	})

	t.Run("NPM with explicit binaries", func(t *testing.T) {
		if _, err := exec.LookPath("npm"); err != nil {
			t.Skip("npm not found in PATH")
		}

		tool := config.Tool{
			Type:     "npm",
			Source:   config.Source{"cowsay"},
			Binaries: []string{"cowsay"},
		}

		if err := m.Install(tool); err != nil {
			t.Fatalf("Failed to install npm tool: %v", err)
		}

		binPath := filepath.Join(tmpDir, ".box", "bin", "cowsay")
		if runtime.GOOS == "windows" {
			binPath += ".exe"
		}
		if _, err := os.Stat(binPath); err != nil {
			t.Errorf("Expected binary at %s, but it does not exist: %v", binPath, err)
		}
	})

	t.Run("Script with missing explicit binary should fail", func(t *testing.T) {
		tool := config.Tool{
			Type:     "script",
			Source:   config.Source{"echo 'doing nothing'"},
			Binaries: []string{"missing-binary"},
		}

		if err := m.Install(tool); err == nil {
			t.Error("Expected error for missing binary, but got nil")
		}
	})
}
