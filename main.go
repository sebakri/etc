package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"box/internal/config"
	"box/internal/doctor"
	"box/internal/installer"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]
	switch command {
	case "install":
		runInstall()
	case "run":
		runExecute()
	case "env":
		runEnv()
	case "add":
		runAdd()
	case "generate":
		runGenerate()
	case "doctor":
		doctor.Run()
	case "help":
		usage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		usage()
		os.Exit(1)
	}
}

func runExecute() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: box run <command> [args...]")
		os.Exit(1)
	}

	commandName := os.Args[2]
	commandArgs := os.Args[3:]

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	configFile := "box.yml"
	cfg, err := config.Load(configFile)
	if err != nil {
		// If box.yml is missing, we can still run if the binary exists, 
		// but we won't have custom env vars.
		cfg = &config.Config{}
	}

	boxDir := filepath.Join(cwd, ".box")
	binDir := filepath.Join(boxDir, "bin")
	binaryPath := filepath.Join(binDir, commandName)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		log.Fatalf("Binary %s not found in .box/bin. Have you run 'box install'?", commandName)
	}

	cmd := exec.Command(binaryPath, commandArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ensure .box/bin is in the PATH for the executed command and add custom env vars
	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if len(e) >= 5 && e[:5] == "PATH=" {
			env[i] = "PATH=" + binDir + string(os.PathListSeparator) + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+binDir)
	}

	env = append(env, fmt.Sprintf("BOX_DIR=%s", boxDir))
	env = append(env, fmt.Sprintf("BOX_BIN_DIR=%s", binDir))

	for k, v := range cfg.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}
	cmd.Env = env

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		log.Fatalf("Failed to execute %s: %v", commandName, err)
	}
}

func runEnv() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	configFile := "box.yml"
	cfg, err := config.Load(configFile)
	if err != nil {
		cfg = &config.Config{}
	}

	boxDir := filepath.Join(cwd, ".box")
	binDir := filepath.Join(boxDir, "bin")

	// Get current environment and merge with box.yml env and updated PATH
	env := os.Environ()
	pathFound := false
	for i, e := range env {
		if len(e) >= 5 && e[:5] == "PATH=" {
			env[i] = "PATH=" + binDir + string(os.PathListSeparator) + e[5:]
			pathFound = true
			break
		}
	}
	if !pathFound {
		env = append(env, "PATH="+binDir)
	}

	// Add custom env vars from box.yml
	// We use a map to handle overrides correctly for display
	envMap := make(map[string]string)
	for _, e := range env {
		pair := strings.SplitN(e, "=", 2)
		if len(pair) == 2 {
			envMap[pair[0]] = pair[1]
		}
	}

	envMap["BOX_DIR"] = boxDir
	envMap["BOX_BIN_DIR"] = binDir

	for k, v := range cfg.Env {
		envMap[k] = v
	}

	// If a specific key is requested
	if len(os.Args) > 2 {
		key := os.Args[2]
		if val, ok := envMap[key]; ok {
			fmt.Print(val) // Print without newline for shell substitution $(bx env BOX_DIR)
			return
		}
		os.Exit(1)
	}

	// Print in KEY=VALUE format
	for k, v := range envMap {
		fmt.Printf("%s=%s\n", k, v)
	}
}

func runAdd() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: box add <type> <source>[@version] [args...]")
		fmt.Println("Example: box add uv ruff@0.1.0")
		os.Exit(1)
	}

	toolType := os.Args[2]
	fullSource := os.Args[3]
	args := os.Args[4:]

	source := fullSource
	version := ""

	if strings.Contains(fullSource, "@") {
		parts := strings.SplitN(fullSource, "@", 2)
		source = parts[0]
		version = parts[1]
	}

	configFile := "box.yml"
	cfg, err := config.Load(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &config.Config{}
		} else {
			log.Fatalf("Failed to load %s: %v", configFile, err)
		}
	}

	// Check if tool already exists by source
	for _, tool := range cfg.Tools {
		if tool.Source == source {
			fmt.Printf("⚠️  Tool with source %s already exists in %s\n", source, configFile)
			os.Exit(0)
		}
	}

	newTool := config.Tool{
		Type:    toolType,
		Source:  source,
		Version: version,
		Args:    args,
	}

	cfg.Tools = append(cfg.Tools, newTool)

	if err := cfg.Save(configFile); err != nil {
		log.Fatalf("Failed to save %s: %v", configFile, err)
	}

	if version != "" {
		fmt.Printf("✅ Added %s (version %s) to %s\n", source, version, configFile)
	} else {
		fmt.Printf("✅ Added %s to %s\n", source, configFile)
	}
	fmt.Printf("Run 'box install' to install it.\n")
}

func runGenerate() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: box generate <type>")
		fmt.Println("Available types: direnv")
		os.Exit(1)
	}

	genType := os.Args[2]
	if genType != "direnv" {
		fmt.Printf("Unknown generation type: %s\n", genType)
		os.Exit(1)
	}

	configFile := "box.yml"
	cfg, err := config.Load(configFile)
	if err != nil {
		cfg = &config.Config{}
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	mgr := installer.New(cwd, cfg.Env)
	if err := mgr.EnsureEnvrc(); err != nil {
		log.Fatalf("Failed to generate .envrc: %v", err)
	}

	fmt.Println("✅ Generated .envrc")
	if err := mgr.AllowDirenv(); err != nil {
		fmt.Printf("⚠️  Failed to run direnv allow: %v\n", err)
	}
}

func runInstall() {
	configFile := "box.yml"
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		log.Fatalf("Configuration file %s not found.", configFile)
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load %s: %v", configFile, err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	mgr := installer.New(cwd, cfg.Env)

	fmt.Println("Starting tool installation...")
	for _, tool := range cfg.Tools {
		if err := mgr.Install(tool); err != nil {
			log.Printf("❌ Failed to install %s: %v", tool.Source, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Successfully installed %s\n", tool.Source)
	}

	fmt.Println("All tools installed successfully.")
}

func usage() {
	fmt.Println("box - Minimalist project-local toolbox")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  box <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install   Install tools defined in box.yml")
	fmt.Println("  add       Add a new tool to box.yml")
	fmt.Println("  run       Execute a binary from .box/bin")
	fmt.Println("  env       Display merged environment variables")
	fmt.Println("  generate  Generate configuration files (e.g., direnv)")
	fmt.Println("  doctor    Check if host tools (go, npm, cargo, uv, gem) are installed")
	fmt.Println("  help      Show this help message")
}