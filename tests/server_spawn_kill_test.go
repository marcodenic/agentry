package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/memory"
	"github.com/marcodenic/agentry/internal/model"
	"github.com/marcodenic/agentry/internal/router"
	"github.com/marcodenic/agentry/internal/server"
	"github.com/marcodenic/agentry/internal/tool"
	"github.com/marcodenic/agentry/pkg/memstore"
)

func TestServerSpawnKill(t *testing.T) {
	store := memstore.NewInMemory()
	route := router.Rules{{IfContains: []string{""}, Client: model.NewMock()}}
	ag := core.New(route, tool.DefaultRegistry(), memory.NewInMemory(), store, memory.NewInMemoryVector(), nil)
	agents := map[string]*core.Agent{"default": ag}

	srv := httptest.NewServer(server.Handler(agents, false, "", "", nil))
	defer srv.Close()

	// spawn a new agent
	resp, err := http.Post(srv.URL+"/spawn", "application/json", bytes.NewBufferString("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out struct {
		AgentID string `json:"agent_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(out.AgentID); err != nil {
		t.Fatalf("bad uuid %s", out.AgentID)
	}

	// invoke the agent (expect async response)
	body := fmt.Sprintf(`{"agent_id":"%s","input":"hi"}`, out.AgentID)
	resp2, err := http.Post(srv.URL+"/invoke", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp2.StatusCode)
	}
	var invokeRes struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&invokeRes); err != nil {
		t.Fatal(err)
	}
	if invokeRes.Status != "queued" {
		t.Fatalf("unexpected status: %s", invokeRes.Status)
	}

	// kill the agent
	killBody := fmt.Sprintf(`{"agent_id":"%s"}`, out.AgentID)
	resp3, err := http.Post(srv.URL+"/kill", "application/json", bytes.NewBufferString(killBody))
	if err != nil {
		t.Fatal(err)
	}
	resp3.Body.Close()

	// should no longer exist
	resp4, err := http.Post(srv.URL+"/invoke", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatal(err)
	}
	if resp4.StatusCode == http.StatusOK {
		t.Fatal("expected error for missing agent")
	}

	// persisted history should be stored
	b, err := store.Get(context.Background(), "history", out.AgentID)
	if err != nil || b == nil {
		t.Fatalf("state not saved: %v", err)
	}
}
