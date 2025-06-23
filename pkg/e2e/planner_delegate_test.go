//go:build integration
// +build integration

package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

type plannerClient struct{ count int }

func (p *plannerClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	p.count++
	if p.count == 1 {
		args, _ := json.Marshal(map[string]any{"input": "subtask"})
		return model.Completion{ToolCalls: []model.ToolCall{{ID: "1", Name: "delegate", Arguments: args}}}, nil
	}
	res := msgs[len(msgs)-1].Content
	return model.Completion{Content: "planner received: " + res}, nil
}

type childClient struct{}

func (childClient) Complete(ctx context.Context, msgs []model.ChatMessage, tools []model.ToolSpec) (model.Completion, error) {
	return model.Completion{Content: "subtask done"}, nil
}

func TestPlannerDelegatesSubtask(t *testing.T) {
	planRoute := router.Rules{{Name: "planner", IfContains: []string{""}, Client: &plannerClient{}}}
	childRoute := router.Rules{{Name: "child", IfContains: []string{""}, Client: childClient{}}}

	planner := core.New(planRoute, nil, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	planner.Tools = tool.Registry{
		"delegate": tool.New("delegate", "Delegate subtask", func(ctx context.Context, args map[string]any) (string, error) {
			input, _ := args["input"].(string)
			child := planner.Spawn()
			child.Route = childRoute
			return child.Run(ctx, input)
		}),
	}

	out, err := planner.Run(context.Background(), "start")
	if err != nil {
		t.Fatal(err)
	}
	if out != "planner received: subtask done" {
		t.Fatalf("unexpected output: %s", out)
	}
}
