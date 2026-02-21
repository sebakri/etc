//go:build !linux && !darwin
package cmd

import (
	"os/exec"
)

func applySandbox(cmd *exec.Cmd, name string, args []string, _ string, _ string) (string, []string) {
	return name, args
}
