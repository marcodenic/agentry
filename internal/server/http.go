package server

import (
	"encoding/json"
	"net/http"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) http.Handler {
	mux := http.NewServeMux()
	if metrics {
		mux.Handle("/metrics", promhttp.Handler())
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
		ag := agents[in.AgentID]
		if ag == nil {
			http.Error(w, "unknown agent", http.StatusBadRequest)
			return
		}
		if in.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			tr := trace.NewSSE(w)
			ag.Tracer = tr
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
			if _, err := ag.Run(r.Context(), in.Input); err != nil {
				http.Error(w, err.Error(), 500)
			}
			return
		}
		if resumeID != "" {
			_ = ag.LoadState(r.Context(), resumeID)
		}
		out, err := ag.Run(r.Context(), in.Input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if saveID != "" {
			_ = ag.SaveState(r.Context(), saveID)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"output": out})
	})
	return mux
}

func Serve(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) error {
	return http.ListenAndServe(":8080", Handler(agents, metrics, saveID, resumeID))
}
