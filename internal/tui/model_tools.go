package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleToolUseMessage processes tool usage messages
func (m Model) handleToolUseMessage(msg toolUseMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	info.CurrentTool = msg.name
	// Show completion message with clean formatting
	completionText := m.formatToolCompletion(msg.name, msg.args)
	commandFormatted := m.formatSingleCommand(completionText)
	// Add consistent spacing
	info.History += "\n" + commandFormatted
	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// handleActionMessage processes action notification messages
func (m Model) handleActionMessage(msg actionMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	// Add action messages with clean formatting
	actionFormatted := m.formatSingleCommand(msg.text)
	// Add consistent spacing
	info.History += "\n" + actionFormatted
	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// handleModelMessage processes model information messages
func (m Model) handleModelMessage(msg modelMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	info.ModelName = msg.name
	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}
