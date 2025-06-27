package tool

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

func init() {
	// Advanced file operation tools
	addFileOperationTools()
}

func addFileOperationTools() {
	// read_lines - Read specific lines from a file
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

	// edit_range - Replace a range of lines in a file atomically
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

	// insert_at - Insert lines at a specific position
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

	// search_replace - Search and replace text in a file with regex support
	builtinMap["search_replace"] = builtinSpec{
		Desc: "Search and replace text in a file with optional regex support",
		Schema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path to edit",
				},
				"search": map[string]any{
					"type":        "string",
					"description": "Text or regex pattern to search for",
				},
				"replace": map[string]any{
					"type":        "string",
					"description": "Replacement text (supports regex capture groups if using regex)",
				},
				"regex": map[string]any{
					"type":        "boolean",
					"description": "Whether to treat search as a regex pattern (default: false)",
					"default":     false,
				},
				"max_replacements": map[string]any{
					"type":        "integer",
					"description": "Maximum number of replacements to make (default: -1 for all)",
					"default":     -1,
				},
			},
			"required": []string{"path", "search", "replace"},
			"example": map[string]any{
				"path":    "src/main.go",
				"search":  "old_function",
				"replace": "new_function",
			},
		},
		Exec: searchReplaceExec,
	}

	// fileinfo - Get comprehensive file information
	builtinMap["fileinfo"] = builtinSpec{
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

	// view - Enhanced file viewing with line numbers and syntax awareness
	builtinMap["view"] = builtinSpec{
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

	// create - Create a new file with content
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

// readLinesExec implements the read_lines tool
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

	// Record that we've viewed this file
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
		"path":        path,
		"start_line":  int(startLine),
		"lines_read":  len(lines),
		"content":     strings.Join(lines, "\n"),
	}
	if hasEndLine {
		result["end_line"] = int(endLine)
	}
	
	jsonResult, _ := json.Marshal(result)
	return string(jsonResult), nil
}

// editRangeExec implements the edit_range tool
func editRangeExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}
	
	startLine, _ := args["start_line"].(float64)
	endLine, _ := args["end_line"].(float64)
	content, _ := args["content"].(string)
	
	if startLine < 1 || endLine < 1 {
		return "", errors.New("line numbers must be >= 1")
	}
	if startLine > endLine {
		return "", errors.New("start_line must be <= end_line")
	}

	path = absPath(path)
	
	// Check for overwrite protection
	if err := checkForOverwrite(path); err != nil {
		return "", err
	}

	// Read all lines from the file
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

	// Validate line numbers
	if int(startLine) > len(allLines) {
		return "", fmt.Errorf("start_line %d exceeds file length %d", int(startLine), len(allLines))
	}
	if int(endLine) > len(allLines) {
		return "", fmt.Errorf("end_line %d exceeds file length %d", int(endLine), len(allLines))
	}

	// Split new content into lines
	newLines := strings.Split(content, "\n")
	
	// Build new file content
	var result []string
	// Lines before the edit range
	result = append(result, allLines[:int(startLine)-1]...)
	// New content
	result = append(result, newLines...)
	// Lines after the edit range
	result = append(result, allLines[int(endLine):]...)
	
	// Write atomically using temp file
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
	
	// Atomic move
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	resultInfo := map[string]any{
		"path":            path,
		"start_line":      int(startLine),
		"end_line":        int(endLine),
		"lines_replaced":  int(endLine) - int(startLine) + 1,
		"lines_inserted":  len(newLines),
		"total_lines":     len(result),
	}
	
	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}

// insertAtExec implements the insert_at tool
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
	
	// Check for overwrite protection
	if err := checkForOverwrite(path); err != nil {
		return "", err
	}

	// Read all lines from the file
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

	// Validate line number
	if int(line) > len(allLines) {
		return "", fmt.Errorf("line %d exceeds file length %d", int(line), len(allLines))
	}

	// Split new content into lines
	newLines := strings.Split(content, "\n")
	
	// Build new file content
	var result []string
	// Lines before insertion point
	result = append(result, allLines[:int(line)]...)
	// New content
	result = append(result, newLines...)
	// Lines after insertion point
	result = append(result, allLines[int(line):]...)
	
	// Write atomically using temp file
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
	
	// Atomic move
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

// searchReplaceExec implements the search_replace tool
func searchReplaceExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}
	
	search, _ := args["search"].(string)
	replace, _ := args["replace"].(string)
	isRegex, _ := args["regex"].(bool)
	maxReplacements := -1
	if mr, ok := args["max_replacements"].(float64); ok {
		maxReplacements = int(mr)
	}

	path = absPath(path)
	
	// Check for overwrite protection
	if err := checkForOverwrite(path); err != nil {
		return "", err
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	originalContent := string(content)
	var newContent string
	var replacements int
	
	if isRegex {
		re, err := regexp.Compile(search)
		if err != nil {
			return "", fmt.Errorf("invalid regex pattern: %w", err)
		}
		
		if maxReplacements == -1 {
			newContent = re.ReplaceAllString(originalContent, replace)
			replacements = len(re.FindAllString(originalContent, -1))
		} else {
			matches := re.FindAllStringIndex(originalContent, -1)
			if len(matches) > maxReplacements {
				matches = matches[:maxReplacements]
			}
			replacements = len(matches)
			
			// Replace from end to beginning to maintain indices
			newContent = originalContent
			for i := len(matches) - 1; i >= 0; i-- {
				match := matches[i]
				matchText := originalContent[match[0]:match[1]]
				replacement := re.ReplaceAllString(matchText, replace)
				newContent = newContent[:match[0]] + replacement + newContent[match[1]:]
			}
		}
	} else {
		if maxReplacements == -1 {
			newContent = strings.ReplaceAll(originalContent, search, replace)
			replacements = strings.Count(originalContent, search)
		} else {
			newContent = originalContent
			for i := 0; i < maxReplacements; i++ {
				if idx := strings.Index(newContent, search); idx != -1 {
					newContent = newContent[:idx] + replace + newContent[idx+len(search):]
					replacements++
				} else {
					break
				}
			}
		}
	}
	
	if replacements == 0 {
		return fmt.Sprintf(`{"path": "%s", "replacements": 0, "message": "No matches found"}`, path), nil
	}
	
	// Write atomically using temp file
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(newContent), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Atomic move
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to move temp file: %w", err)
	}

	resultInfo := map[string]any{
		"path":          path,
		"search":        search,
		"replace":       replace,
		"regex":         isRegex,
		"replacements":  replacements,
		"original_size": len(originalContent),
		"new_size":      len(newContent),
	}
	
	jsonResult, _ := json.Marshal(resultInfo)
	return string(jsonResult), nil
}

// getFileInfoExec implements the fileinfo tool
func getFileInfoExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}

	path = absPath(path)
	
	// Get file stats
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("failed to stat file: %w", err)
	}
	
	// Read file to analyze content
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	
	// Count lines
	lines := strings.Count(string(content), "\n") + 1
	if len(content) == 0 {
		lines = 0
	}
	
	// Detect encoding
	encoding := "ASCII"
	if !utf8.Valid(content) {
		encoding = "Binary"
	} else if len(content) != len(string(content)) {
		encoding = "UTF-8"
	}
	
	// Get file extension
	ext := filepath.Ext(path)
	
	// Determine likely language/type
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

// viewFileExec implements the view tool
func viewFileExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}
	
	startLine := 1.0
	if sl, ok := args["start_line"].(float64); ok {
		startLine = sl
	}
	
	endLine, hasEndLine := args["end_line"].(float64)
	showLineNumbers := true
	if sln, ok := args["show_line_numbers"].(bool); ok {
		showLineNumbers = sln
	}
	
	maxLines := 1000.0
	if ml, ok := args["max_lines"].(float64); ok {
		maxLines = ml
	}

	path = absPath(path)
	
	// Record that we've viewed this file
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

// createFileExec implements the create tool
func createFileExec(ctx context.Context, args map[string]any) (string, error) {
	path, _ := args["path"].(string)
	if path == "" {
		return "", errors.New("missing path")
	}
	
	content, _ := args["content"].(string)
	overwrite, _ := args["overwrite"].(bool)

	path = absPath(path)
	
	// Check if file exists
	if _, err := os.Stat(path); err == nil && !overwrite {
		return "", fmt.Errorf("file %s already exists (use overwrite=true to replace)", path)
	}
	
	// Create directory if needed
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	
	// Write file atomically
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	
	// Atomic move
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
