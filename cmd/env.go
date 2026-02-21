package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sebakri/box/internal/config"

	"github.com/spf13/cobra"
)

// envCmd represents the env command
var envCmd = &cobra.Command{
	Use:   "env [key]",
	Short: "Display the merged list of environment variables",
	RunE: func(_ *cobra.Command, args []string) error {
		configFile, err := findNearestBoxConfig()
		if err != nil {
			return fmt.Errorf("could not find box.yml: %w", err)
		}

		cfg, err := config.Load(configFile)
		if err != nil {
			cfg = &config.Config{}
		}

		boxDir := filepath.Dir(configFile)
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
		if len(args) > 0 {
			key := args[0]
			if val, ok := envMap[key]; ok {
				fmt.Print(val) // Print without newline for shell substitution $(bx env BOX_DIR)
				return nil
			}
			return fmt.Errorf("environment variable %s not found", key)
		}

		// Print in KEY=VALUE format
		for k, v := range envMap {
			fmt.Printf("%s=%s\n", k, v)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(envCmd)
}

func findNearestBoxConfig() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		configPath := filepath.Join(dir, "box.yml")
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // Reached root directory
		}
		dir = parentDir
	}

	return "", fmt.Errorf("box.yml not found in current or parent directories")
}
