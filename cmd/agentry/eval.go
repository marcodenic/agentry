package main

import (
	"context"
	"fmt"
	"github.com/marcodenic/agentry/internal/config"
	"os"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/eval"
)

func runEval(args []string) {
	opts, _ := parseCommon("eval", args)
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}
	applyOverrides(cfg, opts)
	key := os.Getenv("OPENAI_KEY")
	if key != "" {
		for i, m := range cfg.Models {
			if m.Name == "openai" {
				if m.Options == nil {
					m.Options = map[string]string{}
				}
				cfg.Models[i].Options["key"] = key
			}
		}
	}
	ag, err := buildAgent(cfg)
	if err != nil {
		panic(err)
	}
	if opts.maxIter > 0 {
		ag.MaxIterations = opts.maxIter
	}
	if opts.ckptID != "" {
		ag.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(opts.ckptID))
		_ = ag.Resume(context.Background())
	}
	suite := "tests/eval_suite.json"
	if key != "" {
		suite = "tests/openai_eval_suite.json"
	}
	eval.Run(nil, ag, suite)
	if opts.saveID != "" {
		_ = ag.SaveState(context.Background(), opts.saveID)
	}
}
