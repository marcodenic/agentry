// +build integration

package tests

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func TestAgentCheckpointResume(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "mem.db")
	store, err := memstore.NewSQLite(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: recordClient{}}}
	ag := core.New(route, nil, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)
	ag.ID = uuid.New()

	if _, err := ag.Run(context.Background(), "hi"); err != nil {
		t.Fatal(err)
	}

	ag2 := core.New(route, nil, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)
	ag2.ID = ag.ID
	if err := ag2.Resume(context.Background()); err != nil {
		t.Fatal(err)
	}

	hist := ag2.Mem.History()
	if len(hist) != 1 {
		t.Fatalf("expected 1 step, got %d", len(hist))
	}
	if hist[0].Output != "hello" {
		t.Fatalf("unexpected output %s", hist[0].Output)
	}
}
