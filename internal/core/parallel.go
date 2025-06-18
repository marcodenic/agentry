package core

import (
	"context"
	"errors"
	"sync"
)

func RunParallel(ctx context.Context, agents []*Agent, inputs []string) ([]string, error) {
	var wg sync.WaitGroup
	out := make([]string, len(agents))
	errs := make([]error, len(agents))
	for i, ag := range agents {
		wg.Add(1)
		go func(i int, ag *Agent, in string) {
			defer wg.Done()
			out[i], errs[i] = ag.Run(ctx, in)
		}(i, ag, inputs[i])
	}
	wg.Wait()
	return out, errors.Join(errs...)
}
