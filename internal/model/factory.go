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
			modelName = "gpt-4o"
		}
		return NewOpenAI(key, modelName), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", m.Provider)
	}
}
