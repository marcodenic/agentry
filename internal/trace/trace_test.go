package trace

import (
	"bytes"
	"context"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONLWriterEmitsEvent(t *testing.T) {
	buf := new(bytes.Buffer)
	jw := NewJSONL(buf)

	evt := Event{Type: EventFinal, AgentID: "agent-1", Data: map[string]any{"done": true}}
	jw.Write(context.Background(), evt)

	out := buf.String()
	if !strings.Contains(out, "\"type\":\"final\"") {
		t.Fatalf("expected JSON payload to contain event type, got %q", out)
	}
	if !strings.Contains(out, "agent-1") {
		t.Fatalf("expected JSON payload to contain agent id, got %q", out)
	}
}

func TestSSEWriterFormatsDataLines(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := NewSSE(rr)

	evt := Event{Type: EventToolStart, AgentID: "agent-2", Data: map[string]string{"tool": "shell"}}
	sw.Write(context.Background(), evt)

	body := rr.Body.String()
	if !strings.HasPrefix(body, "data: {") {
		t.Fatalf("expected SSE output to start with data prefix, got %q", body)
	}
	if !strings.Contains(body, "\"tool\":\"shell\"") {
		t.Fatalf("expected SSE payload to contain data, got %q", body)
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Fatalf("expected SSE event to be terminated with blank line")
	}
}

func TestNowReturnsUTC(t *testing.T) {
	ts := Now()
	if ts.IsZero() {
		t.Fatalf("expected Now() to return non-zero time")
	}
	if ts.Location().String() != "UTC" {
		t.Fatalf("expected Now() to be in UTC, got %s", ts.Location())
	}
}
