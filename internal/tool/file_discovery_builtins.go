package tool

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcodenic/agentry/internal/sbox"
)

// getFileDiscoveryBuiltins returns file discovery builtin tools
func getFileDiscoveryBuiltins() map[string]builtinSpec {
	return map[string]builtinSpec{
		"find": {
			Desc: "Find files and directories by name pattern",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Starting path to search (default: current directory)",
						"default":     ".",
					},
					"name": map[string]any{
						"type":        "string",
						"description": "File name pattern to search for (supports wildcards)",
					},
					"type": map[string]any{
						"type":        "string",
						"description": "Type of item to find: 'file', 'directory', or 'all'",
						"enum":        []string{"file", "directory", "all"},
						"default":     "all",
					},
					"max_depth": map[string]any{
						"type":        "integer",
						"description": "Maximum depth to search (default: unlimited)",
					},
				},
				"required": []string{"name"},
				"example": map[string]any{
					"name": "*.go",
					"type": "file",
				},
			},
			Exec: findFileExec,
		},
		"grep": {
			Desc: "Search for text patterns in files",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pattern": map[string]any{
						"type":        "string",
						"description": "Text pattern to search for",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "File or directory to search in (default: current directory)",
						"default":     ".",
					},
					"recursive": map[string]any{
						"type":        "boolean",
						"description": "Search recursively in subdirectories",
						"default":     false,
					},
					"file_pattern": map[string]any{
						"type":        "string",
						"description": "File name pattern to include (e.g., '*.go')",
					},
					"ignore_case": map[string]any{
						"type":        "boolean",
						"description": "Ignore case when matching",
						"default":     false,
					},
					"line_numbers": map[string]any{
						"type":        "boolean",
						"description": "Show line numbers in results",
						"default":     true,
					},
				},
				"required": []string{"pattern"},
				"example": map[string]any{
					"pattern":      "func main",
					"path":         ".",
					"recursive":    true,
					"file_pattern": "*.go",
				},
			},
			Exec: grepFileExec,
		},
		"ls": {
			Desc: "List directory contents",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "Directory path to list (default: current directory)",
						"default":     ".",
					},
					"long": map[string]any{
						"type":        "boolean",
						"description": "Show detailed file information",
						"default":     false,
					},
					"all": map[string]any{
						"type":        "boolean",
						"description": "Show hidden files",
						"default":     false,
					},
					"sort_by": map[string]any{
						"type":        "string",
						"description": "Sort by: 'name', 'size', 'modified'",
						"enum":        []string{"name", "size", "modified"},
						"default":     "name",
					},
				},
				"required": []string{},
				"example": map[string]any{
					"path": ".",
					"long": true,
				},
			},
			Exec: listDirExec,
		},
		"glob": {
			Desc: "Find files using glob patterns",
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"pattern": map[string]any{
						"type":        "string",
						"description": "Glob pattern to match files",
					},
					"path": map[string]any{
						"type":        "string",
						"description": "Base directory to search from (default: current directory)",
						"default":     ".",
					},
				},
				"required": []string{"pattern"},
				"example": map[string]any{
					"pattern": "**/*.go",
				},
			},
			Exec: globFileExec,
		},
	}
}

func findFileExec(ctx context.Context, args map[string]any) (string, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("missing required parameter: name")
	}

	basePath := "."
	if p, ok := args["path"].(string); ok && p != "" {
		basePath = p
	}

	fileType := "all"
	if t, ok := args["type"].(string); ok {
		fileType = t
	}

	maxDepth := -1
	if d, ok := args["max_depth"].(float64); ok {
		maxDepth = int(d)
	}

	var results []string
	err := filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		// Check depth limit
		if maxDepth >= 0 {
			rel, _ := filepath.Rel(basePath, path)
			if strings.Count(rel, string(os.PathSeparator)) > maxDepth {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check type filter
		if fileType == "file" && info.IsDir() {
			return nil
		}
		if fileType == "directory" && !info.IsDir() {
			return nil
		}

		// Check name pattern
		matched, _ := filepath.Match(name, filepath.Base(path))
		if matched {
			results = append(results, path)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("error walking directory: %w", err)
	}

	if len(results) == 0 {
		return "No files found matching pattern", nil
	}

	return strings.Join(results, "\n"), nil
}

func grepFileExec(ctx context.Context, args map[string]any) (string, error) {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return "", fmt.Errorf("missing required parameter: pattern")
	}

	// Determine target path or file (support alias 'file')
	target := "."
	if p, ok := args["file"].(string); ok && p != "" {
		target = p
	} else if p, ok := args["path"].(string); ok && p != "" {
		target = p
	}

	// Use shell grep for full functionality; build with safe quoting
	sq := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }

	cmd := "grep"
	if recursive, ok := args["recursive"].(bool); ok && recursive {
		cmd += " -r"
	}
	if ignoreCase, ok := args["ignore_case"].(bool); ok && ignoreCase {
		cmd += " -i"
	}
	if lineNumbers, ok := args["line_numbers"].(bool); ok && lineNumbers {
		cmd += " -n"
	}
	if filePattern, ok := args["file_pattern"].(string); ok && filePattern != "" {
		cmd += " --include=" + sq(filePattern)
	}

	cmd += " -- " + sq(pattern) + " " + sq(target)
	return ExecSandbox(ctx, cmd, sbox.Options{})
}

func listDirExec(ctx context.Context, args map[string]any) (string, error) {
	path := "."
	if p, ok := args["path"].(string); ok && p != "" {
		path = p
	}

	cmd := "ls"

	if long, ok := args["long"].(bool); ok && long {
		cmd += " -l"
	}

	if all, ok := args["all"].(bool); ok && all {
		cmd += " -a"
	}

	if sortBy, ok := args["sort_by"].(string); ok {
		switch sortBy {
		case "size":
			cmd += " -S"
		case "modified":
			cmd += " -t"
		}
	}

	cmd += " " + path

	return ExecSandbox(ctx, cmd, sbox.Options{})
}

func globFileExec(ctx context.Context, args map[string]any) (string, error) {
	pattern, ok := args["pattern"].(string)
	if !ok || pattern == "" {
		return "", fmt.Errorf("missing required parameter: pattern")
	}

	basePath := "."
	if p, ok := args["path"].(string); ok && p != "" {
		basePath = p
	}

	// Change to base directory for glob matching
	origDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	if err := os.Chdir(basePath); err != nil {
		return "", fmt.Errorf("failed to change directory: %w", err)
	}
	defer os.Chdir(origDir)

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("failed to match pattern: %w", err)
	}

	if len(matches) == 0 {
		return "No files found matching pattern", nil
	}

	return strings.Join(matches, "\n"), nil
}
