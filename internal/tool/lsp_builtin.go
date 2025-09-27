package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/marcodenic/agentry/internal/lsp"
)

func lspDiagnosticsSpec() builtinSpec {
	return builtinSpec{
		Desc: "Run language diagnostics (Go: gopls check, TypeScript: tsc --noEmit, Python: pyright, Rust: cargo check, JavaScript: eslint)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"paths": map[string]any{
					"type":        "array",
					"items":       map[string]any{"type": "string"},
					"description": "Optional file paths or globs to check. If omitted, scans the workspace for supported languages.",
				},
				"timeout_ms": map[string]any{
					"type":        "integer",
					"description": "Optional timeout in milliseconds",
				},
			},
		},
		Exec: func(ctx context.Context, args map[string]any) (string, error) {
			// Collect file list
			files, err := expandPaths(args["paths"])
			if err != nil {
				return "", err
			}
			if len(files) == 0 {
				// Auto-discover supported files
				files, _ = discoverWorkspaceFiles()
			}
			if len(files) == 0 {
				return marshal(map[string]any{
					"ok":          true,
					"message":     "no supported files found",
					"languages":   lsp.Languages(),
					"diagnostics": []any{},
					"counts":      map[string]any{"files": 0, "errors": 0, "warnings": 0},
				})
			}

			// Optional timeout (not wired into lsp.Check yet; placeholder for future client-based impl)
			_ = parseTimeout(args["timeout_ms"]) // no-op for now

			out, runErr := lsp.Check(files)
			diags := lsp.ParseDiagnostics(out)
			// aggregate counts
			errCount := 0
			warnCount := 0
			fileSet := map[string]struct{}{}
			for _, d := range diags {
				fileSet[d.File] = struct{}{}
				if strings.EqualFold(d.Severity, "warning") {
					warnCount++
				} else {
					errCount++
				}
			}
			res := map[string]any{
				"ok":          runErr == nil,
				"error":       nil,
				"output":      out,
				"languages":   lsp.Languages(),
				"diagnostics": diags,
				"counts":      map[string]any{"files": len(fileSet), "errors": errCount, "warnings": warnCount},
			}
			if runErr != nil {
				res["error"] = runErr.Error()
			}
			return marshal(res)
		},
	}
}

func registerLSPBuiltins(reg *builtinRegistry) {
	reg.add("lsp_diagnostics", lspDiagnosticsSpec())
}

func expandPaths(v any) ([]string, error) {
	if v == nil {
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, errors.New("paths must be an array of strings")
	}
	var out []string
	for _, it := range arr {
		s, ok := it.(string)
		if !ok {
			return nil, errors.New("paths must be strings")
		}
		// Support simple globs
		if strings.ContainsAny(s, "*?[]") {
			matches, _ := filepath.Glob(s)
			out = append(out, matches...)
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func discoverWorkspaceFiles() ([]string, error) {
	// Basic discovery: common files under current directory
	patterns := []string{
		"**/*.go",
		"**/*.ts", "**/*.tsx",
		"**/*.js", "**/*.jsx", "**/*.mjs", "**/*.cjs",
		"**/*.py",
		"**/*.rs",
	}
	var files []string
	for _, pat := range patterns {
		// Expand ** by using doublestar-like behavior via filepath.Glob fallback
		// Since filepath.Glob doesn't support **, approximate with */*/ forms
		// Fallback: gather common depths
		expanded := approximateGlob(pat)
		for _, g := range expanded {
			matches, _ := filepath.Glob(g)
			files = append(files, matches...)
		}
	}
	return files, nil
}

func approximateGlob(pat string) []string {
	if !strings.Contains(pat, "**/") {
		return []string{pat}
	}
	// crude expansion depths 0-3
	base := strings.ReplaceAll(pat, "**/", "")
	parts := []string{
		base,
		filepath.Join("*", base),
		filepath.Join("*", "*", base),
		filepath.Join("*", "*", "*", base),
	}
	return parts
}

func parseTimeout(v any) time.Duration {
	if v == nil {
		return 0
	}
	switch t := v.(type) {
	case float64:
		return time.Duration(int64(t)) * time.Millisecond
	case int64:
		return time.Duration(t) * time.Millisecond
	default:
		return 0
	}
}

func marshal(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}
	return string(b), nil
}
