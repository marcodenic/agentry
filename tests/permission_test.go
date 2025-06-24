package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestToolManifestAllowed(t *testing.T) {
	m := config.ToolManifest{Name: "echo", Type: "builtin", Permissions: config.ToolPermissions{Allow: boolPtr(true)}}
	tl, err := tool.FromManifest(m)
	if err != nil {
		t.Fatalf("from manifest: %v", err)
	}
	if _, err := tl.Execute(context.Background(), map[string]any{"text": "hi"}); err != nil {
		t.Fatalf("exec: %v", err)
	}
}

func TestToolManifestDenied(t *testing.T) {
	m := config.ToolManifest{Name: "echo", Type: "builtin", Permissions: config.ToolPermissions{Allow: boolPtr(false)}}
	tl, err := tool.FromManifest(m)
	if err != nil {
		t.Fatalf("from manifest: %v", err)
	}
	if _, err := tl.Execute(context.Background(), nil); !errors.Is(err, tool.ErrToolDenied) {
		t.Fatalf("expected denial, got %v", err)
	}
}

func boolPtr(b bool) *bool { return &b }
