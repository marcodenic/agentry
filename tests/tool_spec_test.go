package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestBuildSpecs(t *testing.T) {
	reg := tool.Registry{
		"foo": tool.New("foo", "Foo tool", func(ctx context.Context, args map[string]any) (string, error) { return "", nil }),
		"bar": tool.NewWithSchema("bar", "Bar tool", map[string]any{"type": "object"}, func(ctx context.Context, args map[string]any) (string, error) { return "ok", nil }),
	}

	specs := tool.BuildSpecs(reg)
	if len(specs) != 2 {
		t.Fatalf("expected 2 specs got %d", len(specs))
	}

	found := map[string]bool{}
	for _, s := range specs {
		if s.Name == "foo" && s.Description == "Foo tool" {
			found["foo"] = true
		}
		if s.Name == "bar" && s.Description == "Bar tool" {
			found["bar"] = true
		}
	}

	if !found["foo"] || !found["bar"] {
		t.Fatalf("specs missing: %v", found)
	}
}
