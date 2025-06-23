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

func TestFlowLoadPresetAndInclude(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "templates"), 0755); err != nil {
		t.Fatal(err)
	}
	preset := `agents:
  coder:
    model: gpt-4
tasks:
  - agent: coder
    input: build`
	if err := os.WriteFile(filepath.Join(dir, "templates", "team.yaml"), []byte(preset), 0644); err != nil {
		t.Fatal(err)
	}
	role := `agents:
  reviewer:
    model: gpt-4
tasks:
  - agent: reviewer
    input: check`
	if err := os.WriteFile(filepath.Join(dir, "role.yaml"), []byte(role), 0644); err != nil {
		t.Fatal(err)
	}
	flowYaml := `presets: [team.yaml]
include:
  - role.yaml
agents:
  leader:
    model: gpt-4
tasks:
  - agent: leader
    input: lead`
	if err := os.WriteFile(filepath.Join(dir, ".agentry.flow.yaml"), []byte(flowYaml), 0644); err != nil {
		t.Fatal(err)
	}
	f, err := flow.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Agents) != 3 || len(f.Tasks) != 3 {
		t.Fatalf("unexpected merged data: %#v", f)
	}
}
