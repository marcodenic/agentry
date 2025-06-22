package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestBuiltinsCrossPlatform(t *testing.T) {
	ctx := context.Background()
	for name, tl := range tool.DefaultRegistry() {
		ex, _ := tl.JSONSchema()["example"].(map[string]any)
		if ex == nil {
			ex = map[string]any{}
		}
		if p, ok := ex["path"].(string); ok {
			tmp := filepath.Join(t.TempDir(), filepath.Base(p))
			ex["path"] = tmp
			if name == "view" || name == "grep" || name == "edit" {
				content := []byte("hello")
				if name == "grep" {
					content = []byte("hello world")
				}
				if err := os.WriteFile(tmp, content, 0644); err != nil {
					t.Fatalf("setup file: %v", err)
				}
			}
		}
		t.Run(name, func(t *testing.T) {
			if name == "patch" || name == "ping" || name == "fetch" || name == "sourcegraph" || name == "mcp" {
				t.Skip("skip network-dependent tool")
			}
			if name == "edit" {
				path := ex["path"].(string)
				viewT, _ := tool.DefaultRegistry().Use("view")
				if _, err := viewT.Execute(ctx, map[string]any{"path": path}); err != nil {
					t.Fatalf("setup view: %v", err)
				}
			}
			_, err := tl.Execute(ctx, ex)
			if err != nil {
				t.Fatalf("%s failed: %v", name, err)
			}
		})
	}
}
