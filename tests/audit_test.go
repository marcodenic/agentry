package tests

import (
	"bytes"
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestAuditMiddleware(t *testing.T) {
	reg := tool.DefaultRegistry()
	var buf bytes.Buffer
	reg = tool.WrapWithAudit(reg, &buf)
	tl, _ := reg.Use("echo")
	if _, err := tl.Execute(context.Background(), map[string]any{"text": "hi"}); err != nil {
		t.Fatalf("exec: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("\"tool\":\"echo\"")) {
		t.Fatalf("audit log missing")
	}
}
