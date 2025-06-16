package converse

import (
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
)

var colours = []string{
	"\033[38;5;81m",
	"\033[38;5;118m",
	"\033[38;5;214m",
	"\033[38;5;135m",
	"\033[38;5;203m",
}

const colourReset = "\033[0m"

func colourFor(i int) string { return colours[i%len(colours)] }

func cleanInput(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 && r != '\n' && r != '\t' && r != '\r' {
			return -1
		}
		return r
	}, s)
}

const maxHistoryMsgs = 12

func BuildMessages(hist []memory.Step, input, speaker string, peers []string) []model.ChatMessage {
	input = cleanInput(input)
	if len(hist) > maxHistoryMsgs {
		hist = hist[len(hist)-maxHistoryMsgs:]
	}
	sys := fmt.Sprintf(`You are %s chatting with fellow AIs (%s).
• Keep replies ≤50 words (2–3 quirky sentences).
• Feel free to riff or joke; formal greetings are optional.
• Feel comfortable to refer to, make fun of, agree with, disagree with or otherwise respond to other AIs responses.
• Do not repeat or summarise prior messages; add one fresh angle.
• Mention another agent by name only if it feels natural.
• Plain text only unless calling a tool (JSON arguments required).`,
		speaker, strings.Join(peers, ", "))
	msgs := []model.ChatMessage{{Role: "system", Content: sys}}
	for _, h := range hist {
		msgs = append(msgs, model.ChatMessage{Role: "assistant", Content: h.Output, ToolCalls: h.ToolCalls})
		for id, res := range h.ToolResults {
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: id, Content: res})
		}
	}
	if strings.TrimSpace(input) != "" {
		msgs = append(msgs, model.ChatMessage{Role: "user", Content: input})
	}
	return msgs
}
