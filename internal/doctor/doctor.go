package doctor

import (
	"fmt"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

func Run() {
	fmt.Println(titleStyle.Render("Checking box host environment tools..."))

	tools := []string{"go", "npm", "cargo-binstall", "uv", "gem", "direnv"}
	allFound := true

	for _, tool := range tools {
		path, err := exec.LookPath(tool)
		if err != nil {
			fmt.Printf("%s %-14s : Not found\n", errorStyle.Render("✗"), tool)
			allFound = false
		} else {
			fmt.Printf("%s %-14s : %s\n", successStyle.Render("✓"), tool, dimStyle.Render(path))
		}
	}

	if !allFound {
		fmt.Println(lipgloss.NewStyle().MarginTop(1).Render("Some tools are missing. Please install them to use their respective package managers."))
	} else {
		fmt.Println(lipgloss.NewStyle().MarginTop(1).Foreground(lipgloss.Color("42")).Render("All external tools are ready. ✨"))
	}
}
