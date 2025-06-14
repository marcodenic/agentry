package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
	"ping": func(ctx context.Context, args map[string]any) (string, error) {
		host, _ := args["host"].(string)
		if host == "" {
			return "", errors.New("missing host")
		}
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "ping", "-n", "4", host)
		} else {
			cmd = exec.CommandContext(ctx, "ping", "-c", "4", host)
		}
		out, err := cmd.CombinedOutput()
		return string(out), err
	},
	"bash": func(ctx context.Context, args map[string]any) (string, error) {
		cmdStr, _ := args["cmd"].(string)
		if cmdStr == "" {
			return "", errors.New("missing cmd")
		}
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.CommandContext(ctx, "cmd", "/C", cmdStr)
		} else {
			cmd = exec.CommandContext(ctx, "bash", "-c", cmdStr)
		}
		out, err := cmd.CombinedOutput()
		return string(out), err
	},
	"fetch": func(ctx context.Context, args map[string]any) (string, error) {
		url, _ := args["url"].(string)
		if url == "" {
			return "", errors.New("missing url")
		}
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"ls": func(ctx context.Context, args map[string]any) (string, error) {
		path, _ := args["path"].(string)
		if path == "" {
			path = "."
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}
		names := make([]string, len(entries))
		for i, e := range entries {
			names[i] = e.Name()
		}
		return strings.Join(names, "\n"), nil
	},
	"view": func(ctx context.Context, args map[string]any) (string, error) {
		path, _ := args["path"].(string)
		if path == "" {
			return "", errors.New("missing path")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"write": func(ctx context.Context, args map[string]any) (string, error) {
		path, _ := args["path"].(string)
		text, _ := args["text"].(string)
		if path == "" {
			return "", errors.New("missing path")
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			return "", err
		}
		return "written", nil
	},
	"glob": func(ctx context.Context, args map[string]any) (string, error) {
		pattern, _ := args["pattern"].(string)
		if pattern == "" {
			return "", errors.New("missing pattern")
		}
		files, err := filepath.Glob(pattern)
		if err != nil {
			return "", err
		}
		return strings.Join(files, "\n"), nil
	},
	"grep": func(ctx context.Context, args map[string]any) (string, error) {
		pattern, _ := args["pattern"].(string)
		path, _ := args["path"].(string)
		if path == "" || pattern == "" {
			return "", errors.New("missing args")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		lines := strings.Split(string(b), "\n")
		var out []string
		for _, l := range lines {
			if strings.Contains(l, pattern) {
				out = append(out, l)
			}
		}
		return strings.Join(out, "\n"), nil
	},
	"edit": func(ctx context.Context, args map[string]any) (string, error) {
		path, _ := args["path"].(string)
		text, _ := args["text"].(string)
		if path == "" {
			return "", errors.New("missing path")
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			return "", err
		}
		return "edited", nil
	},
	"sourcegraph": func(ctx context.Context, args map[string]any) (string, error) {
		q, _ := args["query"].(string)
		if q == "" {
			return "", errors.New("missing query")
		}
		url := "https://sourcegraph.com/search?q=" + url.QueryEscape(q) + "&format=json"
		resp, err := http.Get(url)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(b), nil
	},
	"patch": func(ctx context.Context, args map[string]any) (string, error) {
		patchStr, _ := args["patch"].(string)
		if patchStr == "" {
			return "", errors.New("missing patch")
		}
		cmd := exec.CommandContext(ctx, "patch", "-p0")
		cmd.Stdin = strings.NewReader(patchStr)
		out, err := cmd.CombinedOutput()
		return string(out), err
	},
	"agent": func(ctx context.Context, args map[string]any) (string, error) {
		// placeholder that simply echoes the query
		query, _ := args["query"].(string)
		return "agent searched: " + query, nil
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
