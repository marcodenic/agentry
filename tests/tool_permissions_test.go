package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestToolPermissions(t *testing.T) {
	tool.SetPermissions([]string{"echo"})
	defer tool.SetPermissions(nil)

	reg := tool.DefaultRegistry()
	tl, ok := reg.Use("echo")
	if !ok {
		t.Fatal("echo tool missing")
	}
	if _, err := tl.Execute(context.Background(), map[string]any{"text": "hi"}); err != nil {
		t.Fatalf("echo denied: %v", err)
	}

	deny, ok := reg.Use("ls")
	if !ok {
		t.Fatal("ls tool missing")
	}
	if _, err := deny.Execute(context.Background(), nil); !errors.Is(err, tool.ErrToolDenied) {
		t.Fatalf("expected denial, got %v", err)
	}
}
