package tui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/marcodenic/agentry/internal/statusbar"
)

type viewState struct {
	Chat        chatPane
	Tools       list.Model
	Input       inputManager
	Diagnostics diagnosticsView
	Todo        TodoBoard
	Robot       *RobotFace
	Status      statusbar.Model
}

type layoutState struct {
	width     int
	height    int
	activeTab int
}

type chatPane struct {
	Main            viewport.Model
	Debug           viewport.Model
	showInitialLogo bool
	lastWidth       int
}

type diagnosticsView struct {
	Entries []Diag
	Running bool
}

func newChatPane(main, debug viewport.Model, showLogo bool) chatPane {
	return chatPane{Main: main, Debug: debug, showInitialLogo: showLogo}
}

func (c *chatPane) ShowInitialLogo() bool { return c.showInitialLogo }

func (c *chatPane) SetShowInitialLogo(show bool) { c.showInitialLogo = show }

func (c *chatPane) LastWidth() int { return c.lastWidth }

func (c *chatPane) SetLastWidth(width int) { c.lastWidth = width }

func (v *viewState) UpdateDiagnostics(newEntries []Diag, running bool) {
	v.Diagnostics.Entries = newEntries
	v.Diagnostics.Running = running
}

func (v *viewState) AppendDiagnostic(d Diag) {
	v.Diagnostics.Entries = append(v.Diagnostics.Entries, d)
}

func (v *viewState) ClearDiagnostics() {
	v.Diagnostics.Entries = nil
	v.Diagnostics.Running = false
}

func (v *viewState) DiagnosticsActive() bool {
	return len(v.Diagnostics.Entries) > 0 || v.Diagnostics.Running
}

func (v *viewState) SetToolsSize(width, height int) {
	v.Tools.SetSize(width, height)
}

func (v *viewState) SetInputWidth(width int) {
	v.Input.SetWidth(width)
}

func (l *layoutState) Resize(msg tea.WindowSizeMsg) {
	l.width = msg.Width
	l.height = msg.Height
}
