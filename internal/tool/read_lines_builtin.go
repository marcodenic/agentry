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

	// Accept int or float64 for start_line
	startLine, _ := getIntArg(args, "start_line", 1)
	if startLine < 1 {
		return "", errors.New("start_line must be >= 1")
	}

	endLineInt, hasEndLine := getIntArg(args, "end_line", 0)
	maxLinesInt, _ := getIntArg(args, "max_lines", 1000)

	path = absPath(path)
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	if err := recordView(path); err != nil {
		return "", fmt.Errorf("failed to record view: %w", err)
	}

	// Read all lines first for easier slicing and to allow optional context fields
	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	// Compute slice bounds
	startIdx := startLine - 1
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := len(allLines)
	if hasEndLine && endLineInt < endIdx {
		endIdx = endLineInt
	}
	if startIdx > endIdx {
		startIdx = endIdx
	}

	// Apply max lines limit
	limit := startIdx + maxLinesInt
	if limit < endIdx {
		endIdx = limit
	}

	selected := allLines[startIdx:endIdx]

	result := map[string]any{
		"path":       path,
		"start_line": int(startLine),
		"lines_read": len(selected),
		"content":    strings.Join(selected, "\n"),
	}
	if hasEndLine {
		result["end_line"] = int(endLineInt)
	}
	if len(allLines) > 0 {
		result["file_header"] = allLines[0]
	}

	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}
