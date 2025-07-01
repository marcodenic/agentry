package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// handleToolUseMessage processes tool usage messages (tool completion)
func (m Model) handleToolUseMessage(msg toolUseMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	info.CurrentTool = msg.name

	// Complete the progressive status update (add green tick and change bar color)
	info.completeProgressiveStatusUpdate(m)

	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// handleActionMessage processes action notification messages (tool start)
func (m Model) handleActionMessage(msg actionMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]

	// Start progressive status update with orange bar
	info.startProgressiveStatusUpdate(msg.text, m)

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
