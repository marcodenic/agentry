package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/marcodenic/agentry/pkg/flow"
)

func TestFlowParseSuccess(t *testing.T) {
	dir := t.TempDir()
	yaml := `agents:
  coder:
    model: gpt-4
    vars:
      tone: excited
tasks:
  - agent: coder
    input: build
`
	if err := os.WriteFile(filepath.Join(dir, ".agentry.flow.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := flow.Load(dir)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(f.Agents) != 1 || len(f.Tasks) != 1 {
		t.Fatalf("unexpected parsed data: %#v", f)
	}
	if f.Agents["coder"].Vars["tone"] != "excited" {
		t.Fatalf("vars not parsed: %#v", f.Agents["coder"].Vars)
	}
}

func TestFlowParseUndefinedAgent(t *testing.T) {
	dir := t.TempDir()
	yaml := `agents:
  coder:
    model: gpt-4
tasks:
  - agent: missing
    input: build
`
	os.WriteFile(filepath.Join(dir, ".agentry.flow.yaml"), []byte(yaml), 0644)
	_, err := flow.Load(dir)
	if err == nil {
		t.Fatalf("expected error for undefined agent")
	}
}
