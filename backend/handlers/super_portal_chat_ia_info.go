package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	superPortalChatIAInfoKey          = "portal.chat_ia.info_text"
	superPortalChatIAInfoUpdatedByKey = "portal.chat_ia.info_text.updated_by"
)

// SuperPortalChatIAInfoHandler permite editar texto persistente que alimenta el chat público del portal.
// Persistencia: tabla configuraciones (pcs_superadministrador).
func SuperPortalChatIAInfoHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbSuper == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "db_super no disponible"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			raw, _, _, updatedAt, _ := dbpkg.GetConfigEntry(dbSuper, superPortalChatIAInfoKey)
			updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, superPortalChatIAInfoUpdatedByKey)
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":         true,
				"key":        superPortalChatIAInfoKey,
				"value":      strings.TrimSpace(raw),
				"updated_at": strings.TrimSpace(updatedAt),
				"updated_by": strings.TrimSpace(updatedBy),
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				Value string `json:"value"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload inválido", http.StatusBadRequest)
				return
			}
			// Guardamos texto tal cual (sin cifrar) — es contenido editorial, no secreto.
			if err := dbpkg.SetConfigValue(dbSuper, superPortalChatIAInfoKey, strings.TrimSpace(payload.Value), false); err != nil {
				http.Error(w, "No se pudo guardar: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superPortalChatIAInfoUpdatedByKey, strings.TrimSpace(adminEmail), false)
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "updated_at": time.Now().Format("2006-01-02 15:04:05")})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

