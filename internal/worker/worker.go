package worker

import (
	"context"
	"log"

	"github.com/marcodenic/agentry/internal/mocknats"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

// Conn is the subset of nats.Conn used by Start.
type Conn interface {
	QueueSubscribe(subj, queue string, cb mocknats.Handler) error
	Flush() error
}

// DefaultAgent returns a minimal agent with builtin tools and mock model.
func DefaultAgent() *core.Agent {
	reg := tool.DefaultRegistry()
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: model.NewMock()}}
	return core.New(route, reg, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
}

// Start subscribes to the given queue and processes messages with the agent.
func Start(ctx context.Context, nc Conn, queue string, concurrency int, ag *core.Agent) error {
	sem := make(chan struct{}, concurrency)
	err := nc.QueueSubscribe(queue, "workers", func(m *mocknats.Msg) {
		sem <- struct{}{}
		go func(msg *mocknats.Msg) {
			defer func() { <-sem }()
			if _, err := ag.Run(context.Background(), string(msg.Data)); err != nil {
				log.Printf("task error: %v", err)
			}
		}(m)
	})
	if err != nil {
		return err
	}
	if err := nc.Flush(); err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
	}()
	return nil
}
