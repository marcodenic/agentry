package taskqueue

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
)

type Task struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type Queue struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	subj    string
	stream  string
	durable string
}

func NewQueue(url, subject string) (*Queue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	stream := strings.ReplaceAll(subject, ".", "_")
	if _, err := js.StreamInfo(stream); err != nil {
		if _, err := js.AddStream(&nats.StreamConfig{Name: stream, Subjects: []string{subject}}); err != nil {
			nc.Close()
			return nil, err
		}
	}
	return &Queue{conn: nc, js: js, subj: subject, stream: stream, durable: "workers"}, nil
}

func (q *Queue) Publish(ctx context.Context, task Task) error {
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}
	_, err = q.js.Publish(q.subj, b)
	return err
}

func (q *Queue) Subscribe(handler func(Task)) (*nats.Subscription, error) {
	return q.js.QueueSubscribe(q.subj, q.durable, func(msg *nats.Msg) {
		var t Task
		if err := json.Unmarshal(msg.Data, &t); err == nil {
			handler(t)
		}
		_ = msg.Ack()
	}, nats.Durable(q.durable), nats.ManualAck())
}

func (q *Queue) Close() {
	q.conn.Close()
}

// Lag reports the number of pending messages for the queue's consumer.
func (q *Queue) Lag(ctx context.Context) (int, error) {
	info, err := q.js.ConsumerInfo(q.stream, q.durable)
	if err != nil {
		return 0, err
	}
	return int(info.NumPending), nil
}
