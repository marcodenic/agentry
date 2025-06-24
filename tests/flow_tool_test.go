package tests

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestFlowBuiltin(t *testing.T) {
	dir := t.TempDir()
	yaml := `agents:
  tester:
    model: mock
tasks:
  - agent: tester
    input: hi`
	if err := os.WriteFile(filepath.Join(dir, ".agentry.flow.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	tl, ok := tool.DefaultRegistry().Use("flow")
	if !ok {
		t.Fatal("flow tool missing")
	}
	out, err := tl.Execute(context.Background(), map[string]any{"path": dir})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if out != "hello" {
		t.Fatalf("unexpected output %s", out)
	}
}
