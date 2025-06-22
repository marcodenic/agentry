package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/patch"
)

var osType = runtime.GOOS
var shellCmd, shellFlag string

func init() {
	if osType == "windows" {
		shellCmd = "powershell.exe"
		shellFlag = "-Command"
	} else {
		shellCmd = "bash"
		shellFlag = "-c"
	}
}

func absPath(p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	wd, _ := os.Getwd()
	return filepath.Join(wd, p)
}

var ErrUnknownManifest = errors.New("unknown tool manifest")
var ErrUnknownBuiltin = errors.New("unknown builtin tool")

type Tool interface {
	Name() string
	Description() string
	JSONSchema() map[string]any
	Execute(ctx context.Context, args map[string]any) (string, error)
}

type simpleTool struct {
	name   string
	desc   string
	schema map[string]any
	fn     func(context.Context, map[string]any) (string, error)
}

func New(name, desc string, fn func(context.Context, map[string]any) (string, error)) Tool {
	return &simpleTool{name: name, desc: desc, fn: fn, schema: map[string]any{"type": "object"}}
}

func NewWithSchema(name, desc string, schema map[string]any, fn func(context.Context, map[string]any) (string, error)) Tool {
	return &simpleTool{name: name, desc: desc, fn: fn, schema: schema}
}

func (t *simpleTool) Name() string               { return t.name }
func (t *simpleTool) Description() string        { return t.desc }
func (t *simpleTool) JSONSchema() map[string]any { return t.schema }
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

// builtinSpec defines builtin schema and execution.
type builtinSpec struct {
	Desc   string
	Schema map[string]any
	Exec   ExecFn
}

// builtinMap holds safe builtin tools keyed by name.
var builtinMap = map[string]builtinSpec{
	"echo": {
		Desc: "Repeat a string",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"text": map[string]any{"type": "string"}},
			"required":   []string{"text"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			txt, _ := args["text"].(string)
			return txt, nil
		},
	},
	"ping": {
		Desc: "Ping a host",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"host": map[string]any{"type": "string"}},
			"required":   []string{"host"},
			"example":    map[string]any{"host": "example.com"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			host, _ := args["host"].(string)
			if host == "" {
				return "", errors.New("missing host")
			}
			d := net.Dialer{Timeout: 3 * time.Second}
			start := time.Now()
			conn, err := d.DialContext(ctx, "tcp", host+":80")
			if err != nil {
				return "", err
			}
			_ = conn.Close()
			return fmt.Sprintf("pong in %v", time.Since(start)), nil
		},
	},
	"mcp": {
		Desc: "Execute an MCP command",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"host":    map[string]any{"type": "string"},
				"port":    map[string]any{"type": "integer"},
				"command": map[string]any{"type": "string"},
			},
			"required": []string{"host", "port", "command"},
			"example":  map[string]any{"host": "localhost", "port": 1234, "command": "hello"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			host, _ := args["host"].(string)
			var portStr string
			switch p := args["port"].(type) {
			case float64:
				portStr = fmt.Sprintf("%d", int(p))
			case int:
				portStr = fmt.Sprintf("%d", p)
			case string:
				portStr = p
			}
			cmdStr, _ := args["command"].(string)
			if host == "" || portStr == "" || cmdStr == "" {
				return "", errors.New("missing args")
			}
			d := net.Dialer{Timeout: 3 * time.Second}
			conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, portStr))
			if err != nil {
				return "", err
			}
			defer conn.Close()
			_ = conn.SetDeadline(time.Now().Add(3 * time.Second))
			if _, err := conn.Write([]byte(cmdStr)); err != nil {
				return "", err
			}
			resp, err := io.ReadAll(conn)
			if err != nil {
				return "", err
			}
			return string(resp), nil
		},
	},
	"bash": {
		Desc: "Execute a bash command",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"cmd": map[string]any{"type": "string"}},
			"required":   []string{"cmd"},
			"example":    map[string]any{"cmd": "echo hi"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			cmdStr, _ := args["cmd"].(string)
			if cmdStr == "" {
				return "", errors.New("missing cmd")
			}
			if _, err := exec.LookPath(shellCmd); err != nil {
				return "", fmt.Errorf("%s not found – install or use another tool", shellCmd)
			}
			cmd := exec.CommandContext(ctx, shellCmd, shellFlag, cmdStr)
			out, err := cmd.CombinedOutput()
			return string(out), err
		},
	},
	"fetch": {
		Desc: "Download content from a URL",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"url": map[string]any{"type": "string"}},
			"required":   []string{"url"},
			"example":    map[string]any{"url": "https://example.com"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			u, _ := args["url"].(string)
			if u == "" {
				return "", errors.New("missing url")
			}
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if err != nil {
				return "", err
			}
			resp, err := http.DefaultClient.Do(req)
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
	},
	"ls": {
		Desc: "List directory contents",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"path": map[string]any{"type": "string"}},
			"example":    map[string]any{"path": "."},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			path, _ := args["path"].(string)
			if path == "" {
				path = "."
			}
			path = absPath(path)
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
	},
	"view": {
		Desc: "Read a file",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"path": map[string]any{"type": "string"}},
			"required":   []string{"path"},
			"example":    map[string]any{"path": "go.mod"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			path, _ := args["path"].(string)
			if path == "" {
				return "", errors.New("missing path")
			}
			path = absPath(path)
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("file %s does not exist – create it first", path)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	},
	"write": {
		Desc: "Create or overwrite a file",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
				"text": map[string]any{"type": "string"},
			},
			"required": []string{"path", "text"},
			"example":  map[string]any{"path": "tmp.txt", "text": "hi"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			path, _ := args["path"].(string)
			text, _ := args["text"].(string)
			if path == "" {
				return "", errors.New("missing path")
			}
			path = absPath(path)
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				return "", err
			}
			return "written", nil
		},
	},
	"glob": {
		Desc: "Find files by pattern",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"pattern": map[string]any{"type": "string"}},
			"required":   []string{"pattern"},
			"example":    map[string]any{"pattern": "*.go"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
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
	},
	"grep": {
		Desc: "Search file contents",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pattern": map[string]any{"type": "string"},
				"path":    map[string]any{"type": "string"},
			},
			"required": []string{"pattern", "path"},
			"example":  map[string]any{"pattern": "hello", "path": "go.mod"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			pattern, _ := args["pattern"].(string)
			path, _ := args["path"].(string)
			if path == "" || pattern == "" {
				return "", errors.New("missing args")
			}
			path = absPath(path)
			if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
				return "", fmt.Errorf("file %s does not exist – create it first", path)
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
	},
	"edit": {
		Desc: "Update an existing file",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{"type": "string"},
				"text": map[string]any{"type": "string"},
			},
			"required": []string{"path", "text"},
			"example":  map[string]any{"path": "tmp.txt", "text": "bye"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			path, _ := args["path"].(string)
			text, _ := args["text"].(string)
			if path == "" {
				return "", errors.New("missing path")
			}
			path = absPath(path)
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				return "", err
			}
			return "edited", nil
		},
	},
	"sourcegraph": {
		Desc: "Search public repositories",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"query": map[string]any{"type": "string"}},
			"required":   []string{"query"},
			"example":    map[string]any{"query": "grpc"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
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
	},
	"agent": {
		Desc: "Launch a search agent",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"query": map[string]any{"type": "string"}},
			"required":   []string{"query"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			query, _ := args["query"].(string)
			return "agent searched: " + query, nil
		},
	},
}

func init() {
	builtinMap["patch"] = builtinSpec{
		Desc: "Apply a unified diff patch",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"patch": map[string]any{"type": "string"}},
			"required":   []string{"patch"},
			"example":    map[string]any{"patch": ""},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			patchStr, _ := args["patch"].(string)
			if patchStr == "" {
				return "", errors.New("missing patch")
			}
			res, err := patch.Apply(patchStr)
			if err != nil {
				return "", err
			}
			return patch.MarshalResult(res)
		},
	}
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
		return NewWithSchema(m.Name, desc, spec.Schema, spec.Exec), nil
	}

	// HTTP tools
	if m.HTTP != "" {
		return NewWithSchema(m.Name, m.Description, map[string]any{"type": "object", "properties": map[string]any{}}, func(ctx context.Context, args map[string]any) (string, error) {
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
		return NewWithSchema(m.Name, m.Description, map[string]any{"type": "object", "properties": map[string]any{}}, func(ctx context.Context, args map[string]any) (string, error) {
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

func parsePatchFiles(patchStr string) []string {
	var files []string
	for _, line := range strings.Split(patchStr, "\n") {
		if strings.HasPrefix(line, "+++ ") {
			f := strings.TrimPrefix(line, "+++ ")
			f = strings.TrimPrefix(f, "b/")
			if f != "/dev/null" && f != "" {
				files = append(files, f)
			}
		}
	}
	return files
}
