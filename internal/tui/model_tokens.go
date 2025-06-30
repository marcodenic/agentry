package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// handleTokenMessages processes token streaming messages
func (m Model) handleTokenMessages(msg tokenMsg) (Model, tea.Cmd) {
	// Check if agent has been stopped - if so, ignore further tokens
	info := m.infos[msg.id]
	if info.Status == StatusStopped {
		return m, nil
	}

	// ENABLED: Real-time token streaming for smooth UX

	// Stop thinking animation on first token
	if !info.TokensStarted {
		info.TokensStarted = true
		info.StreamingResponse = "" // Initialize streaming response
		// No need to clean up spinners since they were never added to history!
	}

	// Add token to streaming response
	info.StreamingResponse += msg.token
	info.TokenCount++
	info.CurrentActivity++ // Just increment counter, let activityTickMsg handle data points

	// Update viewport with streaming content - OPTIMIZED for performance
	if msg.id == m.active {
		// PERFORMANCE FIX: Update every 10 characters, or immediately on formatting characters
		// This preserves newlines and other formatting while maintaining good performance
		shouldUpdate := len(info.StreamingResponse)%10 == 0 ||
			strings.HasSuffix(msg.token, "\n") ||
			strings.HasSuffix(msg.token, " ") ||
			strings.HasSuffix(msg.token, "\t") ||
			strings.HasSuffix(msg.token, "\r")

		if shouldUpdate {
			// Build display history with properly formatted streaming response
			displayHistory := info.History
			if info.StreamingResponse != "" {
				// FIXED: Always use proper formatting to preserve newlines and wrapping
				formattedStreamingResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
				displayHistory += formattedStreamingResponse
			}
			m.vp.SetContent(displayHistory)
			m.vp.GotoBottom()
		}
	}

	now := time.Now()

	// Token history update for activity monitoring
	if info.LastToken.IsZero() || now.Sub(info.LastToken) > time.Second {
		info.TokenHistory = append(info.TokenHistory, 1)
		if len(info.TokenHistory) > 20 {
			info.TokenHistory = info.TokenHistory[1:]
		}
	} else if len(info.TokenHistory) > 0 {
		info.TokenHistory[len(info.TokenHistory)-1]++
	}
	info.LastToken = now
	m.infos[msg.id] = info // Save the updated info back to the map

	// Continue reading trace stream for more events (including EventFinal)
	return m, m.readCmd(msg.id)
}

// handleTokenStream processes token streaming animation
func (m Model) handleTokenStream(msg startTokenStream) (Model, tea.Cmd) {
	return m, tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
		return tokenStreamTick{id: msg.id, runes: msg.runes, position: 0}
	})
}

// handleTokenStreamTick processes individual token stream ticks
func (m Model) handleTokenStreamTick(msg tokenStreamTick) (Model, tea.Cmd) {
	if msg.position >= len(msg.runes) {
		// Streaming complete
		return m, nil
	}

	token := string(msg.runes[msg.position])

	// Schedule next character
	nextCmd := tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
		return tokenStreamTick{id: msg.id, runes: msg.runes, position: msg.position + 1}
	})

	// Process current token and schedule next
	newModel, _ := m.Update(tokenMsg{id: msg.id, token: token})
	return newModel.(Model), nextCmd
}

// handleFinalMessage processes final completion messages
func (m Model) handleFinalMessage(msg finalMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	
	// Determine if we need spacing before the AI response
	needsSpacing := false
	if info.History != "" {
		// Check if the last content was a status message (contains status bar emoji)
		lines := strings.Split(strings.TrimRight(info.History, "\n"), "\n")
		if len(lines) > 0 {
			lastLine := lines[len(lines)-1]
			// If last line contains status indicators, we need spacing
			if strings.Contains(lastLine, "🤖") || strings.Contains(lastLine, "✅") || strings.Contains(lastLine, "📁") || strings.Contains(lastLine, "⚡") || strings.Contains(lastLine, "🔍") || strings.Contains(lastLine, "🌐") {
				needsSpacing = true
			}
		}
	}
	
	// Add the final AI response with proper spacing
	spacingPrefix := ""
	if needsSpacing {
		spacingPrefix = "\n\n"  // Double spacing after status messages
	} else {
		spacingPrefix = "\n"    // Single spacing for other cases
	}
	
	if info.StreamingResponse != "" {
		formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
		info.History += spacingPrefix + formattedResponse
	} else if msg.text != "" {
		// Fallback to final message text if no streaming occurred
		formattedResponse := m.formatWithBar(m.aiBar(), msg.text, m.vp.Width)
		info.History += spacingPrefix + formattedResponse
	}
	info.StreamingResponse = "" // Clear streaming response

	// Limit history length to prevent unbounded memory growth (keep last ~100KB)
	const maxHistoryLength = 100000
	if len(info.History) > maxHistoryLength {
		// Keep last 75% of history to maintain context
		keepLength := maxHistoryLength * 3 / 4
		info.History = "...[earlier messages truncated]...\n" + info.History[len(info.History)-keepLength:]
	}

	// Set status to idle, clear spinner, and add spacing after AI message for next user input
	info.Status = StatusIdle
	info.TokensStarted = false // Reset streaming state
	info.History += "\n"       // Add spacing after AI response for next user input

	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, nil
}
