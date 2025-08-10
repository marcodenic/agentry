package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/marcodenic/agentry/internal/config"

	"github.com/marcodenic/agentry/internal/trace"
)

func runDev(args []string) {
	opts, _ := parseCommon("dev", args)
	cfg, err := config.Load("examples/.agentry.yaml")
	if err != nil {
		panic(err)
	}
	applyOverrides(cfg, opts)
	ag, err := buildAgent(cfg)
	if err != nil {
		panic(err)
	}
	if opts.maxIter > 0 {
		ag.MaxIterations = opts.maxIter
	}
	if opts.resumeID != "" {
		_ = ag.LoadState(context.Background(), opts.resumeID)
	}

	sc := bufio.NewScanner(os.Stdin)
	fmt.Println("Agentry REPL – Ctrl-D to quit")
	for {
		fmt.Print("> ")
		if !sc.Scan() {
			break
		}
		line := sc.Text()
		if strings.HasPrefix(line, "converse") {
			// Team conversation functionality is being refactored
			fmt.Printf("Team conversation mode temporarily disabled during refactoring\n")
			continue
		}
		col := trace.NewCollector(nil)
		ag.Tracer = col
		out, err := ag.Run(context.Background(), line)
		if err != nil {
			fmt.Println("ERR:", err)
			continue
		}
		sum := trace.Analyze(line, col.Events())
		fmt.Println(out)
		fmt.Printf("input tokens: %d, output tokens: %d, total tokens: %d, cost: $%.6f\n",
			sum.InputTokens, sum.OutputTokens, sum.TotalTokens, sum.Cost)
		if opts.saveID != "" {
			_ = ag.SaveState(context.Background(), opts.saveID)
		}
	}
}

// Reference to avoid unused function warning during refactors
var _ = runDev
