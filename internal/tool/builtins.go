package tool

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/marcodenic/agentry/internal/patch"
	"github.com/marcodenic/agentry/internal/team"
	"github.com/marcodenic/agentry/pkg/sbox"
)

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
	"fetch": {
		Desc: "Download content from a URL",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{"url": map[string]any{"type": "string"}},
			"required":   []string{"url"},
			"example":    map[string]any{"url": "https://example.com"},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			url, _ := args["url"].(string)
			if url == "" {
				return "", errors.New("missing url")
			}
			// Cross-platform URL fetching
			if runtime.GOOS == "windows" {
				// Use PowerShell Invoke-WebRequest
				cmd := fmt.Sprintf("(Invoke-WebRequest -Uri '%s').Content", url)
				return ExecSandbox(ctx, cmd, sbox.Options{})
			} else {
				// Try curl first, fallback to wget if available
				result, err := ExecSandbox(ctx, "curl -s "+url, sbox.Options{})
				if err != nil {
					// Fallback to wget if curl is not available
					result, err = ExecSandbox(ctx, "wget -qO- "+url, sbox.Options{})
				}
				return result, err
			}
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
			port, _ := args["port"].(float64)
			cmd, _ := args["command"].(string)
			addr := net.JoinHostPort(host, fmt.Sprintf("%d", int(port)))
			conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
			if err != nil {
				return "", err
			}
			defer conn.Close()
			_, _ = conn.Write([]byte(cmd + "\n"))
			buf := make([]byte, 1024)
			n, _ := conn.Read(buf)
			return string(buf[:n]), nil
		},
	},
	"agent": {
		Desc: "Delegate to another agent",
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
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			name, _ := args["agent"].(string)
			input, _ := args["input"].(string)
			t, ok := team.FromContext(ctx)
			if !ok {
				return "", errors.New("team not found in context")
			}
			return t.Call(ctx, name, input)
		},
	},
}

func init() {
	// Add patch tool
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

	// Add OS-specific shell tools
	if runtime.GOOS == "windows" {
		builtinMap["powershell"] = builtinSpec{
			Desc: "Execute PowerShell commands on Windows",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "PowerShell command to execute (e.g., 'Get-ChildItem', 'Get-Content file.txt')",
					},
				},
				"required": []string{"command"},
				"example":  map[string]any{"command": "Get-ChildItem -Name '*.go'"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				cmd, _ := args["command"].(string)
				if cmd == "" {
					return "", errors.New("missing command")
				}
				return ExecSandbox(ctx, cmd, sbox.Options{})
			},
		}

		builtinMap["cmd"] = builtinSpec{
			Desc: "Execute cmd.exe commands on Windows",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Command prompt command to execute",
					},
				},
				"required": []string{"command"},
				"example":  map[string]any{"command": "dir *.go"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				cmd, _ := args["command"].(string)
				if cmd == "" {
					return "", errors.New("missing command")
				}
				// Execute using cmd.exe
				cmdLine := fmt.Sprintf("cmd /c %s", cmd)
				return ExecSandbox(ctx, cmdLine, sbox.Options{})
			},
		}
	} else {
		builtinMap["bash"] = builtinSpec{
			Desc: "Execute bash commands on Unix/Linux/macOS",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Bash command to execute (e.g., 'ls -la', 'cat file.txt')",
					},
				},
				"required": []string{"command"},
				"example":  map[string]any{"command": "ls -la *.go"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				cmd, _ := args["command"].(string)
				if cmd == "" {
					return "", errors.New("missing command")
				}
				return ExecSandbox(ctx, cmd, sbox.Options{})
			},
		}

		builtinMap["sh"] = builtinSpec{
			Desc: "Execute sh commands on Unix/Linux/macOS",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"command": map[string]any{
						"type":        "string",
						"description": "Shell command to execute",
					},
				},
				"required": []string{"command"},
				"example":  map[string]any{"command": "find . -name '*.go'"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				cmd, _ := args["command"].(string)
				if cmd == "" {
					return "", errors.New("missing command")
				}
				return ExecSandbox(ctx, cmd, sbox.Options{})
			},
		}
	}
}
