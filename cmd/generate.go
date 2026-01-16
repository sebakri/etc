package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"box/internal/config"
	"box/internal/installer"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate <type>",
	Short: "Generates configuration files",
	Long:  `Generates configuration files for shell integration or containerization (e.g., direnv, dockerfile).`,
	ValidArgs: []string{"direnv", "dockerfile"},
	Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {
		genType := args[0]

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
		mgr.Output = io.Discard

		switch genType {
		case "direnv":
			if err := mgr.EnsureEnvrc(); err != nil {
				log.Fatalf("Failed to generate .envrc: %v", err)
			}
			fmt.Printf("%s Generated .envrc\n", successStyle.Render("✅"))
			if err := mgr.AllowDirenv(); err != nil {
				fmt.Printf("%s Failed to run direnv allow: %v\n", warnStyle.Render("⚠️"), err)
			}
		case "dockerfile":
			if err := mgr.GenerateDockerfile(); err != nil {
				log.Fatalf("Failed to generate Dockerfile: %v", err)
			}
			fmt.Printf("%s Generated Dockerfile\n", successStyle.Render("✅"))
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
}
