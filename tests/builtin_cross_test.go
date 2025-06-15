package tests

import (
	"context"
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
			ex["path"] = filepath.Join("..", p)
		}
		t.Run(name, func(t *testing.T) {
			if name == "patch" || name == "ping" || name == "fetch" || name == "sourcegraph" {
				t.Skip("skip network-dependent tool")
			}
			_, err := tl.Execute(ctx, ex)
			if err != nil {
				t.Fatalf("%s failed: %v", name, err)
			}
		})
	}
}
