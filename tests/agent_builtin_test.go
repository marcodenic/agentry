package tests

import (
	"testing"
)

func TestAgentBuiltin(t *testing.T) {
	// Skip this test - agent tool is now integrated with team system
	// and works through proper delegation channels. See team tests instead.
	t.Skip("agent tool now works through team delegation system")
}
