package tool

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/contracts"
)

type fakeTeamService struct {
	spawned          []string
	roles            []string
	summary          string
	history          []string
	shared           map[string]interface{}
	delegateResponse string
	delegateErr      error
	lastHistoryLimit int
}

func newFakeTeamService() *fakeTeamService {
	return &fakeTeamService{
		shared:  make(map[string]interface{}),
		summary: "",
		history: nil,
	}
}

func (f *fakeTeamService) SpawnedAgentNames() []string {
	return append([]string(nil), f.spawned...)
}

func (f *fakeTeamService) AvailableRoleNames() []string {
	return append([]string(nil), f.roles...)
}

func (f *fakeTeamService) DelegateTask(ctx context.Context, role, task string) (string, error) {
	if f.delegateErr != nil {
		return "", f.delegateErr
	}
	if f.delegateResponse != "" {
		return f.delegateResponse, nil
	}
	return "", nil
}

func (f *fakeTeamService) GetCoordinationSummary() string {
	return f.summary
}

func (f *fakeTeamService) GetCoordinationHistory(limit int) []string {
	f.lastHistoryLimit = limit
	if limit <= 0 || len(f.history) <= limit {
		return append([]string(nil), f.history...)
	}
	return append([]string(nil), f.history[len(f.history)-limit:]...)
}

func (f *fakeTeamService) GetSharedData(key string) (interface{}, bool) {
	v, ok := f.shared[key]
	return v, ok
}

func (f *fakeTeamService) SetSharedData(key string, value interface{}) {
	f.shared[key] = value
}

func (f *fakeTeamService) GetAllSharedData() map[string]interface{} {
	out := make(map[string]interface{}, len(f.shared))
	for k, v := range f.shared {
		out[k] = v
	}
	return out
}

func withTeam(ctx context.Context, svc contracts.TeamService) context.Context {
	return context.WithValue(ctx, contracts.TeamContextKey, svc)
}

func TestTeamStatusSpecUsesTeamService(t *testing.T) {
	svc := newFakeTeamService()
	svc.spawned = []string{"builder", "reviewer"}
	svc.history = []string{"first", "second"}

	ctx := withTeam(context.Background(), svc)

	out, err := teamStatusSpec().Exec(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("team_status exec failed: %v", err)
	}

	if !strings.Contains(out, "builder") || !strings.Contains(out, "reviewer") {
		t.Fatalf("expected spawned agents in output, got: %s", out)
	}

	if !strings.Contains(out, "second") {
		t.Fatalf("expected coordination history in output, got: %s", out)
	}

	if svc.lastHistoryLimit != 5 {
		t.Fatalf("expected history limit 5, got %d", svc.lastHistoryLimit)
	}
}

func TestCheckAgentSpec(t *testing.T) {
	svc := newFakeTeamService()
	svc.spawned = []string{"agent_alpha"}
	ctx := withTeam(context.Background(), svc)

	okMsg, err := checkAgentSpec().Exec(ctx, map[string]any{"agent": "agent_alpha"})
	if err != nil {
		t.Fatalf("checkAgentSpec returned error for existing agent: %v", err)
	}
	if !strings.Contains(okMsg, "✅") {
		t.Fatalf("expected success indicator, got: %s", okMsg)
	}

	missMsg, err := checkAgentSpec().Exec(ctx, map[string]any{"agent": "agent_beta"})
	if err != nil {
		t.Fatalf("checkAgentSpec returned error for missing agent: %v", err)
	}
	if !strings.Contains(missMsg, "❌ Agent 'agent_beta'") {
		t.Fatalf("expected missing agent message, got: %s", missMsg)
	}
}

func TestAvailableRolesSpec(t *testing.T) {
	svc := newFakeTeamService()
	svc.roles = []string{"coder", "reviewer"}
	svc.spawned = []string{"reviewer"}
	ctx := withTeam(context.Background(), svc)

	out, err := availableRolesSpec().Exec(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("availableRolesSpec exec failed: %v", err)
	}

	if !strings.Contains(out, "coder") || !strings.Contains(out, "reviewer") {
		t.Fatalf("expected role names in output, got: %s", out)
	}

	if !strings.Contains(out, "(currently running)") {
		t.Fatalf("expected running role marker, got: %s", out)
	}
}

func TestSharedMemorySpecActions(t *testing.T) {
	t.Run("set_get_list", func(t *testing.T) {
		svc := newFakeTeamService()
		ctx := withTeam(context.Background(), svc)

		if _, err := sharedMemoryExec(ctx, map[string]any{"action": "set", "key": "doc", "value": "draft"}); err != nil {
			t.Fatalf("set action returned error: %v", err)
		}

		val, ok := svc.shared["doc"].(string)
		if !ok || val != "draft" {
			t.Fatalf("expected shared memory to contain value, got %#v", svc.shared["doc"])
		}

		getOut, err := sharedMemoryExec(ctx, map[string]any{"action": "get", "key": "doc"})
		if err != nil {
			t.Fatalf("get action returned error: %v", err)
		}
		if !strings.Contains(getOut, "doc = draft") {
			t.Fatalf("unexpected get output: %s", getOut)
		}

		listOut, err := sharedMemoryExec(ctx, map[string]any{"action": "list"})
		if err != nil {
			t.Fatalf("list action returned error: %v", err)
		}
		if !strings.Contains(listOut, "doc") {
			t.Fatalf("expected list output to include key, got: %s", listOut)
		}
	})

	t.Run("invalid_action", func(t *testing.T) {
		svc := newFakeTeamService()
		ctx := withTeam(context.Background(), svc)

		if _, err := sharedMemoryExec(ctx, map[string]any{"action": "unknown"}); err == nil {
			t.Fatalf("expected error for invalid action")
		}
	})

	t.Run("missing_team", func(t *testing.T) {
		if _, err := sharedMemoryExec(context.Background(), map[string]any{"action": "list"}); err == nil {
			t.Fatalf("expected error when team context missing")
		}
	})
}

func TestCoordinationStatusSpecVariants(t *testing.T) {
	svc := newFakeTeamService()
	svc.summary = "summary text"
	svc.history = make([]string, 12)
	for i := range svc.history {
		svc.history[i] = "event" + strconv.Itoa(i)
	}
	ctx := withTeam(context.Background(), svc)

	summaryOut, err := coordinationStatusSpec().Exec(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("summary detail returned error: %v", err)
	}
	if summaryOut != "summary text" {
		t.Fatalf("unexpected summary output: %s", summaryOut)
	}

	fullOut, err := coordinationStatusSpec().Exec(ctx, map[string]any{"detail": "full"})
	if err != nil {
		t.Fatalf("full detail returned error: %v", err)
	}
	if svc.lastHistoryLimit != 0 {
		t.Fatalf("expected unlimited history request, got %d", svc.lastHistoryLimit)
	}
	if !strings.Contains(fullOut, "event11") {
		t.Fatalf("expected latest event in full output, got: %s", fullOut)
	}

	recentOut, err := coordinationStatusSpec().Exec(ctx, map[string]any{"detail": "recent"})
	if err != nil {
		t.Fatalf("recent detail returned error: %v", err)
	}
	if svc.lastHistoryLimit != 10 {
		t.Fatalf("expected recent history limit 10, got %d", svc.lastHistoryLimit)
	}
	if strings.Contains(recentOut, "event0") {
		t.Fatalf("expected oldest events trimmed, got: %s", recentOut)
	}

	if _, err := coordinationStatusSpec().Exec(ctx, map[string]any{"detail": "bogus"}); err == nil {
		t.Fatalf("expected error for invalid detail")
	}
}

func TestWorkspaceEventsExec(t *testing.T) {
	svc := newFakeTeamService()
	ctx := withTeam(context.Background(), svc)

	ts1 := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	ts2 := time.Date(2024, 1, 1, 12, 1, 0, 0, time.UTC)
	svc.shared["workspace_events"] = []interface{}{
		map[string]interface{}{"timestamp": ts1, "agent_id": "alpha", "description": "opened file"},
		map[string]interface{}{"timestamp": ts2, "agent_id": "beta", "description": "saved file"},
	}

	out, err := workspaceEventsExec(ctx, map[string]any{"limit": 1})
	if err != nil {
		t.Fatalf("workspace events exec failed: %v", err)
	}
	if !strings.Contains(out, "saved file") {
		t.Fatalf("expected most recent event, got: %s", out)
	}
	if strings.Contains(out, "opened file") {
		t.Fatalf("expected older event trimmed by limit, got: %s", out)
	}

	svc.shared["workspace_events"] = []interface{}{}

	emptyOut, err := workspaceEventsExec(ctx, map[string]any{})
	if err != nil {
		t.Fatalf("workspace events exec failed on empty list: %v", err)
	}
	if emptyOut != "📭 No workspace events" {
		t.Fatalf("unexpected empty output: %s", emptyOut)
	}
}
