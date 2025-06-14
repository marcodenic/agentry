package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
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

	return Model{agent: ag, vp: vp, input: ti, tools: l}
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

type errMsg struct{ error }

type finalMsg string

func streamTokens(out string) tea.Cmd {
	runes := []rune(out)
	cmds := make([]tea.Cmd, len(runes)+1)
	for i, r := range runes {
		tok := string(r)
		delay := time.Duration(i*30) * time.Millisecond
		cmds[i] = tea.Tick(delay, func(t time.Time) tea.Msg { return tokenMsg(tok) })
	}
	// emit newline at the end so AI response appears on its own line
	cmds[len(runes)] = tea.Tick(time.Duration(len(runes)*30)*time.Millisecond, func(t time.Time) tea.Msg {
		return tokenMsg("\n")
	})
	return tea.Batch(cmds...)
}

func readEvents(r io.Reader) tea.Cmd {
	return func() tea.Msg {
		dec := json.NewDecoder(bufio.NewReader(r))
		for {
			var ev trace.Event
			if err := dec.Decode(&ev); err != nil {
				if err == io.EOF {
					return nil
				}
				return errMsg{err}
			}
			if ev.Type == trace.EventFinal {
				if s, ok := ev.Data.(string); ok {
					return finalMsg(s)
				}
			}
		}
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
				m.vp.SetContent(m.vp.View() + "You: " + txt + "\n")

				pr, pw := io.Pipe()
				m.agent.Tracer = trace.NewJSONL(pw)
				go func() {
					_, err := m.agent.Run(context.Background(), txt)
					pw.Close()
					if err != nil {
						cmds = append(cmds, func() tea.Msg { return errMsg{err} })
					}
				}()
				return m, readEvents(pr)
			}
		case "up", "k":
			m.tools.CursorUp()
		case "down", "j":
			m.tools.CursorDown()
		}
	case tokenMsg:
		m.vp.SetContent(m.vp.View() + string(msg))
	case finalMsg:
		m.vp.SetContent(m.vp.View() + "AI: ")
		return m, streamTokens(string(msg))
	case errMsg:
		m.err = msg
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Width = int(float64(msg.Width)*0.75) - 2
		m.vp.Height = msg.Height - 5
		m.tools.SetSize(int(float64(msg.Width)*0.25)-2, msg.Height-2)
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
	right := lipgloss.NewStyle().Width(int(float64(m.width) * 0.75)).Render(rightContent)

	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
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
