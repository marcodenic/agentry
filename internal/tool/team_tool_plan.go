package tool

import (
	"context"
	"fmt"
	"strings"
)

func teamPlanSpec() builtinSpec {
	return builtinSpec{
		Desc: "Coordinate a simple ad-hoc team for a topic (demo)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"n":     map[string]any{"type": "integer", "description": "Number of helpers"},
				"topic": map[string]any{"type": "string", "description": "Topic to coordinate on"},
			},
			"required": []string{},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			n, _ := getIntArg(args, "n", 1)
			topic, _ := args["topic"].(string)
			if n < 1 {
				n = 1
			}
			if strings.TrimSpace(topic) == "" {
				topic = "general assistance"
			}
			var b strings.Builder
			b.WriteString("Team plan:\n")
			for i := 1; i <= n; i++ {
				b.WriteString(fmt.Sprintf("- agent_%d handles %s\n", i, topic))
			}
			return b.String(), nil
		},
	}
}
