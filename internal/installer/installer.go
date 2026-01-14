package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"box/internal/config"
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
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	// Ensure directories exist
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	switch tool.Type {
	case "go":
		return m.installGo(tool, binDir)
	case "npm":
		return m.installNpm(tool, boxDir)
	case "cargo":
		return m.installCargo(tool, boxDir)
	case "uv":
		return m.installUv(tool, binDir)
	case "gem":
		return m.installGem(tool, binDir)
	case "script":
		return m.installScript(tool)
	default:
		return fmt.Errorf("unsupported tool type: %s", tool.Type)
	}
}

func (m *Manager) installGo(tool config.Tool, binDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, tool.Version)
	}
	fmt.Printf("Installing %s (go)...\n", source)

	cmd := exec.Command("go", "install", source)

	env := os.Environ()
	env = append(env, fmt.Sprintf("GOBIN=%s", binDir))
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) EnsureEnvrc() error {
	envrcPath := filepath.Join(m.RootDir, ".envrc")
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")
	content := fmt.Sprintf("export BOX_DIR=\"%s\"\n", boxDir)
	content += fmt.Sprintf("export BOX_BIN_DIR=\"%s\"\n", binDir)
	content += "PATH_add .box/bin\n"

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
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, tool.Version)
	}
	fmt.Printf("Installing %s (npm)...\n", source)

	// npm install --prefix .etc -g <package>
	cmd := exec.Command("npm", "install", "--prefix", etcDir, "-g", source)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) installCargo(tool config.Tool, etcDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, tool.Version)
	}
	fmt.Printf("Installing %s (cargo)...\n", source)

	// cargo binstall --root .etc <args> <package>
	args := []string{"binstall", "--root", etcDir, "-y"}
	args = append(args, tool.Args...)
	args = append(args, source)

	cmd := exec.Command("cargo", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) installUv(tool config.Tool, binDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s==%s", tool.Source, tool.Version)
	}
	fmt.Printf("Installing %s (uv)...\n", source)

	boxDir := filepath.Join(m.RootDir, ".box")
	uvDir := filepath.Join(boxDir, "uv")

	// uv tool install --force <package>
	// UV_TOOL_BIN_DIR and UV_TOOL_DIR ensure project-local installation
	args := []string{"tool", "install", "--force"}
	args = append(args, tool.Args...)
	args = append(args, source)

	cmd := exec.Command("uv", args...)

	env := os.Environ()
	env = append(env, fmt.Sprintf("UV_TOOL_BIN_DIR=%s", binDir))
	env = append(env, fmt.Sprintf("UV_TOOL_DIR=%s", uvDir))
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) installGem(tool config.Tool, binDir string) error {
	fmt.Printf("Installing %s %s (gem)...\n", tool.Source, tool.Version)

	boxDir := filepath.Join(m.RootDir, ".box")
	gemDir := filepath.Join(boxDir, "gems")

	// gem install --install-dir .box/gems --bindir .box/bin <gem>
	args := []string{"install", "--install-dir", gemDir, "--bindir", binDir, "--no-document"}
	if tool.Version != "" {
		args = append(args, "-v", tool.Version)
	}
	args = append(args, tool.Args...)
	args = append(args, tool.Source)

	cmd := exec.Command("gem", args...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (m *Manager) installScript(tool config.Tool) error {
	fmt.Printf("Installing via script: %s\n", tool.Source)

	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	cmd := exec.Command("sh", "-c", tool.Source)
	cmd.Dir = m.RootDir

	env := os.Environ()
	env = append(env, fmt.Sprintf("BOX_DIR=%s", boxDir))
	env = append(env, fmt.Sprintf("BOX_BIN_DIR=%s", binDir))
	env = append(env, fmt.Sprintf("PATH=%s%s%s", binDir, string(os.PathListSeparator), os.Getenv("PATH")))
	
	// Add project custom env vars
	for k, v := range m.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
