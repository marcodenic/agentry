package server

import (
        "encoding/json"
        "net/http"
        "os"
        "sort"
        "time"

	"github.com/google/uuid"
	"github.com/marcodenic/agentry/internal/core"
	"github.com/marcodenic/agentry/internal/taskqueue"
	"github.com/marcodenic/agentry/internal/trace"
	"github.com/marcodenic/agentry/ui"

        "github.com/prometheus/client_golang/prometheus"
        "github.com/prometheus/client_golang/prometheus/promauto"
        "github.com/prometheus/client_golang/prometheus/promhttp"
        dto "github.com/prometheus/client_model/go"
)

var (
	httpRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agentry_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"path"})
        httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
                Name:    "agentry_http_request_duration_seconds",
                Help:    "Duration of HTTP requests",
                Buckets: prometheus.DefBuckets,
        }, []string{"path"})
       agentUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
               Name: "agentry_agent_up",
               Help: "Agent running status",
       }, []string{"agent"})
)

func instrument(path string, h http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                start := time.Now()
                h.ServeHTTP(w, r)
                httpRequests.WithLabelValues(path).Inc()
                httpDuration.WithLabelValues(path).Observe(time.Since(start).Seconds())
        })
}

func metricMap(name string) map[string]float64 {
       out := map[string]float64{}
       mfs, _ := prometheus.DefaultGatherer.Gather()
       for _, mf := range mfs {
               if mf.GetName() != name {
                       continue
               }
               for _, m := range mf.GetMetric() {
                       var agent string
                       for _, l := range m.GetLabel() {
                               if l.GetName() == "agent" {
                                       agent = l.GetValue()
                               }
                       }
                       if m.Counter != nil {
                               out[agent] = m.GetCounter().GetValue()
                       } else if m.Gauge != nil {
                               out[agent] = m.GetGauge().GetValue()
                       }
               }
       }
       return out
}

func Handler(agents map[string]*core.Agent, metrics bool, saveID, resumeID string) http.Handler {
	mux := http.NewServeMux()
	var mem *trace.MemoryWriter
	if metrics {
               mux.Handle("/metrics", promhttp.Handler())
               mem = trace.NewMemory(100)
               for id, ag := range agents {
                       agentUp.WithLabelValues(id).Set(1)
                       if ag.Tracer == nil {
                               ag.Tracer = mem
                       } else {
                               ag.Tracer = trace.NewMulti(ag.Tracer, mem)
                       }
               }
               mux.Handle("/traces", instrument("/traces", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                       _ = json.NewEncoder(w).Encode(mem.All())
               })))
	}
	register := func(path string, h http.HandlerFunc) {
		if metrics {
			mux.Handle(path, instrument(path, h))
		} else {
			mux.HandleFunc(path, h)
		}
	}
       register("/agents", func(w http.ResponseWriter, r *http.Request) {
               list := make([]string, 0, len(agents))
               for id := range agents {
                       list = append(list, id)
               }
               sort.Strings(list)
               _ = json.NewEncoder(w).Encode(list)
       })
       register("/token_usage", func(w http.ResponseWriter, r *http.Request) {
               _ = json.NewEncoder(w).Encode(metricMap("agentry_tokens_total"))
       })
       register("/agent_health", func(w http.ResponseWriter, r *http.Request) {
               _ = json.NewEncoder(w).Encode(metricMap("agentry_agent_up"))
       })
       if metrics {
               mux.Handle("/", instrument("/", http.FileServer(http.FS(ui.WebUI))))
       } else {
               mux.Handle("/", http.FileServer(http.FS(ui.WebUI)))
       }

	// NATS queue setup (URL/subject could be from config/env)
	q, err := taskqueue.NewQueue(natsURL(), "agentry.tasks")
	if err != nil {
		panic("NATS unavailable: " + err.Error())
	}

	register("/spawn", func(w http.ResponseWriter, r *http.Request) {
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
               agentUp.WithLabelValues(id).Set(1)
               _ = json.NewEncoder(w).Encode(map[string]string{"agent_id": id})
       })
       register("/kill", func(w http.ResponseWriter, r *http.Request) {
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
               agentUp.WithLabelValues(in.AgentID).Set(0)
               _ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
       })
	register("/invoke", func(w http.ResponseWriter, r *http.Request) {
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
