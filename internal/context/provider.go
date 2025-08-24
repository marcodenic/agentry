package context

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
)

// Provider builds chat messages from prompt, history, and user input.
type Provider struct {
	Prompt  string
	History []memory.Step
}

// Provide assembles context messages with limited history to prevent exponential growth.
func (p Provider) Provide(input string) []model.ChatMessage {
	prompt := p.Prompt
	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}

	// FIXED: Include only the most recent history step to maintain tool call context
	// while preventing exponential growth. Most agents only need the immediate context.
	hist := p.History
	if len(hist) > 0 {
		// Take only the last step to provide tool call context
		lastStep := hist[len(hist)-1]
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: lastStep.Output, ToolCalls: lastStep.ToolCalls})
		for id, res := range lastStep.ToolResults {
			// Truncate large tool results to prevent context bloat
			truncatedRes := res
			if len(res) > 2048 {
				truncatedRes = res[:2048] + "...\n[TRUNCATED: originally " + fmt.Sprintf("%d", len(res)) + " bytes]"
			}
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: truncatedRes})
		}
	}

	msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	return msgs
}
