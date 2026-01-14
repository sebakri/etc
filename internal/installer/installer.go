package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"etc/internal/config"
)

type Manager struct {
	RootDir string
	Env     map[string]string
}

func New(rootDir string, env map[string]string) *Manager {
	return &Manager{
		RootDir: rootDir,
		Env:     env,
	}
}

func (m *Manager) Install(tool config.Tool) error {
	etcDir := filepath.Join(m.RootDir, ".etc")
	binDir := filepath.Join(etcDir, "bin")

	// Ensure directories exist
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	switch tool.Type {
	case "go":
		return m.installGo(tool, binDir)
	case "npm":
		return m.installNpm(tool, etcDir)
	case "cargo":
		return m.installCargo(tool, etcDir)
	case "uv":
		return m.installUv(tool, binDir)
	default:
		return fmt.Errorf("unsupported tool type: %s", tool.Type)
	}
}

func (m *Manager) installGo(tool config.Tool, binDir string) error {
	fmt.Printf("Installing %s (go)...\n", tool.Name)

	cmd := exec.Command("go", "install", tool.Source)

	env := os.Environ()
	env = append(env, fmt.Sprintf("GOBIN=%s", binDir))
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) EnsureEnvrc() error {
	envrcPath := filepath.Join(m.RootDir, ".envrc")
	content := "PATH_add .etc/bin\n"

	for k, v := range m.Env {
		content += fmt.Sprintf("export %s=\"%s\"\n", k, v)
	}

	fmt.Println("Updating .envrc...")
	return os.WriteFile(envrcPath, []byte(content), 0644)
}

func (m *Manager) AllowDirenv() error {
	fmt.Println("Running direnv allow...")
	cmd := exec.Command("direnv", "allow")
	cmd.Dir = m.RootDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *Manager) installNpm(tool config.Tool, etcDir string) error {
	fmt.Printf("Installing %s (npm)...\n", tool.Name)

	// npm install --prefix .etc -g <package>
	// This installs binaries to .etc/bin on Linux/macOS
	cmd := exec.Command("npm", "install", "--prefix", etcDir, "-g", tool.Source)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) installCargo(tool config.Tool, etcDir string) error {
	fmt.Printf("Installing %s (cargo)...\n", tool.Name)

	// cargo binstall --root .etc <args> <package>
	// This installs binaries to .etc/bin
	args := []string{"binstall", "--root", etcDir, "-y"}
	args = append(args, tool.Args...)
	args = append(args, tool.Source)

	cmd := exec.Command("cargo", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) installUv(tool config.Tool, binDir string) error {
	fmt.Printf("Installing %s (uv)...\n", tool.Name)

	etcDir := filepath.Join(m.RootDir, ".etc")
	uvDir := filepath.Join(etcDir, "uv")

	// uv tool install --force <package>
	// UV_TOOL_BIN_DIR and UV_TOOL_DIR ensure project-local installation
	args := []string{"tool", "install", "--force"}
	args = append(args, tool.Args...)
	args = append(args, tool.Source)

	cmd := exec.Command("uv", args...)

	env := os.Environ()
	env = append(env, fmt.Sprintf("UV_TOOL_BIN_DIR=%s", binDir))
	env = append(env, fmt.Sprintf("UV_TOOL_DIR=%s", uvDir))
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
