package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

func runPProfCmd(args []string) {
	fs := flag.NewFlagSet("pprof", flag.ExitOnError)
	httpAddr := fs.String("http", "localhost:8081", "host:port for web UI")
	_ = fs.Parse(args)
	if fs.NArg() < 1 {
		fmt.Println("usage: agentry pprof [-http host:port] profile.out")
		return
	}
	cmd := exec.Command("go", "tool", "pprof", "-http", *httpAddr, fs.Arg(0))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("pprof error: %v\n", err)
		os.Exit(1)
	}
}
