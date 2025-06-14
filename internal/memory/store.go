package memory

// Simple conversation memory.

import "sync"

type Step struct {
	Output     string
	ToolName   string
	ToolResult string
}

type Store interface {
	AddStep(out, tool, result string)
	History() []Step
}

// InMemory is a thread-safe implementation.
type InMemory struct {
	mu    sync.Mutex
	steps []Step
}

func NewInMemory() *InMemory { return &InMemory{} }

func (m *InMemory) AddStep(out, tool, result string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.steps = append(m.steps, Step{Output: out, ToolName: tool, ToolResult: result})
}

func (m *InMemory) History() []Step {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]Step, len(m.steps))
	copy(cp, m.steps)
	return cp
}
