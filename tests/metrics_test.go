package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marcodenic/agentry/internal/server"
)

func TestMetricsEndpoint(t *testing.T) {
	h := server.Handler(nil, true, "", "")
	srv := httptest.NewServer(h)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
