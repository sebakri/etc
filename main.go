package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"etc/internal/config"
	"etc/internal/doctor"
	"etc/internal/installer"
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
		fmt.Println("Usage: etc run <command> [args...]")
		os.Exit(1)
	}

	commandName := os.Args[2]
	commandArgs := os.Args[3:]

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	configFile := "etc.yml"
	cfg, err := config.Load(configFile)
	if err != nil {
		// If etc.yml is missing, we can still run if the binary exists, 
		// but we won't have custom env vars.
		cfg = &config.Config{}
	}

	binDir := filepath.Join(cwd, ".etc", "bin")
	binaryPath := filepath.Join(binDir, commandName)

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		log.Fatalf("Binary %s not found in .etc/bin. Have you run 'etc install'?", commandName)
	}

	cmd := exec.Command(binaryPath, commandArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Ensure .etc/bin is in the PATH for the executed command and add custom env vars
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

func runInstall() {
	configFile := "etc.yml"
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
			log.Printf("❌ Failed to install %s: %v", tool.Name, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Successfully installed %s\n", tool.Name)
	}

	if err := mgr.EnsureEnvrc(); err != nil {
		log.Printf("⚠️  Failed to create .envrc: %v", err)
	} else {
		if err := mgr.AllowDirenv(); err != nil {
			log.Printf("⚠️  Failed to run direnv allow (is direnv installed?): %v", err)
		}
	}

	fmt.Println("All tools installed successfully.")
}

func usage() {
	fmt.Println("etc - Ephemeral Tool Configuration")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  etc <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  install   Install tools defined in etc.yml")
	fmt.Println("  run       Execute a binary from .etc/bin")
	fmt.Println("  doctor    Check if host tools (go, npm, cargo) are installed")
	fmt.Println("  help      Show this help message")
}
