package e2e

import (
	"context"
	"encoding/json"
	"os/exec"
	"reflect"
	"runtime"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/sbox"
)

type cycleClient struct{ count int }

func (c *cycleClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	c.count++
	if c.count == 1 {
		args, _ := json.Marshal(map[string]any{})
		return model.Completion{ToolCalls: []model.ToolCall{{ID: "1", Name: "local", Arguments: args}}}, nil
	}
	return model.Completion{Content: "done"}, nil
}

// TestSandboxedToolE2E ensures shell command tools run through the sandbox engine.
func TestSandboxedToolE2E(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		if runtime.GOOS == "windows" {
			return exec.CommandContext(ctx, "cmd", "/c", "echo", "ok")
		}
		return exec.CommandContext(ctx, "echo", "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()

	manifest := config.ToolManifest{
		Name:    "local",
		Command: "echo hi",
		Net:     "host",
	}
	tl, err := tool.FromManifest(manifest)
	if err != nil {
		t.Fatal(err)
	}

	reg := tool.Registry{"local": tl}
	client := &cycleClient{}
	ag := core.New(client, "mock", reg, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)

	out, err := ag.Run(context.Background(), "start")
	if err != nil {
		t.Fatal(err)
	}
	if out != "done" {
		t.Fatalf("unexpected output: %s", out)
	}

	want := []string{"docker", "run", "--rm", "-v", "/workspace:/workspace", "--network", "host", "alpine", "sh", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("args mismatch: got %v want %v", got, want)
	}
}
