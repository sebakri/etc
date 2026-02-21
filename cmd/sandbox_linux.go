//go:build linux
package cmd

import (
	"os"
	"os/exec"
	"syscall"
)

func applySandbox(cmd *exec.Cmd, name string, args []string, _ string, _ string) (string, []string) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUSER | syscall.CLONE_NEWNS,
		UidMappings: []syscall.UidMap{
			{ContainerID: 0, HostID: os.Getuid(), Size: 1},
		},
		GidMappings: []syscall.GidMap{
			{ContainerID: 0, HostID: os.Getgid(), Size: 1},
		},
	}
	return name, args
}
