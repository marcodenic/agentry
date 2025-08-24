package tool

import (
	"context"
	"os/exec"
	"runtime"
)

// ExecDirect runs a shell command directly
func ExecDirect(ctx context.Context, cmdStr string) (string, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "powershell", "-Command", cmdStr)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}
