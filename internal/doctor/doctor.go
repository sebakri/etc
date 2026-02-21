// Package doctor provides diagnostic tools for verifying the host environment.
package doctor

import (
	"fmt"
	"os/exec"
	"sort"

	"github.com/sebakri/box/internal/installer"

	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).MarginBottom(1)
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
)

// Run executes a series of checks on the host environment to ensure required tools are present.
func Run() {
	fmt.Println(titleStyle.Render("Checking box host environment tools..."))

	allFound := true

	// Get tools from central registry
	var toolNames []string
	toolMap := make(map[string]string)
	for _, t := range installer.SupportedTools {
		toolNames = append(toolNames, t.Name)
		toolMap[t.Name] = t.Name
	}
	// Add direnv specifically as it's an integration, not a tool type
	toolNames = append(toolNames, "direnv")
	sort.Strings(toolNames)

	for _, tool := range toolNames {
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
