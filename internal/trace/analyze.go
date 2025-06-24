package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/marcodenic/agentry/internal/model"
)

// Summary holds token usage and estimated cost for a run.
type Summary struct {
	Tokens int     `json:"tokens"`
	Cost   float64 `json:"cost"`
}

// CostPerToken is the estimated cost for a single token in dollars.
const CostPerToken = 0.000002

// Analyze returns the token count and cost for an input and its trace events.
func Analyze(input string, events []Event) Summary {
	tokens := len(strings.Fields(input))
	for _, ev := range events {
		switch ev.Type {
		case EventStepStart:
			switch d := ev.Data.(type) {
			case model.Completion:
				tokens += len(strings.Fields(d.Content))
			case map[string]any:
				if s, ok := d["Content"].(string); ok {
					tokens += len(strings.Fields(s))
				}
			}
		case EventToolEnd:
			if m, ok := ev.Data.(map[string]any); ok {
				if r, ok := m["result"].(string); ok {
					tokens += len(strings.Fields(r))
				}
			}
		case EventFinal:
			if s, ok := ev.Data.(string); ok {
				tokens += len(strings.Fields(s))
			}
		}
	}
	return Summary{Tokens: tokens, Cost: float64(tokens) * CostPerToken}
}

// ParseLog decodes newline-delimited JSON trace events from r.
func ParseLog(r io.Reader) ([]Event, error) {
	dec := json.NewDecoder(bufio.NewReader(r))
	var evs []Event
	for {
		var ev Event
		if err := dec.Decode(&ev); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}
		evs = append(evs, ev)
	}
	return evs, nil
}

// AnalyzeFile loads a newline-delimited JSON trace log and returns the
// token usage summary. The input text is assumed to be empty.
func AnalyzeFile(path string) (Summary, error) {
	f, err := os.Open(path)
	if err != nil {
		return Summary{}, err
	}
	defer f.Close()
	evs, err := ParseLog(f)
	if err != nil {
		return Summary{}, err
	}
	return Analyze("", evs), nil
}
