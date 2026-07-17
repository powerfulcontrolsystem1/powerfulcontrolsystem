package handlers

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
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
		if err := dbpkg.VerifyPlatformMigrations(ctx, dbEmpresas, dbpkg.MigrationTargetEmpresas); err != nil {
			runtimeProbeJSON(w, r, http.StatusServiceUnavailable, `{"status":"not_ready"}`)
			return
		}
		if err := dbpkg.VerifyPlatformMigrations(ctx, dbSuper, dbpkg.MigrationTargetSuper); err != nil {
			runtimeProbeJSON(w, r, http.StatusServiceUnavailable, `{"status":"not_ready"}`)
			return
		}
		if err := runtimePrivateStorageReady(); err != nil {
			runtimeProbeJSON(w, r, http.StatusServiceUnavailable, `{"status":"not_ready"}`)
			return
		}

		runtimeProbeJSON(w, r, http.StatusOK, `{"status":"ready"}`)
	}
}

func runtimePrivateStorageReady() error {
	root := strings.TrimSpace(os.Getenv("PCS_PRIVATE_STORAGE_DIR"))
	if root == "" {
		root = filepath.Join(resolveProjectRootDir(), "private_storage")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	probe, err := os.CreateTemp(root, ".pcs-ready-*")
	if err != nil {
		return err
	}
	path := probe.Name()
	if err := probe.Chmod(0o600); err != nil {
		_ = probe.Close()
		_ = os.Remove(path)
		return err
	}
	if _, err := probe.Write([]byte("ready")); err != nil {
		_ = probe.Close()
		_ = os.Remove(path)
		return err
	}
	if err := probe.Close(); err != nil {
		_ = os.Remove(path)
		return err
	}
	return os.Remove(path)
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
