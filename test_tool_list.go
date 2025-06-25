package main

import (
	"fmt"

	"github.com/marcodenic/agentry/internal/tool"
)

func main() {
	reg := tool.DefaultRegistry()
	fmt.Println("Available tools:")
	for name := range reg {
		fmt.Printf("  - %s\n", name)
	}
}
