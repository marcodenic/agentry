package server

import (
	"encoding/json"
	"net/http"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"
)

func Serve(agents map[string]*core.Agent) error {
	http.HandleFunc("/invoke", func(w http.ResponseWriter, r *http.Request) {
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
		out, err := ag.Run(r.Context(), in.Input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"output": out})
	})
	return http.ListenAndServe(":8080", nil)
}
