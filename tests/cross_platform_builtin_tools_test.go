package tests

import (
	"context"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestCrossPlatformBuiltinTools(t *testing.T) {
	// Create a context
	ctx := context.Background()

	// No sandboxing setup needed - it's always disabled now

	// Create a temporary working directory for testing
	tempDir, err := os.MkdirTemp("", "agentry-crossplatform-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory for tests
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Initialize tool registry
	reg := tool.DefaultRegistry()
	// Test the write tool - create a test file
	t.Run("write_tool", func(t *testing.T) {
		writeTool, exists := reg.Use("write")
		if !exists {
			t.Skip("write tool not available in registry")
		}

		writeArgs := map[string]any{
			"file":    "test.txt",
			"content": "Hello, cross-platform testing!",
		}

		result, err := writeTool.Execute(ctx, writeArgs)
		if err != nil {
			t.Fatalf("write tool failed: %v", err)
		}
		t.Logf("write result: %s", result)

		// Verify file was created
		if _, err := os.Stat("test.txt"); os.IsNotExist(err) {
			t.Error("write tool did not create the file")
		}
	})

	// Test the view tool - read the file we just created
	t.Run("view_tool", func(t *testing.T) {
		viewTool, exists := reg.Use("view")
		if !exists {
			t.Skip("view tool not available in registry")
		}

		viewArgs := map[string]any{
			"path": "test.txt",
		}

		result, err := viewTool.Execute(ctx, viewArgs)
		if err != nil {
			t.Fatalf("view tool failed: %v", err)
		}
		t.Logf("view result: %s", result)

		// Should contain the content we wrote
		if !strings.Contains(result, "Hello, cross-platform testing!") {
			t.Errorf("view tool did not return expected content, got: %s", result)
		}
	})

	// Test the ls tool - list directory contents
	t.Run("ls_tool", func(t *testing.T) {
		lsTool, exists := reg.Use("ls")
		if !exists {
			t.Skip("ls tool not available in registry")
		}

		lsArgs := map[string]any{
			"path": ".",
		}

		result, err := lsTool.Execute(ctx, lsArgs)
		if err != nil {
			t.Fatalf("ls tool failed: %v", err)
		}
		t.Logf("ls result: %s", result)

		// Should contain our test.txt file
		if !strings.Contains(result, "test.txt") {
			t.Errorf("ls tool did not show test.txt file, got: %s", result)
		}
	})

	// Test the bash tool with a simple command
	t.Run("bash_tool", func(t *testing.T) {
		bashTool, exists := reg.Use("bash")
		if !exists {
			t.Skip("bash tool not available in registry")
		}

		var bashArgs map[string]any
		var expectedContent string

		if runtime.GOOS == "windows" {
			// On Windows, test PowerShell command
			bashArgs = map[string]any{
				"command": "Get-Date -Format 'yyyy-MM-dd'",
			}
			expectedContent = "2" // Should contain year starting with 2
		} else {
			// On Unix, test bash command
			bashArgs = map[string]any{
				"command": "echo 'bash test'",
			}
			expectedContent = "bash test"
		}

		result, err := bashTool.Execute(ctx, bashArgs)
		if err != nil {
			t.Fatalf("bash tool failed: %v", err)
		}
		t.Logf("bash result: %s", result)

		if !strings.Contains(result, expectedContent) {
			t.Errorf("bash tool did not return expected content, got: %s", result)
		}
	})

	// Test the edit tool - modify the existing file
	t.Run("edit_tool", func(t *testing.T) {
		editTool, exists := reg.Use("edit")
		if !exists {
			t.Skip("edit tool not available in registry")
		}

		editArgs := map[string]any{
			"file":    "test.txt",
			"content": "Updated cross-platform content!",
		}

		result, err := editTool.Execute(ctx, editArgs)
		if err != nil {
			t.Fatalf("edit tool failed: %v", err)
		}
		t.Logf("edit result: %s", result)

		// Now view the file to check if it was updated
		viewTool, exists := reg.Use("view")
		if !exists {
			t.Skip("view tool not available for verification")
		}

		viewArgs := map[string]any{
			"path": "test.txt",
		}

		viewResult, err := viewTool.Execute(ctx, viewArgs)
		if err != nil {
			t.Fatalf("view tool failed after edit: %v", err)
		}

		if !strings.Contains(viewResult, "Updated cross-platform content!") {
			t.Errorf("edit tool did not update the file content, got: %s", viewResult)
		}
	})
	// Test the glob tool - find text files
	t.Run("glob_tool", func(t *testing.T) {
		globTool, exists := reg.Use("glob")
		if !exists {
			t.Skip("glob tool not available in registry")
		}

		// First, let's see what directory we're in and what files exist
		lsTool, _ := reg.Use("ls")
		lsResult, _ := lsTool.Execute(ctx, map[string]any{"path": "."})
		t.Logf("Current directory contents: %s", lsResult)

		globArgs := map[string]any{
			"pattern": "*.txt",
		}

		result, err := globTool.Execute(ctx, globArgs)
		if err != nil {
			t.Fatalf("glob tool failed: %v", err)
		}
		t.Logf("glob result: %s", result)

		// Should find our test.txt file
		if !strings.Contains(result, "test.txt") {
			t.Errorf("glob tool did not find test.txt file, got: %s", result)
		}
	})

	// Test the grep tool - search in the file
	t.Run("grep_tool", func(t *testing.T) {
		grepTool, exists := reg.Use("grep")
		if !exists {
			t.Skip("grep tool not available in registry")
		}

		grepArgs := map[string]any{
			"pattern": "Updated",
			"file":    "test.txt",
		}

		result, err := grepTool.Execute(ctx, grepArgs)
		if err != nil {
			t.Fatalf("grep tool failed: %v", err)
		}
		t.Logf("grep result: %s", result)

		// Should find the pattern in our file
		if !strings.Contains(result, "Updated") {
			t.Errorf("grep tool did not find expected pattern, got: %s", result)
		}
	})

	// Test the fetch tool with a simple HTTP request (skip if no internet)
	t.Run("fetch_tool", func(t *testing.T) {
		fetchTool, exists := reg.Use("fetch")
		if !exists {
			t.Skip("fetch tool not available in registry")
		}

		fetchArgs := map[string]any{
			"url": "https://httpbin.org/robots.txt",
		}

		result, err := fetchTool.Execute(ctx, fetchArgs)
		if err != nil {
			t.Skipf("fetch tool failed (likely no internet connection): %v", err)
		}
		t.Logf("fetch result: %s", result)

		// Should get some content from the URL
		if len(strings.TrimSpace(result)) == 0 {
			t.Error("fetch tool returned empty content")
		}
	})
}
