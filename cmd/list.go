package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sebakri/box/internal/config"
	"github.com/sebakri/box/internal/installer"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists installed tools and their binaries",
	RunE: func(_ *cobra.Command, _ []string) error {
		configFile := "box.yml"
		cfg, err := config.Load(configFile)
		if err != nil {
			return fmt.Errorf("failed to load %s: %w", configFile, err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}

		// Create a specific temp directory for this session
		tempDir, err := os.MkdirTemp("", "box-list-*")
		if err != nil {
			return fmt.Errorf("failed to create temporary directory: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()

		mgr := installer.New(cwd, tempDir, cfg.Env, cfg)
		manifest, err := mgr.LoadManifest()
		if err != nil {
			return fmt.Errorf("failed to load manifest: %w", err)
		}

		fmt.Println(titleStyle.Render("Installed tools:"))
		for _, tool := range cfg.Tools {
			fmt.Printf("• %s %s\n", toolStyle.Render(tool.DisplayName()), typeStyle.Render("("+tool.Type+")"))

			if info, ok := manifest.Tools[tool.DisplayName()]; ok {
				binaries := []string{}
				for _, file := range info.Files {
					if strings.HasPrefix(file, ".box/bin/") {
						binaries = append(binaries, filepath.Base(file))
					}
				}

				if len(binaries) > 0 {
					fmt.Printf("  %s %s\n", typeStyle.Render("binaries:"), binStyle.Render(strings.Join(binaries, ", ")))
				}
			}
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
