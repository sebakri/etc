package installer

import (
	"encoding/gob"
	"fmt"
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

	return m.updateManifest(tool.Name, newFiles)
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
		gob.NewDecoder(file).Decode(&manifest)
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
				fmt.Printf("Removing empty directory %s...\n", file)
				os.Remove(fullPath)
			}
		} else {
			fmt.Printf("Removing file %s...\n", file)
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
		fmt.Printf("Removing binary %s...\n", binaryPath)
		os.Remove(binaryPath)
	}

	uvToolDir := filepath.Join(boxDir, "uv", name)
	if _, err := os.Stat(uvToolDir); err == nil {
		fmt.Printf("Removing data directory %s...\n", uvToolDir)
		os.RemoveAll(uvToolDir)
	}

	return nil
}

func (m *Manager) installGo(tool config.Tool, binDir string) error {
	version := tool.Version
	if version != "" && version != "latest" && version != "master" && !strings.HasPrefix(version, "v") {
		// If it looks like a version (starts with a digit), try prepending 'v'
		if len(version) > 0 && version[0] >= '0' && version[0] <= '9' {
			version = "v" + version
		}
	}

	source := tool.Source
	if version != "" {
		source = fmt.Sprintf("%s@%s", tool.Source, version)
	}
	fmt.Printf("Installing %s (go)...\n", tool.Name)

	err := m.runGoInstall(source, binDir, tool.Name)
	
	// If it failed and we didn't have a 'v' prefix, try with it as a fallback
	if err != nil && tool.Version != "" && !strings.HasPrefix(tool.Version, "v") && version == tool.Version {
		fallbackVersion := "v" + tool.Version
		fallbackSource := fmt.Sprintf("%s@%s", tool.Source, fallbackVersion)
		fmt.Printf("Retrying %s with version %s...\n", tool.Name, fallbackVersion)
		err = m.runGoInstall(fallbackSource, binDir, tool.Name)
	}

	return err
}

func (m *Manager) runGoInstall(source string, binDir string, toolName string) error {
	tempDir, err := os.MkdirTemp("", "box-go-install-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Run go install with a temporary GOPATH
	cmd := exec.Command("go", "install", source)
	env := os.Environ()
	env = append(env, fmt.Sprintf("GOPATH=%s", tempDir))
	// Unset GOBIN if it's set in the environment
	newEnv := []string{}
	for _, e := range env {
		if !strings.HasPrefix(e, "GOBIN=") {
			newEnv = append(newEnv, e)
		}
	}
	cmd.Env = newEnv
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
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

	// Find the binary in tempDir/bin
	// It might be in a GOOS_GOARCH subfolder if it's cross-compiling
	srcBinary := ""
	err = filepath.Walk(filepath.Join(tempDir, "bin"), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (info.Name() == binaryName || info.Name() == binaryName+".exe") {
			srcBinary = path
			return filepath.SkipAll
		}
		return nil
	})

	if srcBinary == "" {
		return fmt.Errorf("could not find installed binary %s in %s", binaryName, tempDir)
	}

	destBinary := filepath.Join(binDir, toolName)
	if runtime.GOOS == "windows" && !strings.HasSuffix(destBinary, ".exe") {
		destBinary += ".exe"
	}

	fmt.Printf("Copying %s to %s...\n", srcBinary, destBinary)
	
	input, err := os.ReadFile(srcBinary)
	if err != nil {
		return fmt.Errorf("failed to read installed binary %s: %w", srcBinary, err)
	}
	
	if err := os.WriteFile(destBinary, input, 0755); err != nil {
		return fmt.Errorf("failed to copy binary to .box/bin: %w", err)
	}

	return nil
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
	fmt.Printf("Installing %s (npm)...\n", tool.Name)

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
	fmt.Printf("Installing %s (cargo)...\n", tool.Name)

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
	fmt.Printf("Installing %s (uv)...\n", tool.Name)

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
	fmt.Printf("Installing %s %s (gem)...\n", tool.Name, tool.Version)

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
