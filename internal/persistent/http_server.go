package persistent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/marcodenic/agentry/internal/registry"
)

// startAgentServer starts the HTTP server for an agent
func (pt *PersistentTeam) startAgentServer(agent *PersistentAgent) error {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		agent.mutex.RLock()
		status := agent.Status
		agent.mutex.RUnlock()

		response := map[string]interface{}{
			"status":    status,
			"agent_id":  agent.ID,
			"uptime":    time.Since(agent.StartedAt).Seconds(),
			"last_seen": agent.LastSeen,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Message endpoint - processes tasks through agent
	mux.HandleFunc("/message", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Update last seen and status
		agent.UpdateLastSeen()
		agent.SetStatus(registry.StatusWorking)

		// Parse request body
		var msgRequest struct {
			Input    string            `json:"input"`
			From     string            `json:"from,omitempty"`
			TaskID   string            `json:"task_id,omitempty"`
			Metadata map[string]string `json:"metadata,omitempty"`
		}
		
		if err := json.NewDecoder(r.Body).Decode(&msgRequest); err != nil {
			agent.SetStatus(registry.StatusIdle)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Execute task through the session-aware agent if available, otherwise use regular agent
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()
		
		// Create team context for the agent (required for agent.Run())
		teamCtx := context.WithValue(ctx, "team", pt)
		
		var result string
		var err error
		
		// Use session-aware execution if available
		if agent.SessionAgent != nil && agent.CurrentSession != nil {
			result, err = agent.SessionAgent.RunWithSession(teamCtx, msgRequest.Input)
		} else {
			result, err = agent.Agent.Run(teamCtx, msgRequest.Input)
		}
		
		// Update status back to idle
		agent.SetStatus(registry.StatusIdle)

		if err != nil {
			response := map[string]interface{}{
				"status":   "error",
				"error":    err.Error(),
				"agent_id": agent.ID,
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := map[string]interface{}{
			"status":   "success",
			"result":   result,
			"agent_id": agent.ID,
			"task_id":  msgRequest.TaskID,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Add session management endpoints
	pt.addSessionEndpoints(mux, agent)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", agent.Port),
		Handler: mux,
	}

	agent.Server = server

	// Start server in background
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Agent %s server error: %v\n", agent.ID, err)
		}
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

// addSessionEndpoints adds session management endpoints to the HTTP mux
func (pt *PersistentTeam) addSessionEndpoints(mux *http.ServeMux, agent *PersistentAgent) {
	// Session management endpoints
	mux.HandleFunc("/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// List sessions
			sessions, err := agent.ListSessions(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(sessions)
			
		case "POST":
			// Create new session
			var req struct {
				Name        string `json:"name"`
				Description string `json:"description"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
			
			session, err := agent.CreateSession(r.Context(), req.Name, req.Description)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(session)
			
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/sessions/", func(w http.ResponseWriter, r *http.Request) {
		// Extract session ID from path
		sessionID := r.URL.Path[len("/sessions/"):]
		if sessionID == "" {
			http.Error(w, "Session ID required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case "POST":
			// Load/resume session
			err := agent.LoadSession(r.Context(), sessionID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":     "session loaded",
				"session_id": sessionID,
			})
			
		case "DELETE":
			// Terminate session
			err := agent.TerminateCurrentSession(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"status":     "session terminated",
				"session_id": sessionID,
			})
			
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/sessions/current", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		sessionInfo := agent.GetCurrentSessionInfo()
		if sessionInfo == nil {
			http.Error(w, "No active session", http.StatusNotFound)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessionInfo)
	})
}

// stopAgent stops a single agent and cleans up its resources
func (pt *PersistentTeam) stopAgent(agent *PersistentAgent) {
	agent.mutex.Lock()
	defer agent.mutex.Unlock()

	agent.Status = registry.StatusStopping

	if agent.Server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		agent.Server.Shutdown(ctx)
	}

	agent.Status = registry.StatusStopped
}
