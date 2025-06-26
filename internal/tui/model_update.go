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
		case m.keys.Submit:
			if m.input.Focused() {
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
		info := m.infos[msg.id]
		info.History += msg.token
		info.TokenCount++
		info.CurrentActivity++ // Just increment counter, let activityTickMsg handle data points

		now := time.Now()

		// Legacy token history update (keep for compatibility)
		if info.LastToken.IsZero() || now.Sub(info.LastToken) > time.Second {
			info.TokenHistory = append(info.TokenHistory, 1)
			if len(info.TokenHistory) > 20 {
				info.TokenHistory = info.TokenHistory[1:]
			}
		} else if len(info.TokenHistory) > 0 {
			info.TokenHistory[len(info.TokenHistory)-1]++
		}
		info.LastToken = now
		m.infos[msg.id] = info // IMPORTANT: Save the updated info back to the map
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
	case finalMsg:
		info := m.infos[msg.id]
		info.History += m.aiBar() + " "
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		info.Status = StatusIdle
		m.infos[msg.id] = info
		return m, tea.Batch(streamTokens(msg.id, msg.text+"\n"), m.readCmd(msg.id))
	case toolUseMsg:
		info := m.infos[msg.id]
		info.CurrentTool = msg.name
		// Show completion message
		completionText := m.formatToolCompletion(msg.name, msg.args)
		info.History += fmt.Sprintf("\n%s %s", m.statusBar(), completionText)
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case thinkingMsg:
		info := m.infos[msg.id]
		info.History += fmt.Sprintf("\n%s %s", m.statusBar(), msg.text)
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			m.vp.GotoBottom()
		}
		m.infos[msg.id] = info
		return m, m.readCmd(msg.id)
	case actionMsg:
		info := m.infos[msg.id]
		info.History += fmt.Sprintf("\n%s %s", m.statusBar(), msg.text)
		if msg.id == m.active {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
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
		info := m.infos[msg.id]
		info.Status = StatusIdle
		// Display the final agent response
		if msg.result != "" {
			info.History += fmt.Sprintf("\n%s %s", m.aiBar(), msg.result)
			if msg.id == m.active {
				base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
				m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
				m.vp.GotoBottom()
			}
		}
		m.infos[msg.id] = info
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Width = int(float64(msg.Width)*0.75) - 2 // Calculate viewport height: total height - input line - footer line
		m.vp.Height = msg.Height - 2
		m.tools.SetSize(int(float64(msg.Width)*0.25)-2, msg.Height-2)
		if info, ok := m.infos[m.active]; ok {
			base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
			m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
		}
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	var leftContent string
	if m.activeTab == 0 {
		leftContent = m.vp.View() + "\n" + m.input.View()
	} else {
		if info, ok := m.infos[m.active]; ok {
			leftContent = renderMemory(info.Agent)
		}
	}
	if m.err != nil {
		leftContent += "\nERR: " + m.err.Error()
	}
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
		Background(lipgloss.Color(m.theme.Palette.Background))
	left := base.Copy().Width(int(float64(m.width) * 0.75)).Render(leftContent)
	right := base.Copy().Width(int(float64(m.width) * 0.25)).Render(m.agentPanel())
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)

	tokens := 0
	costVal := 0.0
	if info, ok := m.infos[m.active]; ok && info.Agent.Cost != nil {
		tokens = info.Agent.Cost.TotalTokens()
		costVal = info.Agent.Cost.TotalCost()
	}
	footer := fmt.Sprintf("cwd: %s | agents: %d | tokens: %d cost: $%.4f", m.cwd, len(m.infos), tokens, costVal)
	footer = base.Copy().Width(m.width).Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
}
