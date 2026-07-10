package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

type rustDeskConfigPayload struct {
	Enabled    bool   `json:"enabled"`
	Host       string `json:"host"`
	User       string `json:"user"`
	KeyPath    string `json:"key_path"`
	ServerHost string `json:"server_host"`
	ServerKey  string `json:"server_key"`
}

// RustDeskConfigHandler gestiona la configuración del control RustDesk en VPS por SSH (GET/PUT).
func RustDeskConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			enabledRaw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_enabled")
			if err != nil {
				log.Printf("[rustdesk] read enabled error: %v", err)
			}
			host, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_host")
			user, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_user")
			keyPath, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.vps_ssh_key_path")
			serverHost, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.server_host")
			serverKey, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, "rustdesk.server_key")

			enabled := false
			switch strings.ToLower(strings.TrimSpace(enabledRaw)) {
			case "1", "true", "on", "activo", "enabled":
				enabled = true
			}
			writeJSON(w, http.StatusOK, rustDeskConfigPayload{
				Enabled: enabled,
				Host:    strings.TrimSpace(host),
				User:    strings.TrimSpace(user),
				KeyPath: strings.TrimSpace(keyPath), ServerHost: strings.TrimSpace(serverHost), ServerKey: strings.TrimSpace(serverKey),
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload rustDeskConfigPayload
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "JSON inválido", http.StatusBadRequest)
				return
			}

			payload.Host = strings.TrimSpace(payload.Host)
			payload.User = strings.TrimSpace(payload.User)
			payload.KeyPath = strings.TrimSpace(payload.KeyPath)
			payload.ServerHost = strings.TrimSpace(payload.ServerHost)
			payload.ServerKey = strings.TrimSpace(payload.ServerKey)

			if payload.Enabled {
				if payload.Host == "" || payload.User == "" {
					http.Error(w, "Debe configurar host y usuario para activar el control por SSH.", http.StatusBadRequest)
					return
				}
			}

			enabledValue := "0"
			if payload.Enabled {
				enabledValue = "1"
			}
			if err := dbpkg.SetConfigValue(dbSuper, "rustdesk.vps_ssh_enabled", enabledValue, false); err != nil {
				http.Error(w, "No se pudo guardar rustdesk.vps_ssh_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := dbpkg.SetConfigValue(dbSuper, "rustdesk.vps_ssh_host", payload.Host, false); err != nil {
				http.Error(w, "No se pudo guardar rustdesk.vps_ssh_host: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := dbpkg.SetConfigValue(dbSuper, "rustdesk.vps_ssh_user", payload.User, false); err != nil {
				http.Error(w, "No se pudo guardar rustdesk.vps_ssh_user: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := dbpkg.SetConfigValue(dbSuper, "rustdesk.vps_ssh_key_path", payload.KeyPath, false); err != nil {
				http.Error(w, "No se pudo guardar rustdesk.vps_ssh_key_path: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := dbpkg.SetConfigValue(dbSuper, "rustdesk.server_host", payload.ServerHost, false); err != nil {
				http.Error(w, "No se pudo guardar rustdesk.server_host", http.StatusInternalServerError)
				return
			}
			if err := dbpkg.SetConfigValue(dbSuper, "rustdesk.server_key", payload.ServerKey, false); err != nil {
				http.Error(w, "No se pudo guardar rustdesk.server_key", http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true})
			return

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
