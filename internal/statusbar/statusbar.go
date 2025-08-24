// Package statusbar provides a statusbar bubble which can render
// four different status sections
package statusbar

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
)

// Height represents the height of the statusbar.
const Height = 1

// ColorConfig represents the color configuration for a column.
type ColorConfig struct {
	Foreground lipgloss.AdaptiveColor
	Background lipgloss.AdaptiveColor
}

// Model represents the properties of the statusbar.
type Model struct {
	Width              int
	Height             int
	FirstColumn        string
	SecondColumn       string
	ThirdColumn        string
	FourthColumn       string
	FirstColumnColors  ColorConfig
	SecondColumnColors ColorConfig
	ThirdColumnColors  ColorConfig
	FourthColumnColors ColorConfig
}

// New creates a new instance of the statusbar.
func New(firstColumnColors, secondColumnColors, thirdColumnColors, fourthColumnColors ColorConfig) Model {
	return Model{
		Height:             Height,
		FirstColumnColors:  firstColumnColors,
		SecondColumnColors: secondColumnColors,
		ThirdColumnColors:  thirdColumnColors,
		FourthColumnColors: fourthColumnColors,
	}
}

// SetSize sets the width of the statusbar.
func (m *Model) SetSize(width int) {
	m.Width = width
}

// Update updates the size of the statusbar.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width)
	}

	return m, nil
}

// SetContent sets the content of the statusbar.
func (m *Model) SetContent(firstColumn, secondColumn, thirdColumn, fourthColumn string) {
	m.FirstColumn = firstColumn
	m.SecondColumn = secondColumn
	m.ThirdColumn = thirdColumn
	m.FourthColumn = fourthColumn
}

// SetColors sets the colors of the 4 columns.
func (m *Model) SetColors(firstColumnColors, secondColumnColors, thirdColumnColors, fourthColumnColors ColorConfig) {
	m.FirstColumnColors = firstColumnColors
	m.SecondColumnColors = secondColumnColors
	m.ThirdColumnColors = thirdColumnColors
	m.FourthColumnColors = fourthColumnColors
}

// View returns a string representation of a statusbar with advanced styling.
func (m Model) View() string {
	if m.Width <= 0 {
		return ""
	}

	width := lipgloss.Width

	// Create sophisticated styles for each column with subtle effects
	firstColumn := lipgloss.NewStyle().
		Foreground(m.FirstColumnColors.Foreground).
		Background(m.FirstColumnColors.Background).
		Padding(0, 1).
		Bold(true).
		Render(truncate.StringWithTail(m.FirstColumn, 30, "..."))

	thirdColumn := lipgloss.NewStyle().
		Foreground(m.ThirdColumnColors.Foreground).
		Background(m.ThirdColumnColors.Background).
		Align(lipgloss.Right).
		Padding(0, 1).
		Bold(true).
		Render(m.ThirdColumn)

	fourthColumn := lipgloss.NewStyle().
		Foreground(m.FourthColumnColors.Foreground).
		Background(m.FourthColumnColors.Background).
		Padding(0, 1).
		Bold(true).
		Render(m.FourthColumn)

	// Calculate remaining width for the expandable middle column
	remainingWidth := m.Width - width(firstColumn) - width(thirdColumn) - width(fourthColumn)
	if remainingWidth < 0 {
		remainingWidth = 0
	}

	// Special styling for the CWD section (middle) - make it stand out with italics
	secondColumn := lipgloss.NewStyle().
		Foreground(m.SecondColumnColors.Foreground).
		Background(m.SecondColumnColors.Background).
		Padding(0, 1).
		Width(remainingWidth).
		Italic(true).
		Bold(true).
		Render(truncate.StringWithTail(
			m.SecondColumn,
			uint(remainingWidth-2), // Account for padding
			"..."),
		)

	// Join all columns with subtle transitions
	statusBar := lipgloss.JoinHorizontal(lipgloss.Top,
		firstColumn,
		secondColumn,
		thirdColumn,
		fourthColumn,
	)

	// Apply overall status bar styling with clean, borderless design
	finalStyle := lipgloss.NewStyle().
		Width(m.Width).
		Align(lipgloss.Left)

	return finalStyle.Render(statusBar)
}
