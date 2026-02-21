//go:build !linux

package sandbox

import (
	"os/exec"
	"testing"
)

func checkLinuxSysProcAttr(t *testing.T, cmd *exec.Cmd) {}
