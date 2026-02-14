package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Source []string

func (s *Source) UnmarshalYAML(value *yaml.Node) error {
	var multi []string
	if err := value.Decode(&multi); err == nil {
		*s = multi
		return nil
	}

	var single string
	if err := value.Decode(&single); err == nil {
		*s = []string{single}
		return nil
	}

	return fmt.Errorf("line %d: cannot unmarshal %s into Source (string or array of strings)", value.Line, value.Tag)
}

func (s Source) MarshalYAML() (interface{}, error) {
	if len(s) == 1 {
		return s[0], nil
	}
	return []string(s), nil
}

func (s Source) String() string {
	return strings.Join(s, "\n")
}

type Tool struct {
	Type     string   `yaml:"type"`               // "go", "npm", "cargo", "uv", "gem", "script"
	Source   Source   `yaml:"source"`             // Package path or script command
	Alias    string   `yaml:"alias,omitempty"`    // Optional alias for display
	Version  string   `yaml:"version,omitempty"`  // Optional version (e.g., "latest", "0.1.0")
	Binaries []string `yaml:"binaries,omitempty"` // Optional explicit list of binaries
	Args     []string `yaml:"args,omitempty"`
}

func (t Tool) DisplayName() string {
	if t.Alias != "" {
		return t.Alias
	}
	return t.Source.String()
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
