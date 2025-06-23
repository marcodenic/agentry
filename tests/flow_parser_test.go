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

func TestFlowParseInclude(t *testing.T) {
	dir := t.TempDir()
	role := `name: coder
prompt: |
  testing role
tools:
  - bash
`
	if err := os.WriteFile(filepath.Join(dir, "role.yaml"), []byte(role), 0644); err != nil {
		t.Fatal(err)
	}
	flowYaml := `include:
  - role.yaml
agents:
  coder:
    model: mock
tasks:
  - agent: coder
    input: hi
`
	if err := os.WriteFile(filepath.Join(dir, ".agentry.flow.yaml"), []byte(flowYaml), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := flow.Load(dir)
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	ag, ok := f.Agents["coder"]
	if !ok {
		t.Fatalf("agent not loaded")
	}
	if ag.Prompt == "" || len(ag.Tools) != 1 || ag.Tools[0] != "bash" {
		t.Fatalf("include failed: %#v", ag)
	}
}
