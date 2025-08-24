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

	// Special handling for diagnostics tool to store structured results
	if msg.name == "lsp_diagnostics" {
		if r, ok := msg.args["result"].(map[string]any); ok {
			// reset list
			m.diags = nil
			if arr, ok := r["diagnostics"].([]any); ok {
				for _, it := range arr {
					if m2, ok := it.(map[string]any); ok {
						d := Diag{
							File:     strVal(m2["file"]),
							Line:     intVal(m2["line"]),
							Col:      intVal(m2["col"]),
							Code:     strVal(m2["code"]),
							Severity: strVal(m2["severity"]),
							Message:  strVal(m2["message"]),
						}
						m.diags = append(m.diags, d)
					}
				}
			}
		}
		m.diagRunning = false
	}

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

	// If we have a streamed assistant response that hasn't been finalized yet (because
	// the model produced tool calls), commit it to the permanent history BEFORE we
	// append the tool action status line. Previously the plan/response was only kept
	// in StreamingResponse (ephemeral) and cleared from the viewport once we rendered
	// the first tool action, giving the illusion that the assistant "thought" text
	// disappeared. Persisting it here preserves the initial plan exactly as shown.
	if info.StreamingResponse != "" {
		formatted := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
		info.addContentWithSpacing(formatted, ContentTypeAIResponse)
		info.StreamingResponse = "" // mark as committed
		info.TokensStarted = false  // reset so future streaming cycles behave normally
	}

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

	// Simply update the model name if we have a new one
	if msg.name != "" {
		info.ModelName = msg.name
	}

	m.infos[msg.id] = info
	return m, m.readCmd(msg.id)
}

// Helpers to decode numbers/strings from any
func strVal(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
func intVal(v any) int {
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}
