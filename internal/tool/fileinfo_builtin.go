package tool

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func fileInfoSpec() builtinSpec {
	return builtinSpec{
		Desc: "Get comprehensive information about a file (size, lines, encoding, etc.)",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to analyze",
				},
			},
			"required": []string{"path"},
			"example": map[string]any{
				"path": "src/main.go",
			},
		},
		Exec: getFileInfoExec,
	}
}

func registerFileInfoBuiltins(reg *builtinRegistry) {
	reg.add("fileinfo", fileInfoSpec())
}

func getFileInfoExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	path = absPath(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Count(string(content), "\n") + 1
	if len(content) == 0 {
		lines = 0
	}

	encoding := "ASCII"
	if !utf8.Valid(content) {
		encoding = "Binary"
	} else if len(content) != len(string(content)) {
		encoding = "UTF-8"
	}

	ext := filepath.Ext(path)
	fileType := "Unknown"
	switch ext {
	case ".go":
		fileType = "Go"
	case ".js", ".mjs":
		fileType = "JavaScript"
	case ".ts":
		fileType = "TypeScript"
	case ".py":
		fileType = "Python"
	case ".java":
		fileType = "Java"
	case ".c":
		fileType = "C"
	case ".cpp", ".cxx", ".cc":
		fileType = "C++"
	case ".h", ".hpp":
		fileType = "Header"
	case ".md":
		fileType = "Markdown"
	case ".json":
		fileType = "JSON"
	case ".yaml", ".yml":
		fileType = "YAML"
	case ".xml":
		fileType = "XML"
	case ".html":
		fileType = "HTML"
	case ".css":
		fileType = "CSS"
	case ".sh":
		fileType = "Shell Script"
	case ".ps1":
		fileType = "PowerShell"
	case ".txt":
		fileType = "Text"
	}

	resultInfo := map[string]any{
		"path":         path,
		"size_bytes":   info.Size(),
		"lines":        lines,
		"encoding":     encoding,
		"extension":    ext,
		"file_type":    fileType,
		"modified":     info.ModTime().Format("2006-01-02 15:04:05"),
		"permissions":  info.Mode().String(),
		"is_directory": info.IsDir(),
	}

	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}
