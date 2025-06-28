package tool

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

func init() {
	builtinMap["read_lines"] = builtinSpec{
		Desc: "Read specific lines from a file with line-precise access",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to read",
				},
				"start_line": map[string]any{
					"type":        "integer",
					"description": "Starting line number (1-based, inclusive)",
					"minimum":     1,
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Ending line number (1-based, inclusive). If omitted, reads to end of file",
					"minimum":     1,
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": "Maximum number of lines to read (default: 1000)",
					"minimum":     1,
					"default":     1000,
				},
			},
			"required": []string{"path", "start_line"},
			"example": map[string]any{
				"path":       "src/main.go",
				"start_line": 10,
				"end_line":   20,
			},
		},
		Exec: readLinesExec,
	}
}

func readLinesExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	startLine, _ := args["start_line"].(float64)
	if startLine < 1 {
		return "", errors.New("start_line must be >= 1")
	}

	endLine, hasEndLine := args["end_line"].(float64)
	maxLines := 1000.0
	if ml, ok := args["max_lines"].(float64); ok {
		maxLines = ml
	}

	path = absPath(path)
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if err := recordView(path); err != nil {
		return "", fmt.Errorf("failed to record view: %w", err)
	}

	scanner := bufio.NewScanner(file)
	lines := []string{}
	currentLine := 1
	linesRead := 0

	for scanner.Scan() && linesRead < int(maxLines) {
		if currentLine >= int(startLine) {
			if hasEndLine && currentLine > int(endLine) {
				break
			}
			lines = append(lines, scanner.Text())
			linesRead++
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	result := map[string]any{
		"path":       path,
		"start_line": int(startLine),
		"lines_read": len(lines),
		"content":    strings.Join(lines, "\n"),
	}
	if hasEndLine {
		result["end_line"] = int(endLine)
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
