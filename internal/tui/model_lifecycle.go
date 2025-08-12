package tui

import (
	"fmt"
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

// handleAgentStart processes agent start messages for better UI feedback
func (m Model) handleAgentStart(msg agentStartMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	if info != nil {
		info.Status = StatusRunning
		// Add a status message to the conversation
		statusMsg := fmt.Sprintf("\n\n✨ **%s** (%s) is starting to work...\n", msg.name, msg.role)
		info.History += statusMsg
		m.infos[msg.id] = info
		
		// If this is the active agent, update the viewport
		if msg.id == m.active {
			m.vp.SetContent(info.History)
			m.vp.GotoBottom()
		}
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

	// Dots spinner frames (replacing slash spinner)
	frames := []string{"   ", "•  ", "•• ", "•••"}
	currentSpinner := frames[msg.frame%len(frames)]

	// Build display content WITHOUT modifying history
	displayHistory := info.History

	// Check if we should append spinner to last status message or show on new line
	if info.LastContentType == ContentTypeStatusMessage {
		// Append spinner to the end of the last status message
		displayHistory += " " + currentSpinner
	} else {
		// For user input or other content types, show spinner on new line with AI bar
		if len(displayHistory) > 0 && !strings.HasSuffix(displayHistory, "\n") {
			// Add AI bar and spinner for display only
			displayHistory += "\n" + m.aiBar() + " " + currentSpinner
		} else {
			// Add AI bar and spinner for display only
			displayHistory += m.aiBar() + " " + currentSpinner
		}
	}

	if msg.id == m.active {
		m.vp.SetContent(displayHistory)
		m.vp.GotoBottom()
	}

	// Continue the animation if still running and no tokens have started
	return m, startThinkingAnimation(msg.id)
}
