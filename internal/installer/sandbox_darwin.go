//go:build darwin
package installer

import (
	"fmt"
	"os"
	"os/exec"
)

func applySandbox(cmd *exec.Cmd, name string, args []string, rootDir string) (string, []string) {
	tempDir := os.TempDir()
	profile := fmt.Sprintf(`(version 1)
(allow default)
(deny file-write*)
(allow file-write* (subpath %q))
(allow file-write* (subpath %q))
(allow file-write* (subpath "/private/var/folders"))
`, rootDir, tempDir)

	return "sandbox-exec", append([]string{"-p", profile, name}, args...)
}





