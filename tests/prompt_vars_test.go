package tests

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
)

type promptCheckClient struct{ t *testing.T }

func (p promptCheckClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	if !strings.Contains(msgs[0].Content, "cheerful") {
		p.t.Fatalf("prompt not substituted: %s", msgs[0].Content)
	}
	return model.Completion{Content: "done"}, nil
}

func TestPromptVarSubstitution(t *testing.T) {
	ag := core.New(promptCheckClient{t}, "mock", tool.DefaultRegistry(), memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	ag.Prompt = "You are a {{tone}} bot"
	ag.Vars = map[string]string{"tone": "cheerful"}
	if _, err := ag.Run(context.Background(), "hi"); err != nil {
		t.Fatal(err)
	}
}

type argSubClient struct{ call int }

func (a *argSubClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	a.call++
	if a.call == 1 {
		args, _ := json.Marshal(map[string]string{"text": "{{greet}} world"})
		return model.Completion{ToolCalls: []model.ToolCall{{ID: "1", Name: "echo", Arguments: args}}}, nil
	}
	return model.Completion{Content: "ok"}, nil
}

func TestToolArgVarSubstitution(t *testing.T) {
	ag := core.New(&argSubClient{}, "mock", tool.DefaultRegistry(), memory.NewInMemory(), memory.NewInMemoryVector(), nil)
	ag.Vars = map[string]string{"greet": "hello"}
	out, err := ag.Run(context.Background(), "start")
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("unexpected final output %s", out)
	}
	hist := ag.Mem.History()
	if len(hist) == 0 {
		t.Fatal("no history recorded")
	}
	if res := hist[0].ToolResults["1"]; res != "hello world" {
		t.Fatalf("args not substituted, got %s", res)
	}
}
