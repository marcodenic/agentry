package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileOperationTools(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "agentry_file_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.go")
	
	t.Run("create_file", func(t *testing.T) {
		content := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`

		args := map[string]any{
			"path":    testFile,
			"content": content,
		}

		result, err := createFileExec(context.Background(), args)
		if err != nil {
			t.Fatalf("create_file failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse create_file result: %v", err)
		}

		if resultData["created"] != true {
			t.Error("Expected created=true")
		}

		// Verify file exists and has correct content
		actualContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		if string(actualContent) != content {
			t.Error("File content doesn't match expected content")
		}
	})

	t.Run("view_file", func(t *testing.T) {
		args := map[string]any{
			"path":              testFile,
			"show_line_numbers": true,
		}

		result, err := viewFileExec(context.Background(), args)
		if err != nil {
			t.Fatalf("view_file failed: %v", err)
		}

		lines := strings.Split(result, "\n")
		if len(lines) < 6 {
			t.Errorf("Expected at least 6 lines, got %d", len(lines))
		}

		// Check that line numbers are present
		if !strings.HasPrefix(lines[0], "   1: ") {
			t.Error("Expected line numbers in output")
		}
	})

	t.Run("read_lines", func(t *testing.T) {
		args := map[string]any{
			"path":       testFile,
			"start_line": 3.0,
			"end_line":   4.0,
		}

		result, err := readLinesExec(context.Background(), args)
		if err != nil {
			t.Fatalf("read_lines failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse read_lines result: %v", err)
		}

		content := resultData["content"].(string)
		expectedLines := []string{`import "fmt"`, ""}
		if content != strings.Join(expectedLines, "\n") {
			t.Errorf("Expected lines 3-4, got: %s", content)
		}
	})

	t.Run("get_file_info", func(t *testing.T) {
		args := map[string]any{
			"path": testFile,
		}

		result, err := getFileInfoExec(context.Background(), args)
		if err != nil {
			t.Fatalf("get_file_info failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse get_file_info result: %v", err)
		}

		if resultData["file_type"] != "Go" {
			t.Errorf("Expected file_type=Go, got %v", resultData["file_type"])
		}

		if resultData["lines"].(float64) >= 6.0 {
			// File should have at least 6 lines
		} else {
			t.Errorf("Expected at least 6 lines, got %v", resultData["lines"])
		}
	})

	t.Run("search_replace", func(t *testing.T) {
		args := map[string]any{
			"path":    testFile,
			"search":  "Hello, World!",
			"replace": "Hello, Agentry!",
		}

		result, err := searchReplaceExec(context.Background(), args)
		if err != nil {
			t.Fatalf("search_replace failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse search_replace result: %v", err)
		}

		if resultData["replacements"] != 1.0 {
			t.Errorf("Expected 1 replacement, got %v", resultData["replacements"])
		}

		// Verify the replacement was made
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after replacement: %v", err)
		}

		if !strings.Contains(string(content), "Hello, Agentry!") {
			t.Error("Replacement was not made")
		}
	})

	t.Run("edit_range", func(t *testing.T) {
		// First view the file to update the viewed files map
		viewArgs := map[string]any{
			"path": testFile,
		}
		_, err := viewFileExec(context.Background(), viewArgs)
		if err != nil {
			t.Fatalf("Failed to view file before edit: %v", err)
		}

		// Edit lines 5-6 to change the main function
		args := map[string]any{
			"path":       testFile,
			"start_line": 5.0,
			"end_line":   6.0,
			"content":    "func main() {\n    fmt.Println(\"Hello, Agentry!\")\n    fmt.Println(\"File operations work!\")\n}",
		}

		result, err := editRangeExec(context.Background(), args)
		if err != nil {
			t.Fatalf("edit_range failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse edit_range result: %v", err)
		}

		if resultData["lines_replaced"] != 2.0 {
			t.Errorf("Expected 2 lines replaced, got %v", resultData["lines_replaced"])
		}

		// Verify the edit was made
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after edit: %v", err)
		}

		if !strings.Contains(string(content), "File operations work!") {
			t.Error("Edit was not made correctly")
		}
	})

	t.Run("insert_at", func(t *testing.T) {
		// First view the file to update the viewed files map
		viewArgs := map[string]any{
			"path": testFile,
		}
		_, err := viewFileExec(context.Background(), viewArgs)
		if err != nil {
			t.Fatalf("Failed to view file before insert: %v", err)
		}

		// Insert a comment at line 4 (after the import)
		args := map[string]any{
			"path":    testFile,
			"line":    4.0,
			"content": "// This is a test program",
		}

		result, err := insertAtExec(context.Background(), args)
		if err != nil {
			t.Fatalf("insert_at failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse insert_at result: %v", err)
		}

		if resultData["lines_inserted"] != 1.0 {
			t.Errorf("Expected 1 line inserted, got %v", resultData["lines_inserted"])
		}

		// Verify the insertion was made
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after insertion: %v", err)
		}

		if !strings.Contains(string(content), "// This is a test program") {
			t.Error("Insertion was not made correctly")
		}
	})

	t.Run("regex_search_replace", func(t *testing.T) {
		// First view the file to update the viewed files map
		viewArgs := map[string]any{
			"path": testFile,
		}
		_, err := viewFileExec(context.Background(), viewArgs)
		if err != nil {
			t.Fatalf("Failed to view file before regex replace: %v", err)
		}

		// Replace fmt.Println calls with log.Println using regex
		args := map[string]any{
			"path":    testFile,
			"search":  `fmt\.Println\(([^)]+)\)`,
			"replace": `log.Println($1)`,
			"regex":   true,
		}

		result, err := searchReplaceExec(context.Background(), args)
		if err != nil {
			t.Fatalf("regex search_replace failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse regex search_replace result: %v", err)
		}

		// Should replace 2 fmt.Println calls
		if resultData["replacements"] != 2.0 {
			t.Errorf("Expected 2 replacements, got %v", resultData["replacements"])
		}

		// Verify the replacement was made
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after regex replacement: %v", err)
		}

		if strings.Contains(string(content), "fmt.Println") {
			t.Error("fmt.Println should have been replaced")
		}

		if !strings.Contains(string(content), "log.Println") {
			t.Error("log.Println should be present after replacement")
		}
	})
}

func TestFileOperationToolsInRegistry(t *testing.T) {
	registry := DefaultRegistry()
	
	expectedTools := []string{
		"read_lines",
		"edit_range", 
		"insert_at",
		"search_replace",
		"get_file_info",
		"view_file",
		"create_file",
	}
	
	for _, toolName := range expectedTools {
		tool, exists := registry.Use(toolName)
		if !exists {
			t.Errorf("Tool %s not found in registry", toolName)
			continue
		}
		
		if tool.Name() != toolName {
			t.Errorf("Tool name mismatch: expected %s, got %s", toolName, tool.Name())
		}
		
		schema := tool.JSONSchema()
		if schema == nil {
			t.Errorf("Tool %s has no schema", toolName)
		}
	}
}
