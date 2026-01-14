package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Tool struct {
	Type    string   `yaml:"type"`              // "go", "npm", "cargo", "uv", "gem", "script"
	Source  string   `yaml:"source"`            // Package path or script command
	Version string   `yaml:"version,omitempty"` // Optional version (e.g., "latest", "0.1.0")
	Args    []string `yaml:"args,omitempty"`
}

type Config struct {
	Tools []Tool            `yaml:"tools"`
	Env   map[string]string `yaml:"env,omitempty"`
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

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}
