package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/sebakri/box/internal/config"
)

// ScriptInstaller implements the Installer interface for shell scripts.
type ScriptInstaller struct{}

// Install installs a tool by running a shell script.
func (i *ScriptInstaller) Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error) {
	m.log("Installing via script: %s", tool.DisplayName())

	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	env := os.Environ()
	env = append(env, fmt.Sprintf("BOX_DIR=%s", boxDir))
	env = append(env, fmt.Sprintf("BOX_BIN_DIR=%s", binDir))
	env = append(env, fmt.Sprintf("BOX_OS=%s", runtime.GOOS))
	env = append(env, fmt.Sprintf("BOX_ARCH=%s", runtime.GOARCH))
	env = append(env, fmt.Sprintf("PATH=%s%s%s", binDir, string(os.PathListSeparator), os.Getenv("PATH")))

	if m.TempDir != "" {
		env = append(env, fmt.Sprintf("TMPDIR=%s", m.TempDir))
		env = append(env, fmt.Sprintf("TEMP=%s", m.TempDir))
		env = append(env, fmt.Sprintf("TMP=%s", m.TempDir))
	}

	// Add project custom env vars
	for k, v := range m.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	if err := m.runCommand("sh", []string{"-c", tool.Source.String()}, env, m.RootDir, sandbox); err != nil {
		return nil, err
	}

	// If explicit binaries are specified for a script, we verify they exist in binDir.
	// We don't link them because the script is expected to have put them there (e.g. using $BOX_BIN_DIR).
	createdFiles := []string{}
	for _, name := range tool.Binaries {
		binaryPath := filepath.Join(binDir, name)
		if runtime.GOOS == "windows" && !strings.HasSuffix(binaryPath, ".exe") {
			binaryPath += ".exe"
		}
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("script installation finished but binary %s not found in %s", name, binDir)
		}
		relToRoot, _ := filepath.Rel(m.RootDir, binaryPath)
		createdFiles = append(createdFiles, relToRoot)
	}

	return createdFiles, nil
}
