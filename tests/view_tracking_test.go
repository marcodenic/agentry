package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestViewWriteEditTracking(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "f.txt")
	if err := os.WriteFile(file, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	reg := tool.DefaultRegistry()
	viewT, _ := reg.Use("view")
	writeT, _ := reg.Use("write")
	editT, _ := reg.Use("edit")

	// editing without viewing should fail
	if _, err := editT.Execute(context.Background(), map[string]any{"path": file, "text": "x"}); err == nil {
		t.Fatal("expected error when editing without view")
	}

	// view then edit succeeds
	if _, err := viewT.Execute(context.Background(), map[string]any{"path": file}); err != nil {
		t.Fatalf("view failed: %v", err)
	}
	if _, err := editT.Execute(context.Background(), map[string]any{"path": file, "text": "x"}); err != nil {
		t.Fatalf("edit after view failed: %v", err)
	}

	// external change should trigger error
	time.Sleep(1 * time.Second)
	if err := os.WriteFile(file, []byte("changed"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, err := writeT.Execute(context.Background(), map[string]any{"path": file, "text": "again"}); err == nil {
		t.Fatal("expected error after external change")
	}
}
