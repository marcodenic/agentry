package server

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/internal/taskqueue"
	"github.com/marcodenic/agentry/ui"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) http.Handler {
	mux := http.NewServeMux()
	if metrics {
		mux.Handle("/metrics", promhttp.Handler())
	}
	mux.Handle("/", http.FileServer(http.FS(ui.WebUI)))

	// NATS queue setup (URL/subject could be from config/env)
	q, err := taskqueue.NewQueue(natsURL(), "agentry.tasks")
	if err != nil {
		panic("NATS unavailable: " + err.Error())
	}

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
		base := agents[in.AgentID]
		if base == nil {
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

	// Optionally, add /spawn similarly if needed

	return mux
}

// natsURL returns the NATS server URL (could be env/config driven)
func natsURL() string {
	if u := os.Getenv("NATS_URL"); u != "" {
		return u
	}
	return nats.DefaultURL
}

func Serve(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) error {
	return http.ListenAndServe(":8080", Handler(agents, metrics, saveID, resumeID))
}
