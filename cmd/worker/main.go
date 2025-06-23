package main

import (
	"context"
	"flag"
	"log"

	"github.com/marcodenic/agentry/internal/mocknats"
	"github.com/marcodenic/agentry/internal/worker"
)

type stubConn struct{}

func (stubConn) QueueSubscribe(subj, queue string, cb mocknats.Handler) error { return nil }
func (stubConn) Flush() error                                                 { return nil }

func main() {
	natsURL := flag.String("nats", "", "NATS server URL (unused)")
	queue := flag.String("queue", "tasks", "queue name")
	concurrency := flag.Int("concurrency", 1, "number of concurrent workers")
	flag.Parse()

	_ = natsURL // placeholder

	var nc stubConn

	ag := worker.DefaultAgent()
	if err := worker.Start(context.Background(), nc, *queue, *concurrency, ag); err != nil {
		log.Fatal(err)
	}

	select {}
}
