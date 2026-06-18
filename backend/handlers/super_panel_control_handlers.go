package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

// SuperPanelControlResetHandler ejecuta limpiezas explicitas del centro de mando.
func SuperPanelControlResetHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
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

		action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
		switch action {
		case "metricas", "metrics", "reiniciar_metricas":
			affected, err := dbpkg.ResetMetricsHistory(dbSuper)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudieron reiniciar las metricas"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          true,
				"action":      "metricas",
				"affected":    affected,
				"admin_email": adminEmail,
				"description": "Metricas del panel reiniciadas",
			})
			return
		case "errores", "errors", "reiniciar_errores":
			affected, err := dbpkg.ResetSuperErroresSistema(dbSuper)
			if err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]interface{}{"ok": false, "error": "no se pudieron reiniciar los indicadores de errores"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          true,
				"action":      "errores",
				"affected":    affected,
				"admin_email": adminEmail,
				"description": "Indicadores de errores reiniciados",
			})
			return
		default:
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"ok":    false,
				"error": "action invalida; use metricas o errores",
			})
			return
		}
	}
}
