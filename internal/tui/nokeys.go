package tui

import (
	"github.com/charmbracelet/bubbles/key"
	"github.com/marcodenic/agentry/internal/glyphs"
)

// Enhanced key bindings for cycling agents in the unified TUI model.
var (
	PrevAgentKey = key.NewBinding(
		key.WithKeys("shift+left", "ctrl+p"),
		key.WithHelp("shift+"+glyphs.ArrowLeft+"/ctrl+p", "prev agent"),
	)
	NextAgentKey = key.NewBinding(
		key.WithKeys("shift+right", "ctrl+n"),
		key.WithHelp("shift+"+glyphs.ArrowRight+"/ctrl+n", "next agent"),
	)

	// Additional navigation keys for better UX
	FirstAgentKey = key.NewBinding(
		key.WithKeys("home", "ctrl+a"),
		key.WithHelp("home/ctrl+a", "first agent"),
	)
	LastAgentKey = key.NewBinding(
		key.WithKeys("end", "ctrl+e"),
		key.WithHelp("end/ctrl+e", "last agent"),
	)
)

// NoNavKeyMap disables all navigation keys for the list.
var NoNavKeyMap = struct {
	CursorUp             key.Binding
	CursorDown           key.Binding
	PrevPage             key.Binding
	NextPage             key.Binding
	GoToStart            key.Binding
	GoToEnd              key.Binding
	Filter               key.Binding
	ClearFilter          key.Binding
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding
	ShowFullHelp         key.Binding
	CloseFullHelp        key.Binding
	Quit                 key.Binding
	ForceQuit            key.Binding
}{
	CursorUp:             key.NewBinding(),
	CursorDown:           key.NewBinding(),
	PrevPage:             key.NewBinding(),
	NextPage:             key.NewBinding(),
	GoToStart:            key.NewBinding(),
	GoToEnd:              key.NewBinding(),
	Filter:               key.NewBinding(),
	ClearFilter:          key.NewBinding(),
	CancelWhileFiltering: key.NewBinding(),
	AcceptWhileFiltering: key.NewBinding(),
	ShowFullHelp:         key.NewBinding(),
	CloseFullHelp:        key.NewBinding(),
	Quit:                 key.NewBinding(),
	ForceQuit:            key.NewBinding(),
}
