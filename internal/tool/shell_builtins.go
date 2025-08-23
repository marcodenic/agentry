package tool

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/marcodenic/agentry/internal/patch"
)

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
				return ExecDirect(ctx, cmd)
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
				return ExecDirect(ctx, cmdLine)
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
				return ExecDirect(ctx, cmd)
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
				return ExecDirect(ctx, cmd)
			},
		}
	}
}
