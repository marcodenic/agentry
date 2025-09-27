package core

import (
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tokens"
)

type contextBudgetManager struct {
	modelName string
	specs     []model.ToolSpec
}

func newContextBudgetManager(modelName string, specs []model.ToolSpec) *contextBudgetManager {
	return &contextBudgetManager{modelName: modelName, specs: specs}
}

func (b *contextBudgetManager) Trim(msgs []model.ChatMessage) []model.ChatMessage {
	budget := b.resolveBudget()
	totalTokens := b.countMessageTokens(msgs)
	toolTokens := b.countToolSchemaTokens()
	totalWithTools := totalTokens + toolTokens
	if totalWithTools <= budget.target {
		return msgs
	}
	if len(msgs) < 2 {
		debug.Printf("applyBudget: only %d messages available; skipping mid-trim", len(msgs))
		return msgs
	}
	debug.Printf("Context trimming: initial=%d (msgs=%d + tools≈%d) budget=%d reserve=%d model=%s",
		totalWithTools, totalTokens, toolTokens, budget.target, budget.reserve, b.modelName)
	trimmed := b.trimMidSection(msgs, totalTokens, toolTokens, budget.target)
	return trimmed
}

type budgetWindow struct {
	limit   int
	target  int
	reserve int
}

func (b *contextBudgetManager) resolveBudget() budgetWindow {
	limit := env.Int("AGENTRY_CONTEXT_MAX_TOKENS", 0)
	if limit == 0 {
		pt := cost.NewPricingTable()
		limit = pt.GetContextLimit(b.modelName)
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
		limit = headroom
	}
	if strings.Contains(strings.ToLower(b.modelName), "claude") && limit > 60000 {
		limit = 60000
	}
	reserve := env.Int("AGENTRY_CONTEXT_RESERVE_OUTPUT", 1024)
	if reserve < 256 {
		reserve = 256
	}
	target := limit - reserve
	if target < 1000 {
		target = limit - 500
	}
	return budgetWindow{limit: limit, target: target, reserve: reserve}
}

func (b *contextBudgetManager) countMessageTokens(msgs []model.ChatMessage) int {
	total := 0
	for _, m := range msgs {
		total += tokens.Count(m.Content, b.modelName)
		for _, tc := range m.ToolCalls {
			total += tokens.Count(tc.Name, b.modelName)
			total += tokens.Count(string(tc.Arguments), b.modelName)
		}
	}
	return total
}

func (b *contextBudgetManager) countToolSchemaTokens() int {
	if len(b.specs) == 0 {
		return 0
	}
	total := 0
	for _, spec := range b.specs {
		total += tokens.Count(spec.Name, b.modelName)
		total += tokens.Count(spec.Description, b.modelName)
		for key, val := range spec.Parameters {
			total += tokens.Count(key, b.modelName)
			total += tokens.Count(fmt.Sprintf("%v", val), b.modelName)
		}
	}
	return total
}

func (b *contextBudgetManager) trimMidSection(msgs []model.ChatMessage, totalTokens, toolTokens, target int) []model.ChatMessage {
	systemMsg := msgs[0]
	userMsg := msgs[len(msgs)-1]
	middle := msgs[1 : len(msgs)-1]
	removedCount := 0
	for (totalTokens+toolTokens) > target && removedCount < len(middle) {
		removedTokens := tokens.Count(middle[removedCount].Content, b.modelName)
		for _, tc := range middle[removedCount].ToolCalls {
			removedTokens += tokens.Count(tc.Name, b.modelName)
			removedTokens += tokens.Count(string(tc.Arguments), b.modelName)
		}
		totalTokens -= removedTokens
		middle[removedCount].Content = ""
		middle[removedCount].ToolCalls = nil
		removedCount++
	}
	trimmed := make([]model.ChatMessage, 0, len(middle))
	for _, msg := range middle {
		if strings.TrimSpace(msg.Content) == "" && msg.Role != "system" {
			continue
		}
		trimmed = append(trimmed, msg)
	}
	result := append([]model.ChatMessage{systemMsg}, append(trimmed, userMsg)...)
	debug.Printf("Context trimmed: finalTokens≈%d (msgs) + tools≈%d removedMessages=%d", totalTokens, toolTokens, removedCount)
	return result
}
