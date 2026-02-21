package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary box.yml
	content := []byte(`
tools:
  - type: go
    source: example.com/tool
  - type: uv
    source: ruff
env:
  KEY: value
`)
	tmpfile, err := os.CreateTemp("", "box-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }() // clean up

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading
	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Tools) != 2 {
		t.Errorf("Expected 2 tools, got %d", len(cfg.Tools))
	}

	if cfg.Env["KEY"] != "value" {
		t.Errorf("Expected Env['KEY'] to be 'value', got '%s'", cfg.Env["KEY"])
	}

	if cfg.Tools[0].Source.String() != "example.com/tool" {
		t.Errorf("Expected source 'example.com/tool', got '%s'", cfg.Tools[0].Source.String())
	}
}

func TestLoadMultilineScript(t *testing.T) {
	content := []byte(`
tools:
  - type: script
    source:
      - echo "hello"
      - echo "world"
`)
	tmpfile, err := os.CreateTemp("", "box-test-*.yml")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(cfg.Tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(cfg.Tools))
	}

	expected := "echo \"hello\"\necho \"world\""
	if cfg.Tools[0].Source.String() != expected {
		t.Errorf("Expected source %q, got %q", expected, cfg.Tools[0].Source.String())
	}
}

func TestToolDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		tool     Tool
		expected string
	}{
		{
			name: "with alias",
			tool: Tool{
				Source: Source{"echo hello"},
				Alias:  "hello-script",
			},
			expected: "hello-script",
		},
		{
			name: "without alias",
			tool: Tool{
				Source: Source{"echo hello"},
			},
			expected: "echo hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tool.DisplayName(); got != tt.expected {
				t.Errorf("DisplayName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFindToolForBinary(t *testing.T) {
	cfg := &Config{
		Tools: []Tool{
			{
				Type:     "go",
				Source:   Source{"github.com/go-task/task/v3/cmd/task"},
				Binaries: []string{"task"},
			},
			{
				Type:   "npm",
				Source: Source{"cowsay"},
			},
		},
	}

	tests := []struct {
		name       string
		binaryName string
		found      bool
	}{
		{"explicit binary", "task", true},
		{"detected binary", "cowsay", true},
		{"not found", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.FindToolForBinary(tt.binaryName)
			if (got != nil) != tt.found {
				t.Errorf("FindToolForBinary(%q) found = %v, want %v", tt.binaryName, got != nil, tt.found)
			}
		})
	}
}

func TestIsSandboxEnabled(t *testing.T) {
	tests := []struct {
		name     string
		tool     Tool
		expected bool
	}{
		{"script tool", Tool{Type: "script"}, true},
		{"go tool", Tool{Type: "go"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.tool.IsSandboxEnabled(); got != tt.expected {
				t.Errorf("IsSandboxEnabled() = %v, want %v", got, tt.expected)
			}
		})
	}
}
