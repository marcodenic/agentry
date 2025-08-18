package memory

// Simple conversation memory.

import (
	"sync"

	"github.com/marcodenic/agentry/internal/model"
)

type Step struct {
	Output      string
	ToolCalls   []model.ToolCall
	ToolResults map[string]string
}

type Store interface {
	AddStep(step Step)
	History() []Step
	SetHistory([]Step)
}

// InMemory is a thread-safe implementation.
type InMemory struct {
	mu    sync.Mutex
	steps []Step
}

func NewInMemory() *InMemory { return &InMemory{} }

func (m *InMemory) AddStep(step Step) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps = append(m.steps, step)
}

func (m *InMemory) History() []Step {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Step, len(m.steps))
	copy(cp, m.steps)
	return cp
}

func (m *InMemory) SetHistory(hist []Step) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Step, len(hist))
	copy(cp, hist)
	m.steps = cp
}
