package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/marcodenic/agentry/internal/config"
)

// runtime import kept for platform-specific code elsewhere in package; no local vars needed here.

func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	wd, _ := os.Getwd()
	return filepath.Join(wd, p)
}

var ErrUnknownManifest = errors.New("unknown tool manifest")
var ErrUnknownBuiltin = errors.New("unknown builtin tool")
var ErrToolDenied = errors.New("tool not permitted")

var allowedTools map[string]struct{}

// SetPermissions configures which tools may execute. Nil or empty slice allows all.
func SetPermissions(list []string) {
	if len(list) == 0 {
		allowedTools = nil
		return
	}
	allowedTools = make(map[string]struct{}, len(list))
	for _, n := range list {
		allowedTools[n] = struct{}{}
	}
}

func permitted(name string) bool {
	if allowedTools == nil {
		return true
	}
	_, ok := allowedTools[name]
	return ok
}

// viewedFiles tracks file paths read via the view builtin along with their
// modification time. It is used to prevent overwriting files that have changed
// on disk since they were last viewed.
var viewedFiles sync.Map

func recordView(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	viewedFiles.Store(path, info.ModTime())
	return nil
}

func checkForOverwrite(path string) error {
	// Overwrite checks are disabled to avoid interactive prompts and friction.
	// Git provides safety for undoing changes. We still keep recordView for tooling.
	_ = path
	return nil
}

// IsBuiltinTool checks if the given name is a builtin tool
func IsBuiltinTool(name string) bool {
	_, exists := builtinMap[name]
	return exists
}

// DefaultRegistry returns all builtin tools.
func DefaultRegistry() Registry {
	r := make(Registry, len(builtinMap))
	for n, s := range builtinMap {
		r[n] = NewWithSchema(n, s.Desc, s.Schema, s.Exec)
	}

	return r
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
		spec, ok := builtinMap[m.Name]
		if !ok {
			return nil, ErrUnknownBuiltin
		}
		desc := m.Description
		if desc == "" {
			desc = spec.Desc
		}
		tl := NewWithSchema(m.Name, desc, spec.Schema, spec.Exec)
		if st, ok := tl.(*simpleTool); ok {
			allowed := true
			if m.Permissions.Allow != nil {
				allowed = *m.Permissions.Allow
			}
			st.allowed = allowed
		}
		return tl, nil
	}

	// HTTP tools
	if m.HTTP != "" {
		tl := NewWithSchema(m.Name, m.Description, map[string]any{"type": "object", "properties": map[string]any{}}, func(ctx context.Context, args map[string]any) (string, error) {
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
		})
		if st, ok := tl.(*simpleTool); ok {
			allowed := true
			if m.Permissions.Allow != nil {
				allowed = *m.Permissions.Allow
			}
			st.allowed = allowed
		}
		return tl, nil
	}

	// Shell command tools
	if m.Command != "" {
		tl := NewWithSchema(m.Name, m.Description, map[string]any{"type": "object", "properties": map[string]any{}}, func(ctx context.Context, args map[string]any) (string, error) {
			if !m.Privileged {
				return ExecDirect(ctx, m.Command)
			}
			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(ctx, "cmd", "/C", m.Command)
			} else {
				cmd = exec.CommandContext(ctx, "sh", "-c", m.Command)
			}
			out, err := cmd.CombinedOutput()
			return string(out), err
		})
		if st, ok := tl.(*simpleTool); ok {
			allowed := true
			if m.Permissions.Allow != nil {
				allowed = *m.Permissions.Allow
			}
			st.allowed = allowed
		}
		return tl, nil
	}

	return nil, ErrUnknownManifest
}

// parsePatchFiles moved to patch package and TUI; keep no duplicate here.
