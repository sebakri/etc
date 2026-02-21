package installer

import (
	"testing"
)

func TestDetectBinaryName(t *testing.T) {
	m := &Manager{}
	tests := []struct {
		source   string
		expected string
	}{
		{"github.com/go-task/task/v3/cmd/task", "task"},
		{"github.com/org/tool@v1.0.0", "tool"},
		{"package-name==1.2.3", "package-name"},
		{"github.com/org/repo/v2", "repo"},
		{"simple-tool", "simple-tool"},
	}

	for _, tt := range tests {
		t.Run(tt.source, func(t *testing.T) {
			if got := m.detectBinaryName(tt.source); got != tt.expected {
				t.Errorf("detectBinaryName(%q) = %v, want %v", tt.source, got, tt.expected)
			}
		})
	}
}

func TestPrepareGoEnv(t *testing.T) {
	m := &Manager{}
	goDir := "/tmp/box/go"
	env := m.prepareGoEnv(goDir)

	foundGopath := false
	for _, e := range env {
		if e == "GOPATH="+goDir {
			foundGopath = true
		}
		if e[:6] == "GOBIN=" {
			t.Errorf("prepareGoEnv should have stripped GOBIN, but found %s", e)
		}
	}

	if !foundGopath {
		t.Errorf("prepareGoEnv did not set GOPATH to %s", goDir)
	}
}

func TestShellEscape(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "'simple'"},
		{"with space", "'with space'"},
		{"with'quote", "'with'\\''quote'"},
		{"", "''"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := shellEscape(tt.input); got != tt.expected {
				t.Errorf("shellEscape(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsDigit(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"123", true},
		{"v1", false},
		{"", false},
		{"1.2", false},
	}

	for _, tt := range tests {
		if got := isDigit(tt.input); got != tt.expected {
			t.Errorf("isDigit(%q) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}
