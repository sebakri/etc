package installer

import (
	"os"
	"os/exec"
	"strings"

	"github.com/sebakri/box/internal/config"
	"github.com/sebakri/box/internal/sandbox"
)

// Installer is the interface that all tool installers must implement.
type Installer interface {
	// Install installs the tool and returns a list of files it managed or created.
	Install(tool config.Tool, m *Manager, sandbox bool) ([]string, error)
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
func (m *Manager) runCommand(name string, args []string, env []string, dir string, useSandbox bool) error {
	cmdName := name
	cmdArgs := args

	// We create a dummy command just to pass to applySandbox which might modify its SysProcAttr
	//nolint:gosec
	tempCmd := exec.Command(name, args...)

	if useSandbox {
		cmdName, cmdArgs = sandbox.Apply(tempCmd, name, args, m.RootDir, m.TempDir)
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
