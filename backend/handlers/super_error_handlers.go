package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

func parseSuperErroresInt(raw string, fallback int) int {
	v, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fallback
	}
	return v
}

func parseSuperErroresInt64(raw string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}

// SuperErroresSistemaHandler expone el monitor centralizado de errores para el panel super.
func SuperErroresSistemaHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbSuper == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "conexion de base de datos super no disponible",
			})
			return
		}

		filter := dbpkg.SuperErrorSistemaFiltro{
			EmpresaID: parseSuperErroresInt64(r.URL.Query().Get("empresa_id")),
			Nivel:     strings.TrimSpace(r.URL.Query().Get("nivel")),
			TipoError: strings.TrimSpace(r.URL.Query().Get("tipo_error")),
			Desde:     strings.TrimSpace(r.URL.Query().Get("desde")),
			Hasta:     strings.TrimSpace(r.URL.Query().Get("hasta")),
			Search:    strings.TrimSpace(r.URL.Query().Get("search")),
			Limit:     parseSuperErroresInt(r.URL.Query().Get("limit"), 50),
			Offset:    parseSuperErroresInt(r.URL.Query().Get("offset"), 0),
		}

		items, total, summary, err := dbpkg.ListSuperErroresSistema(dbSuper, filter)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
				"ok":    false,
				"error": "no se pudo consultar el monitor de errores del sistema",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":          true,
			"items":       items,
			"total":       total,
			"summary":     summary,
			"limit":       filter.Limit,
			"offset":      filter.Offset,
			"has_more":    int64(filter.Offset+len(items)) < total,
			"admin_email": adminEmail,
			"filters": map[string]interface{}{
				"empresa_id": filter.EmpresaID,
				"nivel":      filter.Nivel,
				"tipo_error": filter.TipoError,
				"desde":      filter.Desde,
				"hasta":      filter.Hasta,
				"search":     filter.Search,
			},
		})
	}
}