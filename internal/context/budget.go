package context

import (
	"fmt"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tokens"
)

// Budget enforces token limits on message history.
type Budget struct {
	ModelName string
}

// Apply trims messages to fit within the model's context window budget.
func (b Budget) Apply(msgs []model.ChatMessage) []model.ChatMessage {
	startMeasure := time.Now()
	maxContextTokens := env.Int("AGENTRY_CONTEXT_MAX_TOKENS", 0)
	if maxContextTokens == 0 {
		pt := cost.NewPricingTable()
		limit := pt.GetContextLimit(b.ModelName)
		if limit <= 0 {
			limit = 120000
		}
		headroom := int(float64(limit) * 0.85)
		if headroom < 4000 {
			headroom = limit - 2000
		}
		if headroom < 2000 {
			headroom = limit - 1000
		}
		maxContextTokens = headroom
	}
	reserveForOutput := env.Int("AGENTRY_CONTEXT_RESERVE_OUTPUT", 1024)
	if reserveForOutput < 256 {
		reserveForOutput = 256
	}
	targetBudget := maxContextTokens - reserveForOutput
	if targetBudget < 1000 {
		targetBudget = maxContextTokens - 500
	}

	totalTokens := 0
	for _, m := range msgs {
		totalTokens += tokens.Count(m.Content, b.ModelName)
		for _, tc := range m.ToolCalls {
			totalTokens += tokens.Count(tc.Name, b.ModelName)
			totalTokens += tokens.Count(string(tc.Arguments), b.ModelName)
		}
	}
	if totalTokens <= targetBudget {
		return msgs
	}

	debug.Printf("Context trimming: initial=%d budget=%d reserve=%d model=%s", totalTokens, targetBudget, reserveForOutput, b.ModelName)
	systemMsg := msgs[0]
	userMsg := msgs[len(msgs)-1]
	mid := msgs[1 : len(msgs)-1]
	idx := 0
	for totalTokens > targetBudget && idx < len(mid) {
		removed := tokens.Count(mid[idx].Content, b.ModelName)
		for _, tc := range mid[idx].ToolCalls {
			removed += tokens.Count(tc.Name, b.ModelName)
			removed += tokens.Count(string(tc.Arguments), b.ModelName)
		}
		totalTokens -= removed
		mid[idx].Content = ""
		mid[idx].ToolCalls = nil
		idx++
	}
	newMid := make([]model.ChatMessage, 0, len(mid))
	for _, m := range mid {
		if strings.TrimSpace(m.Content) == "" && m.Role != "system" {
			continue
		}
		newMid = append(newMid, m)
	}
	msgs = append([]model.ChatMessage{systemMsg}, append(newMid, userMsg)...)
	debug.Printf("Context trimmed: finalTokens≈%d removedMessages=%d", totalTokens, idx)

	if debug.IsContextDebugEnabled() {
		var sb strings.Builder
		sb.WriteString("[CONTEXT BREAKDOWN]\n")
		for i, m := range msgs {
			role := m.Role
			if role == "system" && i == 0 {
				role = "system(root)"
			}
			sb.WriteString(fmt.Sprintf("%02d %-8s tokens=%d len=%d\n", i, role, tokens.Count(m.Content, b.ModelName), len(m.Content)))
		}
		sb.WriteString(fmt.Sprintf("Total≈%d (budget=%d reserve=%d) buildTime=%s\n", func() int {
			t := 0
			for _, m := range msgs {
				t += tokens.Count(m.Content, b.ModelName)
			}
			return t
		}(), targetBudget, reserveForOutput, time.Since(startMeasure)))
		debug.Printf(sb.String())
	}
	return msgs
}
