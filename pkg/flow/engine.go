package flow

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/memstore"
)

// Run executes the tasks defined in the flow file and returns outputs in order.
func Run(ctx context.Context, f *File, reg tool.Registry, store memstore.KV) ([]string, error) {
	if reg == nil {
		reg = tool.DefaultRegistry()
	}

	spawn := func(name string) (*core.Agent, error) {
		conf, ok := f.Agents[name]
		if !ok {
			return nil, errors.New("unknown agent: " + name)
		}
		tools := tool.Registry{}
		if len(conf.Tools) == 0 {
			// Use full registry if no filter provided
			for n, t := range reg {
				tools[n] = t
			}
		} else {
			for _, n := range conf.Tools {
				if t, ok := reg[n]; ok {
					tools[n] = t
				}
			}
		}
		var client model.Client
		switch conf.Model {
		case "", "mock":
			client = model.NewMock()
		case "openai":
			key := os.Getenv("OPENAI_KEY")
			client = model.NewOpenAI(key, "gpt-4o")
		default:
			client = model.NewMock()
		}
		route := router.Rules{{Name: conf.Model, IfContains: []string{""}, Client: client}}
		ag := core.New(route, tools, memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)
		return ag, nil
	}

	var runTask func(Task) ([]string, error)
	runTask = func(t Task) ([]string, error) {
		switch {
		case t.Agent != "":
			ag, err := spawn(t.Agent)
			if err != nil {
				return nil, err
			}
			out, err := ag.Run(ctx, t.Input)
			if err != nil {
				return nil, err
			}
			return []string{out}, nil
		case len(t.Sequential) > 0:
			var all []string
			for _, st := range t.Sequential {
				r, err := runTask(st)
				if err != nil {
					return all, err
				}
				all = append(all, r...)
			}
			return all, nil
		case len(t.Parallel) > 0:
			res := make([][]string, len(t.Parallel))
			errs := make([]error, len(t.Parallel))
			var wg sync.WaitGroup
			for i, pt := range t.Parallel {
				wg.Add(1)
				go func(i int, pt Task) {
					defer wg.Done()
					r, e := runTask(pt)
					res[i] = r
					errs[i] = e
				}(i, pt)
			}
			wg.Wait()
			var joined []string
			var agg error
			for i := range res {
				if errs[i] != nil {
					agg = errors.Join(agg, errs[i])
				}
				joined = append(joined, res[i]...)
			}
			return joined, agg
		default:
			return nil, nil
		}
	}

	var outputs []string
	var agg error
	for _, t := range f.Tasks {
		r, err := runTask(t)
		if err != nil {
			agg = errors.Join(agg, err)
		}
		outputs = append(outputs, r...)
	}
	return outputs, agg
}
