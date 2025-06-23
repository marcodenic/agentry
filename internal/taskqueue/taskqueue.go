package taskqueue

import (
	"context"
	"encoding/json"
	"github.com/nats-io/nats.go"
)

type Task struct {
	Type string      `json:"type"`
	Payload any      `json:"payload"`
}

type Queue struct {
	conn *nats.Conn
	subj string
}

func NewQueue(url, subject string) (*Queue, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, err
	}
	return &Queue{conn: nc, subj: subject}, nil
}

func (q *Queue) Publish(ctx context.Context, task Task) error {
	b, err := json.Marshal(task)
	if err != nil {
		return err
	}
	return q.conn.Publish(q.subj, b)
}

func (q *Queue) Subscribe(handler func(Task)) (*nats.Subscription, error) {
	return q.conn.Subscribe(q.subj, func(msg *nats.Msg) {
		var t Task
		if err := json.Unmarshal(msg.Data, &t); err == nil {
			handler(t)
		}
	})
}

func (q *Queue) Close() {
	q.conn.Close()
}
