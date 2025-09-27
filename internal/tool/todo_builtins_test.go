package tool

import (
	"context"
	"encoding/json"
	"sync"
	"testing"

	"github.com/marcodenic/agentry/internal/memstore"
	"github.com/marcodenic/agentry/internal/todo"
)

func installIsolatedTodoService(t *testing.T) *todo.Service {
	t.Helper()
	store := memstore.NewMemoryStore()
	svc, err := todo.NewService(store, t.TempDir())
	if err != nil {
		t.Fatalf("new todo service: %v", err)
	}
	todoSvc = svc
	todoErr = nil
	todoOnce = sync.Once{}
	todoOnce.Do(func() {})
	t.Cleanup(func() {
		todoSvc = nil
		todoErr = nil
		todoOnce = sync.Once{}
	})
	return svc
}

type todoAddResponse struct {
	Ok   bool      `json:"ok"`
	ID   string    `json:"id"`
	Item todo.Item `json:"item"`
}

type todoListResponse struct {
	Ok    bool        `json:"ok"`
	Count int         `json:"count"`
	Items []todo.Item `json:"items"`
}

type todoGetResponse struct {
	Ok    bool      `json:"ok"`
	Item  todo.Item `json:"item"`
	Error string    `json:"error"`
}

func TestTodoBuiltinsLifecycle(t *testing.T) {
	svc := installIsolatedTodoService(t)
	ctx := context.Background()

	addOut, err := todoAddExec(ctx, map[string]any{
		"title":       "Write docs",
		"description": "Document the CLI",
		"priority":    "high",
		"tags":        []any{"docs", "cli"},
		"status":      "pending",
	})
	if err != nil {
		t.Fatalf("todoAddExec failed: %v", err)
	}
	var addResp todoAddResponse
	if err := json.Unmarshal([]byte(addOut), &addResp); err != nil {
		t.Fatalf("parse add response: %v", err)
	}
	if !addResp.Ok || addResp.ID == "" {
		t.Fatalf("unexpected add response: %+v", addResp)
	}

	listOut, err := todoListExec(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("todoListExec failed: %v", err)
	}
	var listResp todoListResponse
	if err := json.Unmarshal([]byte(listOut), &listResp); err != nil {
		t.Fatalf("parse list response: %v", err)
	}
	if listResp.Count != 1 || len(listResp.Items) != 1 {
		t.Fatalf("expected one todo item, got %+v", listResp)
	}

	getOut, err := todoGetExec(ctx, map[string]any{"id": addResp.ID})
	if err != nil {
		t.Fatalf("todoGetExec failed: %v", err)
	}
	var getResp todoGetResponse
	if err := json.Unmarshal([]byte(getOut), &getResp); err != nil {
		t.Fatalf("parse get response: %v", err)
	}
	if !getResp.Ok || getResp.Item.Title != "Write docs" {
		t.Fatalf("unexpected get response: %+v", getResp)
	}

	updateOut, err := todoUpdateExec(ctx, map[string]any{
		"id":     addResp.ID,
		"status": "done",
	})
	if err != nil {
		t.Fatalf("todoUpdateExec failed: %v", err)
	}
	var updateResp todoGetResponse
	if err := json.Unmarshal([]byte(updateOut), &updateResp); err != nil {
		t.Fatalf("parse update response: %v", err)
	}
	if !updateResp.Ok || updateResp.Item.Status != "done" {
		t.Fatalf("expected status updated to done, got %+v", updateResp)
	}

	delOut, err := todoDeleteExec(ctx, map[string]any{"id": addResp.ID})
	if err != nil {
		t.Fatalf("todoDeleteExec failed: %v", err)
	}
	var delResp struct {
		Ok      bool   `json:"ok"`
		Deleted string `json:"deleted"`
	}
	if err := json.Unmarshal([]byte(delOut), &delResp); err != nil {
		t.Fatalf("parse delete response: %v", err)
	}
	if !delResp.Ok || delResp.Deleted != addResp.ID {
		t.Fatalf("unexpected delete response: %+v", delResp)
	}

	// Ensure item removed
	items, err := svc.List(todo.Filter{})
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected todo store empty, got %d items", len(items))
	}

	missingGet, err := todoGetExec(ctx, map[string]any{"id": addResp.ID})
	if err != nil {
		t.Fatalf("todoGetExec missing failed: %v", err)
	}
	var missingResp todoGetResponse
	if err := json.Unmarshal([]byte(missingGet), &missingResp); err != nil {
		t.Fatalf("parse missing get response: %v", err)
	}
	if missingResp.Ok {
		t.Fatalf("expected not found response, got %+v", missingResp)
	}
}

func TestTodoBuiltinsValidationErrors(t *testing.T) {
	installIsolatedTodoService(t)
	ctx := context.Background()

	if _, err := todoAddExec(ctx, map[string]any{"title": " "}); err == nil {
		t.Fatalf("expected validation error for empty title")
	}

	if _, err := todoUpdateExec(ctx, map[string]any{"id": "missing"}); err == nil {
		t.Fatalf("expected error when updating missing todo")
	}

	if _, err := todoDeleteExec(ctx, map[string]any{"id": ""}); err == nil {
		t.Fatalf("expected error when deleting without id")
	}
}
