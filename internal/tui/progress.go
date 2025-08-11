package tui

import (
	"github.com/charmbracelet/bubbles/progress"
)

// createTokenProgressBar returns a progress bar configured for token usage.
// Percentage text is rendered separately alongside the bar, so we disable it here.
func createTokenProgressBar() progress.Model {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithoutPercentage(),
	)
	return p
}
