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
	builtinMap["insert_at"] = builtinSpec{
		Desc: "Insert new lines at a specific line position in a file",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to edit",
				},
				"line": map[string]any{
					"type":        "integer",
					"description": "Line number to insert after (1-based). Use 0 to insert at beginning",
					"minimum":     0,
				},
				"content": map[string]any{
					"type":        "string",
					"description": "Content to insert (without trailing newline)",
				},
			},
			"required": []string{"path", "line", "content"},
			"example": map[string]any{
				"path":    "src/main.go",
				"line":    5,
				"content": "import \"fmt\"",
			},
		},
		Exec: insertAtExec,
	}
}

func insertAtExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	line, _ := args["line"].(float64)
	content, _ := args["content"].(string)

	if line < 0 {
		return "", errors.New("line number must be >= 0")
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

	if int(line) > len(allLines) {
		return "", fmt.Errorf("line %d exceeds file length %d", int(line), len(allLines))
	}

	newLines := strings.Split(content, "\n")
	var result []string
	result = append(result, allLines[:int(line)]...)
	result = append(result, newLines...)
	result = append(result, allLines[int(line):]...)

	tempPath := path + ".tmp"
	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	for i, lineText := range result {
		if i > 0 {
			if _, err := tempFile.WriteString("\n"); err != nil {
				tempFile.Close()
				os.Remove(tempPath)
				return "", fmt.Errorf("failed to write to temp file: %w", err)
			}
		}
		if _, err := tempFile.WriteString(lineText); err != nil {
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

	resultInfo := map[string]any{
		"path":           path,
		"insert_after":   int(line),
		"lines_inserted": len(newLines),
		"total_lines":    len(result),
	}

	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}
