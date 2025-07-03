package team

import (
	"context"
	"errors"
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
	return context.WithValue(ctx, ctxKey{}, t)
}

// FromContext retrieves the team stored in ctx.
func FromContext(ctx context.Context) (Caller, bool) {
	t, ok := ctx.Value(ctxKey{}).(Caller)
	return t, ok
}

// Call sends input to the named agent of the team stored in ctx.
func Call(ctx context.Context, name, input string) (string, error) {
	t, ok := FromContext(ctx)
	if !ok || t == nil {
		return "", ErrNoTeam
	}
	return t.Call(ctx, name, input)
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
