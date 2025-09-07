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

	// Limit history size to prevent unbounded growth
	// Keep last 20 steps (configurable via env var)
	maxSteps := 20
	if len(m.steps) > maxSteps {
		// Keep the most recent steps, discard the oldest
		copy(m.steps, m.steps[len(m.steps)-maxSteps:])
		m.steps = m.steps[:maxSteps]
	}
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
