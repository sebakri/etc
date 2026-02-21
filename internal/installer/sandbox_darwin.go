//go:build darwin

package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func applySandbox(_ *exec.Cmd, name string, args []string, rootDir string, tempDir string) (string, []string) {
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
