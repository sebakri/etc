package cmd

import (
	"fmt"
	"io"
	"log"
	"os"

	"box/internal/config"
	"box/internal/installer"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	spinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	doneStyle    = lipgloss.NewStyle().Margin(1, 2)
	checkMark    = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).SetString("✓")
)

type toolStatus int

const (
	statusPending toolStatus = iota
	statusInstalling
	statusDone
	statusFailed
)

type installMsg struct {
	index int
	err   error
}

type toolTask struct {
	name   string
	status toolStatus
	err    error
}

type model struct {
	tasks    []toolTask
	index    int
	spinner  spinner.Model
	quitting bool
	manager  *installer.Manager
	tools    []config.Tool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.installNext())
}

func (m model) installNext() tea.Cmd {
	if m.index >= len(m.tasks) {
		return tea.Quit
	}

	return func() tea.Msg {
		err := m.manager.Install(m.tools[m.index])
		return installMsg{index: m.index, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	case installMsg:
		if msg.err != nil {
			m.tasks[msg.index].status = statusFailed
			m.tasks[msg.index].err = msg.err
			m.quitting = true
			return m, tea.Quit
		}
		m.tasks[msg.index].status = statusDone
		m.index++
		if m.index < len(m.tasks) {
			m.tasks[m.index].status = statusInstalling
			return m, m.installNext()
		}
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) View() string {
	if len(m.tasks) == 0 {
		return "No tools to install.\n"
	}

	s := "\n  " + lipgloss.NewStyle().Bold(true).Render("Installing Tools") + "\n\n"

	// Define a fixed-width style for the status column to ensure alignment
	statusStyle := lipgloss.NewStyle().Width(2).Align(lipgloss.Center)

	for i, t := range m.tasks {
		status := " "
		if t.status == statusInstalling {
			status = m.spinner.View()
		} else if t.status == statusDone {
			status = checkMark.String()
		} else if t.status == statusFailed {
			status = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("✗")
		}

		s += fmt.Sprintf("  %s %s\n", statusStyle.Render(status), t.name)
		if t.err != nil {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("1")).MarginLeft(6).Render(fmt.Sprintf("Error: %v", t.err)) + "\n"
		}
		_ = i
	}

	if m.quitting {
		return s + "\n"
	}

	if m.index >= len(m.tasks) {
		return s + doneStyle.Render("All tools installed successfully! ✨") + "\n"
	}

	return s + helpStyle.Render("Press q to quit") + "\n"
}

var nonInteractive bool

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Installs tools defined in box.yml",
	Run: func(cmd *cobra.Command, args []string) {
		configFile := "box.yml"
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			log.Fatalf("Configuration file %s not found.", configFile)
		}

		cfg, err := config.Load(configFile)
		if err != nil {
			log.Fatalf("Failed to load %s: %v", configFile, err)
		}

		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}

		mgr := installer.New(cwd, cfg.Env)

		if nonInteractive {
			fmt.Println("Starting tool installation (non-interactive)...")
			for _, tool := range cfg.Tools {
				fmt.Printf("• Installing %s...\n", tool.Source)
				if err := mgr.Install(tool); err != nil {
					fmt.Printf("❌ Failed to install %s: %v\n", tool.Source, err)
					os.Exit(1)
				}
				fmt.Printf("✅ Successfully installed %s\n", tool.Source)
			}
			fmt.Println("All tools installed successfully! ✨")
			return
		}

		mgr.Output = io.Discard

		tasks := make([]toolTask, len(cfg.Tools))
		for i, t := range cfg.Tools {
			tasks[i] = toolTask{name: t.Source, status: statusPending}
		}
		if len(tasks) > 0 {
			tasks[0].status = statusInstalling
		}

		s := spinner.New()
		s.Spinner = spinner.Dot
		s.Style = spinnerStyle

		m := model{
			tasks:   tasks,
			spinner: s,
			manager: mgr,
			tools:   cfg.Tools,
		}

		if _, err := tea.NewProgram(m).Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

func init() {
	installCmd.Flags().BoolVarP(&nonInteractive, "non-interactive", "y", false, "Run in non-interactive mode (no TTY required)")
	rootCmd.AddCommand(installCmd)
}
