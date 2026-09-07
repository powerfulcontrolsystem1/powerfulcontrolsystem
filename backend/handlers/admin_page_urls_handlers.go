package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const superAdminPageURLsEnabledConfigKey = "ui.admin_page_urls.enabled"

func AdminPageURLsGlobalConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		if r.Method == http.MethodGet {
			raw, _, _, updatedAt, _ := dbpkg.GetConfigEntry(dbSuper, superAdminPageURLsEnabledConfigKey)
			updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, superAdminPageURLsEnabledConfigKey+".updated_by")
			enabled := parseTruthyConfigValue(raw, false)
			writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
				"ok":          true,
				"enabled":     enabled,
				"updated_at":  updatedAt,
				"updated_by":  strings.TrimSpace(updatedBy),
				"config_key":  superAdminPageURLsEnabledConfigKey,
				"description": "Cuando esta activo, Administrar empresa refleja en la barra del navegador la URL real de la subpagina abierta sin perder el shell al recargar.",
			})
			return
		}

		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			writeAdminAuthError(w, http.StatusMethodNotAllowed, "Metodo no permitido.")
			return
		}

		adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
		if adminEmail == "" || adminEmail == "sistema" {
			writeAdminAuthError(w, http.StatusUnauthorized, "Sesion no autenticada.")
			return
		}

		var payload struct {
			Enabled *bool `json:"enabled"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeAdminAuthError(w, http.StatusBadRequest, "Solicitud de configuracion de URLs invalida.")
			return
		}
		if payload.Enabled == nil {
			writeAdminAuthError(w, http.StatusBadRequest, "Debes enviar enabled.")
			return
		}

		value := "0"
		if *payload.Enabled {
			value = "1"
		}
		if err := dbpkg.SetConfigValue(dbSuper, superAdminPageURLsEnabledConfigKey, value, false); err != nil {
			writeAdminAuthError(w, http.StatusInternalServerError, "No se pudo guardar la configuracion global de URLs.")
			return
		}
		_ = dbpkg.SetConfigValue(dbSuper, superAdminPageURLsEnabledConfigKey+".updated_by", adminEmail, false)
		writeAdminAuthJSON(w, http.StatusOK, map[string]interface{}{
			"ok":      true,
			"enabled": *payload.Enabled,
		})
	}
}
