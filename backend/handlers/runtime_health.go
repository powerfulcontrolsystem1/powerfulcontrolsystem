package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// RuntimeHealthHandler reports that the HTTP process is alive. It deliberately
// does not disclose dependency names, connection strings, or internal errors.
func RuntimeHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}

// RuntimeReadyHandler verifies the API dependencies required for a request.
// It returns a generic response so unauthenticated health probes cannot learn
// database topology or error details.
func RuntimeReadyHandler(databases ...*sql.DB) http.HandlerFunc {
	return RuntimeReadyWithCheck(nil, databases...)
}

// RuntimeReadyWithCheck adds release-specific validation, such as the required
// schema migration. The check must return only an error; its details are never
// sent to an unauthenticated probe.
func RuntimeReadyWithCheck(check func(context.Context) error, databases ...*sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		ready := len(databases) > 0
		if check != nil && check(ctx) != nil {
			ready = false
		}
		for _, database := range databases {
			if database == nil || database.PingContext(ctx) != nil {
				ready = false
				break
			}
		}
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
			if r.Method != http.MethodHead {
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "unavailable"})
			}
			return
		}
		if r.Method == http.MethodHead {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
}
