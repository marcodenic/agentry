package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/marcodenic/agentry/internal/config"
	"os"
	"strconv"
	"strings"

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
			rest := strings.TrimSpace(strings.TrimPrefix(line, "converse"))
			_ = 2 // n was used for agent count
			topic := ""
			if rest != "" {
				fields := strings.Fields(rest)
				if len(fields) > 0 {
					if v, err := strconv.Atoi(fields[0]); err == nil && v > 0 {
						_ = v // n = v
						rest = strings.TrimSpace(rest[len(fields[0]):])
					}
				}
				topic = strings.TrimSpace(rest)
			}
			if topic == "" {
				topic = "Hello agents, let's chat!"
			} else if (strings.HasPrefix(topic, "\"") && strings.HasSuffix(topic, "\"")) ||
				(strings.HasPrefix(topic, "'") && strings.HasSuffix(topic, "'")) {
				topic = strings.Trim(topic, "'\"")
			}
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
		fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
		if opts.saveID != "" {
			_ = ag.SaveState(context.Background(), opts.saveID)
		}
	}
}
