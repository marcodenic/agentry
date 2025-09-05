package team

import (
    "fmt"
    "os"
    "time"
)

// Timer utility for performance debugging
type Timer struct {
    start time.Time
    name  string
}

func StartTimer(name string) *Timer {
    timer := &Timer{start: time.Now(), name: name}
    if isDebug() {
        fmt.Fprintf(os.Stderr, "⏱️  [TIMER] Started: %s\n", name)
    }
    return timer
}

func (t *Timer) Stop() time.Duration {
    elapsed := time.Since(t.start)
    if isDebug() {
        fmt.Fprintf(os.Stderr, "⏱️  [TIMER] %s: %v\n", t.name, elapsed)
    }
    return elapsed
}

func (t *Timer) Checkpoint(checkpoint string) time.Duration {
    elapsed := time.Since(t.start)
    if isDebug() {
        fmt.Fprintf(os.Stderr, "⏱️  [TIMER] %s [%s]: %v\n", t.name, checkpoint, elapsed)
    }
    return elapsed
}
