package installer

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sebakri/box/internal/config"
)

// Installer is the interface that all tool installers must implement.
type Installer interface {
	// Install installs the tool.
	Install(tool config.Tool, m *Manager, sandbox bool) error
}

// ToolType represents a supported tool runtime.
type ToolType struct {
	Name      string
	Installer Installer
}

// SupportedTools is the central registry of all tool types box can handle.
var SupportedTools = map[string]ToolType{
	"go":     {Name: "go", Installer: &GoInstaller{}},
	"npm":    {Name: "npm", Installer: &NpmInstaller{}},
	"cargo":  {Name: "cargo-binstall", Installer: &CargoInstaller{}},
	"uv":     {Name: "uv", Installer: &UvInstaller{}},
	"gem":    {Name: "gem", Installer: &GemInstaller{}},
	"script": {Name: "sh", Installer: &ScriptInstaller{}},
}

// runCommand is a helper to run shell commands with consistent output redirection and environment setup.
func (m *Manager) runCommand(name string, args []string, env []string, dir string, sandbox bool) error {
	cmdName := name
	cmdArgs := args

	// We create a dummy command just to pass to applySandbox which might modify its SysProcAttr
	//nolint:gosec
	tempCmd := exec.Command(name, args...)

	if sandbox {
		cmdName, cmdArgs = applySandbox(tempCmd, name, args, m.RootDir, m.TempDir)
	}

	//nolint:gosec
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.SysProcAttr = tempCmd.SysProcAttr // Transfer modified SysProcAttr
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = m.Output
	cmd.Stderr = m.Output

	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}

	m.log("Running: %s %s", cmdName, strings.Join(cmdArgs, " "))
	return cmd.Run()
}

// GoInstaller implements the Installer interface for Go tools.
type GoInstaller struct{}

// Install installs a Go tool using 'go install'.
func (i *GoInstaller) Install(tool config.Tool, m *Manager, sandbox bool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	return m.installGo(tool, binDir, sandbox)
}

// NpmInstaller implements the Installer interface for NPM packages.
type NpmInstaller struct{}

// Install installs an NPM package.
func (i *NpmInstaller) Install(tool config.Tool, m *Manager, sandbox bool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	return m.installNpm(tool, binDir, sandbox)
}

// CargoInstaller implements the Installer interface for Cargo crates.
type CargoInstaller struct{}

// Install installs a Cargo crate using 'cargo-binstall'.
func (i *CargoInstaller) Install(tool config.Tool, m *Manager, sandbox bool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	return m.installCargo(tool, binDir, sandbox)
}

// UvInstaller implements the Installer interface for Python tools via uv.
type UvInstaller struct{}

// Install installs a Python tool using 'uv tool install'.
func (i *UvInstaller) Install(tool config.Tool, m *Manager, sandbox bool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	return m.installUv(tool, binDir, sandbox)
}

// GemInstaller implements the Installer interface for Ruby gems.
type GemInstaller struct{}

// Install installs a Ruby gem.
func (i *GemInstaller) Install(tool config.Tool, m *Manager, sandbox bool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	return m.installGem(tool, binDir, sandbox)
}

// ScriptInstaller implements the Installer interface for shell scripts.
type ScriptInstaller struct{}

// Install installs a tool by running a shell script.
func (i *ScriptInstaller) Install(tool config.Tool, m *Manager, sandbox bool) error {
	return m.installScript(tool, sandbox)
}
