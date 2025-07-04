//go:build integration
// +build integration

package e2e

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/pkg/memstore"
)

// recordClient returns constant output to simulate a model backend.
type recordClient struct{}

func (recordClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{Content: "hello"}, nil
}

// TestCheckpointResumeE2E verifies that an agent can be resumed from a checkpoint
// and continue processing with its previous state intact.
func TestCheckpointResumeE2E(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "mem.db")
	store, err := memstore.NewSQLite(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	client := recordClient{}
	ag := core.New(client, "mock", nil, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)

	// Run once and create a checkpoint.
	if _, err := ag.Run(context.Background(), "hi"); err != nil {
		t.Fatal(err)
	}

	// Restore from checkpoint using a new agent instance.
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

	// Continue running to ensure state persists across resumes.
	if _, err := ag2.Run(context.Background(), "there"); err != nil {
		t.Fatal(err)
	}

	ag3 := core.New(route, nil, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)
	ag3.ID = ag.ID
	if err := ag3.Resume(context.Background()); err != nil {
		t.Fatal(err)
	}

	hist = ag3.Mem.History()
	if len(hist) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(hist))
	}
}
