package sbox

import (
	"context"
	"fmt"
	"os/exec"
)

// Options controls sandboxed execution.
type Options struct {
	Engine   string // "docker" or "gvisor"
	Net      string
	CPULimit string
	MemLimit string
}

// RunCommand is used to execute external commands. It can be replaced in tests.
var RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// Exec runs a shell command inside Docker or gVisor according to opts.
func Exec(ctx context.Context, cmdStr string, opts Options) (string, error) {
	engine := opts.Engine
	if engine == "" {
		engine = "docker"
	}
	args := buildArgs(engine, cmdStr, opts)
	if len(args) == 0 {
		return "", fmt.Errorf("unknown engine %s", engine)
	}
	cmd := RunCommand(ctx, args[0], args[1:]...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func buildArgs(engine, cmdStr string, opts Options) []string {
	switch engine {
	case "docker":
		args := []string{"docker", "run", "--rm", "-v", "/workspace:/workspace"}
		if opts.Net != "" {
			args = append(args, "--network", opts.Net)
		}
		if opts.CPULimit != "" {
			args = append(args, "--cpus", opts.CPULimit)
		}
		if opts.MemLimit != "" {
			args = append(args, "--memory", opts.MemLimit)
		}
		args = append(args, "alpine", "sh", "-c", cmdStr)
		return args
	case "gvisor":
		return []string{"runsc", "bash", "-c", cmdStr}
	default:
		return nil
	}
}
