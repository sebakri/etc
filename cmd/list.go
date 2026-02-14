package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sebakri/box/internal/config"
	"github.com/sebakri/box/internal/installer"
	"github.com/spf13/cobra"
)

// listCmd represents the list command

var listCmd = &cobra.Command{

	Use: "list",

	Short: "Lists installed tools and their binaries",

	Run: func(cmd *cobra.Command, args []string) {

		configFile := "box.yml"

		cfg, err := config.Load(configFile)

		if err != nil {

			log.Fatalf("Failed to load %s: %v", configFile, err)

		}

		cwd, err := os.Getwd()

		if err != nil {

			log.Fatalf("Failed to get current working directory: %v", err)

		}

		mgr := installer.New(cwd, cfg.Env)

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
