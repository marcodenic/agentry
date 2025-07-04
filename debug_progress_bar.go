package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	prog progress.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m, tea.Quit
	case tea.WindowSizeMsg:
		m.prog.Width = msg.Width - 4
		return m, nil
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("Progress bar test:\n%s\nPress any key to exit", m.prog.View())
}

func main() {
	// Test in terminal mode vs non-terminal mode
	fmt.Println("=== NON-TUI TEST ===")
	prog := progress.New(progress.WithDefaultGradient())
	prog.Width = 30
	prog.SetPercent(0.5)
	fmt.Printf("50%% progress: '%s'\n", prog.View())
	
	// Check if we're in a terminal
	if !isTerminal() {
		fmt.Println("Not in a terminal, skipping TUI test")
		return
	}
	
	fmt.Println("\n=== TUI TEST ===")
	fmt.Println("Starting TUI mode...")
	
	// Test in TUI mode
	prog.Width = 50
	prog.SetPercent(0.75)
	
	p := tea.NewProgram(model{prog: prog})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func isTerminal() bool {
	stat, _ := os.Stdout.Stat()
	return (stat.Mode() & os.ModeCharDevice) != 0
}
