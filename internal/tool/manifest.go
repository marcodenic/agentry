package tool

import (
	"context"
	"errors"
	"os/exec"

	"github.com/yourname/agentry/internal/config"
)

var ErrUnknownManifest = errors.New("unknown tool manifest")

type Tool interface {
	Name() string
	Description() string
	Execute(ctx context.Context, args map[string]any) (string, error)
}

type simpleTool struct {
	name string
	desc string
	fn   func(context.Context, map[string]any) (string, error)
}

func New(name, desc string, fn func(context.Context, map[string]any) (string, error)) Tool {
	return &simpleTool{name: name, desc: desc, fn: fn}
}

func (t *simpleTool) Name() string        { return t.name }
func (t *simpleTool) Description() string { return t.desc }
func (t *simpleTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	return t.fn(ctx, args)
}

type Registry map[string]Tool

func (r Registry) Use(name string) (Tool, bool) {
	t, ok := r[name]
	return t, ok
}

func FromManifest(m config.ToolManifest) (Tool, error) {
	if m.Command != "" {
		return New(m.Name, m.Description, func(ctx context.Context, args map[string]any) (string, error) {
			cmd := exec.CommandContext(ctx, "sh", "-c", m.Command)
			out, err := cmd.CombinedOutput()
			return string(out), err
		}), nil
	}
	// HTTP wrapper left as exercise
	return nil, ErrUnknownManifest
}
