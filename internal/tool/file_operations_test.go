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

	t.Run("create", func(t *testing.T) {
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
			t.Fatalf("create failed: %v", err)
		}

		var resultData map[string]any
		if err := json.Unmarshal([]byte(result), &resultData); err != nil {
			t.Fatalf("Failed to parse create result: %v", err)
		}

		if resultData["created"] != true {
			t.Error("Expected created=true")
		}

		// Verify file exists and has correct content
		if _, err := os.Stat(testFile); os.IsNotExist(err) {
			t.Error("File was not created")
		}

		actualContent, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		if strings.TrimSpace(string(actualContent)) != strings.TrimSpace(content) {
			t.Errorf("File content mismatch.\nExpected:\n%s\nGot:\n%s", content, string(actualContent))
		}
	})

	t.Run("view", func(t *testing.T) {
		args := map[string]any{
			"path": testFile,
		}

		result, err := viewFileExec(context.Background(), args)
		if err != nil {
			t.Fatalf("view failed: %v", err)
		}

		if !strings.Contains(result, "Hello, World!") {
			t.Error("view did not return expected content")
		}
	})

	t.Run("read_lines", func(t *testing.T) {
		args := map[string]any{
			"path":       testFile,
			"start_line": 2,
			"end_line":   4,
		}

		result, err := readLinesExec(context.Background(), args)
		if err != nil {
			t.Fatalf("read_lines failed: %v", err)
		}

		if !strings.Contains(result, "package main") {
			t.Error("read_lines did not return expected content")
		}
	})

	t.Run("fileinfo", func(t *testing.T) {
		args := map[string]any{
			"path": testFile,
		}

		result, err := getFileInfoExec(context.Background(), args)
		if err != nil {
			t.Fatalf("fileinfo failed: %v", err)
		}

		var info map[string]any
		if err := json.Unmarshal([]byte(result), &info); err != nil {
			t.Fatalf("Failed to parse fileinfo result: %v", err)
		}

		if info["size"] != nil && info["size"].(float64) == 0 {
			t.Error("Expected file size > 0")
		}
	})

	t.Run("search_replace", func(t *testing.T) {
		args := map[string]any{
			"path":    testFile,
			"search":  "Hello, World!",
			"replace": "Goodbye, World!",
		}

		result, err := searchReplaceExec(context.Background(), args)
		if err != nil {
			t.Fatalf("search_replace failed: %v", err)
		}

		// Verify the replacement worked
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after replacement: %v", err)
		}

		if !strings.Contains(string(content), "Goodbye, World!") {
			t.Error("search_replace did not work")
		}

		t.Logf("Search replace result: %s", result)
	})

	t.Run("edit_range", func(t *testing.T) {
		args := map[string]any{
			"path":       testFile,
			"start_line": 4,
			"end_line":   5,
			"content":    "func main() {\n    fmt.Println(\"Modified content!\")",
		}

		result, err := editRangeExec(context.Background(), args)
		if err != nil {
			t.Fatalf("edit_range failed: %v", err)
		}

		// Verify the edit worked
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after edit: %v", err)
		}

		if !strings.Contains(string(content), "Modified content!") {
			t.Error("edit_range did not work")
		}

		t.Logf("Edit range result: %s", result)
	})

	t.Run("insert_at", func(t *testing.T) {
		args := map[string]any{
			"path":    testFile,
			"line":    2,
			"content": "// This is a comment",
		}

		result, err := insertAtExec(context.Background(), args)
		if err != nil {
			t.Fatalf("insert_at failed: %v", err)
		}

		// Verify the insert worked
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read file after insert: %v", err)
		}

		if !strings.Contains(string(content), "// This is a comment") {
			t.Error("insert_at did not work")
		}

		t.Logf("Insert at result: %s", result)
	})
}
