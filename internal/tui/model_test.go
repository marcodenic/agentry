package tui

import (
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestNew(t *testing.T) {
	ag := core.New(router.Rules{{IfContains: []string{""}, Client: nil}}, tool.Registry{}, memory.NewInMemory(), nil, memory.NewInMemoryVector(), nil)
	m := New(ag)
	if m.agent != ag {
		t.Fatalf("agent mismatch")
	}
	if len(m.tools.Items()) != 0 {
		t.Fatalf("expected no tools")
	}
}
