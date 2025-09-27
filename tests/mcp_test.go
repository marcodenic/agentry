package tests

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestMcpBuiltin(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if strings.Contains(err.Error(), "not permitted") {
			t.Skipf("skipping MCP test: %v", err)
		}
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 32)
		n, _ := conn.Read(buf)
		conn.Write([]byte("resp:" + string(buf[:n])))
	}()
	port := ln.Addr().(*net.TCPAddr).Port
	tl := tool.DefaultRegistry()["mcp"]
	out, err := tl.Execute(context.Background(), map[string]any{"host": "127.0.0.1", "port": port, "command": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "resp:hi" {
		t.Fatalf("expected resp:hi, got %s", out)
	}
}
