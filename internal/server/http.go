package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/ui"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Handler(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) http.Handler {
	mux := http.NewServeMux()
	if metrics {
		mux.Handle("/metrics", promhttp.Handler())
	}
	mux.Handle("/", http.FileServer(http.FS(ui.WebUI)))
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
		if in.AgentID == "default" {
			ag = ag.Spawn()
		} else {
			_ = ag.Resume(r.Context())
		}
		if in.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			cw := trace.NewCollector(trace.NewSSE(w))
			ag.Tracer = cw
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
			if _, err := ag.Run(r.Context(), in.Input); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			sum := trace.Analyze(in.Input, cw.Events())
			cw.Write(r.Context(), trace.Event{Type: trace.EventSummary, AgentID: ag.ID.String(), Data: sum, Timestamp: trace.Now()})
			return
		}
		if in.AgentID == "default" && resumeID != "" {
			_ = ag.LoadState(r.Context(), resumeID)
		}
		cw := trace.NewCollector(nil)
		ag.Tracer = cw
		out, err := ag.Run(r.Context(), in.Input)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if in.AgentID == "default" && saveID != "" {
			_ = ag.SaveState(r.Context(), saveID)
		}
		sum := trace.Analyze(in.Input, cw.Events())
		_ = json.NewEncoder(w).Encode(map[string]any{"output": out, "summary": sum})
	})
	return mux
}

func Serve(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) error {
	return http.ListenAndServe(":8080", Handler(agents, metrics, saveID, resumeID))
}
