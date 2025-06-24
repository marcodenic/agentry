package tests

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestLoadPermissionsFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "p.yaml")
	if err := os.WriteFile(path, []byte("tools:\n  - echo\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := tool.LoadPermissionsFile(path); err != nil {
		t.Fatalf("load: %v", err)
	}
	reg := tool.DefaultRegistry()
	okTool, _ := reg.Use("echo")
	if _, err := okTool.Execute(context.Background(), map[string]any{"text": "hi"}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	deny, _ := reg.Use("ls")
	if _, err := deny.Execute(context.Background(), nil); !errors.Is(err, tool.ErrToolDenied) {
		t.Fatalf("expected denied, got %v", err)
	}
}
