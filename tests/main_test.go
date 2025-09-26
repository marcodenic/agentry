package tests

import (
	"os"
	"testing"

	"github.com/marcodenic/agentry/internal/env"
)

func TestMain(m *testing.M) {
	env.Load()
	if os.Getenv("AGENTRY_RUN_INTEGRATION_TESTS") == "" {
		// Integration suite requires network and local listeners; skip unless explicitly enabled.
		os.Exit(0)
	}
	os.Exit(m.Run())
}
