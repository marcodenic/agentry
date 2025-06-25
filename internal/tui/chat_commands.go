package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/team"
)

func (m ChatModel) indexOf(id uuid.UUID) int {
	for i, ag := range m.team.Agents() {
		if ag.ID == id {
			return i
		}
	}
	return 0
}

func (m ChatModel) callActive(input string) (ChatModel, tea.Cmd) {
	agName := m.names[m.indexOf(m.active)]
	info := m.infos[m.active]
	info.History += m.userBar() + " " + input + "\n"

	// Update viewport immediately to show user message
	idx := m.indexOf(m.active)
	wrappedContent := lipgloss.NewStyle().Width(m.vps[idx].Width).Render(info.History)
	m.vps[idx].SetContent(wrappedContent)
	m.vps[idx].GotoBottom()
	m.infos[m.active] = info

	ctx := team.WithContext(context.Background(), m.team)
	out, err := m.team.Call(ctx, agName, input)
	if err != nil {
		m.err = err
		return m, nil
	}
	info.History += m.aiBar() + " " + out + "\n"
	m.infos[m.active] = info
	wrappedContent = lipgloss.NewStyle().Width(m.vps[idx].Width).Render(info.History)
	m.vps[idx].SetContent(wrappedContent)
	m.vps[idx].GotoBottom()
	return m, nil
}

type chatCommand struct {
	Name string
	Args []string
}

func parseChatCommand(s string) chatCommand {
	fields := strings.Fields(strings.TrimSpace(s))
	if len(fields) == 0 {
		return chatCommand{}
	}
	name := strings.TrimPrefix(fields[0], "/")
	return chatCommand{Name: name, Args: fields[1:]}
}

func (m ChatModel) handleCommand(cmd string) (ChatModel, tea.Cmd) {
	c := parseChatCommand(cmd)
	switch c.Name {
	case "spawn":
		return m.handleSpawn(c.Args)
	case "switch":
		return m.handleSwitch(c.Args)
	case "stop":
		return m.handleStop(c.Args)
	case "converse":
		return m.handleConverse(c.Args)
	default:
		return m, nil
	}
}

func (m ChatModel) handleSpawn(args []string) (ChatModel, tea.Cmd) {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}
	ag, nm := m.team.AddAgent(name)
	m.names = append(m.names, nm)
	m.infos[ag.ID] = &AgentInfo{Agent: ag, Status: StatusIdle, Name: nm}
	vp := viewport.New(0, 0)
	m.vps = append(m.vps, vp)
	m.active = ag.ID
	return m, nil
}

func (m ChatModel) handleSwitch(args []string) (ChatModel, tea.Cmd) {
	if len(args) == 0 {
		return m, nil
	}
	pref := args[0]
	for id := range m.infos {
		if strings.HasPrefix(id.String(), pref) {
			m.active = id
			break
		}
	}
	return m, nil
}

func (m ChatModel) handleStop(args []string) (ChatModel, tea.Cmd) {
	// No asynchronous runs in this simplified model, but keep status field.
	id := m.active
	if len(args) > 0 {
		pref := args[0]
		for aid := range m.infos {
			if strings.HasPrefix(aid.String(), pref) {
				id = aid
				break
			}
		}
	}
	if info, ok := m.infos[id]; ok {
		info.Status = StatusStopped
		m.infos[id] = info
	}
	return m, nil
}

func (m ChatModel) handleConverse(args []string) (ChatModel, tea.Cmd) {
	// Kick off a round-robin conversation using the existing team.
	ctx := team.WithContext(context.Background(), m.team)
	idx, out, err := m.team.Step(ctx)
	if err != nil {
		m.err = err
		return m, nil
	}
	ag := m.team.Agents()[idx]
	info := m.infos[ag.ID]
	info.History += m.aiBar() + " " + out + "\n"
	m.infos[ag.ID] = info
	m.vps[idx].SetContent(info.History)
	m.vps[idx].GotoBottom()
	return m, nil
}

var _ tea.Model = ChatModel{}

// Helpers copied from model.go
func (m ChatModel) userBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.UserBarColor)).Render("┃")
}

func (m ChatModel) aiBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("┃")
}

// agentPanel creates the right sidebar showing agent status
func (m ChatModel) agentPanel() string {
	lines := []string{}
	for _, ag := range m.team.Agents() {
		info := m.infos[ag.ID]

		dot := m.statusDot(info.Status)
		line := fmt.Sprintf("%s %s", dot, info.Name)
		if ag.ID == m.active {
			line = "*" + line[1:]
		}
		lines = append(lines, line)

		// Token info if available
		if m.parent != nil && m.parent.Cost != nil {
			tokens := m.parent.Cost.TotalTokens()
			tokLine := fmt.Sprintf("  tokens: %d", tokens)
			lines = append(lines, tokLine)
		}
		lines = append(lines, "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

// statusDot returns a colored dot indicating agent status
func (m ChatModel) statusDot(st AgentStatus) string {
	color := m.theme.IdleColor
	switch st {
	case StatusRunning:
		color = m.theme.RunningColor
	case StatusError:
		color = m.theme.ErrorColor
	case StatusStopped:
		color = m.theme.StoppedColor
	}
	return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render("●")
}

// Agents exposes the team's agents for tests.
func (m ChatModel) Agents() map[uuid.UUID]*AgentInfo { return m.infos }
