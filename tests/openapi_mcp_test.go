package tests

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestFromOpenAPI(t *testing.T) {
	spec := `openapi: 3.0.0
info:
  title: Echo API
  version: 1.0.0
servers:
  - url: SERVER_URL
paths:
  /echo:
    post:
      operationId: echo
      summary: Echo input
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                text:
                  type: string
              required: [text]
      responses:
        '200':
          description: ok
`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Write(b)
	}))
	defer srv.Close()
	spec = strings.ReplaceAll(spec, "SERVER_URL", srv.URL)

	reg, err := tool.FromOpenAPI([]byte(spec))
	if err != nil {
		t.Fatal(err)
	}
	tl := reg["echo"]
	out, err := tl.Execute(context.Background(), map[string]any{"text": "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "{\"text\":\"hi\"}" {
		t.Fatalf("unexpected output %s", out)
	}
}

func TestFromMCP(t *testing.T) {
	spec := `{
      "server": "SERVER_URL",
      "commands": [
        {
          "name": "ping",
          "description": "Ping command",
          "method": "POST",
          "path": "/ping",
          "schema": {
            "type": "object",
            "properties": {"msg": {"type": "string"}},
            "required": ["msg"]
          }
        }
      ]
    }`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		w.Write(b)
	}))
	defer srv.Close()
	spec = strings.ReplaceAll(spec, "SERVER_URL", srv.URL)

	reg, err := tool.FromMCP([]byte(spec))
	if err != nil {
		t.Fatal(err)
	}
	tl := reg["ping"]
	out, err := tl.Execute(context.Background(), map[string]any{"msg": "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if out != "{\"msg\":\"hello\"}" {
		t.Fatalf("unexpected output %s", out)
	}
}
