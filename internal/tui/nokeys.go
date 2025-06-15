package tui

import "github.com/charmbracelet/bubbles/key"

// NoNavKeyMap disables all navigation keys for the list.
var NoNavKeyMap = struct {
	CursorUp    key.Binding
	CursorDown  key.Binding
	PrevPage    key.Binding
	NextPage    key.Binding
	GoToStart   key.Binding
	GoToEnd     key.Binding
	Filter      key.Binding
	ClearFilter key.Binding
	CancelWhileFiltering key.Binding
	AcceptWhileFiltering key.Binding
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding
	Quit key.Binding
	ForceQuit key.Binding
}{
	CursorUp:    key.NewBinding(),
	CursorDown:  key.NewBinding(),
	PrevPage:    key.NewBinding(),
	NextPage:    key.NewBinding(),
	GoToStart:   key.NewBinding(),
	GoToEnd:     key.NewBinding(),
	Filter:      key.NewBinding(),
	ClearFilter: key.NewBinding(),
	CancelWhileFiltering: key.NewBinding(),
	AcceptWhileFiltering: key.NewBinding(),
	ShowFullHelp:  key.NewBinding(),
	CloseFullHelp: key.NewBinding(),
	Quit: key.NewBinding(),
	ForceQuit: key.NewBinding(),
}
