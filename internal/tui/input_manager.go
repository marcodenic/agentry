package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

type inputManager struct {
	model        textarea.Model
	history      []string
	historyIndex int
	height       int
}

func newInputManager() inputManager {
	im := inputManager{
		model:        newTextareaModel(),
		historyIndex: -1,
		height:       1,
	}
	im.model.SetHeight(1)
	return im
}

func (im *inputManager) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	im.model, cmd = im.model.Update(msg)
	return cmd
}

func (im *inputManager) Focused() bool { return im.model.Focused() }
func (im *inputManager) Value() string { return im.model.Value() }
func (im *inputManager) View() string  { return im.model.View() }
func (im *inputManager) CursorEnd()    { im.model.CursorEnd() }
func (im *inputManager) SetValue(v string) {
	im.model.SetValue(v)
}

func (im *inputManager) Height() int { return im.height }

func (im *inputManager) SetHeight(h int) {
	im.height = h
	im.model.SetHeight(h)
}

func (im *inputManager) SetWidth(w int) { im.model.SetWidth(w) }

func (im *inputManager) ResetAfterSend() {
	im.model.SetValue("")
	im.SetHeight(1)
	im.historyIndex = len(im.history)
}

func (im *inputManager) PushHistory(entry string) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return
	}
	im.history = append(im.history, entry)
	im.historyIndex = len(im.history)
}

func (im *inputManager) HistoryUp() {
	if len(im.history) == 0 {
		return
	}
	if im.historyIndex <= 0 {
		im.historyIndex = 0
	} else {
		im.historyIndex--
	}
	im.model.SetValue(im.history[im.historyIndex])
	im.model.CursorEnd()
}

func (im *inputManager) HistoryDown() {
	if len(im.history) == 0 {
		return
	}
	if im.historyIndex >= len(im.history)-1 {
		im.historyIndex = len(im.history)
		im.model.SetValue("")
		im.model.CursorEnd()
		return
	}
	im.historyIndex++
	im.model.SetValue(im.history[im.historyIndex])
	im.model.CursorEnd()
}

func (im *inputManager) Model() textarea.Model {
	return im.model
}
