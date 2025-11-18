package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/google/uuid"
)

// ActivityEvent represents a single tool activity entry.
type ActivityEvent struct {
	Timestamp time.Time
	AgentID   uuid.UUID
	AgentName string
	Tool      string
	Summary   string
	Details   string
	Status    string
}

// ActivityFeedStyle controls the palette used when rendering the feed.
type ActivityFeedStyle struct {
	BorderColor        string
	BorderFocusedColor string
	HeaderColor        string
	TimeColor          string
	AgentColor         string
	ToolColor          string
	SummaryColor       string
	TitleColor         string
	TitleBackground    string
}

// ActivityFeed renders a scrollable table of tool activity across agents.
type ActivityFeed struct {
	viewport        viewport.Model
	events          []ActivityEvent
	perAgent        map[uuid.UUID][]ActivityEvent
	lastPerAgent    map[uuid.UUID]ActivityEvent
	maxEvents       int
	maxPerAgent     int
	focused         bool
	style           ActivityFeedStyle
	activeAgent     uuid.UUID
	activeAgentName string
	filtered        bool
	headerHeight    int
}

// NewActivityFeed constructs an ActivityFeed with the provided capacity and style.
func NewActivityFeed(maxEvents int, style ActivityFeedStyle) *ActivityFeed {
	if maxEvents <= 0 {
		maxEvents = 100
	}
	vp := viewport.New(0, 0)
	return &ActivityFeed{
		viewport:     vp,
		maxEvents:    maxEvents,
		maxPerAgent:  maxEvents,
		perAgent:     make(map[uuid.UUID][]ActivityEvent),
		lastPerAgent: make(map[uuid.UUID]ActivityEvent),
		style:        style,
		headerHeight: 2, // title + column header rows
	}
}

// SetSize updates the feed's dimensions. Height accounts for header rows.
func (f *ActivityFeed) SetSize(width, height int) {
	if width <= 0 {
		width = 1
	}
	if height <= f.headerHeight+1 {
		height = f.headerHeight + 1
	}
	f.viewport.Width = width
	f.viewport.Height = height - f.headerHeight
	if f.viewport.Height < 1 {
		f.viewport.Height = 1
	}
	f.rebuild()
}

// SetFocused toggles focused rendering state.
func (f *ActivityFeed) SetFocused(focused bool) {
	f.focused = focused
}

// Update forwards events to the internal viewport when focused.
func (f *ActivityFeed) Update(msg tea.Msg) tea.Cmd {
	switch msg.(type) {
	case tea.KeyMsg, tea.MouseMsg:
		if !f.focused {
			return nil
		}
	}
	var cmd tea.Cmd
	f.viewport, cmd = f.viewport.Update(msg)
	return cmd
}

// Append adds an event to the global and per-agent feeds.
func (f *ActivityFeed) Append(ev ActivityEvent) {
	f.events = append(f.events, ev)
	if len(f.events) > f.maxEvents {
		f.events = f.events[len(f.events)-f.maxEvents:]
	}

	agentEvents := append(f.perAgent[ev.AgentID], ev)
	if len(agentEvents) > f.maxPerAgent {
		agentEvents = agentEvents[len(agentEvents)-f.maxPerAgent:]
	}
	f.perAgent[ev.AgentID] = agentEvents
	f.lastPerAgent[ev.AgentID] = ev

	stickToBottom := !f.focused || f.viewport.AtBottom()
	f.rebuild()
	if stickToBottom {
		f.viewport.GotoBottom()
	}
}

// SetFilter restricts the feed to a single agent's events.
func (f *ActivityFeed) SetFilter(agentID uuid.UUID, agentName string) {
	f.filtered = true
	f.activeAgent = agentID
	f.activeAgentName = agentName
	f.rebuild()
}

// ClearFilter shows the global feed.
func (f *ActivityFeed) ClearFilter() {
	f.filtered = false
	f.activeAgent = uuid.Nil
	f.activeAgentName = ""
	f.rebuild()
}

// ActiveAgent returns the current filter agent (if any).
func (f *ActivityFeed) ActiveAgent() (uuid.UUID, bool) {
	if !f.filtered {
		return uuid.Nil, false
	}
	return f.activeAgent, true
}

// LastEvent returns the latest activity for the requested agent.
func (f *ActivityFeed) LastEvent(agentID uuid.UUID) (ActivityEvent, bool) {
	ev, ok := f.lastPerAgent[agentID]
	return ev, ok
}

// View renders the activity feed as a bordered panel.
func (f *ActivityFeed) View() string {
	title := f.renderTitle()
	header := f.renderHeader()
	body := f.viewport.View()

	borderColor := f.style.BorderColor
	if f.focused && f.style.BorderFocusedColor != "" {
		borderColor = f.style.BorderFocusedColor
	}

	container := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(borderColor)).
		Padding(0, 1)

	content := lipgloss.JoinVertical(lipgloss.Left, title, header, body)
	return container.Width(f.viewport.Width + 2).Render(content)
}

func (f *ActivityFeed) renderTitle() string {
	scope := "GLOBAL"
	if f.filtered && f.activeAgentName != "" {
		scope = strings.ToUpper(f.activeAgentName)
	}
	title := fmt.Sprintf("ACTIVITY • %s", scope)
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.style.TitleColor)).
		Bold(true)
	if f.style.TitleBackground != "" {
		style = style.Background(lipgloss.Color(f.style.TitleBackground))
	}
	return style.Render(title)
}

func (f *ActivityFeed) renderHeader() string {
	widths := f.columnWidths()
	header := fmt.Sprintf("%s  %s  %s  %s",
		padString("TIME", widths[0]),
		padString("AGENT", widths[1]),
		padString("TOOL", widths[2]),
		padString("DETAILS", widths[3]))
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(f.style.HeaderColor)).
		Bold(true).
		Render(header)
}

func (f *ActivityFeed) rebuild() {
	events := f.events
	if f.filtered {
		events = f.perAgent[f.activeAgent]
	}

	widths := f.columnWidths()
	var rows []string
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(f.style.TimeColor))
	agentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(f.style.AgentColor))
	toolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(f.style.ToolColor))
	summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(f.style.SummaryColor))

	for _, ev := range events {
		timeStr := ev.Timestamp.Format("15:04:05")
		agent := truncate(ev.AgentName, widths[1])
		tool := truncate(ev.Tool, widths[2])
		summary := truncate(ev.Summary, widths[3])

		row := lipgloss.JoinHorizontal(lipgloss.Top,
			timeStyle.Render(padString(timeStr, widths[0])),
			"  ",
			agentStyle.Render(padString(agent, widths[1])),
			"  ",
			toolStyle.Render(padString(tool, widths[2])),
			"  ",
			summaryStyle.Render(padString(summary, widths[3])),
		)
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		rows = append(rows, summaryStyle.Faint(true).Render("No activity yet"))
	}

	f.viewport.SetContent(strings.Join(rows, "\n"))
}

func (f *ActivityFeed) columnWidths() [4]int {
	width := f.viewport.Width
	if width <= 0 {
		width = 60
	}
	timeWidth := 8
	agentWidth := 12
	toolWidth := 14
	remaining := width - (timeWidth + agentWidth + toolWidth + 6)
	if remaining < 10 {
		remaining = 10
	}
	return [4]int{timeWidth, agentWidth, toolWidth, remaining}
}

// Events returns a copy of the global activity history.
func (f *ActivityFeed) Events() []ActivityEvent {
	return append([]ActivityEvent(nil), f.events...)
}

// HasEvents reports whether any activity has been recorded.
func (f *ActivityFeed) HasEvents() bool {
	return len(f.events) > 0
}

func padString(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "…"
}
