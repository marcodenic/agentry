package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	builtinMap["create"] = builtinSpec{
		Desc: "Create a new file with specified content",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to create",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Content to write to the file",
				},
				"overwrite": map[string]any{
					"type":        "boolean",
					"description": "Whether to overwrite if file exists (default: false)",
					"default":     false,
				},
			},
			"required": []string{"path", "content"},
			"example": map[string]any{
				"path":    "src/new_file.go",
				"content": "package main\n\nfunc main() {\n    // TODO: implement\n}",
			},
		},
		Exec: createFileExec,
	}
}

func createFileExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	content, _ := args["content"].(string)
	overwrite, _ := args["overwrite"].(bool)

	path = absPath(path)
	if _, err := os.Stat(path); err == nil && !overwrite {
		return "", fmt.Errorf("file %s already exists (use overwrite=true to replace)", path)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	resultInfo := map[string]any{
		"path":       path,
		"size_bytes": len(content),
		"lines":      strings.Count(content, "\n") + 1,
		"created":    true,
	}

	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}
