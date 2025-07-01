package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// handleErrorMessage processes error messages with detailed analysis
func (m Model) handleErrorMessage(msg errMsg) (Model, tea.Cmd) {
	m.err = msg
	if info, ok := m.infos[m.active]; ok {
		// Immediately clear spinner and set error status
		info.Status = StatusError
		info.TokensStarted = false

		// No spinner cleanup needed since spinners are display-only now!

		// Create detailed error message with context and better debugging
		var errorMsg string
		errorStr := msg.Error()

		// Enhanced error analysis with better wrapping
		if strings.Contains(errorStr, "cannot create agent with tool name") && strings.Contains(errorStr, "view") {
			errorMsg = "❌ Error: Agent trying to create 'view' agent\n"
			errorMsg += "   Context: Tool names like 'view', 'write', 'search' are reserved\n"
			errorMsg += "   💡 Fix: Agent should use 'view filename' directly, not create a 'view' agent\n"
			errorMsg += "   💡 This indicates the agent prompt needs adjustment"
		} else if strings.Contains(errorStr, "cannot create agent with tool name") {
			// Extract the tool name from the error
			errorMsg = "❌ Error: Agent trying to create agent with reserved tool name\n"
			errorMsg += "   Context: " + errorStr + "\n"
			errorMsg += "   💡 Fix: Use the tool directly instead of creating an agent with that name"
		} else if strings.Contains(errorStr, "fetch") && strings.Contains(errorStr, "exit status 1") {
			if strings.Contains(errorStr, "roadmap.md") {
				errorMsg = "❌ Error: fetch tool called with local file path instead of URL\n"
				errorMsg += "   Context: Tool 'fetch' requires URLs (http/https), not local file paths\n"
				errorMsg += "   💡 Tip: Use 'view' tool for local files, 'fetch' for web URLs"
			} else {
				errorMsg = "❌ Error: fetch tool execution failed\n"
				errorMsg += "   💡 Tip: Check network connectivity and URL validity"
			}
		} else if strings.Contains(errorStr, "agent") && strings.Contains(errorStr, "tool") && strings.Contains(errorStr, "execution failed") {
			// Split error to show the main error and context separately
			parts := strings.SplitN(errorStr, ": ", 2)
			if len(parts) == 2 {
				errorMsg = fmt.Sprintf("❌ Error: %s\n   Context: %s", parts[0], parts[1])
			} else {
				errorMsg = fmt.Sprintf("❌ Error: %s", errorStr)
			}

			// Add specific tips based on error content
			if strings.Contains(errorStr, "exit status") {
				errorMsg += "\n   💡 Tip: Tool or command execution failed - check syntax and permissions"
			} else if strings.Contains(errorStr, "unknown tool") {
				errorMsg += "\n   💡 Tip: Agent tried to use a tool that doesn't exist"
			}
		} else if strings.Contains(errorStr, "max iterations") {
			errorMsg = fmt.Sprintf("❌ Error: %s", errorStr)
			errorMsg += "\n   💡 Tip: Agent reached iteration limit - try simplifying the request"
		} else {
			errorMsg = fmt.Sprintf("❌ Error: %s", errorStr)
		}

		errorFormatted := m.formatSingleCommand(errorMsg)
		info.addContentWithSpacing(errorFormatted, ContentTypeStatusMessage)

		// Update viewport if this is the active agent
		if m.active == info.Agent.ID {
			m.vp.SetContent(info.History)
			m.vp.GotoBottom()
		}

		m.infos[m.active] = info
	}
	return m, nil
}
