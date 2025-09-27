package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	promptpkg "github.com/marcodenic/agentry/internal/prompt"
)

// promptEnvelope builds the system/user message chain for an agent invocation.
// It isolates the previous inline logic from Agent.buildMessages so message
// construction can be read and tested without scanning the entire agent type.
type promptEnvelope struct {
	agent *Agent
}

func newPromptEnvelope(agent *Agent) *promptEnvelope {
	return &promptEnvelope{agent: agent}
}

// Build assembles the chat message sequence given the base prompt, user input,
// and recent history.
func (p *promptEnvelope) Build(prompt, input string, history []memory.Step) []model.ChatMessage {
	debug.Printf("=== buildMessages START ===")
	debug.Printf("History length: %d steps", len(history))
	for i, step := range history {
		dumpHistoryStep(i, step)
	}
	debug.Printf("=== buildMessages processing ===")

	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		prompt = defaultPrompt()
	}
	prompt = p.sectionizePrompt(prompt)

	msgs := []model.ChatMessage{{Role: "system", Content: prompt}}
	msgs = append(msgs, p.recentHistoryMessages(history)...)
	msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	return msgs
}

func (p *promptEnvelope) sectionizePrompt(prompt string) string {
	extra := p.additionalSections()
	return promptpkg.Sectionize(prompt, p.agent.Tools, extra)
}

func (p *promptEnvelope) additionalSections() map[string]string {
	extras := map[string]string{}
	if p.agent.Vars != nil {
		if s, ok := p.agent.Vars["AGENTS_SECTION"]; ok {
			extras["agents"] = s
		}
	}

	names := make([]string, 0, len(p.agent.Tools))
	for name := range p.agent.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	allowedCommands := []string{"list", "view", "write", "run", "search", "find", "cwd", "env"}
	if guidance := GetPlatformContext(allowedCommands, names); strings.TrimSpace(guidance) != "" {
		extras["tool_guidance"] = guidance
	}
	return extras
}

func (p *promptEnvelope) recentHistoryMessages(history []memory.Step) []model.ChatMessage {
	if len(history) == 0 {
		debug.Printf("No history to include in messages")
		return nil
	}
	lastStep := history[len(history)-1]
	dumpHistorySummary(lastStep)

	var msgs []model.ChatMessage
	if content := strings.TrimSpace(lastStep.Input); content != "" {
		msgs = append(msgs, model.ChatMessage{Role: "user", Content: content})
	}
	if out := strings.TrimSpace(lastStep.Output); out != "" || len(lastStep.ToolCalls) > 0 {
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: lastStep.Output, ToolCalls: lastStep.ToolCalls})
	}
	for id, res := range lastStep.ToolResults {
		msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: truncateToolResult(res)})
	}
	return msgs
}

func dumpHistoryStep(idx int, step memory.Step) {
	debug.Printf("  HISTORY[%d]: Output=%.100s..., ToolCalls=%d, ToolResults=%d",
		idx, step.Output, len(step.ToolCalls), len(step.ToolResults))
}

func dumpHistorySummary(step memory.Step) {
	debug.Printf("Including LAST history step in messages:")
	debug.Printf("  Input: %.200s...", step.Input)
	debug.Printf("  Output: %.200s...", step.Output)
	debug.Printf("  ToolCalls: %d", len(step.ToolCalls))
	debug.Printf("  ToolResults: %d", len(step.ToolResults))
}

func truncateToolResult(res string) string {
	if len(res) <= 2048 {
		return res
	}
	truncated := res[:2048] + "...\n[TRUNCATED: originally " + fmt.Sprintf("%d", len(res)) + " bytes]"
	debug.Printf("  TRUNCATED tool result: %d -> %d chars", len(res), len(truncated))
	return truncated
}
