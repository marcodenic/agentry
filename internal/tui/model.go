package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"
)

// Model is the root TUI model.
type Model struct {
	agent *core.Agent

	vp    viewport.Model
	input textinput.Model
	tools list.Model

	cwd        string
	tokenCount int
	modelName  string
	selected   string

	sc *bufio.Scanner

	history string

	activeTab int
	width     int
	height    int

	err error

	theme Theme
	keys  Keybinds
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

	return Model{agent: ag, vp: vp, input: ti, tools: l, history: "", cwd: cwd, modelName: "unknown", theme: th, keys: th.Keybinds}
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

type tokenMsg string
type toolUseMsg string
type modelMsg string

type errMsg struct{ error }

type finalMsg string

func streamTokens(out string) tea.Cmd {
	runes := []rune(out)
	cmds := make([]tea.Cmd, len(runes))
	for i, r := range runes {
		tok := string(r)
		delay := time.Duration(i*30) * time.Millisecond
		cmds[i] = tea.Tick(delay, func(t time.Time) tea.Msg { return tokenMsg(tok) })
	}
	return tea.Batch(cmds...)
}

func (m *Model) readEvent() tea.Msg {
	if m.sc == nil {
		return nil
	}
	for {
		if !m.sc.Scan() {
			if err := m.sc.Err(); err != nil {
				return errMsg{err}
			}
			return nil
		}
		var ev trace.Event
		if err := json.Unmarshal(m.sc.Bytes(), &ev); err != nil {
			return errMsg{err}
		}
		switch ev.Type {
		case trace.EventFinal:
			if s, ok := ev.Data.(string); ok {
				return finalMsg(s)
			}
		case trace.EventModelStart:
			if name, ok := ev.Data.(string); ok {
				return modelMsg(name)
			}
		case trace.EventToolEnd:
			if m2, ok := ev.Data.(map[string]any); ok {
				if name, ok := m2["name"].(string); ok {
					return toolUseMsg(name)
				}
			}
		default:
			continue
		}
	}
}

func (m *Model) readCmd() tea.Cmd {
	return func() tea.Msg { return m.readEvent() }
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
		switch msg.String() {
		case m.keys.Quit:
			return m, tea.Quit
		case m.keys.ToggleTab:
			m.activeTab = 1 - m.activeTab
		case m.keys.Submit:
			if m.input.Focused() {
				txt := m.input.Value()
				m.input.SetValue("")
				m.history += m.userBar() + " " + txt + "\n"
				m.vp.SetContent(lipgloss.NewStyle().Width(m.vp.Width).Render(m.history))

				pr, pw := io.Pipe()
				errCh := make(chan error, 1)
				m.agent.Tracer = trace.NewJSONL(pw)
				m.sc = bufio.NewScanner(pr)
				go func() {
					_, err := m.agent.Run(context.Background(), txt)
					pw.Close()
					errCh <- err
				}()
				return m, tea.Batch(m.readCmd(), waitErr(errCh))
			}
		}
	case tokenMsg:
		m.history += string(msg)
		m.tokenCount++
		m.vp.SetContent(lipgloss.NewStyle().Width(m.vp.Width).Render(m.history))
		m.vp.GotoBottom()
	case finalMsg:
		m.history += m.aiBar() + " "
		m.vp.SetContent(lipgloss.NewStyle().Width(m.vp.Width).Render(m.history))
		m.vp.GotoBottom()
		return m, tea.Batch(streamTokens(string(msg)+"\n"), m.readCmd())
	case toolUseMsg:
		idx := -1
		for i, it := range m.tools.Items() {
			if li, ok := it.(listItem); ok && li.name == string(msg) {
				idx = i
				break
			}
		}
		if idx >= 0 {
			m.tools.Select(idx)
		}
		return m, m.readCmd()
	case modelMsg:
		m.modelName = string(msg)
		return m, m.readCmd()
	case errMsg:
		m.err = msg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Height = msg.Height - 5
		m.tools.SetSize(int(float64(msg.Width)*0.25)-2, msg.Height-2)
		m.vp.SetContent(lipgloss.NewStyle().Width(m.vp.Width).Render(m.history))
	}

	m.input, _ = m.input.Update(msg)
	m.tools, _ = m.tools.Update(msg)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	left := lipgloss.NewStyle().Width(int(float64(m.width) * 0.25)).Render(m.tools.View())

	var rightContent string
	if m.activeTab == 0 {
		rightContent = m.vp.View() + "\n" + m.input.View()
	} else {
		rightContent = renderMemory(m.agent)
	}
	if m.err != nil {
		rightContent += "\nERR: " + m.err.Error()
	}
	right := lipgloss.NewStyle().Width(int(float64(m.width) * 0.75)).Render(rightContent)
	main := lipgloss.JoinHorizontal(lipgloss.Top, left, right)
	footer := fmt.Sprintf("cwd: %s | tokens: %d | model: %s", m.cwd, m.tokenCount, m.modelName)
	footer = lipgloss.NewStyle().Width(m.width).Render(footer)
	return lipgloss.JoinVertical(lipgloss.Left, main, footer)
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
