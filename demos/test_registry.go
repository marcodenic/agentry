//go:build ignore

package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	registry := tool.DefaultRegistry()
	fmt.Printf("Available tools (%d):\n", len(registry))
	for name := range registry {
		fmt.Printf("  - %s\n", name)
	}
}
