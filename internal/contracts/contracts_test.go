package contracts

import (
	"context"
	"testing"
)

func TestTeamContextKeyRoundTrip(t *testing.T) {
	ctx := context.Background()
	type mockTeam struct{}

	ctx = context.WithValue(ctx, TeamContextKey, mockTeam{})
	if v := ctx.Value(TeamContextKey); v == nil {
		t.Fatal("expected value for TeamContextKey")
	}

	ctx = context.WithValue(ctx, AgentNameContextKey, "agent_1")
	if got, _ := ctx.Value(AgentNameContextKey).(string); got != "agent_1" {
		t.Fatalf("expected agent name round trip, got %q", got)
	}
}
