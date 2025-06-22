package main

import (
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/plugin"
)

func runPluginCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: agentry plugin <command> [args]")
		os.Exit(1)
	}
	switch args[0] {
	case "install":
		if len(args) < 2 {
			fmt.Println("Usage: agentry plugin install <repo>")
			os.Exit(1)
		}
		if err := plugin.Install(args[1]); err != nil {
			fmt.Printf("install failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown plugin command %s\n", args[0])
		os.Exit(1)
	}
}
