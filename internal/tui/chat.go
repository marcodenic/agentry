package tui

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/teamctx"
)

// ChatModel is a unified model that manages a converse.Team of any size.
type ChatModel struct {
	parent *core.Agent
	team   *converse.Team
	infos  map[uuid.UUID]*AgentInfo
	names  []string
	active uuid.UUID

	vps    []viewport.Model
	input  textinput.Model
	width  int
	height int

	theme Theme
	keys  Keybinds
	err   error
}

// NewChat creates a team with n agents talking about topic.
func NewChat(parent *core.Agent, n int, topic string) (ChatModel, error) {
	th := LoadTheme()
	t, err := converse.NewTeam(parent, n, topic)
	if err != nil {
		return ChatModel{}, err
	}
	vps := make([]viewport.Model, n)
	infos := make(map[uuid.UUID]*AgentInfo, n)
	for i, ag := range t.Agents() {
		vps[i] = viewport.New(0, 0)
		infos[ag.ID] = &AgentInfo{Agent: ag, Status: StatusIdle, Name: t.Names()[i]}
	}
	ti := textinput.New()
	ti.Placeholder = "Message"
	ti.Focus()
	return ChatModel{parent: parent, team: t, infos: infos, names: t.Names(), active: t.Agents()[0].ID,
		vps: vps, input: ti, theme: th, keys: th.Keybinds}, nil
}

func (m ChatModel) Init() tea.Cmd { return nil }

// Update handles Bubble Tea messages.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case m.keys.Quit:
			return m, tea.Quit
		case m.keys.Submit:
			if m.input.Focused() {
				txt := m.input.Value()
				m.input.SetValue("")
				if strings.HasPrefix(txt, "/") {
					var cmd tea.Cmd
					m, cmd = m.handleCommand(txt)
					return m, cmd
				}
				return m.callActive(txt)
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		for i := range m.vps {
			m.vps[i].Width = int(float64(msg.Width)*0.75) - 2
			m.vps[i].Height = msg.Height - 5
			if info, ok := m.infos[m.team.Agents()[i].ID]; ok {
				m.vps[i].SetContent(info.History)
			}
		}
	}

	m.input, _ = m.input.Update(msg)
	return m, nil
}

func (m ChatModel) View() string {
	vp := viewport.Model{}
	if info, ok := m.infos[m.active]; ok {
		vp = m.vps[m.indexOf(m.active)]
		vp.SetContent(info.History)
	}
	main := vp.View() + "\n" + m.input.View()
	tokens := 0
	costVal := 0.0
	if m.parent != nil && m.parent.Cost != nil {
		tokens = m.parent.Cost.TotalTokens()
		costVal = m.parent.Cost.TotalCost()
	}
	footer := fmt.Sprintf("agents: %d | tokens: %d cost: $%.4f", len(m.infos), tokens, costVal)
	return fmt.Sprintf("%s\n%s", main, footer)
}

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
	ctx := context.WithValue(context.Background(), teamctx.Key{}, m.team)
	out, err := m.team.Call(ctx, agName, input)
	if err != nil {
		m.err = err
		return m, nil
	}
	info.History += m.aiBar() + " " + out + "\n"
	m.infos[m.active] = info
	idx := m.indexOf(m.active)
	m.vps[idx].SetContent(info.History)
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
	ctx := context.WithValue(context.Background(), teamctx.Key{}, m.team)
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

// Agents exposes the team's agents for tests.
func (m ChatModel) Agents() map[uuid.UUID]*AgentInfo { return m.infos }
