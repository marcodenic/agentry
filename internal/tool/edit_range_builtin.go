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
	builtinMap["edit_range"] = builtinSpec{
		Desc: "Replace a range of lines in a file with new content atomically",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to edit",
				},
				"start_line": map[string]any{
					"type":        "integer",
					"description": "Starting line number to replace (1-based, inclusive)",
					"minimum":     1,
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Ending line number to replace (1-based, inclusive)",
					"minimum":     1,
				},
				"content": map[string]any{
					"type":        "string",
					"description": "New content to replace the range (without trailing newline)",
				},
			},
			"required": []string{"path", "start_line", "end_line", "content"},
			"example": map[string]any{
				"path":       "src/main.go",
				"start_line": 10,
				"end_line":   12,
				"content":    "// New implementation\nfunc main() {\n    fmt.Println(\"Hello World\")",
			},
		},
		Exec: editRangeExec,
	}
}

func editRangeExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	startLine, _ := getIntArg(args, "start_line", 0)
	endLine, _ := getIntArg(args, "end_line", 0)
	content, _ := args["content"].(string)

	if startLine < 1 || endLine < 1 {
		return "", errors.New("line numbers must be >= 1")
	}
	if startLine > endLine {
		return "", errors.New("start_line must be <= end_line")
	}

	path = absPath(path)
	if err := checkForOverwrite(path); err != nil {
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}

	var allLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	file.Close()
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	if int(startLine) > len(allLines) {
		return "", fmt.Errorf("start_line %d exceeds file length %d", int(startLine), len(allLines))
	}
	if int(endLine) > len(allLines) {
		return "", fmt.Errorf("end_line %d exceeds file length %d", int(endLine), len(allLines))
	}

	newLines := strings.Split(content, "\n")
	var result []string
	result = append(result, allLines[:int(startLine)-1]...)
	result = append(result, newLines...)
	result = append(result, allLines[int(endLine):]...)

	tempPath := path + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	for i, line := range result {
		if i > 0 {
			if _, err := tempFile.WriteString("\n"); err != nil {
				tempFile.Close()
				os.Remove(tempPath)
				return "", fmt.Errorf("failed to write to temp file: %w", err)
			}
		}
		if _, err := tempFile.WriteString(line); err != nil {
			tempFile.Close()
			os.Remove(tempPath)
			return "", fmt.Errorf("failed to write to temp file: %w", err)
		}
	}

	if err := tempFile.Close(); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	// Update view record to the new modtime to avoid false "changed since viewed" on follow-ups
	_ = recordView(path)

	resultInfo := map[string]any{
		"path":           path,
		"start_line":     int(startLine),
		"end_line":       int(endLine),
		"lines_replaced": int(endLine) - int(startLine) + 1,
		"lines_inserted": len(newLines),
		"total_lines":    len(result),
	}

	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}
