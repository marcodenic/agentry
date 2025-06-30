package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleAgentComplete processes agent completion messages
func (m Model) handleAgentComplete(msg agentCompleteMsg) (Model, tea.Cmd) {
	// finalMsg already handled completion - this is just cleanup
	info := m.infos[msg.id]
	if info.Status != StatusIdle {
		info.Status = StatusIdle
		m.infos[msg.id] = info
	}
	return m, nil
}

// handleThinkingAnimation processes thinking animation messages
func (m Model) handleThinkingAnimation(msg thinkingAnimationMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	// Stop thinking animation if tokens have started or agent is not running
	if info.Status != StatusRunning || info.TokensStarted {
		// When stopping thinking animation, just refresh display with clean history
		if msg.id == m.active {
			m.vp.SetContent(info.History)
			m.vp.GotoBottom()
		}
		return m, nil
	}

	// ASCII spinner frames
	frames := []string{"|", "/", "-", "\\"}
	currentSpinner := frames[msg.frame]

	// Build display content WITHOUT modifying history
	displayHistory := info.History

	// Add spinner to display only if history doesn't end with newline
	if len(displayHistory) > 0 && !strings.HasSuffix(displayHistory, "\n") {
		// Add AI bar and spinner for display only
		displayHistory += "\n" + m.aiBar() + " " + currentSpinner
	} else {
		// Add AI bar and spinner for display only
		displayHistory += m.aiBar() + " " + currentSpinner
	}

	if msg.id == m.active {
		m.vp.SetContent(displayHistory)
		m.vp.GotoBottom()
	}

	// Continue the animation if still running and no tokens have started
	return m, startThinkingAnimation(msg.id)
}
