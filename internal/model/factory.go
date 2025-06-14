package model

import (
	"fmt"

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
			key = m.Options["api_key"]
		}
		return NewOpenAI(key), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", m.Provider)
	}
}
