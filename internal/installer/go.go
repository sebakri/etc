package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sebakri/box/internal/config"
)

// GoInstaller implements the Installer interface for Go tools.
type GoInstaller struct{}

// Install installs a Go tool using 'go install'.
func (i *GoInstaller) Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error) {
	if tool.Version != "" && !strings.HasPrefix(tool.Version, "v") && len(tool.Version) > 0 && tool.Version[0] >= '0' && tool.Version[0] <= '9' {
		return nil, fmt.Errorf("go tools require a 'v' prefix for versions (e.g., v%s instead of %s)", tool.Version, tool.Version)
	}

	m.log("Installing %s (go)...", tool.DisplayName())

	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	goDir := filepath.Join(boxDir, "go")

	if err := os.MkdirAll(goDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create go dir: %w", err)
	}

	source := tool.Source.String()
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", source, tool.Version)
	}

	// Run go install with a persistent GOPATH in .box/go
	goBinDir := filepath.Join(goDir, "bin")

	newEnv := m.prepareGoEnv(goDir)
	if err := m.runCommand("go", []string{"install", source}, newEnv, "", sandbox); err != nil {
		return nil, err
	}

	binaries := tool.Binaries
	if len(binaries) == 0 {
		binaries = []string{m.detectBinaryName(source)}
	}

	return m.linkBinaries(goBinDir, binDir, binaries)
}
