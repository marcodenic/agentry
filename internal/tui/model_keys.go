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

	// Input history navigation when input is focused
	if m.view.Input.Focused() {
		switch msg.String() {
		case "up":
			m.view.Input.HistoryUp()
			return m, nil
		case "down":
			m.view.Input.HistoryDown()
			return m, nil
		}
	}

	switch msg.String() {
	case m.keys.Quit:
		return m.handleQuit()
	case m.keys.ToggleTab:
		return m.handleToggleTab()
	case m.keys.Pause:
		return m.handlePause()
	case m.keys.Diagnostics:
		return m.handleDiagnostics()
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
	m.layout.activeTab = 1 - m.layout.activeTab
	// When switching to debug tab, update debug viewport content
	if m.layout.activeTab == 1 {
		if info, ok := m.infos[m.active]; ok {
			debugContent := m.renderDetailedMemory(info.Agent)
			m.view.Chat.Debug.SetContent(debugContent)
			m.view.Chat.Debug.GotoBottom() // Start at bottom to see latest events
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
			formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.view.Chat.Main.Width)
			info.addContentWithSpacing(formattedResponse, ContentTypeAIResponse)
			info.StreamingResponse = ""
		}

		info.Status = StatusIdle                                   // Set to idle so new messages can be sent
		info.TokensStarted = false                                 // Reset streaming state
		stopMessage := m.statusBar() + "    Agent stopped by user" // Use 4 spaces for alignment
		info.addContentWithSpacing(stopMessage, ContentTypeStatusMessage)
		m.infos[m.active] = info

		// Update viewport to show the stop message
		m.view.Chat.Main.SetContent(info.History)
		m.view.Chat.Main.GotoBottom()
	}
	return m, nil
}

// handleSubmit processes input submission
func (m Model) handleSubmit() (Model, tea.Cmd) {
	if m.view.Input.Focused() {
		// Check if the active agent is running before accepting new input
		if info, ok := m.infos[m.active]; ok && info.Status == StatusRunning {
			// Agent is busy, ignore input
			return m, nil
		}

		txt := strings.TrimSpace(m.view.Input.Value())
		if txt == "" {
			return m, nil
		}
		m.view.Input.PushHistory(txt)
		m.view.Input.ResetAfterSend()
		// ALL input goes through Agent 0's natural language processing
		// No slash commands - everything is handled by delegation
		return m.startAgent(m.active, txt)
	}
	return m, nil
}
