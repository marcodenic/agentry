package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/internal/tool"
)

// simpleMock returns a simple text completion.
type simpleMock struct{}

func (simpleMock) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{Content: "test response"}, nil
}

func TestBuiltinsCrossPlatform(t *testing.T) {
	// Set up team context for agent tool
	route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: simpleMock{}}}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	tm, err := converse.NewTeam(ag, 1, "test")
	if err != nil {
		t.Fatalf("failed to create team: %v", err)
	}
	ctx := team.WithContext(context.Background(), tm)
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
