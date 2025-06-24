package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/policy"
	"github.com/marcodenic/agentry/internal/tool"
)

type allowCheck struct{}

func (allowCheck) Check(ctx context.Context, req policy.Request) (policy.Decision, error) {
	return policy.Approve, nil
}

type denyCheck struct{}

func (denyCheck) Check(ctx context.Context, req policy.Request) (policy.Decision, error) {
	return policy.Deny, nil
}

func TestPolicyAutoApprove(t *testing.T) {
	cnt := 0
	reg := tool.Registry{"t": tool.New("t", "", func(context.Context, map[string]any) (string, error) { cnt++; return "ok", nil })}
	ap := policy.Manager{Checks: []policy.Checker{allowCheck{}}}
	reg = policy.WrapTools(reg, ap)
	tl, _ := reg.Use("t")
	if _, err := tl.Execute(context.Background(), nil); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("tool not executed")
	}
}

func TestPolicyPrompt(t *testing.T) {
	cnt := 0
	reg := tool.Registry{"t": tool.New("t", "", func(context.Context, map[string]any) (string, error) { cnt++; return "ok", nil })}
	called := false
	ap := policy.Manager{Prompt: func(req policy.Request) bool { called = true; return true }}
	reg = policy.WrapTools(reg, ap)
	tl, _ := reg.Use("t")
	if _, err := tl.Execute(context.Background(), nil); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if !called {
		t.Fatalf("prompt not called")
	}
	if cnt != 1 {
		t.Fatalf("tool not executed")
	}
}

func TestPolicyDeny(t *testing.T) {
	reg := tool.Registry{"t": tool.New("t", "", func(context.Context, map[string]any) (string, error) { return "ok", nil })}
	ap := policy.Manager{Checks: []policy.Checker{denyCheck{}}}
	reg = policy.WrapTools(reg, ap)
	tl, _ := reg.Use("t")
	if _, err := tl.Execute(context.Background(), nil); err == nil {
		t.Fatalf("expected denial")
	}
}
