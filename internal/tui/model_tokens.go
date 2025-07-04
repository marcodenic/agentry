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
		// Initialize live token count based on agent's current count
		if info.Agent != nil && info.Agent.Cost != nil {
			info.StreamingTokenCount = info.Agent.Cost.TotalTokens()
		} else {
			info.StreamingTokenCount = 0
		}
		// No need to clean up spinners since they were never added to history!
	}

	// Add token to streaming response
	info.StreamingResponse += msg.token
	// Count tokens live during streaming for responsive UI
	info.StreamingTokenCount++
	info.CurrentActivity++ // Just increment counter, let activityTickMsg handle data points

	// Save updated info back to map
	m.infos[msg.id] = info

	// Update progress bar to match the percentage that will be shown on tokens line
	var progressCmd tea.Cmd
	if info.StreamingTokenCount%5 == 0 { // Update every 5 tokens for performance
		// Use SAME calculation that agent_panel.go uses for the tokens line
		maxTokens := 8000
		if info.ModelName != "" && strings.Contains(strings.ToLower(info.ModelName), "gpt-4") {
			maxTokens = 128000
		}

		actualTokens := 0
		if info.Agent != nil && info.Agent.Cost != nil {
			if info.TokensStarted && info.StreamingResponse != "" {
				actualTokens = info.StreamingTokenCount
			} else {
				actualTokens = info.Agent.Cost.TotalTokens()
			}
		}

		// Same exact calculation as agent_panel.go
		tokenPct := float64(actualTokens) / float64(maxTokens) * 100
		pct := tokenPct / 100.0
		if pct < 0 {
			pct = 0
		}
		if pct > 1 {
			pct = 1
		}

		// Don't show any filled area until we reach 5% to avoid showing wrong colors
		// The bubbles/progress library shows yellow/orange at very low percentages
		// instead of the expected green, so better to show empty until 5%
		if pct < 0.05 {
			pct = 0
		}

		progressCmd = info.TokenProgress.SetPercent(pct)
	}

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
				// Apply proper spacing logic for streaming AI response
				formattedStreamingResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)

				// Determine spacing based on last content type (same logic as addContentWithSpacing)
				spacing := ""
				switch info.LastContentType {
				case ContentTypeUserInput:
					// User Input → AI Response: No extra spacing during streaming
					spacing = "\n"
				case ContentTypeStatusMessage:
					// Status Message → AI Response: Add spacing during streaming
					spacing = "\n\n"
				default:
					spacing = "\n"
				}

				displayHistory += spacing + formattedStreamingResponse
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

	// Cost is now handled directly by the agent's cost manager
	// No TUI-side cost tracking needed

	m.infos[msg.id] = info // Save the updated info back to the map after token history update

	// Continue reading trace stream for more events (including EventFinal)
	var cmds []tea.Cmd
	cmds = append(cmds, m.readCmd(msg.id))
	if progressCmd != nil {
		cmds = append(cmds, progressCmd)
	}
	return m, tea.Batch(cmds...)
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

	// Add the final AI response using proper content tracking
	if info.StreamingResponse != "" {
		formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
		info.addContentWithSpacing(formattedResponse, ContentTypeAIResponse)
	} else if msg.text != "" {
		// Fallback to final message text if no streaming occurred
		formattedResponse := m.formatWithBar(m.aiBar(), msg.text, m.vp.Width)
		info.addContentWithSpacing(formattedResponse, ContentTypeAIResponse)
	}
	info.StreamingResponse = "" // Clear streaming response

	// Limit history length to prevent unbounded memory growth (keep last ~100KB)
	const maxHistoryLength = 100000
	if len(info.History) > maxHistoryLength {
		// Keep last 75% of history to maintain context
		keepLength := maxHistoryLength * 3 / 4
		info.History = "...[earlier messages truncated]...\n" + info.History[len(info.History)-keepLength:]
		// After truncation, we don't know the last content type, so reset it
		info.LastContentType = ContentTypeEmpty
	}

	// Set status to idle, clear spinner
	info.Status = StatusIdle
	info.TokensStarted = false // Reset streaming state

	// Reconcile streaming token count with final API response
	if info.Agent != nil && info.Agent.Cost != nil {
		// The agent's cost manager has now been updated with real API tokens
		// Reset streaming count to match real count for consistency
		info.StreamingTokenCount = info.Agent.Cost.TotalTokens()
	}

	// Cost is now handled directly by the agent's cost manager
	// No TUI-side cost tracking needed

	if msg.id == m.active {
		m.vp.SetContent(info.History)
		m.vp.GotoBottom()
	}
	m.infos[msg.id] = info
	return m, nil
}
