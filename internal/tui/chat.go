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
	ti.Placeholder = "Message (Shift+Enter for new line)"
	ti.CharLimit = 1000  // Set reasonable character limit
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
		// Set input width to match viewport width
		inputWidth := int(float64(msg.Width)*0.75) - 2
		m.input.Width = inputWidth
		for i := range m.vps {
			m.vps[i].Width = inputWidth
			m.vps[i].Height = msg.Height - 5
			if info, ok := m.infos[m.team.Agents()[i].ID]; ok {
				// Apply text wrapping when setting content
				wrappedContent := lipgloss.NewStyle().Width(m.vps[i].Width).Render(info.History)
				m.vps[i].SetContent(wrappedContent)
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
		// Apply text wrapping when setting content
		wrappedContent := lipgloss.NewStyle().Width(vp.Width).Render(info.History)
		vp.SetContent(wrappedContent)
	}
	
	// Create left panel (chat + input)
	leftContent := vp.View() + "\n" + m.input.View()
	
	// Create right panel (agent sidebar)
	rightContent := m.agentPanel()
	
	// Layout
	base := lipgloss.NewStyle().
		Foreground(lipgloss.Color(m.theme.Palette.Foreground)).
		Background(lipgloss.Color(m.theme.Palette.Background))
	
	left := base.Copy().Width(int(float64(m.width) * 0.75)).Render(leftContent)
	right := base.Copy().Width(int(float64(m.width) * 0.25)).Render(rightContent)
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	
	// Footer
	tokens := 0
	costVal := 0.0
	if m.parent != nil && m.parent.Cost != nil {
		tokens = m.parent.Cost.TotalTokens()
		costVal = m.parent.Cost.TotalCost()
	}
	footer := fmt.Sprintf("agents: %d | tokens: %d cost: $%.4f", len(m.infos), tokens, costVal)
	footer = base.Copy().Width(m.width).Render(footer)
	
	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
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
	
	// Update viewport immediately to show user message
	idx := m.indexOf(m.active)
	wrappedContent := lipgloss.NewStyle().Width(m.vps[idx].Width).Render(info.History)
	m.vps[idx].SetContent(wrappedContent)
	m.vps[idx].GotoBottom()
	m.infos[m.active] = info
	
	ctx := context.WithValue(context.Background(), teamctx.Key{}, m.team)
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
