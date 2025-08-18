package trace

import "context"

// MemoryWriter stores events in memory for debugging.
type MemoryWriter struct {
	Events []Event
	Limit  int
}

func NewMemory(limit int) *MemoryWriter {
	return &MemoryWriter{Limit: limit}
}

func (m *MemoryWriter) Write(_ context.Context, e Event) {
	if m.Limit > 0 && len(m.Events) >= m.Limit {
		copy(m.Events, m.Events[1:])
		m.Events[len(m.Events)-1] = e
	} else {
		m.Events = append(m.Events, e)
	}
}

func (m *MemoryWriter) All() []Event {
	out := make([]Event, len(m.Events))
	copy(out, m.Events)
	return out
}
