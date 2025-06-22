package tests

import (
	"context"
	"os/exec"
	"reflect"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/pkg/sbox"
)

func TestSandboxDockerArgs(t *testing.T) {
	var got []string
	sbox.RunCommand = func(ctx context.Context, name string, args ...string) *exec.Cmd {
		got = append([]string{name}, args...)
		return exec.CommandContext(ctx, "echo", "ok")
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
		return exec.CommandContext(ctx, "echo", "ok")
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
