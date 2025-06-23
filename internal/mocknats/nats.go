package mocknats

// Msg represents a message delivered by NATS.
type Msg struct {
	Data []byte
}

// Handler processes a message.
type Handler func(*Msg)

// Conn is a minimal interface for subscribing to a queue.
type Conn interface {
	QueueSubscribe(subj, queue string, cb Handler) error
}
