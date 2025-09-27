package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/env"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tokens"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

type conversationSession struct {
	agent *Agent
	ctx   context.Context
	input string

	specs []model.ToolSpec
	msgs  []model.ChatMessage

	consecutiveErrors int
	recentToolCalls   []toolCallSignature
	tracer            conversationTracer
}

type toolCallSignature struct {
	Name string
	Args string
}

const (
	maxRecentCalls    = 6
	maxIdenticalCalls = 3
)

func newConversationSession(agent *Agent, ctx context.Context, input string) *conversationSession {
	return &conversationSession{agent: agent, ctx: ctx, input: input, tracer: newConversationTracer()}
}

func (s *conversationSession) Run() (string, error) {
	if err := s.prepare(); err != nil {
		return "", err
	}
	return s.loop()
}

func (s *conversationSession) prepare() error {
	agent := s.agent
	debug.Printf("Agent.Run: Agent ID=%s, Prompt length=%d chars", agent.ID.String()[:8], len(agent.Prompt))
	debug.Printf("Agent.Run: Available tools: %v", agent.toolNames())
	safeInputPreview := sanitizeForLog(s.input[:min(300, len(s.input))])
	debug.Printf("Agent.Run: Input: %s", safeInputPreview)

	agent.Trace(s.ctx, trace.EventModelStart, agent.ModelName)

	prompt := agent.Prompt
	if prompt == "" {
		prompt = defaultPrompt()
	}
	prompt = applyVars(prompt, agent.Vars)

	s.specs = tool.BuildSpecs(agent.Tools)
	history := agent.Mem.History()
	if debug.IsTraceEnabled() {
		debug.Printf("Agent.Run: memory store=%p (history len %d)", agent.Mem, len(history))
	}
	s.msgs = agent.buildMessages(prompt, s.input, history)
	s.msgs = agent.applyBudget(s.msgs, s.specs)

	s.tracer.SessionPrepared(prompt, agent.ModelName, s.msgs, s.specs)
	if debug.IsTraceEnabled() {
		if dbg, ok := s.tracer.(interface {
			SessionPreparedVerbose(string, []model.ChatMessage, []model.ToolSpec)
		}); ok {
			dbg.SessionPreparedVerbose(agent.ModelName, s.msgs, s.specs)
		}
	}
	return nil
}

func (s *conversationSession) loop() (string, error) {
	agent := s.agent
	for i := 0; ; i++ {
		s.tracer.IterationStart(i)
		if agent.MaxIter > 0 && i >= agent.MaxIter {
			return "", fmt.Errorf("iteration cap reached (%d)", agent.MaxIter)
		}

		select {
		case <-s.ctx.Done():
			return "", s.ctx.Err()
		default:
		}

		s.msgs = agent.applyBudget(s.msgs, s.specs)
		s.tracer.IterationMessages(i, s.msgs)

		completion, responseID, err := s.streamOnce()
		if err != nil {
			return "", err
		}

		outcome, done, err := s.handleCompletion(completion, responseID)
		if err != nil {
			return "", err
		}
		if done {
			return outcome, nil
		}
	}
}

func (s *conversationSession) streamOnce() (model.Completion, string, error) {
	return newStreamExecutor(s).Execute()
}

func (s *conversationSession) handleCompletion(res model.Completion, responseID string) (string, bool, error) {
	agent := s.agent

	inTok := res.InputTokens
	if inTok == 0 {
		for _, m := range s.msgs {
			inTok += tokens.Count(m.Content, agent.ModelName)
			for _, tc := range m.ToolCalls {
				inTok += tokens.Count(tc.Name, agent.ModelName)
				inTok += tokens.Count(string(tc.Arguments), agent.ModelName)
			}
		}
	}
	outTok := res.OutputTokens
	if outTok == 0 {
		outTok = tokens.Count(res.Content, agent.ModelName)
	}
	if agent.Cost != nil {
		modelForCost := res.ModelName
		if strings.TrimSpace(modelForCost) == "" {
			modelForCost = agent.ModelName
		}
		agent.Cost.AddModelUsage(modelForCost, inTok, outTok)
		if agent.Cost.OverBudget() && env.Bool("AGENTRY_STOP_ON_BUDGET", false) {
			return "", false, fmt.Errorf("cost or token budget exceeded (tokens=%d cost=$%.4f)", agent.Cost.TotalTokens(), agent.Cost.TotalCost())
		}
	}

	if responseID == "" {
		debug.Printf("Agent.Run: Appending assistant message to local context (no conversation linking)")
		s.msgs = append(s.msgs, model.ChatMessage{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
	} else {
		debug.Printf("Agent.Run: Conversation linking active (responseID: %s), not appending to local context", responseID)
	}

	step := memory.Step{Input: s.input, Output: res.Content, ToolCalls: res.ToolCalls, ToolResults: map[string]string{}}
	stepRecorded := false
	recordStep := func() {
		if !stepRecorded {
			agent.Mem.AddStep(step)
			if debug.IsTraceEnabled() {
				h := agent.Mem.History()
				debug.Printf("Agent.Run: Recorded step (history len now %d)", len(h))
			}
			stepRecorded = true
		}
	}

	if len(res.ToolCalls) == 0 {
		recordStep()
		if env.Bool("AGENTRY_PLAN_HEURISTIC", false) && len(s.specs) > 0 && s.shouldInjectPlanFollowUp(res.Content) {
			s.msgs = append(s.msgs, model.ChatMessage{Role: "system", Content: s.planFollowUpMessage()})
			return "", false, nil
		}

		if err := agent.JSONValidator.ValidateAgentOutput(res.Content); err != nil {
			debug.Printf("Agent.Run: Agent output validation failed: %v", err)
			return fmt.Sprintf("Agent completed task but output validation failed: %v", err), true, nil
		}

		agent.Trace(s.ctx, trace.EventFinal, res.Content)
		s.tracer.Finalised(res.Content)
		return res.Content, true, nil
	}

	s.tracer.DuplicateToolCheck(res.ToolCalls)
	if msg, stop := s.trackRecentToolCalls(res.ToolCalls, recordStep); stop {
		s.tracer.Finalised(msg)
		return msg, true, nil
	}

	toolMsgs, hadErrors, execErr := agent.executeToolCalls(s.ctx, res.ToolCalls, step)
	if execErr != nil {
		return "", false, execErr
	}

	if !hadErrors && agent.allToolCallsTerminal(res.ToolCalls) && len(toolMsgs) > 0 {
		if out := aggregateToolOutputs(toolMsgs); out != "" {
			recordStep()
			agent.Trace(s.ctx, trace.EventFinal, out)
			s.tracer.Finalised(out)
			return out, true, nil
		}
	}

	if responseID == "" {
		debug.Printf("Agent.Run: Appending %d tool results to local context (no conversation linking)", len(toolMsgs))
	} else {
		debug.Printf("Agent.Run: Conversation linking active (responseID: %s), appending %d tool results for function calls", responseID, len(toolMsgs))
	}
	s.msgs = append(s.msgs, toolMsgs...)
	recordStep()
	s.tracer.ToolResultsAppended(len(s.msgs))

	if hadErrors {
		s.consecutiveErrors++
	} else {
		s.consecutiveErrors = 0
	}
	if s.consecutiveErrors > agent.ErrorHandling.MaxErrorRetries {
		return "", false, fmt.Errorf("too many consecutive errors (%d), stopping execution", s.consecutiveErrors)
	}

	return "", false, nil
}

func (s *conversationSession) shouldInjectPlanFollowUp(content string) bool {
	lc := strings.ToLower(content)
	return strings.Contains(lc, "plan") || strings.Contains(content, "I'll ") || strings.Contains(lc, "i will")
}

func (s *conversationSession) planFollowUpMessage() string {
	return "You provided a plan. Now execute the necessary steps using the available tools. For each data collection action, call the appropriate tool (e.g., sysinfo). Then produce the consolidated final report. Respond only with tool calls until data is gathered."
}

func (s *conversationSession) trackRecentToolCalls(calls []model.ToolCall, recordStep func()) (string, bool) {
	if len(calls) == 0 {
		return "", false
	}

	debug.Printf("Agent.Run: Processing %d tool calls for duplicate detection", len(calls))
	for _, tc := range calls {
		signature := toolCallSignature{Name: tc.Name, Args: string(tc.Arguments)}
		debug.Printf("Agent.Run: Tool call signature: %s(%s)", signature.Name, signature.Args)

		s.recentToolCalls = append(s.recentToolCalls, signature)
		if len(s.recentToolCalls) > maxRecentCalls {
			s.recentToolCalls = s.recentToolCalls[1:]
		}

		identicalCount := 0
		for j, recent := range s.recentToolCalls {
			if recent.Name == signature.Name && recent.Args == signature.Args {
				identicalCount++
				debug.Printf("Agent.Run: Found identical call at position %d: %s(%s)", j, recent.Name, recent.Args)
			}
		}
		debug.Printf("Agent.Run: Tool %s has %d identical calls in recent history (max=%d)", signature.Name, identicalCount, maxIdenticalCalls)

		if identicalCount >= maxIdenticalCalls {
			recordStep()
			debug.Printf("Agent.Run: BREAKING LOOP - Detected repeated tool call (%s) %d times", tc.Name, identicalCount)
			return fmt.Sprintf("Task completed. Detected repeated tool execution (%s), stopping to prevent infinite loop.", tc.Name), true
		}
	}

	return "", false
}

func aggregateToolOutputs(toolMsgs []model.ChatMessage) string {
	var b strings.Builder
	for _, m := range toolMsgs {
		if m.Role == "tool" && strings.TrimSpace(m.Content) != "" {
			if b.Len() > 0 {
				b.WriteString("\n")
			}
			b.WriteString(m.Content)
		}
	}
	return strings.TrimSpace(b.String())
}
