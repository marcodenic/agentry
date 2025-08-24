package team

import (
	"context"
	"errors"

	// avoid import cycles by referring to tool package only in type-free way
	"github.com/marcodenic/agentry/internal/tool"
)

// Caller represents a team capable of handling delegated calls.
type Caller interface {
	Call(ctx context.Context, agent, input string) (string, error)
}

// ctxKey is used for storing the current team in a context.
type ctxKey struct{}

// ErrNoTeam indicates the context is missing a team value.
var ErrNoTeam = errors.New("no team in context")

// ErrUnknownAgent is returned when the named agent does not exist.
var ErrUnknownAgent = errors.New("unknown agent")

// WithContext returns a new context carrying t.
func WithContext(ctx context.Context, t Caller) context.Context {
	// Store under local key
	ctx = context.WithValue(ctx, ctxKey{}, t)
	// Also store under tool's context key so builtins can fetch without importing team
	return context.WithValue(ctx, tool.TeamContextKey, t)
}

// FromContext retrieves the team stored in ctx.
func FromContext(ctx context.Context) (Caller, bool) {
	t, ok := ctx.Value(ctxKey{}).(Caller)
	return t, ok
}

// TeamFromContext extracts a Team pointer if present.
// This provides compatibility with legacy converse.TeamFromContext usage.
func TeamFromContext(ctx context.Context) *Team {
	caller, ok := FromContext(ctx)
	if !ok {
		return nil
	}
	team, ok := caller.(*Team)
	if !ok {
		return nil
	}
	return team
}
