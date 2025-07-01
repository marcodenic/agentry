// Package glyphs provides Unicode glyphs for the TUI interface.
// Based on glyphs from github.com/maaslalani/glyphs
package glyphs

import "github.com/charmbracelet/lipgloss"

// Status and progress glyphs
const (
	// Basic shapes
	CircleFilled = "●"
	CircleEmpty  = "○"
	BulletFilled = "•"
	BulletEmpty  = "◦"

	// Checkmarks and crosses
	Checkmark = "✔"
	Crossmark = "✕"

	// Arrows
	ArrowUp    = "↑"
	ArrowRight = "→"
	ArrowDown  = "↓"
	ArrowLeft  = "←"
	Triangle   = "▶"

	// Progress and loading
	Ellipsis         = "⋯"
	VerticalEllipsis = "⋮"

	// Stars and special
	StarFilled = "★"
	StarEmpty  = "☆"
	Star       = "✢"
	Diamond    = "❖"
	Sparkle    = "✦"

	// Symbols
	Lightning = "ϟ"
	Flag      = "⚐"
	Sun       = "☼"
	Moon      = "☾"
	Point     = "☞"

	// Lines and bars
	HorizontalLines = "☰"
)

// Status represents different status states
type Status struct {
	Glyph string
	Name  string
}

// Predefined status types
var (
	StatusPending    = Status{CircleEmpty, "pending"}
	StatusRunning    = Status{CircleFilled, "running"}
	StatusComplete   = Status{Checkmark, "complete"}
	StatusError      = Status{Crossmark, "error"}
	StatusProcessing = Status{BulletFilled, "processing"}
)

// GetStatusGlyph returns the appropriate glyph for a status
func GetStatusGlyph(isComplete bool, hasError bool) string {
	if hasError {
		return StatusError.Glyph
	}
	if isComplete {
		return StatusComplete.Glyph
	}
	return StatusRunning.Glyph
}

// GetProgressGlyph returns the appropriate glyph for progress indication
func GetProgressGlyph(step int) string {
	glyphs := []string{CircleEmpty, BulletEmpty, BulletFilled, CircleFilled}
	if step < 0 {
		step = 0
	}
	if step >= len(glyphs) {
		step = len(glyphs) - 1
	}
	return glyphs[step]
}

// Styled glyph functions with colors and bold text
func GreenCheckmark() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#32CD32")).
		Bold(true).
		Render(Checkmark)
}

func RedCrossmark() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF4444")).
		Bold(true).
		Render(Crossmark)
}

func OrangeTriangle() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8C00")).
		Bold(true).
		Render(Triangle)
}

func BlueCircle() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#4A9FFF")).
		Bold(true).
		Render(CircleFilled)
}

func YellowStar() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700")).
		Bold(true).
		Render(Star)
}

func OrangeLightning() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8C00")).
		Bold(true).
		Render(Lightning)
}
