package tui

import (
	"bufio"
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/trace"
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
func (m Model) handleActivityTick(_ activityTickMsg) (Model, tea.Cmd) {
	var newAgentCmds []tea.Cmd

	// First, check for new agents that may have been spawned by the agent tool
	if m.team != nil {
		teamAgents := m.team.GetTeamAgents() // Use GetTeamAgents to get role info
		for _, teamAgent := range teamAgents {
			if _, exists := m.infos[teamAgent.Agent.ID]; !exists {
				// Found a new agent that's not in our tracking - add it
				sp := spinner.New()
				sp.Spinner = spinner.Dot
				sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))
				agentNumber := len(m.infos) // This gives us the next agent number
				displayName := fmt.Sprintf("Agent %d", agentNumber)

				// Use the actual role from the team agent (this is the correct role!)
				role := teamAgent.Role

				info := &AgentInfo{
					Agent:                  teamAgent.Agent,
					Status:                 StatusIdle,
					Spinner:                sp,
					TokenProgress:          createTokenProgressBar(),
					Name:                   displayName, // Sequential name like "Agent 1"
					Role:                   role,        // Role from team agent (correct role!)
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

				// Set proper token progress bar width based on current panel width
				panelWidth := int(float64(m.width) * 0.25)
				barWidth := panelWidth - 8 // Same calculation as layout and activity chart
				if barWidth < 10 {
					barWidth = 10
				}
				if barWidth > 50 {
					barWidth = 50
				}
				info.TokenProgress.Width = barWidth

				// Get the model name from the agent
				newModelName := teamAgent.Agent.ModelName
				if newModelName == "" {
					newModelName = "unknown"
				}

				// Simply update the model name if we have a new one
				if newModelName != "" {
					info.ModelName = newModelName
				}

				// Set up trace listening for the newly discovered agent
				// This ensures spawned agents' token events are captured by the TUI
				pr, pw := io.Pipe()
				tracer := trace.NewJSONL(pw)
				if teamAgent.Agent.Tracer != nil {
					teamAgent.Agent.Tracer = trace.NewMulti(teamAgent.Agent.Tracer, tracer)
				} else {
					teamAgent.Agent.Tracer = tracer
				}
				info.Scanner = bufio.NewScanner(pr)
				// Bump scanner buffer to avoid "token too long" when spawned agents emit large events
				info.Scanner.Buffer(make([]byte, 0, 256*1024), 4*1024*1024)
				// Store the pipe writer so we can clean it up later if needed
				info.tracePipeWriter = pw

				m.infos[teamAgent.Agent.ID] = info
				m.order = append(m.order, teamAgent.Agent.ID)

				// Emit start message for better UI feedback
				newAgentCmds = append(newAgentCmds, func() tea.Msg {
					return agentStartMsg{
						id:   teamAgent.Agent.ID,
						name: displayName,
						role: role,
					}
				})

				// Schedule readCmd for the new agent to listen to its trace events
				newAgentCmds = append(newAgentCmds, m.readCmd(teamAgent.Agent.ID))
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
				// Keep only last 5 minutes of data to show longer time period in sparkline
				cutoffTime := now.Add(-5 * time.Minute)

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

	// Combine new agent commands with the activity tick timer
	var allCmds []tea.Cmd
	allCmds = append(allCmds, newAgentCmds...)

	// Schedule next tick - use faster polling when agents are running to catch spawned agents quickly
	tickInterval := time.Second
	for _, info := range m.infos {
		if info.Status == StatusRunning {
			// Poll more frequently when agents are running to catch spawned agents
			tickInterval = 200 * time.Millisecond
			break
		}
	}

	allCmds = append(allCmds, tea.Tick(tickInterval, func(t time.Time) tea.Msg {
		return activityTickMsg{}
	}))

	return m, tea.Batch(allCmds...)
}