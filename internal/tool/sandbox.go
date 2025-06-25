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
// the configured default engine is used. If the engine is "disabled" or "none",
// sandboxing is bypassed and the command runs directly.
func ExecSandbox(ctx context.Context, cmdStr string, opts sbox.Options) (string, error) {
	if opts.Engine == "" && defaultEngine != "" {
		opts.Engine = defaultEngine
	}

	// Check if sandboxing is disabled
	if opts.Engine == "disabled" || opts.Engine == "none" || defaultEngine == "disabled" || defaultEngine == "none" {
		// Execute command directly without sandboxing
		return sbox.ExecDirect(ctx, cmdStr)
	}

	return sbox.Exec(ctx, cmdStr, opts)
}
