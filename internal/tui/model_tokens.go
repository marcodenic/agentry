package tui

import (
	"os"
	"strconv"
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

	// Input streaming completes once the model starts returning tokens
	if info.InputActive {
		info.InputActive = false
	}

	// Add token to streaming response
	info.StreamingResponse += msg.token
	// Count tokens live during streaming for responsive UI
	info.StreamingTokenCount++
	info.CurrentOutputActivity++ // Track output-side activity for sparkline updates
	info.OutputActive = true

	// Save updated info back to map
	m.infos[msg.id] = info

	// Update progress bar to match the percentage that will be shown on tokens line
	var progressCmd tea.Cmd
	if info.StreamingTokenCount%5 == 0 { // Update every 5 tokens for performance
		progressCmd = m.updateTokenProgress(info, info.LiveTokenCount())
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
				formattedStreamingResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.view.Chat.Main.Width)

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
				}

				displayHistory += spacing + formattedStreamingResponse
			}

			// Only autoscroll if we were at the bottom prior to this update.
			// If the bubbles version lacks AtBottom(), fall back to always autoscroll.
			wasAtBottom := false
			type atBottomCap interface{ AtBottom() bool }
			if ab, ok := interface{}(m.view.Chat.Main).(atBottomCap); ok {
				wasAtBottom = ab.AtBottom()
			} else {
				// Fallback behavior: assume at bottom (maintains previous behavior)
				wasAtBottom = true
			}
			m.view.Chat.Main.SetContent(displayHistory)
			if wasAtBottom {
				m.view.Chat.Main.GotoBottom()
			}
		}
	}

	now := time.Now()
	info.LastActivity = now

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
	cmds = append(cmds, m.runtime.ReadCmd(&m, msg.id))
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
		formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.view.Chat.Main.Width)
		info.addContentWithSpacing(formattedResponse, ContentTypeAIResponse)
	} else if msg.text != "" {
		// Fallback to final message text if no streaming occurred
		formattedResponse := m.formatWithBar(m.aiBar(), msg.text, m.view.Chat.Main.Width)
		info.addContentWithSpacing(formattedResponse, ContentTypeAIResponse)
	}
	info.StreamingResponse = "" // Clear streaming response

	// Optional: limit history length via env var AGENTRY_HISTORY_LIMIT (bytes)
	if limStr := os.Getenv("AGENTRY_HISTORY_LIMIT"); limStr != "" {
		if maxLen, err := strconv.Atoi(limStr); err == nil && maxLen > 0 {
			if len(info.History) > maxLen {
				// Keep last 75% of history to maintain context
				keepLength := (maxLen * 3) / 4
				if keepLength < 0 {
					keepLength = 0
				}
				if keepLength > len(info.History) {
					keepLength = len(info.History)
				}
				info.History = "...[earlier messages truncated]...\n" + info.History[len(info.History)-keepLength:]
				// After truncation, we don't know the last content type, so reset it
				info.LastContentType = ContentTypeEmpty
			}
		}
	}

	// Set status to idle, clear spinner and elapsed timer
	info.Status = StatusIdle
	info.TokensStarted = false

	// Reconcile streaming token count with final API response
	if info.Agent != nil && info.Agent.Cost != nil {
		info.StreamingTokenCount = info.Agent.Cost.TotalTokens()
	}

	progressCmd := m.updateTokenProgress(info, info.LiveTokenCount())
	info.OutputActive = false

	if msg.id == m.active {
		m.view.Chat.Main.SetContent(info.History)
		m.view.Chat.Main.GotoBottom()
	}
	m.infos[msg.id] = info

	var cmds []tea.Cmd
	if progressCmd != nil {
		cmds = append(cmds, progressCmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleTokenUsageMessage(msg tokenUsageMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	info.InputTokensTotal = msg.totalInput
	info.OutputTokensTotal = msg.totalOutput
	info.HasUsageTotals = true
	info.StreamingTokenCount = msg.totalTokens
	info.InputActive = false
	info.LastActivity = time.Now()

	progressCmd := m.updateTokenProgress(info, info.LiveTokenCount())

	m.infos[msg.id] = info

	cmds := []tea.Cmd{m.runtime.ReadCmd(&m, msg.id)}
	if progressCmd != nil {
		cmds = append(cmds, progressCmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) updateTokenProgress(info *AgentInfo, totalTokens int) tea.Cmd {
	maxTokens := 8000
	if info.ModelName != "" {
		maxTokens = m.pricing.GetContextLimit(info.ModelName)
	}
	if maxTokens <= 0 {
		return nil
	}
	pct := float64(totalTokens) / float64(maxTokens)
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	if pct < 0.05 {
		pct = 0
	}
	return info.TokenProgress.SetPercent(pct)
}
