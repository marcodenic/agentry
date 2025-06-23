package tests

import (
	"context"
	"os/exec"
	"runtime"
)

// echoCmd returns a cross-platform echo command for tests.
func echoCmd(ctx context.Context, args ...string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "cmd", append([]string{"/c", "echo"}, args...)...)
	}
	return exec.CommandContext(ctx, "echo", args...)
}
