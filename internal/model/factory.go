package model

import (
	"fmt"
	"os"
	"strconv"

	"github.com/marcodenic/agentry/internal/config"
)

// FromManifest creates a Client from a config.ModelManifest.
func FromManifest(m config.ModelManifest) (Client, error) {
	switch m.Provider {
	case "mock":
		return NewMock(), nil
	case "openai":
		key := m.Options["key"]
		if key == "" {
			// Standardize on OPENAI_API_KEY
			key = os.Getenv("OPENAI_API_KEY")
		}
		modelName := m.Options["model"]
		if modelName == "" {
			return nil, fmt.Errorf("model name is required for OpenAI provider")
		}
		c := NewOpenAI(key, modelName)
		if tStr := m.Options["temperature"]; tStr != "" {
			if t, err := strconv.ParseFloat(tStr, 64); err == nil {
				// Set pointer only; unsupported models will omit temperature on request
				c.Temperature = &t
			}
		}
		return c, nil
	case "anthropic":
		key := m.Options["key"]
		if key == "" {
			// Standardize on ANTHROPIC_API_KEY
			key = os.Getenv("ANTHROPIC_API_KEY")
		}
		modelName := m.Options["model"]
		if modelName == "" {
			return nil, fmt.Errorf("model name is required for Anthropic provider")
		}
		c := NewAnthropic(key, modelName)
		if tStr := m.Options["temperature"]; tStr != "" {
			if t, err := strconv.ParseFloat(tStr, 64); err == nil {
				// set temperature if supported by client
				// anthropic client uses unexported field; expose via helper if needed in future
				// noop for now
				_ = t
			}
		}
		return c, nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", m.Provider)
	}
}
