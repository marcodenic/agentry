package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
	"github.com/marcodenic/agentry/internal/todo"
)

// TodoBoard manages the TODO board view
// Uses the shared todo.Service so both TUI and tool builtins stay in sync.
type TodoBoard struct {
	list     list.Model
	width    int
	height   int
	selected bool
}

// todoMsg is sent when TODO items are updated
type todoMsg struct {
	items []todo.Item
}

// NewTodoBoard creates a new TODO board component
func NewTodoBoard() TodoBoard {
	items := []list.Item{}

	l := list.New(items, todoItemDelegate{}, 0, 0)
	l.Title = "📋 TODO Board"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230")).
		Padding(0, 1)

	return TodoBoard{
		list: l,
	}
}

// todoItemDelegate defines how TODO items are rendered
type todoItemDelegate struct{}

func (d todoItemDelegate) Height() int                               { return 3 }
func (d todoItemDelegate) Spacing() int                              { return 1 }
func (d todoItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d todoItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(todo.Item)
	if !ok {
		return
	}

	// Status indicator
	var statusGlyph, statusColor string
	switch item.Status {
	case "todo":
		statusGlyph = glyphs.RedCrossmark()
		statusColor = "9" // Red
	case "in_progress":
		statusGlyph = glyphs.YellowStar()
		statusColor = "11" // Yellow
	case "done":
		statusGlyph = glyphs.GreenCheckmark()
		statusColor = "10" // Green
	case "blocked":
		statusGlyph = glyphs.OrangeTriangle()
		statusColor = "208" // Orange
	default:
		statusGlyph = glyphs.BlueCircle()
		statusColor = "12" // Blue
	}

	// Priority indicator
	priorityGlyph := " "
	switch item.Priority {
	case "high":
		priorityGlyph = "🔴"
	case "medium":
		priorityGlyph = "🟡"
	case "low":
		priorityGlyph = "🟢"
	}

	// Agent indicator
	agentInfo := ""
	if item.AgentID != "" {
		agentInfo = fmt.Sprintf(" [%s]", trimTo(item.AgentID, 8))
	}

	// Selected styling
	isSelected := index == m.Index()
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor))
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	metaStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("243"))

	if isSelected {
		titleStyle = titleStyle.Background(lipgloss.Color("62")).Bold(true)
		descStyle = descStyle.Background(lipgloss.Color("62"))
		metaStyle = metaStyle.Background(lipgloss.Color("62"))
	}

	// Format content
	title := fmt.Sprintf("%s %s %s%s", statusGlyph, priorityGlyph, item.Title, agentInfo)
	desc := item.Description
	if len(desc) > 60 {
		desc = desc[:57] + "..."
	}

	// Tags
	tagStr := ""
	if len(item.Tags) > 0 {
		tagStr = fmt.Sprintf(" #%s", strings.Join(item.Tags, " #"))
	}

	meta := fmt.Sprintf("ID: %s | Updated: %s%s",
		trimTo(item.ID, 8),
		item.UpdatedAt.Format("15:04"),
		tagStr)

	content := fmt.Sprintf("%s\n%s\n%s",
		titleStyle.Render(title),
		descStyle.Render(desc),
		metaStyle.Render(meta))

	fmt.Fprint(w, content)
}

// Update handles TODO board updates
func (tb TodoBoard) Update(msg tea.Msg) (TodoBoard, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case todoMsg:
		items := make([]list.Item, len(msg.items))
		for i, todo := range msg.items {
			items[i] = todo
		}
		tb.list.SetItems(items)

	case tea.WindowSizeMsg:
		tb.width = msg.Width
		tb.height = msg.Height
		tb.list.SetWidth(msg.Width)
		tb.list.SetHeight(msg.Height - 2) // Leave space for borders
	}

	tb.list, cmd = tb.list.Update(msg)
	return tb, cmd
}

// View renders the TODO board
func (tb TodoBoard) View() string {
	if tb.width == 0 {
		return "Loading TODO board..."
	}

	return tb.list.View()
}

// SetSelected sets whether this component is selected/focused
func (tb *TodoBoard) SetSelected(selected bool) {
	tb.selected = selected
}

// LoadTodos fetches TODO items using the shared service
func LoadTodos() tea.Cmd {
	return func() tea.Msg {
		svc, err := todo.NewService(nil, "")
		if err != nil {
			return todoMsg{items: []todo.Item{}}
		}

		items, err := svc.List(todo.Filter{})
		if err != nil {
			return todoMsg{items: []todo.Item{}}
		}

		return todoMsg{items: items}
	}
}

func trimTo(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

var _ list.Item = todo.Item{}
