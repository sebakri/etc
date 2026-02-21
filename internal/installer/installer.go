// Package installer provides mechanisms for installing tools from various package managers into the .box directory.
package installer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/sebakri/box/internal/config"
)

// Manager handles tool installations and environment setup.
type Manager struct {
	RootDir      string
	TempDir      string
	Env          map[string]string
	Output       io.Writer
	GlobalConfig *config.Config

	// installers map tool types to their implementation
	installers map[string]Installer
}

// ToolManifest tracks metadata and files installed for a specific tool.
type ToolManifest struct {
	Type      string    `json:"type"`
	Source    string    `json:"source"`
	Version   string    `json:"version,omitempty"`
	Files     []string  `json:"files"`
	Installed time.Time `json:"installed"`
	Updated   time.Time `json:"updated"`
}

// Manifest represents the persistent state of installed tools.
type Manifest struct {
	Tools map[string]ToolManifest `json:"tools"`
}

// New creates a new Manager instance.
func New(rootDir string, tempDir string, env map[string]string, cfg *config.Config) *Manager {
	m := &Manager{
		RootDir:      rootDir,
		TempDir:      tempDir,
		Env:          env,
		Output:       os.Stdout,
		GlobalConfig: cfg,
		installers:   make(map[string]Installer),
	}

	// Register default installers from central registry
	for k, v := range SupportedTools {
		m.RegisterInstaller(k, v.Installer)
	}

	return m
}

// RegisterInstaller adds a new installer for a tool type.
func (m *Manager) RegisterInstaller(toolType string, installer Installer) {
	m.installers[toolType] = installer
}

func (m *Manager) log(format string, a ...any) {
	if m.Output != nil {
		_, _ = fmt.Fprintf(m.Output, format+"\n", a...)
	}
}

// Install installs a tool based on its configuration.
func (m *Manager) Install(tool config.Tool) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	// Ensure directories exist
	if err := os.MkdirAll(binDir, 0700); err != nil {
		return fmt.Errorf("failed to create bin dir: %w", err)
	}

	// Capture state before install
	before, err := m.captureState()
	if err != nil {
		return fmt.Errorf("failed to capture state before install: %w", err)
	}

	installer, ok := m.installers[tool.Type]
	if !ok {
		return fmt.Errorf("unsupported tool type: %s", tool.Type)
	}

	// Determine if sandbox is enabled for this tool (always true for scripts)
	sandboxEnabled := tool.IsSandboxEnabled()

	managedFiles, err := installer.Install(tool, m, sandboxEnabled)
	if err != nil {
		return err
	}

	// Capture state after install and find new files
	after, err := m.captureState()
	if err != nil {
		return fmt.Errorf("failed to capture state after install: %w", err)
	}

	newFilesMap := make(map[string]bool)
	for _, f := range managedFiles {
		newFilesMap[f] = true
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			newFilesMap[path] = true
		}
	}

	newFileList := make([]string, 0, len(newFilesMap))
	for f := range newFilesMap {
		newFileList = append(newFileList, f)
	}
	sort.Strings(newFileList)

	return m.updateManifest(tool, newFileList)
}

func (m *Manager) captureState() (map[string]bool, error) {
	state := make(map[string]bool)
	boxDir := filepath.Join(m.RootDir, ".box")

	if _, err := os.Stat(boxDir); os.IsNotExist(err) {
		return state, nil
	}

	err := filepath.Walk(boxDir, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// We track the relative path from RootDir
		rel, err := filepath.Rel(m.RootDir, path)
		if err != nil {
			return err
		}
		// Skip the .box directory itself and the manifest file
		if rel == ".box" || rel == filepath.Join(".box", "manifest.json") {
			return nil
		}
		state[rel] = true
		return nil
	})

	return state, err
}

func (m *Manager) updateManifest(tool config.Tool, files []string) error {
	manifestPath := filepath.Join(m.RootDir, ".box", "manifest.json")
	manifest := Manifest{Tools: make(map[string]ToolManifest)}

	if data, err := os.ReadFile(filepath.Clean(manifestPath)); err == nil {
		_ = json.Unmarshal(data, &manifest)
	}

	name := tool.DisplayName()
	existing, ok := manifest.Tools[name]

	now := time.Now()
	installed := now
	if ok {
		installed = existing.Installed
		// Merge and deduplicate files
		fileMap := make(map[string]bool)
		for _, f := range existing.Files {
			fileMap[f] = true
		}
		for _, f := range files {
			fileMap[f] = true
		}
		newFileList := make([]string, 0, len(fileMap))
		for f := range fileMap {
			newFileList = append(newFileList, f)
		}
		sort.Strings(newFileList)
		files = newFileList
	}

	manifest.Tools[name] = ToolManifest{
		Type:      tool.Type,
		Source:    tool.Source.String(),
		Version:   tool.Version,
		Files:     files,
		Installed: installed,
		Updated:   now,
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(manifestPath), data, 0600)
}

// LoadManifest reads the installed tools manifest.
func (m *Manager) LoadManifest() (*Manifest, error) {
	manifestPath := filepath.Join(m.RootDir, ".box", "manifest.json")
	manifest := Manifest{Tools: make(map[string]ToolManifest)}

	data, err := os.ReadFile(filepath.Clean(manifestPath))
	if err != nil {
		if os.IsNotExist(err) {
			return &manifest, nil
		}
		return nil, err
	}

	err = json.Unmarshal(data, &manifest)
	return &manifest, err
}

// Uninstall removes an installed tool and its files.
func (m *Manager) Uninstall(name string) error {
	manifest, err := m.LoadManifest()
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

		// Security check: ensure the file is inside the project root
		// This prevents path traversal attacks if the manifest is tampered with.
		rel, err := filepath.Rel(m.RootDir, fullPath)
		if err != nil || strings.HasPrefix(rel, "..") || strings.HasPrefix(rel, "/") || (runtime.GOOS == "windows" && strings.Contains(rel, ":")) {
			m.log("Security Warning: Skipping deletion of unsafe path %s", fullPath)
			continue
		}

		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}

		if info.IsDir() {
			entries, _ := os.ReadDir(fullPath)
			if len(entries) == 0 {
				m.log("Removing empty directory %s...", file)
				_ = os.Remove(fullPath)
			}
		} else {
			m.log("Removing file %s...", file)
			_ = os.Remove(fullPath)
		}
	}

	delete(manifest.Tools, name)

	manifestPath := filepath.Join(m.RootDir, ".box", "manifest.json")
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Clean(manifestPath), data, 0600)
}

func (m *Manager) uninstallBestEffort(name string) error {
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	binaryPath := filepath.Join(binDir, name)
	if _, err := os.Stat(binaryPath); err == nil {
		m.log("Removing binary %s...", binaryPath)
		_ = os.Remove(binaryPath)
	}

	uvToolDir := filepath.Join(boxDir, "uv", name)
	if _, err := os.Stat(uvToolDir); err == nil {
		m.log("Removing data directory %s...", uvToolDir)
		_ = os.RemoveAll(uvToolDir)
	}

	return nil
}

func (m *Manager) prepareGoEnv(goDir string) []string {
	env := os.Environ()
	newEnv := []string{}
	for _, e := range env {
		if !strings.HasPrefix(e, "GOBIN=") && !strings.HasPrefix(e, "GOPATH=") {
			newEnv = append(newEnv, e)
		}
	}
	return append(newEnv, fmt.Sprintf("GOPATH=%s", goDir))
}

func (m *Manager) detectBinaryName(source string) string {
	// The binary name is the last part of the source path (before @ or ==)
	sourcePath := source
	if idx := strings.Index(sourcePath, "@"); idx != -1 {
		sourcePath = sourcePath[:idx]
	}
	if idx := strings.Index(sourcePath, "=="); idx != -1 {
		sourcePath = sourcePath[:idx]
	}

	// Strip major version suffix (e.g. /v2, /v3) if it's the last part of the path
	if parts := strings.Split(sourcePath, "/"); len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if len(lastPart) >= 2 && lastPart[0] == 'v' && isDigit(lastPart[1:]) {
			sourcePath = strings.Join(parts[:len(parts)-1], "/")
		}
	}

	binaryName := sourcePath
	if idx := strings.LastIndex(binaryName, "/"); idx != -1 {
		binaryName = binaryName[idx+1:]
	}
	return binaryName
}

func (m *Manager) linkBinaries(goBinDir, binDir string, binaries []string) ([]string, error) {
	createdFiles := []string{}
	for _, name := range binaries {
		srcBinary, err := m.findBinary(goBinDir, name)
		if err != nil {
			return nil, err
		}

		destBinary := filepath.Join(binDir, name)
		if runtime.GOOS == "windows" && !strings.HasSuffix(destBinary, ".exe") {
			destBinary += ".exe"
		}

		_ = os.Remove(destBinary)

		relPath, err := filepath.Rel(binDir, srcBinary)
		if err == nil {
			m.log("Symlinking %s to %s...", relPath, destBinary)
			err = os.Symlink(relPath, destBinary)
			if err == nil {
				relToRoot, _ := filepath.Rel(m.RootDir, destBinary)
				createdFiles = append(createdFiles, relToRoot)
				continue
			}
			m.log("Symlink failed, falling back to copy: %v", err)
		}

		m.log("Copying %s to %s...", srcBinary, destBinary)
		input, err := os.ReadFile(filepath.Clean(srcBinary))
		if err != nil {
			return nil, fmt.Errorf("failed to read installed binary %s: %w", srcBinary, err)
		}

		if err := os.WriteFile(destBinary, input, 0600); err != nil {
			return nil, fmt.Errorf("failed to copy binary to .box/bin: %w", err)
		}
		relToRoot, _ := filepath.Rel(m.RootDir, destBinary)
		createdFiles = append(createdFiles, relToRoot)
	}
	return createdFiles, nil
}

func (m *Manager) findBinary(searchDir, name string) (string, error) {
	var newestBinary string
	var newestModTime os.FileInfo

	err := filepath.Walk(searchDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == name || info.Name() == name+".exe") {
			if newestBinary == "" || info.ModTime().After(newestModTime.ModTime()) {
				newestBinary = path
				newestModTime = info
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if newestBinary != "" {
		return newestBinary, nil
	}

	return "", fmt.Errorf("could not find installed binary %s in %s", name, searchDir)
}

// EnsureEnvrc creates or updates the .envrc file.
func (m *Manager) EnsureEnvrc() error {
	envrcPath := filepath.Join(m.RootDir, ".envrc")
	boxDir := filepath.Join(m.RootDir, ".box")
	binDir := filepath.Join(boxDir, "bin")

	content := fmt.Sprintf("export BOX_DIR=%s\n", shellEscape(boxDir))
	content += fmt.Sprintf("export BOX_BIN_DIR=%s\n", shellEscape(binDir))
	content += fmt.Sprintf("export BOX_OS=%s\n", shellEscape(runtime.GOOS))
	content += fmt.Sprintf("export BOX_ARCH=%s\n", shellEscape(runtime.GOARCH))
	content += "PATH_add .box/bin\n"

	// Sort keys for deterministic output
	keys := make([]string, 0, len(m.Env))
	for k := range m.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := m.Env[k]
		content += fmt.Sprintf("export %s=%s\n", k, shellEscape(v))
	}

	m.log("Updating .envrc...")
	return os.WriteFile(filepath.Clean(envrcPath), []byte(content), 0600)
}

func shellEscape(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// AllowDirenv runs direnv allow in the project directory.
func (m *Manager) AllowDirenv() error {
	m.log("Running direnv allow...")
	return m.runCommand("direnv", []string{"allow"}, nil, m.RootDir, false)
}

// GenerateDockerfile creates a Dockerfile for the project.
// GenerateDockerfile creates a Dockerfile for the project.
func (m *Manager) GenerateDockerfile() error {
	dockerfilePath := filepath.Join(m.RootDir, "Dockerfile")
	content := `FROM debian:bookworm-slim

# Package manager feature flags
ARG INSTALL_GO=true
ARG INSTALL_NODE=true
ARG INSTALL_CARGO=true
ARG INSTALL_UV=true
ARG INSTALL_RUBY=true
ARG INSTALL_PIP=true

# Install system dependencies and selected package managers
RUN apt-get update && \
    PACKAGES="curl ca-certificates git build-essential direnv" && \
    if [ "$INSTALL_NODE" = "true" ]; then PACKAGES="$PACKAGES nodejs npm"; fi && \
    if [ "$INSTALL_RUBY" = "true" ]; then PACKAGES="$PACKAGES ruby-full"; fi && \
    if [ "$INSTALL_PIP" = "true" ]; then PACKAGES="$PACKAGES python3-pip"; fi && \
    apt-get install -y --no-install-recommends $PACKAGES && \
    rm -rf /var/lib/apt/lists/*

# Install latest Go if enabled
RUN if [ "$INSTALL_GO" = "true" ]; then \
    curl -LsSf https://go.dev/dl/go1.24.0.linux-amd64.tar.gz | tar -C /usr/local -xz; \
    fi
ENV PATH="/usr/local/go/bin:${PATH}"

# Install cargo-binstall if enabled
RUN if [ "$INSTALL_CARGO" = "true" ]; then \
    curl -L --proto '=https' --tlsv1.2 -sSf https://raw.githubusercontent.com/cargo-bins/cargo-binstall/main/install.sh | sh && \
    if [ -f "$HOME/.cargo/bin/cargo-binstall" ]; then mv "$HOME/.cargo/bin/cargo-binstall" /usr/local/bin/; fi; \
    fi

# Install uv globally if enabled
RUN if [ "$INSTALL_UV" = "true" ]; then \
    curl -LsSf https://astral.sh/uv/install.sh | UV_INSTALL_DIR=/usr/local/bin sh; \
    fi

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

CMD ["/bin/bash"]
`
	m.log("Generating Dockerfile...")
	return os.WriteFile(dockerfilePath, []byte(content), 0600)
}

func isDigit(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
