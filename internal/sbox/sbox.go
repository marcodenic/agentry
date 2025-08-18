package sbox

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
)

// Options controls sandboxed execution.
type Options struct {
	Engine   string // "docker", "gvisor", "firecracker", or "cri"
	Image    string
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

// ExecDirect runs a shell command directly without any sandboxing.
// This is used when sandboxing is disabled.
func ExecDirect(ctx context.Context, cmdStr string) (string, error) {
	// Use runtime.GOOS to determine the platform
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Windows - use PowerShell with full path for better compatibility
		powershellPath := "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"
		cmd = RunCommand(ctx, powershellPath, "-Command", cmdStr)
	} else {
		// Unix-like systems - use sh
		cmd = RunCommand(ctx, "sh", "-c", cmdStr)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func buildArgs(engine, cmdStr string, opts Options) []string {
	switch engine {
	case "docker":
		img := opts.Image
		if img == "" {
			img = "alpine"
		}
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
		args = append(args, img, "sh", "-c", cmdStr)
		return args
	case "gvisor":
		img := opts.Image
		if img == "" {
			return []string{"runsc", "bash", "-c", cmdStr}
		}
		return []string{"runsc", "--rootfs", img, "bash", "-c", cmdStr}
	case "firecracker":
		img := opts.Image
		if img == "" {
			img = "alpine"
		}
		return []string{"ignite", "run", "--quiet", img, "--", "sh", "-c", cmdStr}
	case "cri":
		return []string{"cri-shim", "run", "--", "sh", "-c", cmdStr}
	default:
		return nil
	}
}
