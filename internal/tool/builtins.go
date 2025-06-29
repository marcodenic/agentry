package tool

import (
	"context"
	"errors"
	"fmt"
	"net"
	"runtime"
	"strings"
	"time"

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
	"sysinfo": {
		Desc: "Get system information including CPU, memory, disk usage, OS details, and hardware specs",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
			"example":    map[string]any{},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			// Cross-platform system information gathering
			if runtime.GOOS == "windows" {
				// Use simpler PowerShell commands for Windows
				return ExecSandbox(ctx, "powershell -Command \"Get-ComputerInfo | Select-Object WindowsProductName, WindowsVersion, TotalPhysicalMemory\"", sbox.Options{})
			} else {
				// Use standard Unix commands for system info
				return ExecSandbox(ctx, "uname -a && free -h", sbox.Options{})
			}
		},
	},
	"team_status": {
		Desc: "Get the current status of all team agents",
		Schema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
			"required":   []string{},
			"example":    map[string]any{},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			_, ok := team.FromContext(ctx)
			if !ok {
				return "No team context available", nil
			}
			
			// Return basic team info for now
			// TODO: Integrate with orchestrator when context supports it
			return "Team coordination active - use other tools to manage agents", nil
		},
	},
	"send_message": {
		Desc: "Send a message to another team agent",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"to": map[string]any{
					"type":        "string",
					"description": "Name of the agent to send message to, or 'all' for broadcast",
				},
				"message": map[string]any{
					"type":        "string",
					"description": "The message content to send",
				},
				"type": map[string]any{
					"type":        "string",
					"description": "Message type: 'info', 'task', 'question', 'status'",
					"default":     "info",
				},
			},
			"required": []string{"to", "message"},
			"example": map[string]any{
				"to":      "coder",
				"message": "Please create a new file called test.txt",
				"type":    "task",
			},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			to, _ := args["to"].(string)
			message, _ := args["message"].(string)
			msgType, _ := args["type"].(string)
			if msgType == "" {
				msgType = "info"
			}
			
			if to == "" || message == "" {
				return "", errors.New("missing required parameters: to and message")
			}
			
			// For now, return a confirmation
			// TODO: Integrate with actual team messaging system
			return fmt.Sprintf("Message sent to %s: %s (type: %s)", to, message, msgType), nil
		},
	},
	"assign_task": {
		Desc: "Assign a specific task to a team agent",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{
					"type":        "string",
					"description": "Name of the agent to assign the task to",
				},
				"task": map[string]any{
					"type":        "string",
					"description": "Description of the task to assign",
				},
				"priority": map[string]any{
					"type":        "string",
					"description": "Task priority: 'low', 'normal', 'high', 'urgent'",
					"default":     "normal",
				},
			},
			"required": []string{"agent", "task"},
			"example": map[string]any{
				"agent":    "coder",
				"task":     "Create a new Python script to parse CSV files",
				"priority": "normal",
			},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			agent, _ := args["agent"].(string)
			task, _ := args["task"].(string)
			priority, _ := args["priority"].(string)
			if priority == "" {
				priority = "normal"
			}
			
			if agent == "" || task == "" {
				return "", errors.New("missing required parameters: agent and task")
			}
			
			// For now, return a confirmation
			// TODO: Integrate with actual task assignment system
			return fmt.Sprintf("Task assigned to %s (priority: %s): %s", agent, priority, task), nil
		},
	},
	"check_agent": {
		Desc: "Check the status and availability of a specific agent",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"agent": map[string]any{
					"type":        "string",
					"description": "Name of the agent to check",
				},
			},
			"required": []string{"agent"},
			"example": map[string]any{
				"agent": "coder",
			},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			agent, _ := args["agent"].(string)
			if agent == "" {
				return "", errors.New("missing required parameter: agent")
			}
			
			t, ok := team.FromContext(ctx)
			if !ok {
				return "No team context available", nil
			}
			
			// Try to call the agent to see if it exists
			_, err := t.Call(ctx, agent, "status check")
			if err != nil {
				return fmt.Sprintf("Agent '%s' not available or not found", agent), nil
			}
			
			return fmt.Sprintf("Agent '%s' is available", agent), nil
		},
	},
	"project_tree": {
		Desc: "Get intelligent project structure with smart filtering (ignores node_modules, .git, dist, etc.)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"depth": map[string]any{
					"type":        "integer",
					"description": "Maximum tree depth to show (default: 3)",
					"default":     3,
				},
				"path": map[string]any{
					"type":        "string", 
					"description": "Root path to analyze (default: current directory)",
					"default":     ".",
				},
				"show_files": map[string]any{
					"type":        "boolean",
					"description": "Include files in tree (default: true)",
					"default":     true,
				},
			},
			"required": []string{},
			"example": map[string]any{
				"depth":      3,
				"path":       ".",
				"show_files": true,
			},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			depth := 3
			if d, ok := args["depth"].(float64); ok {
				depth = int(d)
			}
			
			rootPath := "."
			if p, ok := args["path"].(string); ok && p != "" {
				rootPath = p
			}
			
			showFiles := true
			if sf, ok := args["show_files"].(bool); ok {
				showFiles = sf
			}
			
			// Use find command with smart filtering
			ignorePatterns := []string{
				"node_modules", ".git", "dist", "build", "target", "vendor", 
				".next", "__pycache__", ".pytest_cache", "coverage",
				"*.egg-info", ".venv", "venv", ".env", "tmp", ".tmp",
			}
			
			var cmd strings.Builder
			cmd.WriteString("find ")
			cmd.WriteString(rootPath)
			cmd.WriteString(" -type d \\( ")
			for i, pattern := range ignorePatterns {
				if i > 0 {
					cmd.WriteString(" -o ")
				}
				cmd.WriteString("-name '")
				cmd.WriteString(pattern)
				cmd.WriteString("'")
			}
			cmd.WriteString(" \\) -prune -o ")
			
			if showFiles {
				cmd.WriteString("-type f")
			} else {
				cmd.WriteString("-type d")
			}
			
			cmd.WriteString(" -print | head -50 | sort")
			
			result, err := ExecSandbox(ctx, cmd.String(), sbox.Options{})
			if err != nil {
				return "", fmt.Errorf("failed to get project tree: %w", err)
			}
			
			// Format the output nicely
			lines := strings.Split(strings.TrimSpace(result), "\n")
			var output strings.Builder
			output.WriteString("📂 Project Structure:\n")
			output.WriteString("==================\n")
			
			for _, line := range lines {
				if line == "" {
					continue
				}
				
				// Calculate indentation based on depth
				parts := strings.Split(line, "/")
				currentDepth := len(parts) - 1
				
				if currentDepth > depth {
					continue
				}
				
				indent := strings.Repeat("  ", currentDepth)
				filename := parts[len(parts)-1]
				
				// Add emoji based on file type
				if strings.Contains(line, ".") {
					// It's a file
					if strings.HasSuffix(filename, ".go") {
						output.WriteString(indent + "📄 " + filename + " (Go)\n")
					} else if strings.HasSuffix(filename, ".js") || strings.HasSuffix(filename, ".ts") {
						output.WriteString(indent + "📄 " + filename + " (JavaScript)\n")
					} else if strings.HasSuffix(filename, ".py") {
						output.WriteString(indent + "📄 " + filename + " (Python)\n")
					} else if strings.HasSuffix(filename, ".md") {
						output.WriteString(indent + "📖 " + filename + " (Markdown)\n")
					} else if strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".yaml") || strings.HasSuffix(filename, ".yml") {
						output.WriteString(indent + "⚙️ " + filename + " (Config)\n")
					} else {
						output.WriteString(indent + "📄 " + filename + "\n")
					}
				} else {
					// It's a directory
					output.WriteString(indent + "📁 " + filename + "/\n")
				}
			}
			
			return output.String(), nil
		},
	},
}
