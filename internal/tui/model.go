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

	dec *json.Decoder

	history string

	activeTab int
	width     int
	height    int

	err error
}

// New creates a new TUI model bound to an Agent.
func New(ag *core.Agent) Model {
	items := []list.Item{}
	for name, tl := range ag.Tools {
		items = append(items, listItem{name: name, desc: tl.Description()})
	}
	l := list.New(items, listItemDelegate{}, 0, 0)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Title = "Tools"

	ti := textinput.New()
	ti.Placeholder = "Message"
	ti.Focus()

	vp := viewport.New(0, 0)
	cwd, _ := os.Getwd()

	return Model{agent: ag, vp: vp, input: ti, tools: l, history: "", cwd: cwd, modelName: "unknown"}
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
	if m.dec == nil {
		return nil
	}
	for {
		var ev trace.Event
		if err := m.dec.Decode(&ev); err != nil {
			if err == io.EOF {
				return nil
			}
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
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.activeTab = 1 - m.activeTab
		case "enter":
			if m.input.Focused() {
				txt := m.input.Value()
				m.input.SetValue("")
				m.history += userBar() + " " + txt + "\n"
				m.vp.SetContent(lipgloss.NewStyle().Width(m.vp.Width).Render(m.history))

				pr, pw := io.Pipe()
				errCh := make(chan error, 1)
				m.agent.Tracer = trace.NewJSONL(pw)
				m.dec = json.NewDecoder(bufio.NewReader(pr))
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
		m.history += aiBar() + " "
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

func userBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#8B5CF6")).Render("│")
}

func aiBar() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("│")
}

func renderMemory(ag *core.Agent) string {
	hist := ag.Mem.History()
	var b bytes.Buffer
	for i, s := range hist {
		b.WriteString("Step ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(": ")
		b.WriteString(s.Output)
		if s.ToolName != "" {
			b.WriteString(" -> ")
			b.WriteString(s.ToolName)
			b.WriteString(": ")
			b.WriteString(s.ToolResult)
		}
		b.WriteString("\n")
	}
	return b.String()
}
