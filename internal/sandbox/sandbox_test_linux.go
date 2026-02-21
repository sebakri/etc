//go:build linux

package sandbox

import (
	"os/exec"
	"testing"
)

func checkLinuxSysProcAttr(t *testing.T, cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		t.Error("SysProcAttr not set on linux")
	}
	if cmd.SysProcAttr.Cloneflags == 0 {
		t.Error("Cloneflags not set on linux")
	}
}
