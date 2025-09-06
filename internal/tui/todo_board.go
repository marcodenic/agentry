package tui

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/marcodenic/agentry/internal/glyphs"
	"github.com/marcodenic/agentry/internal/memstore"
)

// TodoItem represents a single TODO item for display
type TodoItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	AgentID     string    `json:"agent_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Tags        []string  `json:"tags"`
}

// Implement list.Item interface
func (t TodoItem) FilterValue() string { return t.Title }

// TodoBoard manages the TODO board view
type TodoBoard struct {
	list     list.Model
	width    int
	height   int
	selected bool
}

// todoMsg is sent when TODO items are updated
type todoMsg struct {
	items []TodoItem
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

func (d todoItemDelegate) Height() int                             { return 3 }
func (d todoItemDelegate) Spacing() int                            { return 1 }
func (d todoItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d todoItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(TodoItem)
	if !ok {
		return
	}

	// Status indicator
	var statusGlyph, statusColor string
	switch item.Status {
	case "todo":
		statusGlyph = glyphs.RedCrossmark()
		statusColor = "9"  // Red
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
		agentInfo = fmt.Sprintf(" [%s]", item.AgentID[:8])
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
		item.ID[:8], 
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

// LoadTodos fetches TODO items from the memstore
func LoadTodos() tea.Cmd {
	return func() tea.Msg {
		// Get TODO namespace (same logic as in todo_builtins.go)
		ns := todoNamespace()
		
		// List all TODO items
		keys, err := memstore.Get().Keys(ns)
		if err != nil {
			return todoMsg{items: []TodoItem{}}
		}

		var items []TodoItem
		for _, key := range keys {
			if !strings.HasPrefix(key, "item:") {
				continue
			}

			data, ok, err := memstore.Get().Get(ns, key)
			if err != nil || !ok {
				continue
			}

			var item TodoItem
			if err := json.Unmarshal(data, &item); err != nil {
				continue
			}

			items = append(items, item)
		}

		return todoMsg{items: items}
	}
}

// Helper function to get TODO namespace (mirrors todo_builtins.go)
func todoNamespace() string {
	cwd, _ := os.Getwd()
	abs, _ := filepath.Abs(cwd)
	h := sha1.Sum([]byte(abs))
	return "todo:project:" + hex.EncodeToString(h[:8])
}
