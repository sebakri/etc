package installer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sebakri/box/internal/config"
)

// UvInstaller implements the Installer interface for Python tools via uv.
type UvInstaller struct{}

// Install installs a Python tool using 'uv tool install'.
func (i *UvInstaller) Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error) {
	source := tool.Source.String()
	if tool.Version != "" {
		source = fmt.Sprintf("%s==%s", source, tool.Version)
	}
	m.log("Installing %s (uv)...", tool.DisplayName())

	boxDir := filepath.Join(m.RootDir, ".box")
	uvDir := filepath.Join(boxDir, "uv")
	uvBinDir := filepath.Join(uvDir, "bin")
	binDir := filepath.Join(boxDir, "bin")

	// uv tool install --force <package>
	// UV_TOOL_BIN_DIR and UV_TOOL_DIR ensure project-local installation
	args := []string{"tool", "install", "--force"}
	args = append(args, tool.Args...)
	args = append(args, source)

	env := os.Environ()
	env = append(env, fmt.Sprintf("UV_TOOL_BIN_DIR=%s", uvBinDir))
	env = append(env, fmt.Sprintf("UV_TOOL_DIR=%s", uvDir))

	if err := m.runCommand("uv", args, env, "", sandbox); err != nil {
		return nil, err
	}

	binaries := tool.Binaries
	if len(binaries) == 0 {
		binaries = []string{m.detectBinaryName(source)}
	}

	return m.linkBinaries(uvBinDir, binDir, binaries)
}
