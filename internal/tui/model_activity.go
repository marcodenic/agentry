package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/tool"
)

// handleSpinnerTick processes spinner animation updates
func (m Model) handleSpinnerTick(msg spinner.TickMsg) (Model, []tea.Cmd) {
	var cmds []tea.Cmd
	for id, ag := range m.infos {
		// Only update spinner for agents that are actually running and not finished streaming
		if ag.Status == StatusRunning && !ag.TokensStarted {
			var c tea.Cmd
			ag.Spinner, c = ag.Spinner.Update(msg)
			cmds = append(cmds, c)
			m.infos[id] = ag
		}
	}
	return m, cmds
}

// handleActivityTick processes activity monitoring and agent discovery
func (m Model) handleActivityTick(msg activityTickMsg) (Model, tea.Cmd) {
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
					Agent:                  agent,
					Status:                 StatusIdle,
					Spinner:                sp,
					TokenProgress:          createTokenProgressBar(),
					Name:                   displayName, // Sequential name like "Agent 1"
					Role:                   role,        // Role from team, validated not to be a tool
					ActivityData:           make([]float64, 0),
					ActivityTimes:          make([]time.Time, 0),
					CurrentActivity:        0,
					LastActivity:           time.Time{},
					TokenHistory:           []int{},
					TokensStarted:          false,
					StreamingResponse:      "",
					DebugTrace:             make([]DebugTraceEvent, 0), // Initialize debug trace
					CurrentStep:            0,
					DebugStreamingResponse: "", // Initialize debug streaming response
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
	return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	})
}
