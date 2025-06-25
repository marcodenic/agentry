package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/server"
)

func TestServerConcurrentInvoke(t *testing.T) {
	ag := newAgent("ok", nil)
	agents := map[string]*core.Agent{"a": ag}

	h, err := server.Handler(agents, false, "", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(h)
	defer srv.Close()

	const n = 10
	var wg sync.WaitGroup
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			body := bytes.NewBufferString(`{"agent_id":"a","input":"hi"}`)
			resp, err := http.Post(srv.URL+"/invoke", "application/json", body)
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusAccepted {
				errs <- fmt.Errorf("expected 202, got %d", resp.StatusCode)
				return
			}
			var out struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				errs <- err
				return
			}
			if out.Status != "queued" {
				errs <- fmt.Errorf("unexpected status %s", out.Status)
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Error(err)
		}
	}
}
