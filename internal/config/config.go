// Package config handles the loading and saving of box.yml configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Source represents a single string or a list of strings for tool sources.
type Source []string

// UnmarshalYAML implements custom unmarshaling for Source.
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

// MarshalYAML implements custom marshaling for Source.
func (s Source) MarshalYAML() (interface{}, error) {
	if len(s) == 1 {
		return s[0], nil
	}
	return []string(s), nil
}

func (s Source) String() string {
	return strings.Join(s, "\n")
}

// IsSandboxEnabled returns true if sandboxing should be used for this binary.
func (c *Config) IsSandboxEnabled(binaryName string) bool {
	if t := c.FindToolForBinary(binaryName); t != nil {
		return t.IsSandboxEnabled()
	}
	return false
}

// Tool defines a single tool to be installed by box.
type Tool struct {
	Type     string   `yaml:"type"`               // "go", "npm", "cargo", "uv", "gem", "script"
	Source   Source   `yaml:"source"`             // Package path or script command
	Alias    string   `yaml:"alias,omitempty"`    // Optional alias for display
	Version  string   `yaml:"version,omitempty"`  // Optional version (e.g., "latest", "0.1.0")
	Binaries []string `yaml:"binaries,omitempty"` // Optional explicit list of binaries
	Args     []string `yaml:"args,omitempty"`
}

// IsSandboxEnabled returns true if sandboxing is enabled for this tool.
// It is always true for 'script' type, and false for others.
func (t Tool) IsSandboxEnabled() bool {
	return t.Type == "script"
}

// FindToolForBinary looks up which tool definition produces the given binary name.
func (c *Config) FindToolForBinary(binaryName string) *Tool {
	for i := range c.Tools {
		t := &c.Tools[i]
		// Check explicit binaries
		for _, b := range t.Binaries {
			if b == binaryName {
				return t
			}
		}
		// If no explicit binaries, try to match the source-based name
		if len(t.Binaries) == 0 {
			if t.detectBinaryName() == binaryName {
				return t
			}
		}
	}
	return nil
}

func (t Tool) detectBinaryName() string {
	source := t.Source.String()
	// The binary name is the last part of the source path (before @ or ==)
	sourcePath := source
	if idx := strings.Index(sourcePath, "@"); idx != -1 {
		sourcePath = sourcePath[:idx]
	}
	if idx := strings.Index(sourcePath, "=="); idx != -1 {
		sourcePath = sourcePath[:idx]
	}

	// Strip major version suffix (e.g. /v2, /v3) if it's the last part of the path
	if parts := strings.Split(sourcePath, "/"); len(parts) > 1 {
		lastPart := parts[len(parts)-1]
		if len(lastPart) >= 2 && lastPart[0] == 'v' && isDigit(lastPart[1:]) {
			sourcePath = strings.Join(parts[:len(parts)-1], "/")
		}
	}

	binaryName := sourcePath
	if idx := strings.LastIndex(binaryName, "/"); idx != -1 {
		binaryName = binaryName[idx+1:]
	}
	return binaryName
}

func isDigit(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// DisplayName returns a human-readable name for the tool.
func (t Tool) DisplayName() string {
	if t.Alias != "" {
		return t.Alias
	}
	return t.Source.String()
}

// Config represents the top-level box configuration.
type Config struct {
	Tools []Tool            `yaml:"tools"`
	Env   map[string]string `yaml:"env,omitempty"`
}

// Load loads the configuration from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to the given path.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Clean(path), data, 0600)
}
