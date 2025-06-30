package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// handleKeyMessages processes keyboard input and navigation
func (m Model) handleKeyMessages(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Enhanced agent navigation
	if key.Matches(msg, PrevAgentKey) {
		return m.cycleActive(-1), nil
	} else if key.Matches(msg, NextAgentKey) {
		return m.cycleActive(1), nil
	} else if key.Matches(msg, FirstAgentKey) {
		return m.jumpToAgent(0), nil
	} else if key.Matches(msg, LastAgentKey) {
		return m.jumpToAgent(len(m.order) - 1), nil
	}

	switch msg.String() {
	case m.keys.Quit:
		return m.handleQuit()
	case m.keys.ToggleTab:
		return m.handleToggleTab()
	case m.keys.Pause:
		return m.handlePause()
	case m.keys.Submit:
		return m.handleSubmit()
	}

	return m, nil
}

// handleQuit processes quit command and cleans up running agents
func (m Model) handleQuit() (Model, tea.Cmd) {
	// Clean up all running agents before quitting
	for id, info := range m.infos {
		if info.Cancel != nil {
			info.Cancel() // Cancel all running agent contexts
		}
		if info.Status == StatusRunning {
			info.Status = StatusStopped
			m.infos[id] = info
		}
	}
	return m, tea.Quit
}

// handleToggleTab switches between chat and debug tabs
func (m Model) handleToggleTab() (Model, tea.Cmd) {
	m.activeTab = 1 - m.activeTab
	// When switching to debug tab, update debug viewport content
	if m.activeTab == 1 {
		if info, ok := m.infos[m.active]; ok {
			debugContent := m.renderDetailedMemory(info.Agent)
			m.debugVp.SetContent(debugContent)
			m.debugVp.GotoBottom() // Start at bottom to see latest events
		}
	}
	return m, nil
}

// handlePause stops the active agent if running
func (m Model) handlePause() (Model, tea.Cmd) {
	// Stop the active agent if it's running or streaming
	if info, ok := m.infos[m.active]; ok && (info.Status == StatusRunning || info.TokensStarted) {
		if info.Cancel != nil {
			info.Cancel() // Cancel the agent's context
		}

		// Clean up streaming response if in progress
		if info.StreamingResponse != "" {
			// Add the partial streaming response to history before stopping
			formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
			info.History += formattedResponse
			info.StreamingResponse = ""
		}

		info.Status = StatusIdle   // Set to idle so new messages can be sent
		info.TokensStarted = false // Reset streaming state
		info.History += "\n\n" + m.statusBar() + " Agent stopped by user\n"
		m.infos[m.active] = info

		// Update viewport to show the stop message
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	return m, nil
}

// handleSubmit processes input submission
func (m Model) handleSubmit() (Model, tea.Cmd) {
	if m.input.Focused() {
		// Check if the active agent is running before accepting new input
		if info, ok := m.infos[m.active]; ok && info.Status == StatusRunning {
			// Agent is busy, ignore input
			return m, nil
		}

		txt := m.input.Value()
		m.input.SetValue("")
		if strings.HasPrefix(txt, "/") {
			var cmd tea.Cmd
			m, cmd = m.handleCommand(txt)
			return m, cmd
		}
		return m.startAgent(m.active, txt)
	}
	return m, nil
}
