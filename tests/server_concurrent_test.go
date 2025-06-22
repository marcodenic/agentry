package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

	srv := httptest.NewServer(server.Handler(agents, false))
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
			if resp.StatusCode != http.StatusOK {
				b, _ := io.ReadAll(resp.Body)
				errs <- fmt.Errorf("status %d: %s", resp.StatusCode, string(b))
				return
			}
			var out struct {
				Output string `json:"output"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
				errs <- err
				return
			}
			if out.Output != "ok" {
				errs <- fmt.Errorf("unexpected output %s", out.Output)
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
