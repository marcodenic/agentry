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
		key := ""
		if envVar, ok := m.Options["env_key"]; ok && envVar != "" {
			key = os.Getenv(envVar)
		} else {
			key = os.Getenv("OPENAI_KEY")
		}
		return NewOpenAI(key), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", m.Provider)
	}
}
