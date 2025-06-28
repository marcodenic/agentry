package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/tool"
)

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Enhanced agent navigation
		if key.Matches(msg, PrevAgentKey) {
			m = m.cycleActive(-1)
		} else if key.Matches(msg, NextAgentKey) {
			m = m.cycleActive(1)
		} else if key.Matches(msg, FirstAgentKey) {
			m = m.jumpToAgent(0)
		} else if key.Matches(msg, LastAgentKey) {
			m = m.jumpToAgent(len(m.order) - 1)
		}
		switch msg.String() {
		case m.keys.Quit:
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
		case m.keys.ToggleTab:
			m.activeTab = 1 - m.activeTab
		case m.keys.Pause:
			// Stop the active agent if it's running or streaming
			if info, ok := m.infos[m.active]; ok && (info.Status == StatusRunning || info.TokensStarted) {
				if info.Cancel != nil {
					info.Cancel() // Cancel the agent's context
				}
				
				// Clean up streaming response if in progress
				if info.StreamingResponse != "" {
					// Add the partial streaming response to history before stopping
					formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
					info.History += formattedResponse
					info.StreamingResponse = ""
				}
				
				info.Status = StatusIdle // Set to idle so new messages can be sent
				info.TokensStarted = false // Reset streaming state
				info.History += fmt.Sprintf("\n\n%s Agent stopped by user\n", m.statusBar())
				m.infos[m.active] = info
				
				// Update viewport to show the stop message
				m.vp.SetContent(info.History)
				m.vp.GotoBottom()
			}
		case m.keys.Submit:
			if m.input.Focused() {
				// Check if the active agent is running before accepting new input
				if info, ok := m.infos[m.active]; ok && info.Status == StatusRunning {
					// Agent is busy, ignore input
					return m, nil
				}
				
				txt := m.input.Value()
				m.input.SetValue("")
				if strings.HasPrefix(txt, "/") {
					var cmd tea.Cmd
					m, cmd = m.handleCommand(txt)
					return m, cmd
				}
				return m.startAgent(m.active, txt)
			}
		}
	case tokenMsg:
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
			// PERFORMANCE FIX: Only format and update every 10 characters or on newlines
			// This reduces expensive formatting calls by 90%
			shouldUpdate := len(info.StreamingResponse)%10 == 0 || 
				strings.HasSuffix(msg.token, "\n") || 
				strings.HasSuffix(msg.token, " ")
			
			if shouldUpdate {
				// Build display history with current streaming response
				displayHistory := info.History
				if info.StreamingResponse != "" {
					// Cache the AI bar to avoid repeated method calls
					aiBar := m.aiBar()
					// Simple concatenation without expensive formatting during streaming
					displayHistory += aiBar + " " + info.StreamingResponse
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
	case startTokenStream:
		// Start optimized token streaming - single timer approach
		return m, tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
			return tokenStreamTick{id: msg.id, text: msg.text, position: 0}
		})
	case tokenStreamTick:
		// Handle optimized token streaming
		if msg.position >= len([]rune(msg.text)) {
			// Streaming complete
			return m, nil
		}
		
		// Send current character
		runes := []rune(msg.text)
		token := string(runes[msg.position])
		
		// Schedule next character
		nextCmd := tea.Tick(30*time.Millisecond, func(t time.Time) tea.Msg {
			return tokenStreamTick{id: msg.id, text: msg.text, position: msg.position + 1}
		})
		
		// Process current token and schedule next
		newModel, _ := m.Update(tokenMsg{id: msg.id, token: token})
		return newModel, nextCmd
	case finalMsg:
		info := m.infos[msg.id]
		// Add the final AI response with proper formatting
		if info.StreamingResponse != "" {
			formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
			info.History += formattedResponse
			info.StreamingResponse = "" // Clear streaming response
		}
		
		// Limit history length to prevent unbounded memory growth (keep last ~100KB)
		const maxHistoryLength = 100000
		if len(info.History) > maxHistoryLength {
			// Keep last 75% of history to maintain context
			keepLength := maxHistoryLength * 3 / 4
			info.History = "...[earlier messages truncated]...\n" + info.History[len(info.History)-keepLength:]
		}
		
		// Set status to idle, clear spinner, and add proper spacing after AI message
		info.Status = StatusIdle
		info.TokensStarted = false // Reset streaming state
		info.History += "\n\n" // Add extra spacing after AI response
		
		if msg.id == m.active {
			m.vp.SetContent(info.History)
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, nil
	case toolUseMsg:
		info := m.infos[msg.id]
		info.CurrentTool = msg.name
		// Show completion message with better formatting
		completionText := m.formatToolCompletion(msg.name, msg.args)
		commandFormatted := m.formatSingleCommand(completionText)
		info.History += commandFormatted
		if msg.id == m.active {
			m.vp.SetContent(info.History)
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case actionMsg:
		info := m.infos[msg.id]
		// Add action messages with better spacing
		actionFormatted := m.formatSingleCommand(msg.text)
		info.History += actionFormatted
		if msg.id == m.active {
			m.vp.SetContent(info.History)
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case modelMsg:
		info := m.infos[msg.id]
		info.ModelName = msg.name
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case spinner.TickMsg:
		for id, ag := range m.infos {
			// Only update spinner for agents that are actually running and not finished streaming
			if ag.Status == StatusRunning && !ag.TokensStarted {
				var c tea.Cmd
				ag.Spinner, c = ag.Spinner.Update(msg)
				cmds = append(cmds, c)
				m.infos[id] = ag
			}
		}
	case activityTickMsg:
		// First, check for new agents that may have been spawned by the agent tool
		if m.team != nil {
			teamAgents := m.team.Agents()
			teamNames := m.team.Names()
			for i, agent := range teamAgents {
				if _, exists := m.infos[agent.ID]; !exists {
					// Found a new agent that's not in our tracking - add it
					sp := spinner.New()
					sp.Spinner = spinner.Line
					sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))
					
					// Generate sequential agent name and use team name as role
					agentNumber := len(m.infos) // This gives us the next agent number
					displayName := fmt.Sprintf("Agent %d", agentNumber)
					
					// Use team name as role, but validate it's not a tool name
					role := "agent" // default role
					if i < len(teamNames) {
						candidateRole := teamNames[i]
						// Only use the team name as role if it's not a builtin tool
						if !tool.IsBuiltinTool(candidateRole) {
							role = candidateRole
						}
					}
					
					info := &AgentInfo{
						Agent:           agent,
						Status:          StatusIdle,
						Spinner:         sp,
						Name:            displayName, // Sequential name like "Agent 1"
						Role:            role,        // Role from team, validated not to be a tool
						ActivityData:    make([]float64, 0),
						ActivityTimes:   make([]time.Time, 0),
						CurrentActivity: 0,
						LastActivity:    time.Time{},
						TokenHistory:    []int{},
						TokensStarted:   false,
						StreamingResponse: "",
					}
					m.infos[agent.ID] = info
					m.order = append(m.order, agent.ID)
				}
			}
		}
		
		// Update activity data for all agents - OPTIMIZED to reduce overhead
		now := time.Now()
		
		// Only process agents that have recent activity to avoid unnecessary work
		for id, info := range m.infos {
			// Skip inactive agents to improve performance
			if info.Status == StatusIdle && info.CurrentActivity == 0 && 
			   !info.LastActivity.IsZero() && now.Sub(info.LastActivity) > 10*time.Second {
				continue
			}

			shouldAddDataPoint := false

			if len(info.ActivityTimes) == 0 {
				// First data point
				shouldAddDataPoint = true
			} else {
				// Check if we need a new data point (more than 1 second since last)
				lastTime := info.ActivityTimes[len(info.ActivityTimes)-1]
				if now.Sub(lastTime) >= time.Second {
					shouldAddDataPoint = true
				}
			}

			if shouldAddDataPoint {
				// Add activity level (either current activity or 0.0 for idle)
				activityLevel := 0.0
				if info.CurrentActivity > 0 {
					// Normalize current activity (10 tokens/sec = 100%)
					activityLevel = float64(info.CurrentActivity) / 10.0
					if activityLevel > 1.0 {
						activityLevel = 1.0
					}
					info.CurrentActivity = 0 // Reset for next second
				}

				info.ActivityData = append(info.ActivityData, activityLevel)
				info.ActivityTimes = append(info.ActivityTimes, now)

				// Only clean up activity data every 5 seconds to reduce overhead
				if len(info.ActivityData)%5 == 0 {
					// Keep only last 30 seconds of data to prevent memory growth
					cutoffTime := now.Add(-30 * time.Second)
					
					// Use more efficient cleanup - find cutoff index first
					cutoffIndex := -1
					for i := len(info.ActivityTimes) - 1; i >= 0; i-- {
						if info.ActivityTimes[i].Before(cutoffTime) {
							cutoffIndex = i
							break
						}
					}
					
					if cutoffIndex >= 0 {
						// Remove old data efficiently using slicing
						info.ActivityData = info.ActivityData[cutoffIndex+1:]
						info.ActivityTimes = info.ActivityTimes[cutoffIndex+1:]
					}
				}

				info.LastActivity = now
				m.infos[id] = info
			}
		}
		// Schedule next tick - ONLY ONE TIMER to prevent exponential growth
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return activityTickMsg{}
		}))
	case refreshMsg:
		// This just causes a re-render to update the footer with live token/cost data
		// Do NOT schedule another refresh - let activityTickMsg handle all timing
	case errMsg:
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
			info.History += errorFormatted
			
			// Update viewport if this is the active agent
			if m.active == info.Agent.ID {
				m.vp.SetContent(info.History)
				m.vp.GotoBottom()
			}
			
			m.infos[m.active] = info
		}
	case agentCompleteMsg:
		// finalMsg already handled completion - this is just cleanup
		info := m.infos[msg.id]
		if info.Status != StatusIdle {
			info.Status = StatusIdle
			m.infos[msg.id] = info
		}
	case thinkingAnimationMsg:
		info := m.infos[msg.id]
		// Stop thinking animation if tokens have started or agent is not running
		if info.Status != StatusRunning || info.TokensStarted {
			// When stopping thinking animation, just refresh display with clean history
			if msg.id == m.active {
				m.vp.SetContent(info.History)
				m.vp.GotoBottom()
			}
			return m, nil
		}
		
		// ASCII spinner frames
		frames := []string{"|", "/", "-", "\\"}
		currentSpinner := frames[msg.frame]
		
		// Build display content WITHOUT modifying history
		displayHistory := info.History
		
		// Add spinner to display only if history doesn't end with newline
		if len(displayHistory) > 0 && !strings.HasSuffix(displayHistory, "\n") {
			// Add AI bar and spinner for display only
			displayHistory += "\n" + m.aiBar() + " " + currentSpinner
		} else {
			// Add AI bar and spinner for display only
			displayHistory += m.aiBar() + " " + currentSpinner
		}
		
		if msg.id == m.active {
			m.vp.SetContent(displayHistory)
			m.vp.GotoBottom()
		}
		
		// Continue the animation if still running and no tokens have started
		return m, startThinkingAnimation(msg.id)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 2  // Full width minus padding for input
		
		// Calculate chat area dimensions
		chatWidth := int(float64(msg.Width)*0.75) - 2  // 75% width for chat area
		
		// Calculate viewport height more accurately:
		// Total height - top section margin - horizontal separator (1) - input section height - footer section height - padding
		viewportHeight := msg.Height - 5  // Leave space for separator, input, footer, and padding
		
		m.vp.Width = chatWidth
		m.vp.Height = viewportHeight
		
		// Set agent panel size (25% width)
		m.tools.SetSize(int(float64(msg.Width)*0.25)-2, viewportHeight)
		
		// Refresh the viewport content with proper sizing - avoid expensive reformatting
		if info, ok := m.infos[m.active]; ok {
			// For normal history, only re-wrap if width changed significantly  
			if !m.showInitialLogo && info.History != "" {
				// Only reformat if width changed by more than 10 characters to avoid constant reformatting
				if m.lastWidth == 0 || (chatWidth != m.lastWidth && abs(chatWidth-m.lastWidth) > 10) {
					reformattedHistory := m.formatHistoryWithBars(info.History, chatWidth)
					m.vp.SetContent(reformattedHistory)
					m.lastWidth = chatWidth
				} else {
					// Width didn't change much, just update viewport size without reformatting
					m.vp.SetContent(info.History)
				}
			} else {
				// For initial logo or empty history, use as-is
				m.vp.SetContent(info.History)
			}
			m.vp.GotoBottom() // Ensure we're at the bottom after resize
		}
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var chatContent string
	if m.activeTab == 0 {
		// Use viewport content directly for proper scrolling
		chatContent = m.vp.View()
		
		// Special handling for centered logo
		if m.showInitialLogo {
			if info, ok := m.infos[m.active]; ok {
				// Apply centering to the logo content
				logoStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
					Width(m.vp.Width).
					Height(m.vp.Height).
					Align(lipgloss.Center, lipgloss.Center)
				chatContent = logoStyle.Render(info.History)
			}
		}
	} else {
		if info, ok := m.infos[m.active]; ok {
			chatContent = renderMemory(info.Agent)
		}
	}
	
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground))
	
	// Create top section with chat (left) and agents (right)
	// Don't apply extra width constraints to chatContent - let viewport handle it
	left := base.Width(int(float64(m.width) * 0.75)).Render(chatContent)
	right := base.Width(int(float64(m.width) * 0.25)).Render(m.agentPanel())
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	
	// Add full-width horizontal line
	horizontalLine := strings.Repeat("─", m.width)
	if m.width < 0 {
		horizontalLine = ""
	}
	
	// Add full-width input
	inputSection := base.Width(m.width).Render(m.input.View())
	
	// Stack everything vertically
	content := lipgloss.JoinVertical(lipgloss.Left, topSection, horizontalLine, inputSection)

	// Calculate total tokens and cost across all agents
	totalTokens := 0
	totalCost := 0.0
	for _, info := range m.infos {
		if info.Agent.Cost != nil {
			totalTokens += info.Agent.Cost.TotalTokens()
			totalCost += info.Agent.Cost.TotalCost()
		}
	}
	
	footerText := fmt.Sprintf("cwd: %s | agents: %d | tokens: %d cost: $%.4f", m.cwd, len(m.infos), totalTokens, totalCost)
	footer := base.Width(m.width).Align(lipgloss.Right).Render(footerText)
	
	// Add empty line between input and footer
	return lipgloss.JoinVertical(lipgloss.Left, content, "", footer)
}
