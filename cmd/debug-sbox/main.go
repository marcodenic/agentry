package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/marcodenic/agentry/pkg/sbox"
)

func main() {
	fmt.Printf("Testing on %s\n", runtime.GOOS)
	
	// Test direct sbox.ExecDirect with 'ls'
	fmt.Println("\n=== Testing sbox.ExecDirect with 'ls' ===")
	result, err := sbox.ExecDirect(context.Background(), "ls")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %s\n", result)
	}
	
	// Test direct sbox.ExecDirect with Get-ChildItem
	fmt.Println("\n=== Testing sbox.ExecDirect with 'Get-ChildItem' ===")
	result, err = sbox.ExecDirect(context.Background(), "Get-ChildItem")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Success: %s\n", result)
	}
}
