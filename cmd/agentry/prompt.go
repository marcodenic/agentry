package main

import (
	"context"
	"fmt"
	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/trace"
	"os"
	"strings"
)

func runAnalyzeCmd(args []string) {
	if len(args) < 1 {
		fmt.Println("usage: agentry analyze trace.log")
		return
	}
	sum, err := trace.AnalyzeFile(args[0])
	if err != nil {
		fmt.Println("analyze error:", err)
		os.Exit(1)
	}
	fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
}

func runPrompt(cmd string, args []string) {
	opts, _ := parseCommon(cmd, args)
	prompt := cmd
	if len(args) > 0 {
		prompt = cmd + " " + strings.Join(args, " ")
	}
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
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
	col := trace.NewCollector(nil)
	ag.Tracer = col
	out, err := ag.Run(context.Background(), prompt)
	if err != nil {
		fmt.Printf("ERR: %v\n", err)
		os.Exit(1)
	}
	sum := trace.Analyze(prompt, col.Events())
	fmt.Println(out)
	fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
