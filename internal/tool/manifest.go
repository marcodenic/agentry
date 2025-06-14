package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/marcodenic/agentry/internal/config"
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

// ExecFn defines the signature for tool execution functions.
type ExecFn func(context.Context, map[string]any) (string, error)

// builtinMap holds safe builtin tools keyed by name.
var builtinMap = map[string]ExecFn{
	"echo": func(ctx context.Context, args map[string]any) (string, error) {
		txt, _ := args["text"].(string)
		return txt, nil
	},
}

func FromManifest(m config.ToolManifest) (Tool, error) {
	// ensure only one of builtin, http or command is specified
	count := 0
	if m.Type != "" {
		count++
	}
	if m.HTTP != "" {
		count++
	}
	if m.Command != "" {
		count++
	}
	if count != 1 {
		return nil, ErrUnknownManifest
	}

	// Builtin Go tools
	if m.Type == "builtin" {
		fn, ok := builtinMap[m.Name]
		if !ok {
			return nil, errors.New("unknown builtin tool")
		}
		return New(m.Name, m.Description, fn), nil
	}

	// HTTP tools
	if m.HTTP != "" {
		return New(m.Name, m.Description, func(ctx context.Context, args map[string]any) (string, error) {
			b, err := json.Marshal(args)
			if err != nil {
				return "", err
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.HTTP, bytes.NewReader(b))
			if err != nil {
				return "", err
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return "", err
			}
			defer resp.Body.Close()
			rb, err := io.ReadAll(resp.Body)
			if err != nil {
				return "", err
			}
			return string(rb), nil
		}), nil
	}

	// Shell command tools (advanced use, may behave differently across OSes)
	if m.Command != "" {
		return New(m.Name, m.Description, func(ctx context.Context, args map[string]any) (string, error) {
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(ctx, "cmd", "/C", m.Command)
			} else {
				cmd = exec.CommandContext(ctx, "sh", "-c", m.Command)
			}
			out, err := cmd.CombinedOutput()
			return string(out), err
		}), nil
	}

	return nil, ErrUnknownManifest
}
