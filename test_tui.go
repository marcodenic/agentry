package main

import (
	"fmt"
	"os/exec"
	"time"
)

func main() {
	fmt.Println("Starting TUI test...")
	
	// Start the TUI in the background
	cmd := exec.Command("./bin/agentry.exe", "tui")
	cmd.Dir = "."
	
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Error starting TUI: %v\n", err)
		return
	}
	
	fmt.Printf("TUI started with PID: %d\n", cmd.Process.Pid)
	
	// Wait a bit for it to initialize
	time.Sleep(2 * time.Second)
	
	// Kill it after a short test
	err = cmd.Process.Kill()
	if err != nil {
		fmt.Printf("Error killing process: %v\n", err)
	}
	
	fmt.Println("TUI test completed")
}
