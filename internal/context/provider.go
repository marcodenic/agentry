package context

import (
	"sort"
	"strconv"
	"strings"

	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
)

// Provider builds chat messages from prompt, history, and user input.
type Provider struct {
	Prompt  string
	History []memory.Step
}

// Provide assembles context messages including optional history compaction.
func (p Provider) Provide(input string) []model.ChatMessage {
	prompt := p.Prompt
	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}

	hist := p.History
	compactAfter := env.Int("AGENTRY_HISTORY_COMPACT_AFTER", 0)
	keepNewest := env.Int("AGENTRY_HISTORY_KEEP", 8)
	if compactAfter > 0 && len(hist) > compactAfter {
		if keepNewest < 1 {
			keepNewest = 1
		}
		newHist, summary := compactHistory(hist, keepNewest)
		if summary != "" {
			msgs = append(msgs, model.ChatMessage{Role: "system", Content: summary})
		}
		hist = newHist
	}

	for _, h := range hist {
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: h.Output, ToolCalls: h.ToolCalls})
		for id, res := range h.ToolResults {
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: res})
		}
	}
	msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	return msgs
}

// compactHistory summarizes older history while keeping recent steps.
func compactHistory(hist []memory.Step, keepNewest int) ([]memory.Step, string) {
	if keepNewest >= len(hist) {
		return hist, ""
	}
	cut := len(hist) - keepNewest
	older := hist[:cut]
	newer := hist[cut:]

	var totalToolCalls int
	toolFreq := make(map[string]int, 8)
	for _, step := range older {
		totalToolCalls += len(step.ToolCalls)
		for _, tc := range step.ToolCalls {
			toolFreq[tc.Name]++
		}
	}

	type kv struct {
		name string
		n    int
	}
	tops := make([]kv, 0, len(toolFreq))
	for k, v := range toolFreq {
		tops = append(tops, kv{name: k, n: v})
	}
	sort.Slice(tops, func(i, j int) bool { return tops[i].n > tops[j].n })
	if len(tops) > 5 {
		tops = tops[:5]
	}

	var b strings.Builder
	b.WriteString("[HISTORY COMPACTED] Earlier ")
	b.WriteString(strconv.Itoa(cut))
	b.WriteString(" steps summarized (" + strconv.Itoa(totalToolCalls) + " tool calls). Top tools: ")
	if len(tops) == 0 {
		b.WriteString("none")
	} else {
		for i, t := range tops {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t.name + "x" + strconv.Itoa(t.n))
		}
	}
	b.WriteString(". Recent context retained.")
	return newer, b.String()
}
