package teamruntime

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/debug"
)

type Notifier interface {
	User(format string, args ...interface{})
	File(format string, args ...interface{})
}

type TeamNotifier struct{}

func NewNotifier() TeamNotifier { return TeamNotifier{} }

var logToFileFunc = debug.LogToFile

func SetLogToFile(fn func(level, format string, args ...interface{})) {
	if fn == nil {
		logToFileFunc = debug.LogToFile
		return
	}
	logToFileFunc = fn
}


func (TeamNotifier) User(format string, args ...interface{}) {
	if IsTUI() {
		return
	}
	fmt.Fprintf(os.Stderr, format, args...)
}

func (TeamNotifier) File(format string, args ...interface{}) {
	LogToFile(fmt.Sprintf(format, args...))
}

func IsTUI() bool { return os.Getenv("AGENTRY_TUI_MODE") == "1" }

func LogToFile(message string) {
	logToFileFunc("TEAM", "%s", message)
}

func Debugf(format string, args ...interface{}) {
	debug.Printf(format, args...)
}

func IsDebug() bool { return debug.DebugEnabled }

type CoordinationLogger interface {
	LogCoordinationEvent(eventType, from, to, content string, metadata map[string]interface{})
}

type Timer interface {
	Checkpoint(label string) time.Duration
}

type DelegationTelemetry struct {
	agentID  string
	task     string
	notifier Notifier
	logger   CoordinationLogger
	timer    Timer
}

func NewDelegationTelemetry(agentID, task string, logger CoordinationLogger, notifier Notifier, timer Timer) DelegationTelemetry {
	if notifier == nil {
		notifier = NewNotifier()
	}
	return DelegationTelemetry{agentID: agentID, task: task, notifier: notifier, logger: logger, timer: timer}
}

func (d DelegationTelemetry) Start() {
	Debugf("\n🔄 AGENT DELEGATION: Agent 0 -> %s\n", d.agentID)
	Debugf("📝 Task: %s\n", d.task)
	Debugf("⏰ Timestamp: %s\n", time.Now().Format("15:04:05"))
	d.notifier.User("🔄 Delegating to %s agent...\n", d.agentID)
	if d.logger != nil {
		d.logger.LogCoordinationEvent("delegation", "agent_0", d.agentID, d.task, map[string]interface{}{
			"task_length": len(d.task),
			"agent_type":  d.agentID,
		})
	}
	if d.timer != nil {
		d.timer.Checkpoint("coordination logged")
	}
}

func (d DelegationTelemetry) WorkStart() {
	Debugf("🚀 Starting task execution on agent %s...\n", d.agentID)
	d.notifier.User("🚀 %s agent working on task...\n", d.agentID)
}

func (d DelegationTelemetry) LogTaskFile() {
	d.notifier.File("DELEGATION: Agent 0 -> %s | Task: %s", d.agentID, d.task)
	if d.timer != nil {
		d.timer.Checkpoint("logging completed")
	}
}

func (d DelegationTelemetry) RunAgentStart() {
	Debugf("🔧 Call: About to call runAgent for %s", d.agentID)
}

func (d DelegationTelemetry) RunAgentComplete(duration time.Duration) {
	Debugf("🔧 Call: runAgent completed for %s in %s", d.agentID, duration)
}

func (d DelegationTelemetry) TimeoutWithWork(timeout time.Duration) string {
	msg := fmt.Sprintf("✅ %s agent completed the work successfully (response generation timed out after %s but files were created)", d.agentID, timeout)
	d.notifier.User("✅ %s agent completed work successfully (response timed out)\n", d.agentID)
	if d.logger != nil {
		d.logger.LogCoordinationEvent("delegation_success_timeout", d.agentID, "agent_0", msg, map[string]interface{}{"timeout": timeout.String()})
	}
	return msg
}

func (d DelegationTelemetry) TimeoutWithoutWork(timeout time.Duration) string {
	msg := fmt.Sprintf("⏳ Delegation to '%s' timed out after %s without completing work. Consider simplifying the task, choosing a different agent, or increasing AGENTRY_DELEGATION_TIMEOUT.", d.agentID, timeout)
	d.notifier.User("⏳ %s agent timed out without completing work\n", d.agentID)
	if d.logger != nil {
		d.logger.LogCoordinationEvent("delegation_timeout", d.agentID, "agent_0", msg, map[string]interface{}{"timeout": timeout.String()})
	}
	return msg
}

func (d DelegationTelemetry) RecordFailure(err error) {
	Debugf("❌ Agent %s failed: %v\n", d.agentID, err)
	LogToFile(fmt.Sprintf("DELEGATION FAILED: %s | Error: %v", d.agentID, err))
	if d.logger != nil {
		d.logger.LogCoordinationEvent("delegation_failed", d.agentID, "agent_0", err.Error(), map[string]interface{}{"error": err.Error()})
	}
}

func (d DelegationTelemetry) RecordSuccess(result string) {
	Debugf("✅ Agent %s completed successfully\n", d.agentID)
	d.notifier.User("✅ %s agent completed task\n", d.agentID)
	Debugf("📤 Result length: %d characters\n", len(result))
	if d.logger != nil {
		d.logger.LogCoordinationEvent("delegation_success", d.agentID, "agent_0", "Task completed", map[string]interface{}{"result_length": len(result), "agent_type": d.agentID})
	}
	Debugf("🏁 Delegation complete: Agent 0 <- %s\n\n", d.agentID)
}

type WorkspaceEvent struct {
	AgentID     string
	Type        string
	Description string
	Timestamp   time.Time
}

func BuildWorkspaceContext(events []WorkspaceEvent) string {
	if len(events) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("\n\nRECENT WORKSPACE EVENTS:\n")
	for _, e := range events {
		sb.WriteString("- ")
		if !e.Timestamp.IsZero() {
			sb.WriteString("[" + e.Timestamp.Format("15:04:05") + "] ")
		}
		if e.AgentID != "" {
			sb.WriteString(e.AgentID + " | ")
		}
		sb.WriteString(e.Type)
		if e.Description != "" {
			sb.WriteString(": " + e.Description)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
