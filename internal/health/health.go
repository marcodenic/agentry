// Package health provides a basic HTTP health check endpoint.
//
// Usage example:
//   mux := http.NewServeMux()
//   health.Register(mux)
//   http.ListenAndServe(":8080", mux)
//   // GET/HEAD http://localhost:8080/healthz returns 200 OK with body "ok"
package health

import (
	"net/http"
)

// Handler returns an HTTP handler that responds with 200 OK and body "ok"
// for GET and HEAD requests. Other methods return 405 Method Not Allowed.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead:
			w.WriteHeader(http.StatusOK)
			if r.Method == http.MethodGet {
				w.Write([]byte("ok"))
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

// Register registers the health check handler on the provided ServeMux
// at the /healthz path.
func Register(mux *http.ServeMux) {
	mux.HandleFunc("/healthz", Handler())
}