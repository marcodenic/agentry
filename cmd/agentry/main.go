package main

import (
	"fmt"
	"os"

	agentry "github.com/marcodenic/agentry/internal"
	"github.com/marcodenic/agentry/internal/env"
)

func main() {
	env.Load()
	if len(os.Args) < 2 {
		runTui([]string{})
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "dev":
		runDev(args)
	case "serve":
		runServe(args)
	case "eval", "test":
		runEval(args)
	case "flow":
		runFlow(args)
	case "tui":
		runTui(args)
	case "cost":
		runCostCmd(args)
	case "pprof":
		runPProfCmd(args)
	case "plugin":
		runPluginCmd(args)
	case "tool":
		runToolCmd(args)
	case "analyze":
		runAnalyzeCmd(args)
	case "version":
		fmt.Printf("agentry %s\n", agentry.Version)
	default:
		runPrompt(cmd, args)
	}
}
