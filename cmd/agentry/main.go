package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/yourname/agentry/internal/config"
	"github.com/yourname/agentry/internal/core"
	"github.com/yourname/agentry/internal/memory"
	"github.com/yourname/agentry/internal/model"
	"github.com/yourname/agentry/internal/router"
	"github.com/yourname/agentry/internal/server"
	"github.com/yourname/agentry/internal/tool"
)

func main() {
	mode := flag.String("mode", "dev", "dev|serve|eval")
	conf := flag.String("config", "", "path to .agentry.yaml")
	flag.Parse()

	switch *mode {
	case "dev":
		// start REPL (omitted for brevity)
	case "serve":
		if *conf == "" {
			fmt.Println("need --config")
			os.Exit(1)
		}
		cfg, err := config.Load(*conf)
		if err != nil {
			panic(err)
		}
		reg := tool.Registry{}
		for _, m := range cfg.Tools {
			tl, _ := tool.FromManifest(m)
			reg[m.Name] = tl
		}
		r := router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
		ag := core.New(r, reg, memory.NewInMemory(), nil)
		agents := map[string]*core.Agent{"default": ag}
		server.Serve(agents)
	case "eval":
		// load suite and run eval on default agent
	default:
		fmt.Println("unknown mode")
	}
}
