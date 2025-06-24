package tui

import (
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestNewTeam(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	cm, err := NewChat(ag, 2, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cm.vps) != 2 {
		t.Fatalf("expected 2 panes")
	}
}
