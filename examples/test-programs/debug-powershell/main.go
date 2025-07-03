package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/sbox"
)

func main() {
	// Set sandbox to disabled
	tool.SetSandboxEngine("disabled")
	
	// Create a test file
	content := "Hello cross-platform world!"
	err := os.WriteFile("debug_test.txt", []byte(content), 0644)
	if err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove("debug_test.txt")
	
	ctx := context.Background()
	
	// Test write command
	fmt.Println("Testing write command...")
	cmd := "Set-Content -Path 'debug_test2.txt' -Value 'Test content from PowerShell'"
	result, err := sbox.ExecDirect(ctx, cmd)
	if err != nil {
		fmt.Printf("Write command failed: %v\n", err)
		fmt.Printf("Output: %s\n", result)
	} else {
		fmt.Printf("Write command succeeded: %s\n", result)
	}
	
	// Test view command
	fmt.Println("Testing view command...")
	cmd = "Get-Content debug_test.txt"
	result, err = sbox.ExecDirect(ctx, cmd)
	if err != nil {
		fmt.Printf("View command failed: %v\n", err)
		fmt.Printf("Output: %s\n", result)
	} else {
		fmt.Printf("View command succeeded: %s\n", result)
	}
	
	// Test ls command
	fmt.Println("Testing ls command...")
	cmd = "dir ."
	result, err = sbox.ExecDirect(ctx, cmd)
	if err != nil {
		fmt.Printf("Ls command failed: %v\n", err)
		fmt.Printf("Output: %s\n", result)
	} else {
		fmt.Printf("Ls command succeeded: %s\n", result)
	}
}
