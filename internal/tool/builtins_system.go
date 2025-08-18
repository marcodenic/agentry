package tool

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/marcodenic/agentry/internal/sbox"
)

// getSystemBuiltins returns system information builtin tools
func getSystemBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
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
}
