package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Tool struct {
	Name   string   `yaml:"name"`
	Type   string   `yaml:"type"`   // "go", "npm", "cargo", "uv"
	Source string   `yaml:"source"` // Package path (e.g., "golang.org/x/tools/gopls@latest")
	Args   []string `yaml:"args,omitempty"`
}

type Config struct {
	Tools []Tool `yaml:"tools"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
