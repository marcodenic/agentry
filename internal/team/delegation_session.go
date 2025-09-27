package team

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	teamruntime "github.com/marcodenic/agentry/internal/teamruntime"
)

type delegationSession struct {
	team    *Team
	agentID string
	input   string

	timer     *Timer
	startTime time.Time

	agent            *Agent
	workspaceContext string
	timeout          time.Duration
	notifier         teamruntime.Notifier
	telemetry        teamruntime.DelegationTelemetry

	ctx    context.Context
	runCtx context.Context
	cancel context.CancelFunc
}

func newDelegationSession(team *Team, agentID, input string) *delegationSession {
	return &delegationSession{team: team, agentID: agentID, input: input, notifier: teamruntime.NewNotifier()}
}

func (s *delegationSession) Run(ctx context.Context) (string, error) {
	s.ctx = ctx
	s.timer = StartTimer(fmt.Sprintf("Call(%s)", s.agentID))
	defer s.timer.Stop()

	s.telemetry = teamruntime.NewDelegationTelemetry(s.agentID, s.input, s.team, s.notifier, s.timer)
	s.telemetry.Start()

	if err := s.ensureAgent(); err != nil {
		return "", err
	}

	s.agent.SetStatus("working")
	s.telemetry.WorkStart()
	s.telemetry.LogTaskFile()

	events := s.team.GetWorkspaceEvents(5)
	s.timer.Checkpoint("inbox processing completed")
	s.workspaceContext = teamruntime.BuildWorkspaceContext(convertWorkspaceEvents(events))
	s.timer.Checkpoint("context and events prepared")

	s.configureTimeout()
	if s.cancel != nil {
		defer s.cancel()
	}
	s.publishStartEvent()

	s.startTime = time.Now()
	s.telemetry.RunAgentStart()
	result, err := runAgent(s.runCtx, s.agent.Agent, s.augmentedInput(), s.agentID, s.team.GetAgents())
	duration := time.Since(s.startTime)
	s.timer.Checkpoint("runAgent completed")
	s.telemetry.RunAgentComplete(duration)

	outcome, outcomeErr := s.processOutcome(result, err)
	s.timer.Checkpoint("cleanup completed")
	return outcome, outcomeErr
}

func (s *delegationSession) ensureAgent() error {
	s.team.mutex.RLock()
	agent, exists := s.team.agentsByName[s.agentID]
	s.team.mutex.RUnlock()
	s.timer.Checkpoint("agent lookup completed")

	if !exists {
		teamruntime.Debugf("🆕 Creating new agent: %s\n", s.agentID)
		s.notifier.User("🆕 Creating %s agent...\n", s.agentID)

		spawnedAgent, err := s.team.SpawnAgent(s.ctx, s.agentID, s.agentID)
		if err != nil {
			teamruntime.Debugf("❌ Failed to spawn agent %s: %v\n", s.agentID, err)
			return fmt.Errorf("failed to spawn agent %s: %w", s.agentID, err)
		}

		s.timer.Checkpoint("new agent spawned")
		agent = spawnedAgent
		teamruntime.Debugf("✅ Agent %s created and ready\n", s.agentID)
		s.notifier.User("✅ %s agent ready\n", s.agentID)
	} else {
		s.timer.Checkpoint("existing agent found")
		teamruntime.Debugf("♻️  Using existing agent: %s (Status: %s)\n", s.agentID, agent.Status)
	}

	s.agent = agent
	return nil
}

func (s *delegationSession) configureTimeout() {
	timeout := 15 * time.Minute
	if v := os.Getenv("AGENTRY_DELEGATION_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			timeout = d
		}
	}

	s.timeout = timeout
	teamruntime.Debugf("🔧 Call: Creating context with timeout %s for agent %s", timeout, s.agentID)
	s.runCtx, s.cancel = context.WithTimeout(s.ctx, timeout)
}

func convertWorkspaceEvents(events []WorkspaceEvent) []teamruntime.WorkspaceEvent {
	out := make([]teamruntime.WorkspaceEvent, 0, len(events))
	for _, e := range events {
		out = append(out, teamruntime.WorkspaceEvent{
			AgentID:     e.AgentID,
			Type:        e.Type,
			Description: e.Description,
			Timestamp:   e.Timestamp,
		})
	}
	return out
}

func (s *delegationSession) publishStartEvent() {
	if teamruntime.IsTUI() {
		return
	}

	s.team.PublishWorkspaceEvent(
		"agent_0",
		"delegation_started",
		fmt.Sprintf("Delegated to %s", s.agentID),
		map[string]interface{}{"agent": s.agentID, "timeout": s.timeout.String()},
	)
}

func (s *delegationSession) augmentedInput() string {
	if strings.TrimSpace(s.workspaceContext) == "" {
		return s.input
	}
	return s.input + s.workspaceContext
}

func (s *delegationSession) processOutcome(result string, runErr error) (string, error) {
	if runErr != nil {
		return s.handleError(runErr)
	}
	return s.handleSuccess(result), nil
}

func (s *delegationSession) handleError(runErr error) (string, error) {
	teamruntime.Debugf("❌ Call: runAgent failed for %s: %v", s.agentID, runErr)
	if errors.Is(runErr, context.DeadlineExceeded) || errors.Is(runErr, context.Canceled) {
		if s.team.checkWorkCompleted(s.agentID, s.input) {
			msg := s.telemetry.TimeoutWithWork(s.timeout)
			return msg, nil
		}

		msg := s.telemetry.TimeoutWithoutWork(s.timeout)
		if !teamruntime.IsTUI() {
			s.team.PublishWorkspaceEvent("agent_0", "delegation_timeout", msg, map[string]interface{}{"agent": s.agentID})
		}
		return "", errors.New(msg)
	}

	s.agent.SetStatus("error")
	s.telemetry.RecordFailure(runErr)
	errorFeedback := fmt.Sprintf("❌ Agent '%s' encountered an error: %v\n\nSuggestions:\n- Try a different approach\n- Simplify the request\n- Use alternative tools\n- Break the task into smaller steps", s.agentID, runErr)
	return "", errors.New(errorFeedback)
}

func (s *delegationSession) handleSuccess(result string) string {
	s.agent.SetStatus("ready")
	s.telemetry.RecordSuccess(result)
	s.team.SetSharedData(fmt.Sprintf("last_result_%s", s.agentID), result)
	s.team.SetSharedData(fmt.Sprintf("last_task_%s", s.agentID), s.input)
	return result
}
