package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/marcodenic/agentry/internal/debug"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

// toolExecutor isolates tool-call handling from Agent.Run for easier testing.
type toolExecutor struct {
	agent *Agent
	note  toolUserNotifier
}

func newToolExecutor(agent *Agent) *toolExecutor {
	return &toolExecutor{agent: agent, note: newToolUserNotifier()}
}

func (e *toolExecutor) execute(ctx context.Context, calls []model.ToolCall, step memory.Step) ([]model.ChatMessage, bool, error) {
	msgs := make([]model.ChatMessage, 0, len(calls))
	hadErrors := false

	for _, tc := range calls {
		select {
		case <-ctx.Done():
			return msgs, hadErrors, ctx.Err()
		default:
		}

		toolInstance, args, err := e.prepareCall(tc)
		if err != nil {
			content := err.Error()
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: content})
			step.ToolResults[tc.ID] = content
			hadErrors = true
			continue
		}

		res, softFailure, err := e.runTool(ctx, tc, toolInstance, args)
		if err != nil {
			hadErrors = true
			return msgs, hadErrors, err
		}
		if softFailure {
			hadErrors = true
			msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: res})
			step.ToolResults[tc.ID] = res
			continue
		}

		step.ToolResults[tc.ID] = res
		msgs = append(msgs, model.ChatMessage{Role: "tool", ToolCallID: tc.ID, Content: res})
	}

	return msgs, hadErrors, nil
}

func (e *toolExecutor) prepareCall(tc model.ToolCall) (tool.Tool, map[string]any, error) {
	toolInstance, ok := e.agent.Tools.Use(tc.Name)
	if !ok {
		errMsg := fmt.Sprintf("Error: Unknown tool '%s'. Available tools: %v", tc.Name, getToolNames(e.agent.Tools))
		return nil, nil, errors.New(errMsg)
	}

	var args map[string]any
	if err := json.Unmarshal(tc.Arguments, &args); err != nil {
		return nil, nil, fmt.Errorf("Error: Invalid tool arguments for '%s': %w", tc.Name, err)
	}

	applyVarsMap(args, e.agent.Vars)
	if err := e.agent.JSONValidator.ValidateToolArgs(args); err != nil {
		return nil, nil, fmt.Errorf("Error: Invalid tool arguments for '%s': %w", tc.Name, err)
	}

	return toolInstance, args, nil
}

func (e *toolExecutor) runTool(ctx context.Context, tc model.ToolCall, toolInstance tool.Tool, args map[string]any) (string, bool, error) {
	if b, _ := json.Marshal(args); len(b) > 0 {
		debug.Printf("Agent '%s' executing tool '%s' with args: %s", e.agent.ID, tc.Name, sanitizeForLog(string(b)))
	} else {
		debug.Printf("Agent '%s' executing tool '%s'", e.agent.ID, tc.Name)
	}

	e.logToolStart(tc, args)

	e.agent.Trace(ctx, trace.EventToolStart, map[string]any{"name": tc.Name, "args": args})

	argSummary := getToolArgSummary(tc.Name, args)
	e.note.Start(e.agent.ID.String(), tc.Name, argSummary)

	result, err := toolInstance.Execute(ctx, args)
	e.logToolEnd(tc, args, result, err)
	if err != nil {
		e.agent.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "error": err.Error()})
	} else {
		e.agent.Trace(ctx, trace.EventToolEnd, map[string]any{"name": tc.Name, "result": result})
	}

	if err != nil {
		e.note.Failure(e.agent.ID.String(), tc.Name, err)

		errorMsg := e.formatError(tc.Name, args, err)
		if e.agent.ErrorHandling.TreatErrorsAsResults {
			return errorMsg, true, nil
		}
		return "", true, errors.New(errorMsg)
	}

	e.note.Success(e.agent.ID.String(), tc.Name)

	if err := e.agent.JSONValidator.ValidateToolResponse(result); err != nil {
		return "", false, fmt.Errorf("Error: Tool '%s' produced invalid response: %v", tc.Name, err)
	}

	return e.normalizeResult(tc.Name, args, result), false, nil
}

func (e *toolExecutor) logToolStart(tc model.ToolCall, args map[string]any) {
	debug.LogToolCall(tc.Name, args, "", nil)
	debug.LogAgentAction(e.agent.ID.String(), "tool_execution_start", map[string]interface{}{
		"tool_name":    tc.Name,
		"tool_call_id": tc.ID,
		"args_count":   len(args),
	})
}

func (e *toolExecutor) logToolEnd(tc model.ToolCall, args map[string]any, result string, err error) {
	debug.LogToolCall(tc.Name, args, result, err)
	debug.LogAgentAction(e.agent.ID.String(), "tool_execution_complete", map[string]interface{}{
		"tool_name":     tc.Name,
		"tool_call_id":  tc.ID,
		"success":       err == nil,
		"result_length": len(result),
		"has_error":     err != nil,
	})
}

func (e *toolExecutor) formatError(name string, args map[string]any, err error) string {
	if !e.agent.ErrorHandling.IncludeErrorContext {
		return fmt.Sprintf("Error executing tool '%s': %v", name, err)
	}

	return fmt.Sprintf("Error executing tool '%s': %v\n\nContext:\n- Tool: %s\n- Arguments: %v\n- Suggestion: Please try a different approach or check the tool usage.", name, err, name, args)
}

func (e *toolExecutor) normalizeResult(name string, args map[string]any, result string) string {
	if strings.TrimSpace(result) != "" {
		return result
	}

	switch name {
	case "bash", "sh":
		return "Command executed successfully."
	case "create":
		if path, ok := args["path"].(string); ok {
			return fmt.Sprintf("File '%s' created successfully.", path)
		}
		return "File created successfully."
	case "edit_range", "search_replace":
		return "File edited successfully."
	default:
		return "Operation completed successfully."
	}
}

type toolUserNotifier interface {
	Start(agentID, toolName, summary string)
	Failure(agentID, toolName string, err error)
	Success(agentID, toolName string)
}

func newToolUserNotifier() toolUserNotifier {
	if os.Getenv("AGENTRY_TUI_MODE") == "1" {
		return noopToolNotifier{}
	}
	return stderrToolNotifier{}
}

type stderrToolNotifier struct{}
type noopToolNotifier struct{}

func (stderrToolNotifier) Start(agentID, toolName, summary string) {
	id := shortenID(agentID)
	if summary != "" {
		fmt.Fprintf(os.Stderr, "🔧 %s: %s %s\n", id, toolName, summary)
		return
	}
	fmt.Fprintf(os.Stderr, "🔧 %s: %s\n", id, toolName)
}

func (stderrToolNotifier) Failure(agentID, toolName string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s: %s failed: %v\n", shortenID(agentID), toolName, err)
}

func (stderrToolNotifier) Success(agentID, toolName string) {
	fmt.Fprintf(os.Stderr, "✅ %s: %s completed\n", shortenID(agentID), toolName)
}

func (noopToolNotifier) Start(string, string, string)  {}
func (noopToolNotifier) Failure(string, string, error) {}
func (noopToolNotifier) Success(string, string)        {}

func shortenID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
