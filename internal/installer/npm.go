package installer

import (
	"fmt"
	"path/filepath"

	"github.com/sebakri/box/internal/config"
)

// NpmInstaller implements the Installer interface for NPM packages.
type NpmInstaller struct{}

// Install installs an NPM package.
func (i *NpmInstaller) Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error) {
	source := tool.Source.String()
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", source, tool.Version)
	}
	m.log("Installing %s (npm)...", tool.DisplayName())

	boxDir := filepath.Join(m.RootDir, ".box")
	npmDir := filepath.Join(boxDir, "npm")
	npmBinDir := filepath.Join(npmDir, "bin")
	binDir := filepath.Join(boxDir, "bin")

	// npm install --prefix .box/npm -g <package>
	if err := m.runCommand("npm", []string{"install", "--prefix", npmDir, "-g", source}, nil, "", sandbox); err != nil {
		return nil, err
	}

	binaries := tool.Binaries
	if len(binaries) == 0 {
		binaries = []string{m.detectBinaryName(source)}
	}

	return m.linkBinaries(npmBinDir, binDir, binaries)
}
