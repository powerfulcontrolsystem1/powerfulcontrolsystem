package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	dbpkg "github.com/you/pos-backend/db"
)

// MetricsCurrentHandler devuelve la última muestra de métricas
func MetricsCurrentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := dbpkg.GetLatestMetric(db)
		if err != nil {
			http.Error(w, "failed to get latest metric: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(m)
	}
}

// MetricsHistoryHandler devuelve histórico de métricas (limit opcional)
func MetricsHistoryHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		limit := 200
		if lstr := q.Get("limit"); lstr != "" {
			if l, err := strconv.Atoi(lstr); err == nil && l > 0 && l <= 5000 {
				limit = l
			}
		}
		hist, err := dbpkg.GetMetricsHistory(db, limit)
		if err != nil {
			http.Error(w, "failed to get metrics history: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(hist)
	}
}
