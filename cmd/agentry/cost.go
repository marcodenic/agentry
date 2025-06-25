//go:build tools
// +build tools

package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/marcodenic/agentry/internal/trace"
)

func runCostCmd(args []string) {
	fs := flag.NewFlagSet("cost", flag.ExitOnError)
	input := fs.String("input", "", "original user prompt")
	_ = fs.Parse(args)

	var r io.Reader
	if fs.NArg() == 0 || fs.Arg(0) == "-" {
		r = os.Stdin
	} else {
		f, err := os.Open(fs.Arg(0))
		if err != nil {
			fmt.Printf("open log: %v\n", err)
			os.Exit(1)
		}
		defer f.Close()
		r = f
	}

	events, err := trace.ParseLog(r)
	if err != nil {
		fmt.Printf("parse log: %v\n", err)
		os.Exit(1)
	}
	sum := trace.Analyze(*input, events)
	fmt.Printf("tokens: %d cost: $%.4f\n", sum.Tokens, sum.Cost)
}
