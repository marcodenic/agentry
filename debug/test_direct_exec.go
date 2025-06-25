//go:build ignore

package main

import (
	"context"
	"fmt"

	"github.com/marcodenic/agentry/pkg/sbox"
)

func main() {
	ctx := context.Background()

	// Test direct execution with the same command pattern
	fmt.Println("Testing ExecDirect with simple command...")
	result, err := sbox.ExecDirect(ctx, "echo 'hello world'")
	if err != nil {
		fmt.Printf("ERROR with simple command: %v\n", err)
	} else {
		fmt.Printf("SUCCESS with simple command: %s\n", result)
	}

	// Test reading a file
	fmt.Println("\nTesting ExecDirect with file read...")
	result, err = sbox.ExecDirect(ctx, "Get-Content ../README.md | Select-Object -First 3")
	if err != nil {
		fmt.Printf("ERROR reading file: %v\n", err)
	} else {
		fmt.Printf("SUCCESS reading file: %s\n", result)
	}

	// Test what might be causing exit 127
	fmt.Println("\nTesting potential problem command...")
	result, err = sbox.ExecDirect(ctx, "Get-ChildItem")
	if err != nil {
		fmt.Printf("ERROR with Get-ChildItem: %v\n", err)
	} else {
		fmt.Printf("SUCCESS with Get-ChildItem: %s\n", result)
	}
}
