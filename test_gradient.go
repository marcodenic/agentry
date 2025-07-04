package main

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type testModel struct {
	tokenCount int
	maxTokens  int
}

func (m testModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return "tick"
	})
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	case string:
		if msg == "tick" {
			m.tokenCount += 100
			if m.tokenCount > m.maxTokens {
				m.tokenCount = 0
			}
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return "tick"
			})
		}
	}
	return m, nil
}

func (m testModel) View() string {
	pct := float64(m.tokenCount) / float64(m.maxTokens)

	// Simulate the color logic from renderTokenBar
	var fillColor string
	if pct <= 0.3 {
		fillColor = "#22C55E" // Green
	} else if pct <= 0.7 {
		fillColor = "#EAB308" // Yellow/Orange
	} else {
		fillColor = "#EF4444" // Red
	}

	// Create a simple bar representation
	barWidth := 20
	filledWidth := int(pct * float64(barWidth))

	bar := ""
	for i := 0; i < filledWidth; i++ {
		bar += "█"
	}
	for i := filledWidth; i < barWidth; i++ {
		bar += "░"
	}

	coloredBar := lipgloss.NewStyle().Foreground(lipgloss.Color(fillColor)).Render(bar)

	return fmt.Sprintf("Token Progress Test\n\nTokens: %d/%d (%.1f%%)\nColor: %s\n\n%s\n\nPress 'q' to quit",
		m.tokenCount, m.maxTokens, pct*100, fillColor, coloredBar)
}

func main() {
	p := tea.NewProgram(testModel{
		tokenCount: 0,
		maxTokens:  8000,
	})

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
