//go:build !linux && !darwin

package sandbox

import (
	"os/exec"
)

// Apply is a no-op on unsupported platforms.
func Apply(cmd *exec.Cmd, name string, args []string, _ string, _ string) (string, []string) {
	return name, args
}
