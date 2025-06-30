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
	
	// Group status messages together - only add spacing if we're starting a new group
	lastChar := ""
	if len(info.History) > 0 {
		lastChar = info.History[len(info.History)-1:]
	}
	
	// If the last thing was a user message or AI response (ends with newline), add spacing before status group
	if lastChar == "\n" || info.History == "" {
		info.History += "\n" + commandFormatted
	} else {
		// If we're continuing status messages, just add to the group
		info.History += "\n" + commandFormatted
	}
	
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
	
	// Group status messages together - only add spacing if we're starting a new group
	lastChar := ""
	if len(info.History) > 0 {
		lastChar = info.History[len(info.History)-1:]
	}
	
	// If the last thing was a user message or AI response (ends with newline), add spacing before status group
	if lastChar == "\n" || info.History == "" {
		info.History += "\n" + actionFormatted
	} else {
		// If we're continuing status messages, just add to the group
		info.History += "\n" + actionFormatted
	}
	
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
