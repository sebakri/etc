package doctor

import (
	"fmt"
	"os/exec"
)

func Run() {
	fmt.Println("Checking host environment tools...")

	tools := []string{"go", "npm", "cargo", "cargo-binstall", "uv", "direnv"}
	allFound := true

	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			if tool == "cargo-binstall" {
				fmt.Printf("❌ %-14s : Not found (required for cargo tools)\n", tool)
			} else {
				fmt.Printf("❌ %-14s : Not found\n", tool)
			}
			allFound = false
		} else {
			fmt.Printf("✅ %-14s : %s\n", tool, path)
		}
	}

	if !allFound {
		fmt.Println("\nSome tools are missing. Please install them to use their respective package managers.")
	} else {
		fmt.Println("\nAll external tools are ready.")
	}
}
