package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func parseSuperPaymentAuditFilters(r *http.Request) (dbpkg.SuperPaymentAuditFilters, string) {
	filters := dbpkg.SuperPaymentAuditFilters{Limit: 50}
	query := r.URL.Query()
	filters.Provider = strings.ToLower(strings.TrimSpace(query.Get("provider")))
	if filters.Provider != "" && filters.Provider != "wompi" && filters.Provider != "epayco" {
		return filters, "proveedor invalido"
	}
	filters.Status = strings.TrimSpace(query.Get("status"))
	if len(filters.Status) > 40 {
		return filters, "estado invalido"
	}
	filters.Search = strings.TrimSpace(query.Get("q"))
	if len(filters.Search) > 100 {
		return filters, "busqueda demasiado larga"
	}
	if raw := strings.TrimSpace(query.Get("empresa_id")); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || value <= 0 {
			return filters, "empresa_id invalido"
		}
		filters.EmpresaID = value
	}
	if raw := strings.TrimSpace(query.Get("limit")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 200 {
			return filters, "limit invalido"
		}
		filters.Limit = value
	}
	if raw := strings.TrimSpace(query.Get("offset")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 0 || value > 1000000 {
			return filters, "offset invalido"
		}
		filters.Offset = value
	}
	return filters, ""
}

// SuperPaymentAuditHandler exposes a read-only, sanitized operational ledger.
// The route must be wrapped with WithSuperAuditoria in main.go.
func SuperPaymentAuditHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		filters, validationError := parseSuperPaymentAuditFilters(r)
		if validationError != "" {
			http.Error(w, validationError, http.StatusBadRequest)
			return
		}
		result, err := dbpkg.ListSuperPaymentAudit(dbSuper, filters)
		if err != nil {
			http.Error(w, "no se pudo consultar la auditoria de pagos", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(result)
	}
}
