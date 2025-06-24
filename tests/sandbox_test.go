package tests

import (
	"context"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/sbox"
)

func TestSandboxDockerArgs(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return echoCmd(ctx, "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()

	out, err := sbox.Exec(context.Background(), "echo hi", sbox.Options{Net: "host", CPULimit: "1", MemLimit: "512m"})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	want := []string{"docker", "run", "--rm", "-v", "/workspace:/workspace", "--network", "host", "--cpus", "1", "--memory", "512m", "alpine", "sh", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSandboxGVisor(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return echoCmd(ctx, "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()

	out, err := sbox.Exec(context.Background(), "echo hi", sbox.Options{Engine: "gvisor"})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	want := []string{"runsc", "bash", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSandboxCRI(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return echoCmd(ctx, "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()

	out, err := sbox.Exec(context.Background(), "echo hi", sbox.Options{Engine: "cri"})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	want := []string{"cri-shim", "run", "--", "sh", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestSandboxFirecracker(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return echoCmd(ctx, "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()

	out, err := sbox.Exec(context.Background(), "echo hi", sbox.Options{Engine: "firecracker"})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	want := []string{"ignite", "run", "--quiet", "alpine", "--", "sh", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestExecSandboxGVisor(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return echoCmd(ctx, "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()
	tool.SetSandboxEngine("gvisor")
	defer tool.SetSandboxEngine("")
	out, err := tool.ExecSandbox(context.Background(), "echo hi", sbox.Options{})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	want := []string{"runsc", "bash", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestExecSandboxFirecracker(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return echoCmd(ctx, "ok")
	}
	defer func() {
		sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
			return exec.CommandContext(ctx, name, args...)
		}
	}()
	tool.SetSandboxEngine("firecracker")
	defer tool.SetSandboxEngine("")
	out, err := tool.ExecSandbox(context.Background(), "echo hi", sbox.Options{})
	if err != nil {
		t.Fatalf("exec failed: %v", err)
	}
	if strings.TrimSpace(out) != "ok" {
		t.Fatalf("unexpected output: %q", out)
	}
	want := []string{"ignite", "run", "--quiet", "alpine", "--", "sh", "-c", "echo hi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
