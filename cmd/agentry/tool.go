//go:build tools
// +build tools

package main

import (
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/plugin"
	"github.com/marcodenic/agentry/internal/tool"
	"gopkg.in/yaml.v3"
)

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
	case "openapi":
		if len(args) < 2 {
			fmt.Println("Usage: agentry tool openapi <spec.yaml>")
			os.Exit(1)
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("read spec: %v\n", err)
			os.Exit(1)
		}
		reg, err := tool.FromOpenAPI(data)
		if err != nil {
			fmt.Printf("parse spec: %v\n", err)
			os.Exit(1)
		}
		specs := tool.BuildSpecs(reg)
		b, _ := yaml.Marshal(specs)
		os.Stdout.Write(b)
	case "mcp":
		if len(args) < 2 {
			fmt.Println("Usage: agentry tool mcp <spec.json>")
			os.Exit(1)
		}
		data, err := os.ReadFile(args[1])
		if err != nil {
			fmt.Printf("read spec: %v\n", err)
			os.Exit(1)
		}
		reg, err := tool.FromMCP(data)
		if err != nil {
			fmt.Printf("parse spec: %v\n", err)
			os.Exit(1)
		}
		specs := tool.BuildSpecs(reg)
		b, _ := yaml.Marshal(specs)
		os.Stdout.Write(b)
	default:
		fmt.Printf("unknown tool command %s\n", args[0])
		os.Exit(1)
	}
}
