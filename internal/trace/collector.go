package trace

import "context"

// Collector captures trace events and optionally forwards them to another Writer.
type Collector struct {
	events []Event
	next   Writer
}

// NewCollector returns a Collector that forwards events to next.
func NewCollector(next Writer) *Collector { return &Collector{next: next} }

// Write appends the event and forwards it.
func (c *Collector) Write(ctx context.Context, e Event) {
	c.events = append(c.events, e)
	if c.next != nil {
		c.next.Write(ctx, e)
	}
}

// Events returns all captured events.
func (c *Collector) Events() []Event { return c.events }
