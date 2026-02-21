//go:build !linux

package sandbox

import (
	"os/exec"
	"testing"
)

func checkLinuxSysProcAttr(_ *testing.T, _ *exec.Cmd) {}
