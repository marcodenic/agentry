package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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
				base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
				m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
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
			
			// Clean up any leftover spinner characters from the end of History
			if len(info.History) > 0 {
				// Check for and remove spinner characters at the end
				lastChar := info.History[len(info.History)-1:]
				for lastChar == "|" || lastChar == "/" || lastChar == "-" || lastChar == "\\" {
					info.History = info.History[:len(info.History)-1]
					if len(info.History) == 0 {
						break
					}
					lastChar = info.History[len(info.History)-1:]
				}
			}
		}
		
		// Add token to streaming response
		info.StreamingResponse += msg.token
		info.TokenCount++
		info.CurrentActivity++ // Just increment counter, let activityTickMsg handle data points

		// Update viewport with formatted streaming content in real-time
		if msg.id == m.active {
			// Build display history with current streaming response
			displayHistory := info.History
			if info.StreamingResponse != "" {
				formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
				displayHistory += formattedResponse
			}
			
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(displayHistory))
			m.vp.GotoBottom()
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
	case finalMsg:
		info := m.infos[msg.id]
		// Finalize the AI response with proper formatting
		if info.StreamingResponse != "" {
			formattedResponse := m.formatWithBar(m.aiBar(), info.StreamingResponse, m.vp.Width)
			info.History += formattedResponse
			info.StreamingResponse = "" // Clear streaming response
		}
		
		// Set status to idle and add spacing after AI message
		info.Status = StatusIdle
		info.History += "\n\n" // Add extra spacing after AI response
		
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, nil
	case toolUseMsg:
		info := m.infos[msg.id]
		info.CurrentTool = msg.name
		// Show completion message
		completionText := m.formatToolCompletion(msg.name, msg.args)
		info.History += fmt.Sprintf("\n%s %s", m.statusBar(), completionText)
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case actionMsg:
		info := m.infos[msg.id]
		info.History += fmt.Sprintf("\n%s %s", m.statusBar(), msg.text)
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
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
			if ag.Status == StatusRunning {
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
					agentName := "agent"
					if i < len(teamNames) {
						agentName = teamNames[i]
					}
					info := &AgentInfo{
						Agent:           agent,
						Status:          StatusIdle,
						Spinner:         sp,
						Name:            agentName,
						Role:            agentName,
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
		
		// Update activity data for all agents (to make chart scroll even when idle)
		now := time.Now()
		for id, info := range m.infos {
			// Always add a data point every second to make the chart scroll
			// If there was activity in this second, it will already be recorded
			// Otherwise, add a zero point to show time progression

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

				// Keep only last 60 seconds of data
				cutoffTime := now.Add(-60 * time.Second)
				var newData []float64
				var newTimes []time.Time

				for i, t := range info.ActivityTimes {
					if t.After(cutoffTime) {
						newData = append(newData, info.ActivityData[i])
						newTimes = append(newTimes, info.ActivityTimes[i])
					}
				}

				info.ActivityData = newData
				info.ActivityTimes = newTimes
				info.LastActivity = now
				m.infos[id] = info
			}
		}
		// Schedule next tick
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return activityTickMsg{}
		}))
	case errMsg:
		m.err = msg
		if info, ok := m.infos[m.active]; ok {
			info.Status = StatusError
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
			return m, nil
		}
		
		// ASCII spinner frames
		frames := []string{"|", "/", "-", "\\"}
		
		// Check if we need to replace an existing spinner or this is the first spinner
		if len(info.History) > 0 {
			lastChar := info.History[len(info.History)-1:]
			if lastChar == "|" || lastChar == "/" || lastChar == "-" || lastChar == "\\" {
				// Replace the existing spinner character
				info.History = info.History[:len(info.History)-1]
			} else if lastChar == " " && strings.HasSuffix(info.History, " ") {
				// This is the first spinner - replace the trailing space after AI bar
				// Check if this looks like an AI bar line (ends with space after non-space)
				if len(info.History) >= 2 && info.History[len(info.History)-2] != ' ' {
					info.History = info.History[:len(info.History)-1] // Remove the space
				}
			}
		}
		// Add the new spinner frame
		info.History += frames[msg.frame]
		
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		// Continue the animation if still running and no tokens have started
		return m, startThinkingAnimation(msg.id)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 2  // Full width minus padding for input
		m.vp.Width = int(float64(msg.Width)*0.75) - 2  // 75% width for chat area
		// Calculate viewport height: total height - horizontal separator line - input line - footer line
		m.vp.Height = msg.Height - 4   // Account for separator + input + footer + spacing
		m.tools.SetSize(int(float64(msg.Width)*0.25)-2, msg.Height-4) // 25% width for agent panel
		if info, ok := m.infos[m.active]; ok {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
		}
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var chatContent string
	if m.activeTab == 0 {
		chatContent = m.vp.View()
		
		// Center the logo if we're showing the initial logo
		if m.showInitialLogo {
			if info, ok := m.infos[m.active]; ok {
				// Apply centering to the logo content
				logoStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
					Width(int(float64(m.width) * 0.75)).
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
	if m.err != nil {
		chatContent += "\nERR: " + m.err.Error()
	}
	
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground))
	
	// Create top section with chat (left) and agents (right)
	left := base.Copy().Width(int(float64(m.width) * 0.75)).Render(chatContent)
	right := base.Copy().Width(int(float64(m.width) * 0.25)).Render(m.agentPanel())
	topSection := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	
	// Add full-width horizontal line
	horizontalLine := strings.Repeat("─", m.width)
	if m.width < 0 {
		horizontalLine = ""
	}
	
	// Add full-width input
	inputSection := base.Copy().Width(m.width).Render(m.input.View())
	
	// Stack everything vertically
	content := lipgloss.JoinVertical(lipgloss.Left, topSection, horizontalLine, inputSection)

	tokens := 0
	costVal := 0.0
	if info, ok := m.infos[m.active]; ok && info.Agent.Cost != nil {
		tokens = info.Agent.Cost.TotalTokens()
		costVal = info.Agent.Cost.TotalCost()
	}
	footerText := fmt.Sprintf("cwd: %s | agents: %d | tokens: %d cost: $%.4f", m.cwd, len(m.infos), tokens, costVal)
	footer := base.Copy().Width(m.width).Align(lipgloss.Right).Render(footerText)
	
	// Add empty line between input and footer
	return lipgloss.JoinVertical(lipgloss.Left, content, "", footer)
}
