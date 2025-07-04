package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"

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

					// Calculate cost using the model name from the completion
					if d.ModelName != "" {
						cost := pricing.CalculateCost(d.ModelName, d.InputTokens, d.OutputTokens)
						totalCost += cost

						usage := modelUsage[d.ModelName]
						usage.InputTokens += d.InputTokens
						usage.OutputTokens += d.OutputTokens
						modelUsage[d.ModelName] = usage
					}
				}
				// Note: No fallback to word-based counting - only use actual API token counts
			}
		// Note: Tool and final events are not counted separately
		// Their token usage is included in the API response token counts
		}
	}

	// Calculate total cost using the cost manager's pricing table
	for modelName, usage := range modelUsage {
		// Use the same pricing table as the cost manager
		costManager := cost.New(0, 0) // Create temporary cost manager to access pricing
		modelCost := costManager.GetModelCost(modelName)
		if modelCost == 0 {
			// If we can't price this model, add its tokens to get a cost estimate
			costManager.AddModelUsage(modelName, usage.InputTokens, usage.OutputTokens)
			totalCost += costManager.GetModelCost(modelName)
		} else {
			totalCost += modelCost
		}
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
