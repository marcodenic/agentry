package tests

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/internal/trace"
)

func TestInvokeStreaming(t *testing.T) {
	reg := tool.DefaultRegistry()

	route := router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
	ag := core.New(route, reg, memory.NewInMemory(), nil)
	agents := map[string]*core.Agent{"a": ag}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			AgentID string `json:"agent_id"`
			Input   string `json:"input"`
			Stream  bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		ag := agents[in.AgentID]
		if ag == nil {
			http.Error(w, "unknown agent", http.StatusBadRequest)
			return
		}
		if in.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			tr := trace.NewSSE(w)
			ag.Tracer = tr
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
			if _, err := ag.Run(r.Context(), in.Input); err != nil {
				http.Error(w, err.Error(), 500)
			}
			return
		}
		out, err := ag.Run(r.Context(), in.Input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"output": out})
	})

	srv := httptest.NewServer(handler)
	defer srv.Close()

	body := bytes.NewBufferString(`{"agent_id":"a","input":"hi","stream":true}`)
	resp, err := http.Post(srv.URL+"/invoke", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	sc := bufio.NewScanner(resp.Body)
	sc.Scan()
	line := sc.Text()
	if !bytes.HasPrefix([]byte(line), []byte("data:")) {
		t.Fatalf("expected data line, got %s", line)
	}
	var ev trace.Event
	if err := json.Unmarshal(bytes.TrimSpace(bytes.TrimPrefix([]byte(line), []byte("data:"))), &ev); err != nil {
		t.Fatalf("bad event: %v", err)
	}
	if ev.Type == "" {
		t.Fatalf("empty event type")
	}
}
