package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/teamctx"
	"github.com/marcodenic/agentry/internal/trace"
)

// Model is the root TUI model.
type Model struct {
	agents []*core.Agent
	infos  map[uuid.UUID]*AgentInfo
	order  []uuid.UUID
	active uuid.UUID

	team *converse.Team

	vp    viewport.Model
	input textinput.Model
	tools list.Model

	cwd string

	activeTab int
	width     int
	height    int

	err error

	theme Theme
	keys  Keybinds
}

type AgentStatus int

const (
	StatusIdle AgentStatus = iota
	StatusRunning
	StatusError
	StatusStopped
)

type AgentInfo struct {
	Agent        *core.Agent
	History      string
	Status       AgentStatus
	CurrentTool  string
	TokenCount   int
	TokenHistory []int
	LastToken    time.Time
	ModelName    string
	Scanner      *bufio.Scanner
	Cancel       context.CancelFunc
	Spinner      spinner.Model
	Name         string
}

// New creates a new TUI model bound to an Agent.
func New(ag *core.Agent) Model {
	th := LoadTheme()
	items := []list.Item{}
	for name, tl := range ag.Tools {
		items = append(items, listItem{name: name, desc: tl.Description()})
	}
	l := list.New(items, listItemDelegate{}, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = "Tools"
	l.SetShowHelp(false) // Hide help bar with navigation hints
	// Assign custom keymap to disable navigation keys
	l.KeyMap.CursorUp = NoNavKeyMap.CursorUp
	l.KeyMap.CursorDown = NoNavKeyMap.CursorDown
	l.KeyMap.PrevPage = NoNavKeyMap.PrevPage
	l.KeyMap.NextPage = NoNavKeyMap.NextPage
	l.KeyMap.GoToStart = NoNavKeyMap.GoToStart
	l.KeyMap.GoToEnd = NoNavKeyMap.GoToEnd
	l.KeyMap.Filter = NoNavKeyMap.Filter
	l.KeyMap.ClearFilter = NoNavKeyMap.ClearFilter
	l.KeyMap.CancelWhileFiltering = NoNavKeyMap.CancelWhileFiltering
	l.KeyMap.AcceptWhileFiltering = NoNavKeyMap.AcceptWhileFiltering
	l.KeyMap.ShowFullHelp = NoNavKeyMap.ShowFullHelp
	l.KeyMap.CloseFullHelp = NoNavKeyMap.CloseFullHelp
	l.KeyMap.Quit = NoNavKeyMap.Quit
	l.KeyMap.ForceQuit = NoNavKeyMap.ForceQuit

	ti := textinput.New()
	ti.Placeholder = "Message"
	ti.Focus()

	vp := viewport.New(0, 0)
	cwd, _ := os.Getwd()

	info := &AgentInfo{Agent: ag, Status: StatusIdle, Spinner: spinner.New(), Name: "master"}
	infos := map[uuid.UUID]*AgentInfo{ag.ID: info}

	tm, err := converse.NewTeam(ag, 1, "")
	if err != nil {
		tm = &converse.Team{}
	}

	m := Model{
		agents: []*core.Agent{ag},
		infos:  infos,
		order:  []uuid.UUID{ag.ID},
		active: ag.ID,
		team:   tm,
		vp:     vp,
		input:  ti,
		tools:  l,
		cwd:    cwd,
		theme:  th,
		keys:   th.Keybinds,
	}
	return m
}

type listItem struct{ name, desc string }

func (l listItem) Title() string       { return l.name }
func (l listItem) Description() string { return l.desc }
func (l listItem) FilterValue() string { return l.name }

type listItemDelegate struct{}

func (d listItemDelegate) Height() int                               { return 1 }
func (d listItemDelegate) Spacing() int                              { return 0 }
func (d listItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d listItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it := item.(listItem)
	style := lipgloss.NewStyle()
	if index == m.Index() {
		style = style.Bold(true)
		io.WriteString(w, style.Render("> "+it.name))
	} else {
		io.WriteString(w, "  "+it.name)
	}
}

type tokenMsg struct {
	id    uuid.UUID
	token string
}

type toolUseMsg struct {
	id   uuid.UUID
	name string
}

type modelMsg struct {
	id   uuid.UUID
	name string
}

type errMsg struct{ error }

type finalMsg struct {
	id   uuid.UUID
	text string
}

func streamTokens(id uuid.UUID, out string) tea.Cmd {
	runes := []rune(out)
	cmds := make([]tea.Cmd, len(runes))
	for i, r := range runes {
		tok := string(r)
		delay := time.Duration(i*30) * time.Millisecond
		cmds[i] = tea.Tick(delay, func(t time.Time) tea.Msg { return tokenMsg{id: id, token: tok} })
	}
	return tea.Batch(cmds...)
}

func (m *Model) readEvent(id uuid.UUID) tea.Msg {
	info := m.infos[id]
	if info == nil || info.Scanner == nil {
		return nil
	}
	for {
		if !info.Scanner.Scan() {
			if err := info.Scanner.Err(); err != nil {
				return errMsg{err}
			}
			return nil
		}
		var ev trace.Event
		if err := json.Unmarshal(info.Scanner.Bytes(), &ev); err != nil {
			return errMsg{err}
		}
		switch ev.Type {
		case trace.EventFinal:
			if s, ok := ev.Data.(string); ok {
				return finalMsg{id: id, text: s}
			}
		case trace.EventModelStart:
			if name, ok := ev.Data.(string); ok {
				return modelMsg{id: id, name: name}
			}
		case trace.EventToolEnd:
			if m2, ok := ev.Data.(map[string]any); ok {
				if name, ok := m2["name"].(string); ok {
					return toolUseMsg{id: id, name: name}
				}
			}
		default:
			continue
		}
	}
}

func (m *Model) readCmd(id uuid.UUID) tea.Cmd {
	return func() tea.Msg { return m.readEvent(id) }
}

func waitErr(ch <-chan error) tea.Cmd {
	return func() tea.Msg {
		if err := <-ch; err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, PrevAgentKey) {
			m = m.cycleActive(-1)
		} else if key.Matches(msg, NextAgentKey) {
			m = m.cycleActive(1)
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
		now := time.Now()
		if info.LastToken.IsZero() || now.Sub(info.LastToken) > time.Second {
			info.TokenHistory = append(info.TokenHistory, 1)
			if len(info.TokenHistory) > 10 {
				info.TokenHistory = info.TokenHistory[1:]
			}
		} else if len(info.TokenHistory) > 0 {
			info.TokenHistory[len(info.TokenHistory)-1]++
		}
		info.LastToken = now
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
	case errMsg:
		m.err = msg
		if info, ok := m.infos[m.active]; ok {
			info.Status = StatusError
			m.infos[m.active] = info
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Height = msg.Height - 5
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
	help := lipgloss.NewStyle().Width(m.width).Render(helpView())

	tokens := 0
	costVal := 0.0
	if info, ok := m.infos[m.active]; ok && info.Agent.Cost != nil {
		tokens = info.Agent.Cost.TotalTokens()
		costVal = info.Agent.Cost.TotalCost()
	}
	footer := fmt.Sprintf("cwd: %s | agents: %d | tokens: %d cost: $%.4f", m.cwd, len(m.infos), tokens, costVal)
	footer = base.Copy().Width(m.width).Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, main, help, footer)
}

func (m Model) userBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.UserBarColor)).Render("┃")
}

func (m Model) aiBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor)).Render("┃")
}

func renderMemory(ag *core.Agent) string {
	hist := ag.Mem.History()
	var b bytes.Buffer
	for i, s := range hist {
		b.WriteString("Step ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(": ")
		b.WriteString(s.Output)
		for _, tc := range s.ToolCalls {
			if r, ok := s.ToolResults[tc.ID]; ok {
				b.WriteString(" -> ")
				b.WriteString(tc.Name)
				b.WriteString(": ")
				b.WriteString(r)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

func (m Model) agentPanel() string {
	lines := []string{}
	for _, id := range m.order {
		ag := m.infos[id]

		dot := m.statusDot(ag.Status)
		line := fmt.Sprintf("%s %s %s", dot, ag.Spinner.View(), ag.Name)
		if id == m.active {
			line = "*" + line[1:]
		}
		lines = append(lines, line)

		tokLine := fmt.Sprintf("  tokens: %d", ag.TokenCount)
		bar := m.renderTokenBar(ag.TokenCount, 1000)
		lines = append(lines, tokLine)
		lines = append(lines, "  "+bar)
		if spark := m.renderSparkline(ag.TokenHistory); spark != "" {
			lines = append(lines, "  "+spark)
		}
		lines = append(lines, "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (m Model) startAgent(id uuid.UUID, input string) (Model, tea.Cmd) {
	info := m.infos[id]
	info.Status = StatusRunning
	info.Spinner = spinner.New()
	info.Spinner.Spinner = spinner.Line
	info.Spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))
	info.History += m.userBar() + " " + input + "\n"
	base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
	m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))

	pr, pw := io.Pipe()
	errCh := make(chan error, 1)
	tracer := trace.NewJSONL(pw)
	if info.Agent.Tracer != nil {
		info.Agent.Tracer = trace.NewMulti(info.Agent.Tracer, tracer)
	} else {
		info.Agent.Tracer = tracer
	}
	info.Scanner = bufio.NewScanner(pr)
	ctx := context.WithValue(context.Background(), teamctx.Key{}, m.team)
	ctx, cancel := context.WithCancel(ctx)
	info.Cancel = cancel
	m.infos[id] = info
	go func() {
		_, err := info.Agent.Run(ctx, input)
		pw.Close()
		errCh <- err
	}()
	m.infos[id] = info
	return m, tea.Batch(m.readCmd(id), waitErr(errCh), info.Spinner.Tick)
}

func (m Model) handleCommand(cmd string) (Model, tea.Cmd) {
	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return m, nil
	}
	switch fields[0] {
	case "/spawn":
		return m.handleSpawn(fields[1:])
	case "/switch":
		return m.handleSwitch(fields[1:])
	case "/stop":
		return m.handleStop(fields[1:])
	case "/converse":
		return m.handleConverse(fields[1:])
	default:
		return m, nil
	}
}

func (m Model) handleSpawn(args []string) (Model, tea.Cmd) {
	name := "agent"
	if len(args) > 0 {
		name = args[0]
	}
	if len(m.agents) == 0 {
		return m, nil
	}
	ag := m.agents[0].Spawn()
	sp := spinner.New()
	sp.Spinner = spinner.Line
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.AIBarColor))
	info := &AgentInfo{Agent: ag, Status: StatusIdle, Spinner: sp, Name: name}
	m.infos[ag.ID] = info
	m.order = append(m.order, ag.ID)
	if m.team != nil {
		m.team.Add(name, ag)
	}
	m.active = ag.ID
	m.vp.SetContent("")
	return m, nil
}

func (m Model) handleSwitch(args []string) (Model, tea.Cmd) {
	if len(args) == 0 {
		return m, nil
	}
	prefix := args[0]
	for _, id := range m.order {
		if strings.HasPrefix(id.String(), prefix) {
			m.active = id
			if info, ok := m.infos[id]; ok {
				base := lipgloss.NewStyle().Foreground(lipgloss.Color(m.theme.Palette.Foreground)).Background(lipgloss.Color(m.theme.Palette.Background))
				m.vp.SetContent(base.Copy().Width(m.vp.Width).Render(info.History))
			}
			break
		}
	}
	return m, nil
}

func (m Model) handleStop(args []string) (Model, tea.Cmd) {
	id := m.active
	if len(args) > 0 {
		pref := args[0]
		for _, aid := range m.order {
			if strings.HasPrefix(aid.String(), pref) {
				id = aid
				break
			}
		}
	}
	if info, ok := m.infos[id]; ok {
		if info.Cancel != nil {
			info.Cancel()
		}
		info.Status = StatusStopped
		m.infos[id] = info
	}
	return m, nil
}

func (m Model) handleConverse(args []string) (Model, tea.Cmd) {
	if len(args) < 2 {
		return m, nil
	}
	n, err := strconv.Atoi(args[0])
	if err != nil {
		return m, nil
	}
	topic := strings.Join(args[1:], " ")
	if len(m.agents) == 0 {
		return m, nil
	}
	tm, err := NewTeam(m.agents[0], n, topic)
	if err != nil {
		m.err = err
		return m, nil
	}
	go func() { _ = tea.NewProgram(tm).Start() }()
	return m, nil
}

func (m Model) cycleActive(delta int) Model {
	if len(m.order) == 0 {
		return m
	}
	idx := 0
	for i, id := range m.order {
		if id == m.active {
			idx = i
			break
		}
	}
	idx = (idx + delta + len(m.order)) % len(m.order)
	m.active = m.order[idx]
	if info, ok := m.infos[m.active]; ok {
		m.vp.SetContent(info.History)
	}
	return m
}

func helpView() string {
	return strings.Join([]string{
		"/spawn <n>    - create a new agent",
		"/switch <prefix> - focus an agent",
		"/stop <prefix>   - stop an agent",
		"/converse <n> <topic> - side conversation",
	}, "\n")
}

func (m Model) statusDot(st AgentStatus) string {
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

func (m Model) renderTokenBar(count, max int) string {
	if max <= 0 {
		max = 1
	}
	pct := float64(count) / float64(max)
	if pct > 1 {
		pct = 1
	}
	filled := int(pct * 10)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
	return fmt.Sprintf("%s %d%%", bar, int(pct*100))
}

func (m Model) renderSparkline(history []int) string {
	if len(history) == 0 {
		return ""
	}
	chars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	min, max := history[0], history[0]
	for _, v := range history {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	var b strings.Builder
	for _, v := range history {
		if max == min {
			b.WriteString(chars[0])
		} else {
			n := float64(v-min) / float64(max-min)
			idx := int(n * float64(len(chars)-1))
			b.WriteString(chars[idx])
		}
	}
	return b.String()
}
