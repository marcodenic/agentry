package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/mocknats"
	"github.com/marcodenic/agentry/internal/worker"
)

type fakeConn struct {
	mu sync.Mutex
	cb mocknats.Handler
}

func (f *fakeConn) QueueSubscribe(subj, queue string, cb mocknats.Handler) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.cb = cb
	return nil
}

func (f *fakeConn) Flush() error { return nil }

func (f *fakeConn) publish(msg string) {
	f.mu.Lock()
	cb := f.cb
	f.mu.Unlock()
	if cb != nil {
		cb(&mocknats.Msg{Data: []byte(msg)})
	}
}

func TestWorkerProcessesMessages(t *testing.T) {
	fc := &fakeConn{}
	ag := worker.DefaultAgent()
	if err := worker.Start(context.Background(), fc, "jobs", 1, ag); err != nil {
		t.Fatal(err)
	}
	fc.publish("hi")
	time.Sleep(20 * time.Millisecond)
	hist := ag.Mem.History()
	if len(hist) == 0 || hist[len(hist)-1].Output != "hello" {
		t.Fatalf("unexpected history: %+v", hist)
	}
}
