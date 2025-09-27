package tool

import (
	"context"
	"errors"
	"fmt"
	"runtime"

	"github.com/marcodenic/agentry/internal/patch"
)

func patchSpec() builtinSpec {
	return builtinSpec{
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

func windowsShellSpec(name, desc, example string) builtinSpec {
	return builtinSpec{
		Desc: desc,
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "Command to execute",
				},
			},
			"required": []string{"command"},
			"example":  map[string]any{"command": example},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			cmd, _ := args["command"].(string)
			if cmd == "" {
				return "", errors.New("missing command")
			}
			if name == "cmd" {
				return ExecDirect(ctx, fmt.Sprintf("cmd /c %s", cmd))
			}
			return ExecDirect(ctx, cmd)
		},
	}
}

func unixShellSpec(desc string) builtinSpec {
	return builtinSpec{
		Desc: desc,
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "Command to execute",
				},
			},
			"required": []string{"command"},
			"example":  map[string]any{"command": "ls -la"},
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

func registerShellBuiltins(reg *builtinRegistry) {
	reg.add("patch", patchSpec())
	if runtime.GOOS == "windows" {
		reg.add("powershell", windowsShellSpec("powershell", "Execute PowerShell commands on Windows", "Get-ChildItem -Name '*.go'"))
		reg.add("cmd", windowsShellSpec("cmd", "Execute cmd.exe commands on Windows", "dir *.go"))
		return
	}
	reg.add("bash", unixShellSpec("Execute bash commands on Unix/Linux/macOS"))
	reg.add("sh", unixShellSpec("Execute sh commands on Unix/Linux/macOS"))
}
