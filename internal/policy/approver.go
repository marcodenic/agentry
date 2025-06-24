package policy

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/marcodenic/agentry/internal/tool"
)

// Request describes a pending tool execution to be approved.
type Request struct {
	AgentID string
	Tool    string
	Args    map[string]any
}

// Decision indicates the result of a Check.
type Decision int

const (
	// NoOpinion means the checker neither approves nor denies.
	NoOpinion Decision = iota
	// Approve allows the tool execution.
	Approve
	// Deny rejects the tool execution.
	Deny
)

// Checker evaluates whether a request should be approved.
type Checker interface {
	Check(ctx context.Context, req Request) (Decision, error)
}

// Approver decides if a request may proceed.
type Approver interface {
	Approve(ctx context.Context, req Request) (bool, error)
}

// PromptFunc requests manual approval from a user.
type PromptFunc func(req Request) bool

// Manager runs one or more checks and optionally prompts the user.
type Manager struct {
	Checks []Checker
	Prompt PromptFunc
}

func (m Manager) Approve(ctx context.Context, req Request) (bool, error) {
	for _, c := range m.Checks {
		d, err := c.Check(ctx, req)
		if err != nil {
			return false, err
		}
		switch d {
		case Approve:
			return true, nil
		case Deny:
			return false, nil
		}
	}
	if m.Prompt != nil {
		return m.Prompt(req), nil
	}
	return false, nil
}

// CLIPrompt prompts on stdin when manual approval is required.
func CLIPrompt(req Request) bool {
	fmt.Printf("approve tool %s with args %#v? [y/N]: ", req.Tool, req.Args)
	rd := bufio.NewReader(os.Stdin)
	line, _ := rd.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// WrapTools wraps a registry with approval checks.
func WrapTools(reg tool.Registry, ap Approver) tool.Registry {
	if ap == nil {
		return reg
	}
	out := tool.Registry{}
	for name, t := range reg {
		out[name] = approvalTool{Tool: t, ap: ap}
	}
	return out
}

type approvalTool struct {
	tool.Tool
	ap Approver
}

func (a approvalTool) Execute(ctx context.Context, args map[string]any) (string, error) {
	ok, err := a.ap.Approve(ctx, Request{Tool: a.Name(), Args: args})
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("tool %s not approved", a.Name())
	}
	return a.Tool.Execute(ctx, args)
}
