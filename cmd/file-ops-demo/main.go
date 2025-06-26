package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	// Demonstrate the new advanced file operation tools
	tempDir, err := os.MkdirTemp("", "agentry_demo")
	if err != nil {
		log.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "demo.go")
	
	// Get the default registry with all builtin tools
	registry := tool.DefaultRegistry()
	
	fmt.Println("🎯 Agentry Advanced File Operations Demo")
	fmt.Println("=======================================")
	
	// 1. Create a file
	createTool, _ := registry.Use("create_file")
	createArgs := map[string]any{
		"path": testFile,
		"content": `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
    // TODO: Add more functionality
}`,
	}
	
	fmt.Println("\n📝 Creating file...")
	result, err := createTool.Execute(context.Background(), createArgs)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	fmt.Printf("Result: %s\n", result)
	
	// 2. Get file info
	infoTool, _ := registry.Use("get_file_info")
	infoArgs := map[string]any{
		"path": testFile,
	}
	
	fmt.Println("\nℹ️ Getting file info...")
	result, err = infoTool.Execute(context.Background(), infoArgs)
	if err != nil {
		log.Fatalf("Failed to get file info: %v", err)
	}
	
	var fileInfo map[string]any
	json.Unmarshal([]byte(result), &fileInfo)
	fmt.Printf("File type: %s, Lines: %.0f, Size: %.0f bytes\n", 
		fileInfo["file_type"], fileInfo["lines"], fileInfo["size_bytes"])
	
	// 3. View file with line numbers
	viewTool, _ := registry.Use("view_file")
	viewArgs := map[string]any{
		"path":              testFile,
		"show_line_numbers": true,
	}
	
	fmt.Println("\n👀 Viewing file with line numbers...")
	result, err = viewTool.Execute(context.Background(), viewArgs)
	if err != nil {
		log.Fatalf("Failed to view file: %v", err)
	}
	fmt.Printf("Content:\n%s\n", result)
	
	// 4. Read specific lines
	readTool, _ := registry.Use("read_lines")
	readArgs := map[string]any{
		"path":       testFile,
		"start_line": 3.0,
		"end_line":   5.0,
	}
	
	fmt.Println("\n📖 Reading lines 3-5...")
	result, err = readTool.Execute(context.Background(), readArgs)
	if err != nil {
		log.Fatalf("Failed to read lines: %v", err)
	}
	
	var readResult map[string]any
	json.Unmarshal([]byte(result), &readResult)
	fmt.Printf("Lines 3-5:\n%s\n", readResult["content"])
	
	// 5. Insert at line
	insertTool, _ := registry.Use("insert_at")
	insertArgs := map[string]any{
		"path":    testFile,
		"line":    6.0,
		"content": "    // Enhanced with Agentry!",
	}
	
	fmt.Println("\n➕ Inserting comment at line 6...")
	result, err = insertTool.Execute(context.Background(), insertArgs)
	if err != nil {
		log.Fatalf("Failed to insert: %v", err)
	}
	fmt.Printf("Result: %s\n", result)
	
	// Re-view file after insert to update protection
	fmt.Println("\n👀 Re-viewing file after insert...")
	result, err = viewTool.Execute(context.Background(), viewArgs)
	if err != nil {
		log.Fatalf("Failed to re-view file: %v", err)
	}
	
	// 6. Search and replace with regex
	replaceTool, _ := registry.Use("search_replace")
	replaceArgs := map[string]any{
		"path":    testFile,
		"search":  `fmt\.Println\("([^"]+)"\)`,
		"replace": `log.Println("$1")`,
		"regex":   true,
	}
	
	fmt.Println("\n🔍 Replacing fmt.Println with log.Println using regex...")
	result, err = replaceTool.Execute(context.Background(), replaceArgs)
	if err != nil {
		log.Fatalf("Failed to replace: %v", err)
	}
	fmt.Printf("Result: %s\n", result)
	
	// Re-view file after replace to update protection  
	fmt.Println("\n👀 Re-viewing file after replace...")
	result, err = viewTool.Execute(context.Background(), viewArgs)
	if err != nil {
		log.Fatalf("Failed to re-view file: %v", err)
	}
	
	// 7. Edit range
	editTool, _ := registry.Use("edit_range")
	editArgs := map[string]any{
		"path":       testFile,
		"start_line": 3.0,
		"end_line":   3.0,
		"content":    `import (\n    "fmt"\n    "log"\n)`,
	}
	
	fmt.Println("\n✏️ Editing import statement (lines 3-3)...")
	result, err = editTool.Execute(context.Background(), editArgs)
	if err != nil {
		log.Fatalf("Failed to edit range: %v", err)
	}
	fmt.Printf("Result: %s\n", result)
	
	// 8. Final view
	fmt.Println("\n🎉 Final file content:")
	result, err = viewTool.Execute(context.Background(), viewArgs)
	if err != nil {
		log.Fatalf("Failed to view final file: %v", err)
	}
	fmt.Printf("%s\n", result)
	
	fmt.Println("\n✅ Demo complete! All advanced file operations working correctly.")
}
