package tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/marcodenic/agentry/internal/tool"
)

func TestFetchCanceledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	tl, ok := tool.DefaultRegistry().Use("fetch")
	if !ok {
		t.Fatal("fetch tool not found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := tl.Execute(ctx, map[string]any{"url": srv.URL})
	if err == nil {
		t.Fatal("expected error")
	}

	t.Logf("Got error: %v (type: %T)", err, err)

	// Check for context cancellation or deadline exceeded
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, context.Canceled) {
		// On Windows, PowerShell might return different error types
		// Accept any error that suggests cancellation/timeout
		errStr := err.Error()
		if !strings.Contains(errStr, "timeout") &&
			!strings.Contains(errStr, "cancel") &&
			!strings.Contains(errStr, "deadline") &&
			!strings.Contains(errStr, "context") {
			t.Fatalf("expected context cancellation error, got %v", err)
		}
	}
}
