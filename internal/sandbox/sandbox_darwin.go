//go:build darwin

// Package sandbox provides platform-specific mechanisms for isolating tool execution.
package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Apply configures the command to run within a sandbox on macOS using sandbox-exec.
// It allows write access to the project root, the .box directory, and the specified tempDir.
func Apply(_ *exec.Cmd, name string, args []string, rootDir string, tempDir string) (string, []string) {
	if tempDir == "" {
		tempDir = os.TempDir()
	}

	// Resolve symlinks (especially for macOS /var -> /private/var)
	resolvedRoot, err := filepath.EvalSymlinks(rootDir)
	if err != nil {
		resolvedRoot = rootDir
	}
	resolvedTemp, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		resolvedTemp = tempDir
	}

	// Always allow the system temp dir as well, as some tools (like mktemp on macOS)
	// may ignore TMPDIR or require access to the parent temp hierarchy.
	systemTemp := os.TempDir()
	resolvedSystemTemp, err := filepath.EvalSymlinks(systemTemp)
	if err != nil {
		resolvedSystemTemp = systemTemp
	}

	profile := fmt.Sprintf(`(version 1)
(allow default)
(deny file-write*)
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
(allow file-write* (literal "/dev/null"))
(allow file-write* (literal "/dev/zero"))
(allow file-write* (literal "/dev/stdout"))
(allow file-write* (literal "/dev/stderr"))
`, resolvedRoot, tempDir, resolvedTemp, systemTemp, resolvedSystemTemp)

	return "sandbox-exec", append([]string{"-p", profile, name}, args...)
}
