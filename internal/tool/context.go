package tool

import "context"

// SpawnFn represents a function that runs a query using a new agent derived
// from the parent.
type SpawnFn func(context.Context, string) (string, error)

type ctxKey struct{}

// WithSpawn returns a context carrying fn so tools can launch sub-agents.
func WithSpawn(ctx context.Context, fn SpawnFn) context.Context {
	return context.WithValue(ctx, ctxKey{}, fn)
}

// SpawnFromContext retrieves the SpawnFn stored in ctx.
func SpawnFromContext(ctx context.Context) (SpawnFn, bool) {
	fn, ok := ctx.Value(ctxKey{}).(SpawnFn)
	return fn, ok
}
