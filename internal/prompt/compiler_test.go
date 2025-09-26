package prompt

import (
	"context"
	"sort"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

// simple fake tool for test
type fakeTool struct{ name, desc string }

func (f fakeTool) Name() string                                            { return f.name }
func (f fakeTool) Description() string                                     { return f.desc }
func (f fakeTool) JSONSchema() map[string]any                              { return map[string]any{"type": "object"} }
func (f fakeTool) Execute(context.Context, map[string]any) (string, error) { return "", nil }

func TestSectionize_OrderAndTags(t *testing.T) {
	reg := tool.Registry{}
	reg["grep"] = fakeTool{name: "grep", desc: "search files"}
	reg["read_lines"] = fakeTool{name: "read_lines", desc: "read file lines"}
	reg["agent"] = fakeTool{name: "agent", desc: "delegate to another agent"}

	base := "You are Agent0. Keep answers concise."
	out := Sectionize(base, reg, map[string]string{"agents": "coder, tester"})

	// Basic structure
	if !strings.Contains(out, "<prompt>\n"+base+"\n</prompt>") {
		t.Fatalf("prompt section missing or incorrect: %q", out)
	}
	if !strings.Contains(out, "<tools>") || !strings.Contains(out, "</tools>") {
		t.Fatalf("tools section missing: %q", out)
	}
	if !strings.Contains(out, "<agents>\ncoder, tester\n</agents>") {
		t.Fatalf("agents extra section missing: %q", out)
	}

	// Tools should be sorted by name for stability
	names := []string{"agent", "grep", "read_lines"}
	sort.Strings(names)
	idx := 0
	for _, n := range names {
		pos := strings.Index(out, "- "+n)
		if pos < idx {
			t.Fatalf("tool %s appears out of order", n)
		}
		idx = pos
	}
}
