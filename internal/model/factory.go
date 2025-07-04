package model

import (
	"fmt"
	"os"

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
			key = os.Getenv("OPENAI_KEY")
		}
		modelName := m.Options["model"]
		if modelName == "" {
			return nil, fmt.Errorf("model name is required for OpenAI provider")
		}
		return NewOpenAI(key, modelName), nil
	case "anthropic":
		key := m.Options["key"]
		if key == "" {
			key = os.Getenv("ANTHROPIC_KEY")
		}
		modelName := m.Options["model"]
		if modelName == "" {
			return nil, fmt.Errorf("model name is required for Anthropic provider")
		}
		return NewAnthropic(key, modelName), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", m.Provider)
	}
}
