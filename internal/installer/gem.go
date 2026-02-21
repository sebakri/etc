package installer

import (
	"path/filepath"

	"github.com/sebakri/box/internal/config"
)

// GemInstaller implements the Installer interface for Ruby gems.
type GemInstaller struct{}

// Install installs a Ruby gem.
func (i *GemInstaller) Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error) {
	m.log("Installing %s %s (gem)...", tool.DisplayName(), tool.Version)

	boxDir := filepath.Join(m.RootDir, ".box")
	gemDir := filepath.Join(boxDir, "gems")
	gemBinDir := filepath.Join(gemDir, "bin")
	binDir := filepath.Join(boxDir, "bin")

	// gem install --install-dir .box/gems --bindir .box/gems/bin <gem>
	args := []string{"install", "--install-dir", gemDir, "--bindir", gemBinDir, "--no-document"}
	if tool.Version != "" {
		args = append(args, "-v", tool.Version)
	}
	args = append(args, tool.Args...)
	args = append(args, tool.Source.String())

	if err := m.runCommand("gem", args, nil, "", sandbox); err != nil {
		return nil, err
	}

	binaries := tool.Binaries
	if len(binaries) == 0 {
		binaries = []string{m.detectBinaryName(tool.Source.String())}
	}

	return m.linkBinaries(gemBinDir, binDir, binaries)
}
