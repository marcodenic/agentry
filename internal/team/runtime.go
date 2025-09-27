package team

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/internal/contracts"
	"github.com/marcodenic/agentry/internal/core"
	runtime "github.com/marcodenic/agentry/internal/team/runtime"
)

// runAgent executes an agent with the given input, similar to converse.runAgent
func runAgent(ctx context.Context, ag *core.Agent, input, name string, peers []string) (string, error) {
	timer := StartTimer(fmt.Sprintf("runAgent(%s)", name))
	defer timer.Stop()

	// Attach agent name into context for builtins to use sensible defaults
	ctx = context.WithValue(ctx, contracts.AgentNameContextKey, name)
	timer.Checkpoint("context prepared")

	result, err := ag.Run(ctx, input)
	timer.Checkpoint("agent.Run completed")

	runtime.DebugPrintf("🏁 runAgent: ag.Run completed for agent %s", name)
	runtime.DebugPrintf("🏁 runAgent: Result length: %d", len(result))
	runtime.DebugPrintf("🏁 runAgent: Error: %v", err)
	runtime.DebugPrintf("🏁 runAgent: Agent %s tokens after: %d", name, func() int {
		if ag.Cost != nil {
			return ag.Cost.TotalTokens()
		}
		return 0
	}())
	runtime.DebugPrintf("🏁 runAgent: Agent %s context final state: %v", name, ctx.Err())

	return result, err
}
