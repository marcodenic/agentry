package tool

import (
	"context"
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/contracts"
)

func sharedMemoryBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"shared_memory": sharedMemorySpec(),
	}
}

func sharedMemorySpec() builtinSpec {
	return builtinSpec{
		Desc: "Access shared memory between agents",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"action": map[string]any{
					"type":        "string",
					"enum":        []string{"get", "set", "list"},
					"description": "Action to perform: get, set, or list",
				},
				"key": map[string]any{
					"type":        "string",
					"description": "Key for get/set operations",
				},
				"value": map[string]any{
					"type":        "string",
					"description": "Value to set (for action 'set')",
				},
			},
			"required": []string{"action"},
		},
		Exec: sharedMemoryExec,
	}
}

func sharedMemoryExec(ctx context.Context, args map[string]any) (string, error) {
	action := stringArg(args, "action")
	key := stringArg(args, "key")
	value := stringArg(args, "value")
	if strings.TrimSpace(action) == "" {
		return "", fmt.Errorf("action is required")
	}
	tv := ctx.Value(contracts.TeamContextKey)
	t, _ := tv.(contracts.TeamService)
	if t == nil {
		return "", fmt.Errorf("no team in context")
	}
	switch action {
	case "get":
		if key == "" {
			return "", fmt.Errorf("key required for get operation")
		}
		if v, ok := t.GetSharedData(key); ok {
			return fmt.Sprintf("📊 %s = %v", key, v), nil
		}
		return fmt.Sprintf("📊 %s not set", key), nil
	case "set":
		if key == "" || value == "" {
			return "", fmt.Errorf("key and value required for set operation")
		}
		t.SetSharedData(key, value)
		return fmt.Sprintf("✅ Stored '%s' in shared memory", key), nil
	case "list":
		data := t.GetAllSharedData()
		if len(data) == 0 {
			return "📋 Shared memory is empty", nil
		}
		var b strings.Builder
		b.WriteString("📋 Shared memory keys:\n")
		for k := range data {
			b.WriteString("- ")
			b.WriteString(k)
			b.WriteString("\n")
		}
		return b.String(), nil
	default:
		return "", fmt.Errorf("invalid action. Use: get, set, or list")
	}
}
