package tool

import (
	"context"
	"fmt"
)

// Tool represents an executable action available to agents.
type Tool interface {
	Name() string
	Description() string
	JSONSchema() map[string]any
	Execute(ctx context.Context, args map[string]any) (string, error)
}

type simpleTool struct {
	name    string
	desc    string
	schema  map[string]any
	fn      func(context.Context, map[string]any) (string, error)
	allowed bool
}

func New(name, desc string, fn func(context.Context, map[string]any) (string, error)) Tool {
	return &simpleTool{name: name, desc: desc, fn: fn, schema: map[string]any{"type": "object"}, allowed: true}
}

func NewWithSchema(name, desc string, schema map[string]any, fn func(context.Context, map[string]any) (string, error)) Tool {
	return &simpleTool{name: name, desc: desc, fn: fn, schema: schema, allowed: true}
}

func (t *simpleTool) Name() string               { return t.name }
func (t *simpleTool) Description() string        { return t.desc }
func (t *simpleTool) JSONSchema() map[string]any { return t.schema }
func (t *simpleTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	if !t.allowed {
		return "", fmt.Errorf("%w: %s", ErrToolDenied, t.name)
	}
	if !permitted(t.name) {
		return "", fmt.Errorf("%w: %s", ErrToolDenied, t.name)
	}
	return t.fn(ctx, args)
}

type Registry map[string]Tool

func (r Registry) Use(name string) (Tool, bool) {
	t, ok := r[name]
	return t, ok
}

// ExecFn defines the signature for tool execution functions.
type ExecFn func(context.Context, map[string]any) (string, error)
