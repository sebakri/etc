//go:build darwin

package sandbox

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Apply configures the command to run within a sandbox on macOS using sandbox-exec.
// It allows write access to the project root, the .box directory, and the specified tempDir.
func Apply(cmd *exec.Cmd, name string, args []string, rootDir string, tempDir string) (string, []string) {
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

	boxDir := filepath.Join(resolvedRoot, ".box")

	profile := fmt.Sprintf(`(version 1)
(allow default)
(deny file-write*)
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
`, resolvedRoot, boxDir, resolvedTemp)

	return "sandbox-exec", append([]string{"-p", profile, name}, args...)
}
