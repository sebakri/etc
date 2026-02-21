package installer

import (
	"fmt"
	"path/filepath"

	"github.com/sebakri/box/internal/config"
)

// CargoInstaller implements the Installer interface for Cargo crates.
type CargoInstaller struct{}

// Install installs a Cargo crate using 'cargo-binstall'.
func (i *CargoInstaller) Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error) {
	source := tool.Source.String()
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", source, tool.Version)
	}
	m.log("Installing %s (cargo)...", tool.DisplayName())

	boxDir := filepath.Join(m.RootDir, ".box")
	cargoDir := filepath.Join(boxDir, "cargo")
	cargoBinDir := filepath.Join(cargoDir, "bin")
	binDir := filepath.Join(boxDir, "bin")

	// cargo-binstall --root .box/cargo <args> <package>
	args := []string{"--root", cargoDir, "-y"}
	args = append(args, tool.Args...)
	args = append(args, source)

	if err := m.runCommand("cargo-binstall", args, nil, "", sandbox); err != nil {
		return nil, err
	}

	binaries := tool.Binaries
	if len(binaries) == 0 {
		binaries = []string{m.detectBinaryName(source)}
	}

	return m.linkBinaries(cargoBinDir, binDir, binaries)
}
