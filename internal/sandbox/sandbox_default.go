//go:build !linux && !darwin

// Package sandbox provides platform-specific mechanisms for isolating tool execution.
package sandbox

import (
	"os/exec"
)

// Apply is a no-op on unsupported platforms.
func Apply(_ *exec.Cmd, name string, args []string, _ string, _ string) (string, []string) {
	return name, args
}
