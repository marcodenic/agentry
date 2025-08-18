package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/audit"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestAuditMiddleware(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.jsonl")
	logWriter, err := audit.Open(path, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer logWriter.Close()

	reg := tool.WrapWithAudit(tool.DefaultRegistry(), logWriter)
	tl, _ := reg.Use("echo")
	if _, err := tl.Execute(context.Background(), map[string]any{"text": "hi"}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	parts := bytes.Split(bytes.TrimSpace(data), []byte("\n"))
	if len(parts) != 1 {
		t.Fatalf("expected 1 event got %d", len(parts))
	}
	var ev tool.AuditEvent
	if err := json.Unmarshal(parts[0], &ev); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if ev.Tool != "echo" || ev.Args["text"] != "hi" || ev.Timestamp.IsZero() {
		t.Fatalf("bad event: %+v", ev)
	}
}
