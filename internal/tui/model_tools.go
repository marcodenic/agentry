package tui

import (
	"strings"
	
	tea "github.com/charmbracelet/bubbletea"
)

// handleToolUseMessage processes tool usage messages
func (m Model) handleToolUseMessage(msg toolUseMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	info.CurrentTool = msg.name
	// Show completion message with better formatting
	completionText := m.formatToolCompletion(msg.name, msg.args)
	commandFormatted := m.formatSingleCommand(completionText)
	// Add spacing before first status message in a sequence
	if !strings.HasSuffix(info.History, "\n") && info.History != "" {
		info.History += "\n"
	}
	info.History += commandFormatted
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
	// Add action messages with better spacing
	actionFormatted := m.formatSingleCommand(msg.text)
	// Add spacing before first status message in a sequence
	if !strings.HasSuffix(info.History, "\n") && info.History != "" {
		info.History += "\n"
	}
	info.History += actionFormatted
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
