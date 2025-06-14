package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/config"
	"github.com/marcodenic/agentry/internal/tool"
)

func TestFromManifestBuiltin(t *testing.T) {
	m := config.ToolManifest{Name: "echo", Description: "", Type: "builtin"}
	tl, err := tool.FromManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tl.Execute(context.Background(), map[string]any{"text": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "hello" {
		t.Errorf("expected hello, got %s", out)
	}
}

func TestFromManifestHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Write(b)
	}))
	defer srv.Close()
	m := config.ToolManifest{Name: "ping", HTTP: srv.URL, Description: ""}
	tl, err := tool.FromManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tl.Execute(context.Background(), map[string]any{"x": "y"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "\"x\":\"y\"") {
		t.Errorf("unexpected http output: %s", out)
	}
}

func TestFromManifestCommand(t *testing.T) {
	m := config.ToolManifest{Name: "local", Command: "echo hi", Description: ""}
	tl, err := tool.FromManifest(m)
	if err != nil {
		t.Fatal(err)
	}
	out, err := tl.Execute(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(out) != "hi" {
		t.Errorf("expected hi, got %s", out)
	}
}
