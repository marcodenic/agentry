package main

import (
	"context"
	"fmt"
	"os"

	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/flow"
)

func runFlow(args []string) {
	opts, _ := parseCommon("flow", args)
	f, err := flow.Load(opts.configPath)
	if err != nil {
		fmt.Printf("failed to load flow: %v\n", err)
		os.Exit(1)
	}
	outs, err := flow.Run(context.Background(), f, tool.DefaultRegistry(), nil)
	if err != nil {
		fmt.Printf("flow error: %v\n", err)
		os.Exit(1)
	}
	for _, o := range outs {
		fmt.Println(o)
	}
}
