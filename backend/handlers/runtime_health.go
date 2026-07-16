package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

const runtimeProbeTimeout = 3 * time.Second

// RuntimeHealthHandler confirms that the HTTP process can accept requests.
// It deliberately does not depend on authentication or database availability.
func RuntimeHealthHandler(w http.ResponseWriter, r *http.Request) {
	if !runtimeProbeMethodAllowed(w, r) {
		return
	}
	runtimeProbeJSON(w, r, http.StatusOK, `{"status":"ok"}`)
}

// RuntimeReadyHandler confirms that the HTTP process and both application
// databases are available before a load balancer sends normal traffic.
func RuntimeReadyHandler(dbEmpresas, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !runtimeProbeMethodAllowed(w, r) {
			return
		}
		if dbEmpresas == nil || dbSuper == nil {
			runtimeProbeJSON(w, r, http.StatusServiceUnavailable, `{"status":"not_ready"}`)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), runtimeProbeTimeout)
		defer cancel()
		if err := dbEmpresas.PingContext(ctx); err != nil {
			runtimeProbeJSON(w, r, http.StatusServiceUnavailable, `{"status":"not_ready"}`)
			return
		}
		if err := dbSuper.PingContext(ctx); err != nil {
			runtimeProbeJSON(w, r, http.StatusServiceUnavailable, `{"status":"not_ready"}`)
			return
		}

		runtimeProbeJSON(w, r, http.StatusOK, `{"status":"ready"}`)
	}
}

func runtimeProbeMethodAllowed(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func runtimeProbeJSON(w http.ResponseWriter, r *http.Request, status int, body string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte(body))
	}
}
