package tests

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/marcodenic/agentry/internal/server"
)

func TestMetricsEndpoint(t *testing.T) {
	h := server.Handler(nil, true, "", "")
	srv := httptest.NewServer(h)
	defer srv.Close()

	// generate some metrics
	if _, err := http.Get(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}

	resp, err := http.Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "agentry_http_requests_total") {
		t.Fatalf("metrics missing: %s", b[:100])
	}
}
