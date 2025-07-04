package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/marcodenic/agentry/internal/cost"
	"github.com/marcodenic/agentry/internal/model"
)

// Summary holds token usage and estimated cost for a run.
type Summary struct {
	InputTokens  int                        `json:"input_tokens"`
	OutputTokens int                        `json:"output_tokens"`
	TotalTokens  int                        `json:"total_tokens"`
	Cost         float64                    `json:"cost"`
	ModelUsage   map[string]cost.TokenUsage `json:"model_usage"`
}

// Analyze returns the token count and cost for an input and its trace events.
func Analyze(input string, events []Event) Summary {
	pricing := cost.NewPricingTable()
	modelUsage := make(map[string]cost.TokenUsage)
	totalInputTokens := 0
	totalOutputTokens := 0
	totalCost := 0.0

	// Only count actual API token usage - no word-based estimates
	for _, ev := range events {
		switch ev.Type {
		case EventStepStart:
			switch d := ev.Data.(type) {
			case model.Completion:
				// Use actual token counts from API response if available
				if d.InputTokens > 0 || d.OutputTokens > 0 {
					totalInputTokens += d.InputTokens
					totalOutputTokens += d.OutputTokens

					// We need to determine the model name from context
					// For now, we'll use a default model for cost calculation
					// In a real implementation, this would be tracked in the trace
					modelName := "gpt-4o" // Default fallback
					cost := pricing.CalculateCost(modelName, d.InputTokens, d.OutputTokens)
					totalCost += cost

					usage := modelUsage[modelName]
					usage.InputTokens += d.InputTokens
					usage.OutputTokens += d.OutputTokens
					modelUsage[modelName] = usage
				}
				// Note: No fallback to word-based counting - only use actual API token counts
			}
		// Note: Tool and final events are not counted separately
		// Their token usage is included in the API response token counts
		}
	}

	// If we didn't get any actual token counts, estimate from input only
	if totalInputTokens == 0 && totalOutputTokens == 0 {
		totalInputTokens = len(strings.Fields(input))
		totalCost = float64(totalInputTokens) * cost.CostPerToken
	}

	return Summary{
		InputTokens:  totalInputTokens,
		OutputTokens: totalOutputTokens,
		TotalTokens:  totalInputTokens + totalOutputTokens,
		Cost:         totalCost,
		ModelUsage:   modelUsage,
	}
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
