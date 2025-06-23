package tests

import (
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/trace"
)

func TestAnalyzeEvents(t *testing.T) {
	events := []trace.Event{
		{Type: trace.EventStepStart, Data: model.Completion{Content: "hello world"}},
		{Type: trace.EventToolEnd, Data: map[string]any{"result": "foo bar"}},
		{Type: trace.EventFinal, Data: "bye"},
	}
	sum := trace.Analyze("input tokens", events)
	if sum.Tokens != 7 { // input(2) + hello world(2) + foo bar(2) + bye(1)
		t.Fatalf("expected 7 tokens got %d", sum.Tokens)
	}
}

func TestParseLog(t *testing.T) {
	log := `{"type":"step_start","data":{"Content":"hi there"}}
{"type":"tool_end","data":{"result":"ok"}}`
	evs, err := trace.ParseLog(strings.NewReader(log))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 2 {
		t.Fatalf("want 2 events got %d", len(evs))
	}
	sum := trace.Analyze("hello", evs)
	if sum.Tokens != 4 { // hello(1) + hi there(2) + ok(1)
		t.Fatalf("expected 4 tokens got %d", sum.Tokens)
	}
}
