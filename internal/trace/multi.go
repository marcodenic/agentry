package trace

import "context"

// MultiWriter dispatches events to multiple writers.
type MultiWriter struct{ writers []Writer }

func NewMulti(w ...Writer) *MultiWriter { return &MultiWriter{writers: w} }

func (m *MultiWriter) Write(ctx context.Context, e Event) {
	for _, w := range m.writers {
		w.Write(ctx, e)
	}
}
