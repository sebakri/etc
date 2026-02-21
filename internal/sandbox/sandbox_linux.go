//go:build linux

// Package sandbox provides platform-specific mechanisms for isolating tool execution.
package sandbox

import (
	"os"
	"os/exec"
	"syscall"
)

// Apply configures the command to run within a sandbox on Linux using User Namespaces.
// It maps the current user and group to root inside the namespace but disables setgroups.
func Apply(cmd *exec.Cmd, name string, args []string, _ string, _ string) (string, []string) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		UidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.SysProcIDMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
		GidMappingsEnableSetgroups: false,
	}
	return name, args
}
