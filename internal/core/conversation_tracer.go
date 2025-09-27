package core

import (
	"time"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/model"
)

// conversationTracer centralises verbose logging so the core loop stays readable.
type conversationTracer interface {
	SessionPrepared(prompt string, modelName string, msgs []model.ChatMessage, specs []model.ToolSpec)
	IterationStart(iteration int)
	IterationMessages(iteration int, msgs []model.ChatMessage)
	StreamInvocation(msgs []model.ChatMessage)
	StreamChunk(count int)
	StreamCompleted(count int, duration time.Duration)
	CompletionReady(res model.Completion)
	DuplicateToolCheck(calls []model.ToolCall)
	ToolResultsAppended(msgCount int)
	Finalised(output string)
}

type debugConversationTracer struct{}

type noopConversationTracer struct{}

func newConversationTracer() conversationTracer {
	if debug.IsTraceEnabled() {
		return debugConversationTracer{}
	}
	return noopConversationTracer{}
}

func (debugConversationTracer) SessionPrepared(prompt string, modelName string, msgs []model.ChatMessage, specs []model.ToolSpec) {
	debug.Printf("Agent.Run: Agent prompt length=%d chars", len(prompt))
	debug.Printf("Agent.Run: Built %d messages (post-trim), %d tool specs", len(msgs), len(specs))
	debug.Printf("Agent.Run: About to call model client with model %s", modelName)
	debug.Printf("Agent.Run: ENTERING MAIN LOOP NOW - BEFORE FOR LOOP")
	debug.Printf("=== AVAILABLE TOOLS ===")
	for _, spec := range specs {
		debug.Printf("Tool: %s", spec.Name)
	}
	debug.Printf("=== END TOOLS ===")
}

func (debugConversationTracer) IterationStart(iteration int) {
	debug.Printf("Agent.Run: *** ITERATION %d START ***", iteration)
}

func (debugConversationTracer) IterationMessages(iteration int, msgs []model.ChatMessage) {
	debug.Printf("Agent.Run: Current message count: %d", len(msgs))
	if iteration == 0 {
		return
	}
	start := len(msgs) - 3
	if start < 0 {
		start = 0
	}
	debug.Printf("Agent.Run: Recent messages in iteration %d:", iteration)
	for i := start; i < len(msgs); i++ {
		debug.Printf("  [%d] Role: %s, Content: %.100s...", i, msgs[i].Role, msgs[i].Content)
	}
}

func (debugConversationTracer) StreamInvocation(msgs []model.ChatMessage) {
	debug.Printf("Agent.Run: About to call model with %d messages", len(msgs))
	debug.Printf("Agent.Run: WHAT TRIGGERS NEW CALL? Messages context:")
	for j, msg := range msgs {
		debug.Printf("  MSG[%d] Role:%s ToolCalls:%d Content:%.150s...", j, msg.Role, len(msg.ToolCalls), msg.Content)
	}
	debug.Printf("Agent.Run: CALLING MODEL CLIENT NOW - BEFORE STREAM")
}

func (debugConversationTracer) StreamChunk(count int) {
	debug.Printf("Agent.Run: Received chunk %d", count)
}

func (debugConversationTracer) StreamCompleted(count int, duration time.Duration) {
	debug.Printf("Agent.Run: Stream reading completed with %d chunks, read_duration=%v", count, duration)
}

func (debugConversationTracer) CompletionReady(res model.Completion) {
	debug.Printf("Agent.Run: Streaming completed with %d tool calls", len(res.ToolCalls))
	debug.Printf("Agent.Run: Agent response content: '%.200s...'", res.Content)
	if len(res.ToolCalls) > 0 {
		debug.Printf("Agent.Run: AGENT IS MAKING TOOL CALLS - WHY?")
		for i, tc := range res.ToolCalls {
			debug.Printf("  TOOL_CALL[%d]: %s with args %s", i, tc.Name, string(tc.Arguments))
		}
	}
}

func (debugConversationTracer) DuplicateToolCheck(calls []model.ToolCall) {
	debug.Printf("Agent.Run: Processing %d tool calls for duplicate detection", len(calls))
	for _, tc := range calls {
		debug.Printf("Agent.Run: Tool call signature: %s(%s)", tc.Name, string(tc.Arguments))
	}
}

func (debugConversationTracer) ToolResultsAppended(msgCount int) {
	debug.Printf("Agent.Run: Messages after tool execution (count=%d):", msgCount)
}

func (debugConversationTracer) Finalised(output string) {
	debug.Printf("Agent.Run: Finalising with output snippet: %.200s", output)
}

func (noopConversationTracer) SessionPrepared(string, string, []model.ChatMessage, []model.ToolSpec) {
}
func (noopConversationTracer) IterationStart(int)                         {}
func (noopConversationTracer) IterationMessages(int, []model.ChatMessage) {}
func (noopConversationTracer) StreamInvocation([]model.ChatMessage)       {}
func (noopConversationTracer) StreamChunk(int)                            {}
func (noopConversationTracer) StreamCompleted(int, time.Duration)         {}
func (noopConversationTracer) CompletionReady(model.Completion)           {}
func (noopConversationTracer) DuplicateToolCheck([]model.ToolCall)        {}
func (noopConversationTracer) ToolResultsAppended(int)                    {}
func (noopConversationTracer) Finalised(string)                           {}

func (debugConversationTracer) logMessagesDebug(msgs []model.ChatMessage) {
	debug.Printf("=== FULL MESSAGE PAYLOAD TO API ===")
	totalChars := 0
	for i, msg := range msgs {
		msgSize := len(msg.Content)
		totalChars += msgSize
		debug.Printf("[MSG %d] Role: %s, Size: %d chars, ToolCalls: %d", i, msg.Role, msgSize, len(msg.ToolCalls))
		if msg.Role == "system" {
			debug.Printf("  SYSTEM CONTENT (first 500 chars): %.500s...", msg.Content)
		} else if msg.Role == "user" {
			debug.Printf("  USER CONTENT: %s", msg.Content)
		} else if msg.Role == "assistant" {
			debug.Printf("  ASSISTANT CONTENT: %.200s...", msg.Content)
			for j, tc := range msg.ToolCalls {
				debug.Printf("    TOOL_CALL[%d]: %s(%s)", j, tc.Name, string(tc.Arguments))
			}
		} else if msg.Role == "tool" {
			debug.Printf("  TOOL RESULT (ID: %s): %.200s...", msg.ToolCallID, msg.Content)
		}
	}
	debug.Printf("=== TOTAL PAYLOAD: %d messages, %d total chars ===", len(msgs), totalChars)
}

func (d debugConversationTracer) SessionPreparedVerbose(modelName string, msgs []model.ChatMessage, specs []model.ToolSpec) {
	d.logMessagesDebug(msgs)
	d.SessionPrepared("", modelName, msgs, specs)
}
