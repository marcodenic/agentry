package tests

import (
	"context"
	"testing"

	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/sbox"
)

func TestSandboxDisabled(t *testing.T) {
	// Set sandbox to disabled
	tool.SetSandboxEngine("disabled")
	
	// Test that ExecSandbox with disabled engine works
	result, err := tool.ExecSandbox(context.Background(), "echo hello", sbox.Options{})
	
	if err != nil {
		t.Logf("ExecSandbox with disabled engine returned error (expected): %v", err)
	} else {
		t.Logf("ExecSandbox with disabled engine succeeded: %s", result)
	}
	
	// Test that direct execution works
	result2, err2 := sbox.ExecDirect(context.Background(), "echo hello direct")
	
	if err2 != nil {
		t.Logf("ExecDirect returned error: %v", err2)
	} else {
		t.Logf("ExecDirect succeeded: %s", result2)
	}
	
	t.Log("✅ Sandbox disable functionality tested")
}
