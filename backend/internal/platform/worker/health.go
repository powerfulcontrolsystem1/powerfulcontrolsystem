package worker

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const workerHealthProbeTimeout = 3 * time.Second

// HealthState carries only process-level status. It deliberately excludes job
// payloads, tenant identifiers and provider errors from the health surface.
type HealthState struct {
	mu          sync.RWMutex
	ready       bool
	lastSuccess time.Time
	lastFailure time.Time
}

func (s *HealthState) MarkBatchSuccess(at time.Time) {
	if s == nil {
		return
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	s.mu.Lock()
	s.ready = true
	s.lastSuccess = at.UTC()
	s.mu.Unlock()
}

func (s *HealthState) MarkBatchFailure(at time.Time) {
	if s == nil {
		return
	}
	if at.IsZero() {
		at = time.Now().UTC()
	}
	s.mu.Lock()
	s.ready = false
	s.lastFailure = at.UTC()
	s.mu.Unlock()
}

func (s *HealthState) Ready() bool {
	if s == nil {
		return false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ready && !s.lastSuccess.IsZero()
}

// StartHealthServer binds only to loopback. It is intended for the Docker
// healthcheck, not a public status endpoint or an operational control plane.
func StartHealthServer(ctx context.Context, addr string, dbConn *sql.DB, state *HealthState) (<-chan error, error) {
	if ctx == nil {
		return nil, fmt.Errorf("worker health context is required")
	}
	if dbConn == nil || state == nil {
		return nil, fmt.Errorf("worker health requires database and state")
	}
	addr = strings.TrimSpace(addr)
	if err := validateLoopbackHealthAddr(addr); err != nil {
		return nil, err
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("listen worker health: %w", err)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if !workerHealthMethodAllowed(w, r) {
			return
		}
		workerHealthResponse(w, r, http.StatusOK, "ok")
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if !workerHealthMethodAllowed(w, r) {
			return
		}
		if !state.Ready() {
			workerHealthResponse(w, r, http.StatusServiceUnavailable, "not_ready")
			return
		}
		probeCtx, cancel := context.WithTimeout(r.Context(), workerHealthProbeTimeout)
		defer cancel()
		if err := dbConn.PingContext(probeCtx); err != nil {
			workerHealthResponse(w, r, http.StatusServiceUnavailable, "not_ready")
			return
		}
		workerHealthResponse(w, r, http.StatusOK, "ready")
	})
	server := &http.Server{Handler: mux, ReadHeaderTimeout: workerHealthProbeTimeout}
	errs := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			errs <- err
		}
		close(errs)
	}()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), workerHealthProbeTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	return errs, nil
}

func workerHealthMethodAllowed(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func validateLoopbackHealthAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || strings.TrimSpace(port) == "" {
		return fmt.Errorf("worker health address must include loopback host and port")
	}
	host = strings.TrimSpace(host)
	if strings.EqualFold(host, "localhost") {
		return nil
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return fmt.Errorf("worker health address must bind only to loopback")
	}
	return nil
}

func workerHealthResponse(w http.ResponseWriter, r *http.Request, status int, state string) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if r.Method == http.MethodHead {
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]string{"status": state})
}
