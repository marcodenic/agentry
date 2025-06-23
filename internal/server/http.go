package server

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/taskqueue"
	"github.com/marcodenic/agentry/ui"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) http.Handler {
	mux := http.NewServeMux()
	var mem *trace.MemoryWriter
	if metrics {
		mux.Handle("/metrics", promhttp.Handler())
		mem = trace.NewMemory(100)
		mux.HandleFunc("/traces", func(w http.ResponseWriter, r *http.Request) {
			_ = json.NewEncoder(w).Encode(mem.All())
		})
	}
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		list := make([]string, 0, len(agents))
		for id := range agents {
			list = append(list, id)
		}
		sort.Strings(list)
		_ = json.NewEncoder(w).Encode(list)
	})
	mux.Handle("/", http.FileServer(http.FS(ui.WebUI)))

	// NATS queue setup (URL/subject could be from config/env)
	q, err := taskqueue.NewQueue(natsURL(), "agentry.tasks")
	if err != nil {
		panic("NATS unavailable: " + err.Error())
	}

	mux.HandleFunc("/spawn", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			Template string `json:"template"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if in.Template == "" {
			in.Template = "default"
		}
		base := agents[in.Template]
		if base == nil {
			http.Error(w, "unknown template", http.StatusBadRequest)
			return
		}
		ag := base.Spawn()
		id := uuid.New().String()
		ag.ID = uuid.MustParse(id)
		agents[id] = ag
		_ = json.NewEncoder(w).Encode(map[string]string{"agent_id": id})
	})
	mux.HandleFunc("/kill", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			AgentID string `json:"agent_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		ag := agents[in.AgentID]
		if ag == nil {
			http.Error(w, "unknown agent", http.StatusBadRequest)
			return
		}
		_ = ag.SaveState(r.Context(), in.AgentID)
		delete(agents, in.AgentID)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	mux.HandleFunc("/invoke", func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			AgentID string `json:"agent_id"`
			Input   string `json:"input"`
			Stream  bool   `json:"stream"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		ag := agents[in.AgentID]
		if ag == nil {
			http.Error(w, "unknown agent", http.StatusBadRequest)
			return
		}
		// Publish to NATS instead of running synchronously
		task := taskqueue.Task{
			Type: "invoke",
			Payload: map[string]any{
				"agent_id": in.AgentID,
				"input":    in.Input,
				"stream":   in.Stream,
			},
		}
		if err := q.Publish(r.Context(), task); err != nil {
			http.Error(w, "queue error", 500)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "queued"})
	})
	return mux
}

// natsURL returns the NATS server URL (could be env/config driven)
func natsURL() string {
	if u := os.Getenv("NATS_URL"); u != "" {
		return u
	}
	return "nats://localhost:4222"
}

func Serve(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) error {
	return http.ListenAndServe(":8080", Handler(agents, metrics, saveID, resumeID))
}
