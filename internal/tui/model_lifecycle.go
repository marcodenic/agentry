package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) handleAgentComplete(msg agentCompleteMsg) (Model, tea.Cmd) {
	// finalMsg already handled completion - this is just cleanup
	info := m.infos[msg.id]
	if info.Status != StatusIdle {
		info.Status = StatusIdle
		m.infos[msg.id] = info
	}
	info.InputActive = false
	info.OutputActive = false
	m.infos[msg.id] = info
	return m, nil
}

func (m Model) handleAgentStart(msg agentStartMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	if info != nil {
		info.Status = StatusRunning
		// Add a status message to the conversation
		statusMsg := fmt.Sprintf("\n\n✨ **%s** (%s) is starting to work...\n", msg.name, msg.role)
		info.History += statusMsg
		m.infos[msg.id] = info

		// If this is the active agent, update the viewport
		if msg.id == m.active {
			m.view.Chat.Main.SetContent(info.History)
			m.view.Chat.Main.GotoBottom()
		}
	}
	return m, nil
}

// handleThinkingAnimation processes thinking animation messages
func (m Model) handleThinkingAnimation(msg thinkingAnimationMsg) (Model, tea.Cmd) {
	info := m.infos[msg.id]
	// Stop thinking animation if tokens have started or agent is not running
	if info.Status != StatusRunning || info.TokensStarted {
		// When stopping thinking animation, just refresh display with clean history
		if msg.id == m.active {
			m.view.Chat.Main.SetContent(info.History)
			m.view.Chat.Main.GotoBottom()
		}
		return m, nil
	}

	// Dots spinner frames (replacing slash spinner)
	// Use fixed width of 4 spaces so alignment matches user input (which uses 4 spaces after the bar)
	frames := []string{"    ", "•   ", "••  ", "••• "}
	currentSpinner := frames[msg.frame%len(frames)]

	// Build display content WITHOUT modifying history
	displayHistory := info.History

	// Check if we should append spinner to last status message or show on new line
	if info.LastContentType == ContentTypeStatusMessage {
		// Append spinner to the end of the last status message, align with 4-space content indent
		displayHistory += "    " + currentSpinner
	} else {
		// For user input or other content types, show spinner on new line with AI bar
		// Use four spaces after the bar to match user input indentation (and give a bit more offset)
		if len(displayHistory) > 0 && !strings.HasSuffix(displayHistory, "\n") {
			displayHistory += "\n" + m.aiBar() + "    " + currentSpinner
		} else {
			displayHistory += m.aiBar() + "    " + currentSpinner
		}
	}

	if msg.id == m.active {
		m.view.Chat.Main.SetContent(displayHistory)
		m.view.Chat.Main.GotoBottom()
	}

	// Continue the animation if still running and no tokens have started
	return m, m.runtime.StartThinkingAnimation(msg.id)
}

// handleDelegationLifecycle processes delegation lifecycle events from subagents
func (m Model) handleDelegationLifecycle(msg delegationLifecycleMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	event := msg.event
	
	switch event.Type {
	case "spawn":
		// Create a new agent entry when a subagent is spawned
		if m.team != nil {
			teamAgents := m.team.GetTeamAgents()
			for _, teamAgent := range teamAgents {
				if teamAgent.Name == event.AgentName {
					// Create AgentInfo for this subagent
					info := newAgentInfo(teamAgent.Agent, "", m.layout.width)
					info.History = fmt.Sprintf("🚀 Subagent **%s** spawned for: %s\n", event.AgentName, event.Input)
					info.Status = StatusIdle
					info.LastContentType = ContentTypeStatusMessage
					info.Role = event.Role
					info.Name = event.AgentName
					info.ModelName = teamAgent.Agent.ModelName
					
					m.infos[teamAgent.Agent.ID] = info
					m.order = append(m.order, teamAgent.Agent.ID)
					cmds = append(cmds, info.Spinner.Tick)
					break
				}
			}
		}
		
	case "session_start":
		// Find the agent and mark it as running
		for id, info := range m.infos {
			if info.Name == event.AgentName {
				info.Status = StatusRunning
				info.History += fmt.Sprintf("\n▶️  Working on: %s\n", event.Input)
				m.infos[id] = info
				
				// Start thinking animation for this agent
				cmds = append(cmds, m.runtime.StartThinkingAnimation(id))
				break
			}
		}
		
	case "session_complete":
		// Mark agent as idle and show result
		for id, info := range m.infos {
			if info.Name == event.AgentName {
				info.Status = StatusIdle
				resultPreview := event.Result
				if len(resultPreview) > 100 {
					resultPreview = resultPreview[:100] + "..."
				}
				info.History += fmt.Sprintf("\n✅ Completed: %s\n", resultPreview)
				m.infos[id] = info
				break
			}
		}
		
	case "session_error":
		// Mark agent as error state
		for id, info := range m.infos {
			if info.Name == event.AgentName {
				info.Status = StatusError
				info.History += fmt.Sprintf("\n❌ Error: %s\n", event.Err)
				m.infos[id] = info
				break
			}
		}
	}
	
	// Continue listening for delegation events
	if m.delegationEvents != nil {
		cmds = append(cmds, m.runtime.WaitDelegationEvent(m.delegationEvents))
	}
	
	return m, tea.Batch(cmds...)
}
