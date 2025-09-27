package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
)

type runtimeService struct{}

func newRuntimeService() runtimeService {
	return runtimeService{}
}

func (runtimeService) ReadCmd(model *Model, id uuid.UUID) tea.Cmd {
	return func() tea.Msg {
		return model.readEvent(id)
	}
}

func (runtimeService) WaitErr(ch <-chan error) tea.Cmd {
	return func() tea.Msg {
		if err := <-ch; err != nil {
			return errMsg{err}
		}
		return nil
	}
}

func (runtimeService) WaitComplete(id uuid.UUID, ch <-chan string) tea.Cmd {
	return func() tea.Msg {
		result := <-ch
		return agentCompleteMsg{id: id, result: result}
	}
}

func (runtimeService) StartThinkingAnimation(id uuid.UUID) tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		frame := int(t.UnixMilli()/100) % len(spinnerFrames)
		return thinkingAnimationMsg{id: id, frame: frame}
	})
}
