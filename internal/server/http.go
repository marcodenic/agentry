package server

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/trace"
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
		ag := base.Spawn()
		writers := []trace.Writer{}
		if in.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.Header().Set("Cache-Control", "no-cache")
			writers = append(writers, trace.NewSSE(w))
			if fl, ok := w.(http.Flusher); ok {
				fl.Flush()
			}
		}
		if metrics {
			writers = append(writers, trace.NewOTel())
			if mem != nil {
				writers = append(writers, mem)
			}
		}
		if len(writers) > 0 {
			ag.Tracer = trace.NewMulti(writers...)
		}
		if in.Stream {
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
