package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sebakri/box/internal/config"
	"github.com/sebakri/box/internal/installer"
)

// listCmd represents the list command

var listCmd = &cobra.Command{

	Use: "list",

	Short: "Lists installed tools and their binaries",

	Run: func(_ *cobra.Command, _ []string) {

		configFile := "box.yml"

		cfg, err := config.Load(configFile)

		if err != nil {

			log.Fatalf("Failed to load %s: %v", configFile, err)

		}

				cwd, err := os.Getwd()
				if err != nil {
					log.Fatalf("Failed to get current working directory: %v", err)
				}
		
				// Create a specific temp directory for this session
				tempDir, err := os.MkdirTemp("", "box-list-*")
				if err != nil {
					log.Fatalf("Failed to create temporary directory: %v", err)
				}
				defer func() {
					_ = os.RemoveAll(tempDir)
				}()
		
				mgr := installer.New(cwd, tempDir, cfg.Env, cfg)
		

		manifest, err := mgr.LoadManifest()

		if err != nil {

			log.Fatalf("Failed to load manifest: %v", err)

		}

		fmt.Println(titleStyle.Render("Installed tools:"))

		for _, tool := range cfg.Tools {

			fmt.Printf("• %s %s\n", toolStyle.Render(tool.DisplayName()), typeStyle.Render("("+tool.Type+")"))

			if info, ok := manifest.Tools[tool.Source.String()]; ok {

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

	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
