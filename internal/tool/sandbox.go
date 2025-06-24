package tool

import (
	"context"

	"github.com/marcodenic/agentry/pkg/sbox"
)

var defaultEngine string

// SetSandboxEngine configures the default sandbox engine used when executing
// unprivileged commands. Empty string falls back to docker.
func SetSandboxEngine(engine string) { defaultEngine = engine }

// ExecSandbox runs cmdStr via the sandbox using opts. If opts.Engine is empty,
// the configured default engine is used.
func ExecSandbox(ctx context.Context, cmdStr string, opts sbox.Options) (string, error) {
	if opts.Engine == "" && defaultEngine != "" {
		opts.Engine = defaultEngine
	}
	return sbox.Exec(ctx, cmdStr, opts)
}
