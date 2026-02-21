package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sebakri/box/internal/config"
	"github.com/sebakri/box/internal/sandbox"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:                "run <command> [args...]",
	Short:              "Execute a binary from the local .box/bin directory",
	DisableFlagParsing: true,
	SilenceUsage:       true,
	Args:               cobra.MinimumNArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		commandName := args[0]
		if commandName != filepath.Base(commandName) {
			return fmt.Errorf("invalid command name %q: path separators are not allowed", commandName)
		}
		commandArgs := args[1:]

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
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
			return fmt.Errorf("binary %s not found in .box/bin. Have you run 'box install'?", commandName)
		}

		// Create a specific temp directory for this run
		tempDir, err := os.MkdirTemp("", "box-run-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		finalCmdName := binaryPath
		finalCmdArgs := commandArgs
		//nolint:gosec
		tempCmd := exec.Command(binaryPath, commandArgs...)

		if cfg.IsSandboxEnabled(commandName) {
			finalCmdName, finalCmdArgs = sandbox.Apply(tempCmd, binaryPath, commandArgs, cwd, tempDir)
		}

		//nolint:gosec
		execCmd := exec.Command(finalCmdName, finalCmdArgs...)
		execCmd.Stdin = os.Stdin
		execCmd.Stdout = os.Stdout
		execCmd.Stderr = os.Stderr
		execCmd.SysProcAttr = tempCmd.SysProcAttr

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

		// Set isolated temp directory
		env = append(env, fmt.Sprintf("TMPDIR=%s", tempDir))
		env = append(env, fmt.Sprintf("TEMP=%s", tempDir))
		env = append(env, fmt.Sprintf("TMP=%s", tempDir))

		for k, v := range cfg.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		execCmd.Env = env

		if err := execCmd.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				os.Exit(exitError.ExitCode())
			}
			return fmt.Errorf("failed to execute %s: %w", commandName, err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
}
