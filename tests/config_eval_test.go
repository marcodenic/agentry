package tests

import (
	"testing"

	"github.com/yourname/agentry/internal/config"
	"github.com/yourname/agentry/internal/core"
	"github.com/yourname/agentry/internal/eval"
	"github.com/yourname/agentry/internal/memory"
	"github.com/yourname/agentry/internal/model"
	"github.com/yourname/agentry/internal/router"
	"github.com/yourname/agentry/internal/tool"
)

func TestConfigBootAndEval(t *testing.T) {
	cfg, err := config.Load("../examples/.agentry.yaml")
	if err != nil {
		t.Fatal(err)
	}
	reg := tool.Registry{}
	for _, m := range cfg.Tools {
		tl, _ := tool.FromManifest(m)
		reg[m.Name] = tl
	}
	mockModel := model.NewMock()
	r := router.Rules{{IfContains: []string{""}, Client: mockModel}}
	ag := core.New(r, reg, memory.NewInMemory(), nil)
	eval.Run(t, ag, "../tests/eval_suite.json")
}
