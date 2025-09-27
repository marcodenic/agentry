package todo

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/memstore"
)

func TestNewServiceEnsuresSchema(t *testing.T) {
	store := memstore.NewMemoryStore()
	svc, err := NewService(store, "/tmp/project")
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	if svc.Namespace() == "" {
		t.Fatalf("expected namespace to be set")
	}

	payload, ok, err := store.Get(svc.Namespace(), schemaKey)
	if err != nil {
		t.Fatalf("schema read error: %v", err)
	}
	if !ok {
		t.Fatalf("expected schema version to be written")
	}
	if got := strings.TrimSpace(string(payload)); got != SchemaVersion {
		t.Fatalf("schema mismatch: got %s want %s", got, SchemaVersion)
	}
}

func TestServiceCreateListLifecycle(t *testing.T) {
	store := memstore.NewMemoryStore()
	svc, err := NewService(store, "/workspace/demo")
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	first, err := svc.Create(CreateParams{Title: "Alpha", Tags: []string{" dev ", "Dev"}})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	if first.ID == "" {
		t.Fatalf("expected generated id")
	}
	if first.Status != "pending" {
		t.Fatalf("expected default status pending, got %s", first.Status)
	}
	if len(first.Tags) != 1 || first.Tags[0] != "dev" {
		t.Fatalf("expected normalized tags, got %#v", first.Tags)
	}

	time.Sleep(time.Nanosecond)

	second, err := svc.Create(CreateParams{Title: "Beta", Status: "done", Priority: "high", Tags: []string{"ops"}})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	items, err := svc.List(Filter{})
	if err != nil {
		t.Fatalf("List all: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != first.ID || items[1].ID != second.ID {
		t.Fatalf("unexpected ordering: %+v", items)
	}

	done, err := svc.List(Filter{Status: "done"})
	if err != nil {
		t.Fatalf("List status: %v", err)
	}
	if len(done) != 1 || done[0].ID != second.ID {
		t.Fatalf("status filter failed: %+v", done)
	}

	high, err := svc.List(Filter{Priority: "high"})
	if err != nil {
		t.Fatalf("List priority: %v", err)
	}
	if len(high) != 1 || high[0].ID != second.ID {
		t.Fatalf("priority filter failed: %+v", high)
	}

	byTag, err := svc.List(Filter{Tags: []string{"dev"}})
	if err != nil {
		t.Fatalf("List tags: %v", err)
	}
	if len(byTag) != 1 || byTag[0].ID != first.ID {
		t.Fatalf("tag filter failed: %+v", byTag)
	}

	none, err := svc.List(Filter{Tags: []string{"not-real"}})
	if err != nil {
		t.Fatalf("List missing tags: %v", err)
	}
	if len(none) != 0 {
		t.Fatalf("expected no matches for missing tags")
	}
}

func TestServiceUpdateAndDelete(t *testing.T) {
	store := memstore.NewMemoryStore()
	svc, err := NewService(store, "/workspace/demo2")
	if err != nil {
		t.Fatalf("NewService error: %v", err)
	}

	item, err := svc.Create(CreateParams{Title: "Task"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	title := "Updated"
	desc := "Detailed description"
	priority := "low"
	tags := []string{"refined"}
	agent := "agent-1234"
	status := "done"

	updated, err := svc.Update(item.ID, UpdateParams{
		Title:       &title,
		Description: &desc,
		Priority:    &priority,
		Tags:        &tags,
		AgentID:     &agent,
		Status:      &status,
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != title || updated.Description != desc || updated.Priority != priority || updated.Status != status {
		t.Fatalf("update failed: %+v", updated)
	}
	if len(updated.Tags) != 1 || updated.Tags[0] != "refined" {
		t.Fatalf("tags not updated: %#v", updated.Tags)
	}
	if updated.AgentID != agent {
		t.Fatalf("agent not updated: %s", updated.AgentID)
	}
	if !updated.UpdatedAt.After(item.UpdatedAt) {
		t.Fatalf("expected updated timestamp to advance")
	}

	fetched, ok, err := svc.Get(item.ID)
	if err != nil || !ok {
		t.Fatalf("Get after update failed: %v ok=%v", err, ok)
	}
	if fetched.Title != title {
		t.Fatalf("fetched item mismatch: %+v", fetched)
	}

	if err := svc.Delete(item.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if _, ok, err := svc.Get(item.ID); err != nil || ok {
		t.Fatalf("expected item to be removed, err=%v ok=%v", err, ok)
	}

	if err := svc.Delete(" "); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for empty delete, got %v", err)
	}

	if _, err := svc.Update("missing", UpdateParams{}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for missing update, got %v", err)
	}
}

func TestNewServiceSchemaMismatch(t *testing.T) {
	store := memstore.NewMemoryStore()
	ns := namespaceForPath("/workspace/schema")
	if err := store.Set(ns, schemaKey, []byte("999"), 0); err != nil {
		t.Fatalf("seed schema: %v", err)
	}

	if _, err := NewService(store, "/workspace/schema"); !errors.Is(err, ErrSchemaUnknown) {
		t.Fatalf("expected schema error, got %v", err)
	}
}

func TestNormalizeTagsDeterministic(t *testing.T) {
	tags := []string{"Beta", "alpha", "", "ALPHA", "dev"}
	out := normalizeTags(tags)
	expected := []string{"Beta", "alpha", "dev"}
	if len(out) != len(expected) {
		t.Fatalf("expected %d tags, got %d: %#v", len(expected), len(out), out)
	}
	for i, tag := range expected {
		if out[i] != tag {
			t.Fatalf("index %d: expected %s got %s", i, tag, out[i])
		}
	}
}

func TestFilterMatches(t *testing.T) {
	item := Item{Status: "done", Priority: "high", Tags: []string{"dev", "ops"}}
	if !(Filter{}).matches(item) {
		t.Fatalf("empty filter should match")
	}
	if (Filter{Status: "pending"}).matches(item) {
		t.Fatalf("status filter should fail")
	}
	if !(Filter{Status: "done"}).matches(item) {
		t.Fatalf("status filter should match")
	}
	if !(Filter{Priority: "high"}).matches(item) {
		t.Fatalf("priority filter should match")
	}
	if (Filter{Tags: []string{"missing"}}).matches(item) {
		t.Fatalf("tag filter should fail")
	}
	if !(Filter{Tags: []string{"DEV"}}).matches(item) {
		t.Fatalf("tag filter should be case-insensitive")
	}
}
