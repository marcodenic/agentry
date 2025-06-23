//go:build integration
// +build integration

package tests

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/pkg/memstore"
)

// simple client returning constant output
type recordClient struct{}

func (recordClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{Content: "hello"}, nil
}

func TestAgentSaveLoad(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "mem.db")
	store, err := memstore.NewSQLite(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: recordClient{}}}
	ag := core.New(route, nil, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)

	if _, err := ag.Run(context.Background(), "hi"); err != nil {
		t.Fatal(err)
	}

	if err := ag.SaveState(context.Background(), "run1"); err != nil {
		t.Fatal(err)
	}

	ag2 := core.New(route, nil, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)
	if err := ag2.LoadState(context.Background(), "run1"); err != nil {
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
