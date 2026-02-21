package sandbox

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestApply(t *testing.T) {
	cmd := exec.Command("ls")
	name, args := Apply(cmd, "ls", []string{"-la"}, "/root", "/tmp")

	switch runtime.GOOS {
	case "darwin":
		if name != "sandbox-exec" {
			t.Errorf("Expected sandbox-exec on darwin, got %s", name)
		}
		foundProfile := false
		for i, arg := range args {
			if arg == "-p" && i+1 < len(args) {
				profile := args[i+1]
				if !strings.Contains(profile, "(allow default)") {
					t.Errorf("Sandbox profile missing (allow default): %s", profile)
				}
				foundProfile = true
			}
		}
		if !foundProfile {
			t.Error("Sandbox profile not found in args")
		}
	case "linux":
		checkLinuxSysProcAttr(t, cmd)
	default:
		if name != "ls" {
			t.Errorf("Expected unchanged name on %s, got %s", runtime.GOOS, name)
		}
	}
}
