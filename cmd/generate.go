package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sebakri/box/internal/config"
	"github.com/sebakri/box/internal/installer"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:       "generate <type>",
	Short:     "Generates configuration files",
	Long:      `Generates configuration files for shell integration or containerization (e.g., direnv, dockerfile).`,
	ValidArgs: []string{"direnv", "dockerfile"},
	Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(_ *cobra.Command, args []string) error {
		genType := args[0]

		configFile := "box.yml"
		cfg, err := config.Load(configFile)
		if err != nil {
			cfg = &config.Config{}
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Create a specific temp directory for this session
		tempDir, err := os.MkdirTemp("", "box-generate-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		mgr := installer.New(cwd, tempDir, cfg.Env, cfg)
		mgr.Output = io.Discard

		switch genType {
		case "direnv":
			if err := mgr.EnsureEnvrc(); err != nil {
				return fmt.Errorf("failed to generate .envrc: %w", err)
			}
			fmt.Printf("%s Generated .envrc\n", successStyle.Render("✅"))
			if err := mgr.AllowDirenv(); err != nil {
				fmt.Printf("%s Failed to run direnv allow: %v\n", warnStyle.Render("⚠️"), err)
			}
		case "dockerfile":
			if err := mgr.GenerateDockerfile(); err != nil {
				return fmt.Errorf("failed to generate Dockerfile: %w", err)
			}
			fmt.Printf("%s Generated Dockerfile\n", successStyle.Render("✅"))
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(generateCmd)
}
