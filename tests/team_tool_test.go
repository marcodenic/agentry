package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestTeamBuiltin(t *testing.T) {
	tl, ok := tool.DefaultRegistry().Use("team")
	if !ok {
		t.Fatal("team tool missing")
	}
	out, err := tl.Execute(context.Background(), map[string]any{"n": 2, "topic": "hi"})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if lines := strings.Split(strings.TrimSpace(out), "\n"); len(lines) == 0 {
		t.Fatal("no output")
	}
}
