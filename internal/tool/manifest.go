package tool

import (
	"bufio"
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
	"sync"
	"time"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/converse"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/patch"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/pkg/flow"
	"github.com/marcodenic/agentry/pkg/sbox"
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

// confirmOverwrite toggles interactive confirmation before overwriting. Set the
// AGENTRY_CONFIRM environment variable to any non-empty value to enable.
var confirmOverwrite = os.Getenv("AGENTRY_CONFIRM") != ""

func recordView(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	viewedFiles.Store(path, info.ModTime())
	return nil
}

func confirm(msg string) bool {
	if !confirmOverwrite {
		return false
	}
	fmt.Printf("%s [y/N]: ", msg)
	rd := bufio.NewReader(os.Stdin)
	line, _ := rd.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

func checkForOverwrite(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	v, ok := viewedFiles.Load(path)
	if !ok {
		if confirm("overwrite " + path + " without viewing?") {
			return nil
		}
		return fmt.Errorf("file %s must be viewed before modification", path)
	}
	if mod := v.(time.Time); !info.ModTime().Equal(mod) {
		if confirm("file " + path + " changed since viewed, overwrite?") {
			return nil
		}
		return fmt.Errorf("file %s changed since viewed", path)
	}
	return nil
}

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
		}, Exec: func(ctx context.Context, args map[string]any) (string, error) {
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
	"branch-tidy": {
		Desc: "Delete all local Git branches except the current one",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"force": map[string]any{
					"type":        "boolean",
					"description": "Use -D flag to force delete branches (default: false uses -d)",
					"default":     false,
				},
				"dry-run": map[string]any{
					"type":        "boolean",
					"description": "Show which branches would be deleted without actually deleting them",
					"default":     false,
				},
			},
			"example": map[string]any{"force": false, "dry-run": true},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			force, _ := args["force"].(bool)
			dryRun, _ := args["dry-run"].(bool)

			// Get current branch
			currentCmd := exec.CommandContext(ctx, "git", "rev-parse", "--abbrev-ref", "HEAD")
			currentOut, err := currentCmd.Output()
			if err != nil {
				return "", fmt.Errorf("failed to get current branch: %v", err)
			}
			current := strings.TrimSpace(string(currentOut))
			if current == "" {
				return "", errors.New("could not determine current branch")
			}

			// Get all local branches
			branchCmd := exec.CommandContext(ctx, "git", "branch")
			branchOut, err := branchCmd.Output()
			if err != nil {
				return "", fmt.Errorf("failed to list branches: %v", err)
			}

			var result strings.Builder
			var deleted []string
			var skipped []string

			lines := strings.Split(string(branchOut), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}

				// Skip current branch (marked with * or just the name)
				branch := strings.TrimPrefix(line, "* ")
				if branch == current || strings.HasPrefix(line, "* ") {
					skipped = append(skipped, branch)
					continue
				}

				// Skip branches that match common protected patterns
				if branch == "main" || branch == "master" || branch == "develop" || branch == "development" {
					skipped = append(skipped, branch)
					continue
				}

				if dryRun {
					deleted = append(deleted, branch)
				} else {
					// Delete the branch
					deleteFlag := "-d"
					if force {
						deleteFlag = "-D"
					}

					deleteCmd := exec.CommandContext(ctx, "git", "branch", deleteFlag, branch)
					if err := deleteCmd.Run(); err != nil {
						result.WriteString(fmt.Sprintf("Failed to delete branch '%s': %v\n", branch, err))
					} else {
						deleted = append(deleted, branch)
					}
				}
			}

			// Build result message
			if dryRun {
				result.WriteString("DRY RUN - branches that would be deleted:\n")
				for _, branch := range deleted {
					result.WriteString(fmt.Sprintf("  - %s\n", branch))
				}
			} else {
				result.WriteString(fmt.Sprintf("Successfully deleted %d branches:\n", len(deleted)))
				for _, branch := range deleted {
					result.WriteString(fmt.Sprintf("  - %s\n", branch))
				}
			}

			if len(skipped) > 0 {
				result.WriteString("\nSkipped branches:\n")
				for _, branch := range skipped {
					result.WriteString(fmt.Sprintf("  - %s (protected or current)\n", branch))
				}
			}

			if len(deleted) == 0 && !dryRun {
				result.WriteString("No branches to delete.")
			}

			return result.String(), nil
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
			if err := recordView(path); err != nil {
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
			if err := checkForOverwrite(path); err != nil {
				return "", err
			}
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				return "", err
			}
			_ = recordView(path)
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
			if err := checkForOverwrite(path); err != nil {
				return "", err
			}
			if err := os.WriteFile(path, []byte(text), 0644); err != nil {
				return "", err
			}
			_ = recordView(path)
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
	}, "agent": {
		Desc: "Send a message to another agent",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{"type": "string"},
				"input": map[string]any{"type": "string"},
			},
			"required": []string{"agent", "input"},
			"example": map[string]any{
				"agent": "Agent1",
				"input": "Hello, how are you?",
			},
		}, Exec: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["agent"].(string)
			input, _ := args["input"].(string)

			// Look for team interface in context
			t, ok := team.FromContext(ctx)
			if !ok {
				return "", errors.New("team not found in context")
			}
			return t.Call(ctx, name, input)
		},
	},
	"flow": {
		Desc: "Run a flow file",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"path": map[string]any{"type": "string"}},
			"required":   []string{"path"},
			"example":    map[string]any{"path": "examples/flows/research_task"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			p, _ := args["path"].(string)
			if p == "" {
				return "", errors.New("missing path")
			}
			p = absPath(p)
			f, err := flow.Load(p)
			if err != nil {
				return "", err
			}
			outs, err := flow.Run(ctx, f, DefaultRegistry(), nil)
			if err != nil {
				return "", err
			}
			return strings.Join(outs, "\n"), nil
		},
	},
	"team": {
		Desc: "Run a multi-agent conversation",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"n":     map[string]any{"type": "integer"},
				"topic": map[string]any{"type": "string"},
			},
			"required": []string{"n"},
			"example":  map[string]any{"n": 2, "topic": "discuss"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			n, _ := args["n"].(int)
			if n == 0 {
				if f, ok := args["n"].(float64); ok {
					n = int(f)
				}
			}
			topic, _ := args["topic"].(string)
			if n <= 0 {
				return "", errors.New("n must be > 0")
			}
			route := router.Rules{{Name: "mock", IfContains: []string{""}, Client: model.NewMock()}}
			ag := core.New(route, DefaultRegistry(), memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
			outs, err := converse.Run(ctx, ag, n, topic)
			if err != nil {
				return "", err
			}
			return strings.Join(outs, "\n"), nil
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
				return ExecSandbox(ctx, m.Command, sbox.Options{
					Engine:   m.Engine,
					Net:      m.Net,
					CPULimit: m.CPULimit,
					MemLimit: m.MemLimit,
				})
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
