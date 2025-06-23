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
	case "fetch":
		if len(args) < 3 {
			fmt.Println("Usage: agentry plugin fetch <index> <name>")
			os.Exit(1)
		}
		if _, err := plugin.Fetch(args[1], args[2]); err != nil {
			fmt.Printf("fetch failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown plugin command %s\n", args[0])
		os.Exit(1)
	}
}

func runToolCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: agentry tool <command> [args]")
		os.Exit(1)
	}
	switch args[0] {
	case "init":
		if len(args) < 2 {
			fmt.Println("Usage: agentry tool init <name>")
			os.Exit(1)
		}
		if err := plugin.InitTool(args[1]); err != nil {
			fmt.Printf("init failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown tool command %s\n", args[0])
		os.Exit(1)
	}
}
