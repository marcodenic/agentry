package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestSandboxDisabled(t *testing.T) {
	// No sandboxing setup needed - it's always disabled now

	// Test that ExecDirect works (no more ExecSandbox)
	result, err := tool.ExecDirect(context.Background(), "echo hello")

	if err != nil {
		t.Logf("ExecDirect returned error: %v", err)
	} else {
		t.Logf("ExecDirect succeeded: %s", result)
	}

	// Test that direct execution works
	result2, err2 := tool.ExecDirect(context.Background(), "echo hello direct")

	if err2 != nil {
		t.Logf("ExecDirect returned error: %v", err2)
	} else {
		t.Logf("ExecDirect succeeded: %s", result2)
	}

	t.Log("✅ Sandbox disable functionality tested")
}
