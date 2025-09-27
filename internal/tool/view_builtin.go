package tool

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

func viewSpec() builtinSpec {
	return builtinSpec{
		Desc: "View file contents with line numbers and optional syntax highlighting information",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to view",
				},
				"start_line": map[string]any{
					"type":        "integer",
					"description": "Starting line number (1-based, inclusive). Default: 1",
					"minimum":     1,
					"default":     1,
				},
				"end_line": map[string]any{
					"type":        "integer",
					"description": "Ending line number (1-based, inclusive). If omitted, shows entire file",
					"minimum":     1,
				},
				"show_line_numbers": map[string]any{
					"type":        "boolean",
					"description": "Whether to show line numbers (default: true)",
					"default":     true,
				},
				"max_lines": map[string]any{
					"type":        "integer",
					"description": "Maximum number of lines to show (default: 1000)",
					"minimum":     1,
					"default":     1000,
				},
			},
			"required": []string{"path"},
			"example": map[string]any{
				"path":              "src/main.go",
				"start_line":        1,
				"end_line":          50,
				"show_line_numbers": true,
			},
		},
		Exec: viewFileExec,
	}
}

func registerViewBuiltins(reg *builtinRegistry) {
	reg.add("view", viewSpec())
}

func viewFileExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	startLine, _ := getIntArg(args, "start_line", 1)

	endLine, hasEndLine := getIntArg(args, "end_line", 0)
	showLineNumbers := true
	if sln, ok := args["show_line_numbers"].(bool); ok {
		showLineNumbers = sln
	}

	maxLines, _ := getIntArg(args, "max_lines", 1000)

	path = absPath(path)
	if err := recordView(path); err != nil {
		return "", fmt.Errorf("failed to record view: %w", err)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lines []string
	currentLine := 1
	linesRead := 0

	for scanner.Scan() && linesRead < int(maxLines) {
		if currentLine >= int(startLine) {
			if hasEndLine && currentLine > int(endLine) {
				break
			}
			lineText := scanner.Text()
			if showLineNumbers {
				lineText = fmt.Sprintf("%4d: %s", currentLine, lineText)
			}
			lines = append(lines, lineText)
			linesRead++
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	return strings.Join(lines, "\n"), nil
}
