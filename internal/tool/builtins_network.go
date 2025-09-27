package tool

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"
)

// getNetworkBuiltins returns network-related builtin tools
func getNetworkBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"fetch": {
			Desc: "Download content from HTTP/HTTPS URLs (web pages, APIs, etc.). ONLY for web URLs - NEVER use for local files! Use 'view' tool for reading local files.",
			Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{"url": map[string]any{"type": "string", "description": "HTTP or HTTPS URL to fetch (must start with http:// or https://)"}},
				"required":   []string{"url"},
				"example":    map[string]any{"url": "https://api.github.com/repos/owner/repo"},
			},
			Exec: func(ctx context.Context, args map[string]any) (string, error) {
				url, _ := args["url"].(string)
				if url == "" {
					return "", errors.New("missing url")
				}

				// Validate that this is actually a URL and not a file path
				if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
					return "", fmt.Errorf("fetch tool requires HTTP/HTTPS URLs, got '%s'. Use 'view' tool for local files", url)
				}

				// Cross-platform URL fetching
				if runtime.GOOS == "windows" {
					// Use PowerShell Invoke-WebRequest
					cmd := fmt.Sprintf("(Invoke-WebRequest -Uri '%s').Content", url)
					return ExecDirect(ctx, cmd)
				} else {
					// Try curl first, fallback to wget if available
					result, err := ExecDirect(ctx, "curl -s "+url)
					if err != nil {
						// Fallback to wget if curl is not available
						result, err = ExecDirect(ctx, "wget -qO- "+url)
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
				p, _ := getIntArg(args, "port", 0)
				cmd, _ := args["command"].(string)
				addr := net.JoinHostPort(host, fmt.Sprintf("%d", p))
				conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
				if err != nil {
					return "", err
				}
				defer conn.Close()
				_, _ = conn.Write([]byte(cmd + "\n"))
				buf := make([]byte, 1024)
				n, _ := conn.Read(buf)
				return strings.TrimRight(string(buf[:n]), "\r\n"), nil
			},
		},
	}
}

func registerNetworkBuiltins(reg *builtinRegistry) {
	reg.addAll(getNetworkBuiltins())
}
