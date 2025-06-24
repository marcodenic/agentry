package tests

import (
	"context"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/session"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func TestSessionGCRemovesExpired(t *testing.T) {
	store := memstore.NewInMemory()

	if err := store.Set(context.Background(), "history", "sess1", []byte("data")); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session.Start(ctx, store, 10*time.Millisecond, 10*time.Millisecond)

	time.Sleep(30 * time.Millisecond)

	b, err := store.Get(context.Background(), "history", "sess1")
	if err != nil {
		t.Fatal(err)
	}
	if b != nil {
		t.Fatalf("expected session to be GC'ed, found %v", b)
	}
}
