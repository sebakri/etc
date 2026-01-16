package installer

import (
	"encoding/gob"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"box/internal/config"
)

type Manager struct {
	RootDir string
	Env     map[string]string
	Output  io.Writer
}

type ToolManifest struct {
	Files []string
}

type Manifest struct {
	Tools map[string]ToolManifest
}

func New(rootDir string, env map[string]string) *Manager {
	return &Manager{
		RootDir: rootDir,
		Env:     env,
		Output:  os.Stdout,
	}
}

func (m *Manager) log(format string, a ...any) {
	if m.Output != nil {
		fmt.Fprintf(m.Output, format+"\n", a...)
	}
}

func (m *Manager) Install(tool config.Tool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	// Ensure directories exist
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	// Capture state before install

	before, err := m.captureState()
	if err != nil {
		return fmt.Errorf("failed to capture state before install: %w", err)
	}

	var installErr error
	switch tool.Type {
	case "go":
		installErr = m.installGo(tool, binDir)
	case "npm":
		installErr = m.installNpm(tool, boxDir)
	case "cargo":
		installErr = m.installCargo(tool, boxDir)
	case "uv":
		installErr = m.installUv(tool, binDir)
	case "gem":
		installErr = m.installGem(tool, binDir)
	case "script":
		installErr = m.installScript(tool)
	default:
		return fmt.Errorf("unsupported tool type: %s", tool.Type)
	}

	if installErr != nil {
		return installErr
	}

	// Capture state after install and find new files
	after, err := m.captureState()
	if err != nil {
		return fmt.Errorf("failed to capture state after install: %w", err)
	}

	newFiles := []string{}
	for path := range after {
		if _, ok := before[path]; !ok {
			newFiles = append(newFiles, path)
		}
	}
	sort.Strings(newFiles)

	return m.updateManifest(tool.Source, newFiles)
}

func (m *Manager) captureState() (map[string]bool, error) {
	state := make(map[string]bool)
	boxDir := filepath.Join(m.RootDir, ".box")

	if _, err := os.Stat(boxDir); os.IsNotExist(err) {
		return state, nil
	}

	err := filepath.Walk(boxDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// We track the relative path from RootDir
		rel, err := filepath.Rel(m.RootDir, path)
		if err != nil {
			return err
		}
		state[rel] = true
		return nil
	})

	return state, err
}

func (m *Manager) updateManifest(name string, files []string) error {
	manifestPath := filepath.Join(m.RootDir, ".box", "manifest.bin")
	manifest := Manifest{Tools: make(map[string]ToolManifest)}

	if file, err := os.Open(manifestPath); err == nil {
		_ = gob.NewDecoder(file).Decode(&manifest)
		file.Close()
	}

	manifest.Tools[name] = ToolManifest{Files: files}

	file, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return gob.NewEncoder(file).Encode(manifest)
}

func (m *Manager) LoadManifest() (*Manifest, error) {
	manifestPath := filepath.Join(m.RootDir, ".box", "manifest.bin")
	manifest := Manifest{Tools: make(map[string]ToolManifest)}

	file, err := os.Open(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &manifest, nil
		}
		return nil, err
	}
	defer file.Close()

	err = gob.NewDecoder(file).Decode(&manifest)
	return &manifest, err
}

func (m *Manager) Uninstall(name string) error {
	manifestPath := filepath.Join(m.RootDir, ".box", "manifest.bin")
	manifest := Manifest{Tools: make(map[string]ToolManifest)}

	file, err := os.Open(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Fallback for old installations
			return m.uninstallBestEffort(name)
		}
		return err
	}
	err = gob.NewDecoder(file).Decode(&manifest)
	file.Close()
	if err != nil {
		return err
	}

	toolInfo, ok := manifest.Tools[name]
	if !ok {
		return m.uninstallBestEffort(name)
	}

	// Remove files in reverse order
	sort.Sort(sort.Reverse(sort.StringSlice(toolInfo.Files)))

	for _, file := range toolInfo.Files {
		fullPath := filepath.Join(m.RootDir, file)
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			entries, _ := os.ReadDir(fullPath)
			if len(entries) == 0 {
				m.log("Removing empty directory %s...", file)
				os.Remove(fullPath)
			}
		} else {
			m.log("Removing file %s...", file)
			os.Remove(fullPath)
		}
	}

	delete(manifest.Tools, name)

	outFile, err := os.Create(manifestPath)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return gob.NewEncoder(outFile).Encode(manifest)
}

func (m *Manager) uninstallBestEffort(name string) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	binaryPath := filepath.Join(binDir, name)
	if _, err := os.Stat(binaryPath); err == nil {
		m.log("Removing binary %s...", binaryPath)
		os.Remove(binaryPath)
	}

	uvToolDir := filepath.Join(boxDir, "uv", name)
	if _, err := os.Stat(uvToolDir); err == nil {
		m.log("Removing data directory %s...", uvToolDir)
		os.RemoveAll(uvToolDir)
	}

	return nil
}

func (m *Manager) installGo(tool config.Tool, binDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, tool.Version)
	}
	m.log("Installing %s (go)...", tool.Source)

	err := m.runGoInstall(source, binDir)
	if err != nil {
		if tool.Version != "" && !strings.HasPrefix(tool.Version, "v") && len(tool.Version) > 0 && tool.Version[0] >= '0' && tool.Version[0] <= '9' {
			m.log("Hint: Go tools often require a 'v' prefix for versions (e.g., v%s instead of %s)", tool.Version, tool.Version)
		}
		return err
	}

	return nil
}

func (m *Manager) runGoInstall(source string, binDir string) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	goDir := filepath.Join(boxDir, "go")
	if err := os.MkdirAll(goDir, 0755); err != nil {
		return fmt.Errorf("failed to create go dir: %w", err)
	}

	// Run go install with a persistent GOPATH in .box/go
	cmd := exec.Command("go", "install", source)
	env := os.Environ()

	// Explicitly set GOBIN to .box/go/bin to ensure we know where it lands
	goBinDir := filepath.Join(goDir, "bin")

	// Filter out GOBIN and GOPATH from existing env to ensure ours take precedence cleanly
	newEnv := []string{}
	for _, e := range env {
		if !strings.HasPrefix(e, "GOBIN=") && !strings.HasPrefix(e, "GOPATH=") {
			newEnv = append(newEnv, e)
		}
	}

	newEnv = append(newEnv, fmt.Sprintf("GOPATH=%s", goDir))
	// Do not set GOBIN, rely on GOPATH/bin to avoid "cross-compiled" errors
	// newEnv = append(newEnv, fmt.Sprintf("GOBIN=%s", goBinDir))

	cmd.Env = newEnv
	cmd.Stdout = m.Output
	cmd.Stderr = m.Output
	if err := cmd.Run(); err != nil {
		return err
	}

	// The binary name is the last part of the source path (before @)
	// Strip version if present in source for binary name detection
	sourcePath := source
	if idx := strings.Index(sourcePath, "@"); idx != -1 {
		sourcePath = sourcePath[:idx]
	}

	binaryName := sourcePath
	if idx := strings.LastIndex(binaryName, "/"); idx != -1 {
		binaryName = binaryName[idx+1:]
	}

	// On Windows, append .exe
	if runtime.GOOS == "windows" && !strings.HasSuffix(binaryName, ".exe") {
		binaryName += ".exe"
	}

	srcBinary, err := m.findBinary(goBinDir, binaryName)
	if err != nil {
		return err
	}

	destBinary := filepath.Join(binDir, binaryName)
	if runtime.GOOS == "windows" && !strings.HasSuffix(destBinary, ".exe") {
		destBinary += ".exe"
	}

	m.log("Copying %s to %s...", srcBinary, destBinary)

	input, err := os.ReadFile(srcBinary)
	if err != nil {
		return fmt.Errorf("failed to read installed binary %s: %w", srcBinary, err)
	}

	if err := os.WriteFile(destBinary, input, 0755); err != nil {
		return fmt.Errorf("failed to copy binary to .box/bin: %w", err)
	}

	return nil
}

func (m *Manager) findBinary(searchDir, name string) (string, error) {
	var srcBinary string
	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == name || info.Name() == name+".exe") {
			srcBinary = path
			return io.EOF
		}
		return nil
	})

	if err == io.EOF {
		return srcBinary, nil
	}
	if err != nil {
		return "", err
	}

	return "", fmt.Errorf("could not find installed binary %s in %s", name, searchDir)
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

	m.log("Updating .envrc...")
	return os.WriteFile(envrcPath, []byte(content), 0644)
}

func (m *Manager) AllowDirenv() error {
	m.log("Running direnv allow...")
	cmd := exec.Command("direnv", "allow")
	cmd.Dir = m.RootDir
	cmd.Stdout = m.Output
	cmd.Stderr = m.Output
	return cmd.Run()
}

func (m *Manager) GenerateDockerfile() error {
	dockerfilePath := filepath.Join(m.RootDir, "Dockerfile")
	content := `FROM debian:bookworm-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl ca-certificates git build-essential \
    nodejs npm ruby-full \
    && rm -rf /var/lib/apt/lists/*

# Install latest Go
RUN curl -LsSf https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xz
ENV PATH="/usr/local/go/bin:${PATH}"

# Install cargo-binstall
RUN curl -L --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/cargo-bins/cargo-binstall/main/install.sh | sh

# Install uv globally (accessible by all users)
ENV UV_INSTALL_DIR="/usr/local/bin"
RUN curl -LsSf https://astral.sh/uv/install.sh | sh

# Copy box binary
COPY --link --chmod=755 box /usr/local/bin/box

# Set up user and workspace
RUN useradd -m -s /bin/bash box
USER box
WORKDIR /home/box

# Copy configuration and install tools
COPY --chown=box:box box.yml .
ENV CGO_ENABLED=0
RUN box install --non-interactive

# Add box binaries to PATH
ENV PATH="/home/box/.box/bin:${PATH}"

ENTRYPOINT ["/bin/bash"]
`
	m.log("Generating Dockerfile...")
	return os.WriteFile(dockerfilePath, []byte(content), 0644)
}

func (m *Manager) installNpm(tool config.Tool, etcDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, tool.Version)
	}
	m.log("Installing %s (npm)...", tool.Source)

	// npm install --prefix .etc -g <package>
	cmd := exec.Command("npm", "install", "--prefix", etcDir, "-g", source)

	cmd.Stdout = m.Output
	cmd.Stderr = m.Output

	return cmd.Run()
}

func (m *Manager) installCargo(tool config.Tool, etcDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, tool.Version)
	}
	m.log("Installing %s (cargo)...", tool.Source)

	// cargo-binstall --root .etc <args> <package>
	args := []string{"--root", etcDir, "-y"}
	args = append(args, tool.Args...)
	args = append(args, source)

	cmd := exec.Command("cargo-binstall", args...)

	cmd.Stdout = m.Output
	cmd.Stderr = m.Output

	return cmd.Run()
}

func (m *Manager) installUv(tool config.Tool, binDir string) error {
	source := tool.Source
	if tool.Version != "" {
		source = fmt.Sprintf("%s==%s", tool.Source, tool.Version)
	}
	m.log("Installing %s (uv)...", tool.Source)

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

	cmd.Stdout = m.Output
	cmd.Stderr = m.Output

	return cmd.Run()
}

func (m *Manager) installGem(tool config.Tool, binDir string) error {
	m.log("Installing %s %s (gem)...", tool.Source, tool.Version)

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

	cmd.Stdout = m.Output
	cmd.Stderr = m.Output

	return cmd.Run()
}

func (m *Manager) installScript(tool config.Tool) error {
	m.log("Installing via script: %s", tool.Source)

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

	cmd.Stdout = m.Output
	cmd.Stderr = m.Output

	return cmd.Run()
}
