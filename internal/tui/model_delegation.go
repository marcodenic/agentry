package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/team"
)

func (m Model) handleDelegationLifecycle(msg delegationLifecycleMsg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	ev := msg.event

	if m.team != nil && ev.TeamAgentID != "" {
		if teamAgent := m.team.GetAgent(ev.TeamAgentID); teamAgent != nil {
			var added bool
			var newCmds []tea.Cmd
			m, newCmds, added = m.attachTeamAgent(teamAgent)
			if added {
				cmds = append(cmds, newCmds...)
				agentID := teamAgent.Agent.ID
				info := m.infos[agentID]
				cmds = append(cmds, func() tea.Msg {
					return agentStartMsg{
						id:   agentID,
						name: info.Name,
						role: info.Role,
					}
				})
			}
		}
	}

	if ev.CoreAgentID != "" {
		if agentUUID, err := uuid.Parse(ev.CoreAgentID); err == nil {
			if info, ok := m.infos[agentUUID]; ok {
				switch ev.Type {
				case team.DelegationEventSessionStart:
					info.Status = StatusRunning
					info.InputActive = true
					info.OutputActive = false
					info.LastActivity = time.Now()
					info.CurrentTool = ""
					m.infos[agentUUID] = info
				case team.DelegationEventSessionComplete:
					if info.Status != StatusError {
						info.Status = StatusIdle
					}
					info.InputActive = false
					info.OutputActive = false
					info.CurrentTool = ""
					m.infos[agentUUID] = info
				case team.DelegationEventSessionError:
					info.Status = StatusError
					info.InputActive = false
					info.OutputActive = false
					info.CurrentTool = ""
					m.infos[agentUUID] = info
				}
			}
		}
	}

	if m.delegationEvents != nil {
		if waitCmd := m.runtime.WaitDelegationEvent(m.delegationEvents); waitCmd != nil {
			cmds = append(cmds, waitCmd)
		}
	}

	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
}
