package installer

import (
	"os"

	"os/exec"

	"path/filepath"

	"strings"

	"testing"

	"box/internal/config"
)

func TestIntegrationInstallation(t *testing.T) {

	if testing.Short() {

		t.Skip("skipping integration test in short mode")

	}

	// Build the latest box binary for testing

	boxBin := filepath.Join(t.TempDir(), "box")

	buildCmd := exec.Command("go", "build", "-o", boxBin, "../../main.go")

	if output, err := buildCmd.CombinedOutput(); err != nil {

		t.Fatalf("Failed to build box binary: %v\nOutput: %s", err, string(output))

	}

	// Create a temporary project directory

	projectDir := t.TempDir()

	// Use a dedicated GOPATH for the test to avoid caching issues
	testGoPath := filepath.Join(t.TempDir(), "gopath")
	os.Setenv("GOPATH", testGoPath)

	// Helper to fix permissions for cleanup

	t.Cleanup(func() {

		_ = filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {

			if err == nil {

				_ = os.Chmod(path, 0777)

			}

			return nil

		})

	})

	// Copy the integration test config to the project directory
	configSource, err := os.ReadFile("testdata/integration_test.yml")
	if err != nil {
		t.Fatalf("Failed to read integration test config: %v", err)
	}

	configPath := filepath.Join(projectDir, "box.yml")

	if err := os.WriteFile(configPath, configSource, 0644); err != nil {

		t.Fatalf("Failed to write box.yml: %v", err)

	}

	// Run box install

	installCmd := exec.Command(boxBin, "install", "--non-interactive")

	installCmd.Dir = projectDir

	if output, err := installCmd.CombinedOutput(); err != nil {

		t.Fatalf("box install failed: %v\nOutput: %s", err, string(output))

	}

	// Verify the tool was installed

	cfg, err := config.Load(configPath)

	if err != nil {

		t.Fatalf("Failed to load config: %v", err)

	}

	for _, tool := range cfg.Tools {
		binaryNames := tool.Binaries
		if len(binaryNames) == 0 {
			sourcePath := tool.Source.String()

			// Strip major version suffix (e.g. /v2, /v3) if it's the last part of the path
			if parts := strings.Split(sourcePath, "/"); len(parts) > 1 {
				lastPart := parts[len(parts)-1]
				if len(lastPart) >= 2 && lastPart[0] == 'v' && isDigit(lastPart[1:]) {
					sourcePath = strings.Join(parts[:len(parts)-1], "/")
				}
			}

			binaryName := sourcePath

			if idx := strings.LastIndex(binaryName, "/"); idx != -1 {

				binaryName = binaryName[idx+1:]

			}
			binaryNames = []string{binaryName}
		}

		for _, binaryName := range binaryNames {
			binPath := filepath.Join(projectDir, ".box", "bin", binaryName)

			if _, err := os.Stat(binPath); err != nil {

				t.Errorf("Expected binary for %s at %s, but not found", tool.Source, binPath)

			}

			// Verify version if it's task (as in our test file)

			if binaryName == "task" && tool.Version != "" {

				versionCmd := exec.Command(binPath, "--version")

				output, err := versionCmd.CombinedOutput()

				if err != nil {

					t.Errorf("Failed to run installed tool %s: %v", binaryName, err)

				}

				if !strings.Contains(string(output), tool.Version) {

					t.Errorf("Tool version mismatch. Expected %s, got output: %s", tool.Version, string(output))

				}

			}
		}

	}

}
