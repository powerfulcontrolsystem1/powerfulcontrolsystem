package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	nextcloudEnabledConfigKey      = "nextcloud.enabled"
	nextcloudBaseURLConfigKey      = "nextcloud.base_url" // ej: https://nextcloud.powerfulcontrolsystem.com
	nextcloudAdminUserConfigKey    = "nextcloud.admin_user"
	nextcloudAdminSecretConfigKey  = "nextcloud.admin_secret" // cifrado (password o app token)
	defaultNextcloudQuotaGBPerEmpresa = int64(1)
)

func isNextcloudEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return false
	}
	value, err := getDecryptedConfigValue(dbSuper, nextcloudEnabledConfigKey)
	if err != nil {
		value, _, _, _, _ = dbpkg.GetConfigEntry(dbSuper, nextcloudEnabledConfigKey)
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "on", "activo", "enabled":
		return true
	default:
		return false
	}
}

func nextcloudResolveBaseURL(dbSuper *sql.DB) (string, error) {
	v, err := getDecryptedConfigValue(dbSuper, nextcloudBaseURLConfigKey)
	if err != nil || strings.TrimSpace(v) == "" {
		v, _, _, _, _ = dbpkg.GetConfigEntry(dbSuper, nextcloudBaseURLConfigKey)
	}
	v = strings.TrimRight(strings.TrimSpace(v), "/")
	return v, nil
}

func nextcloudResolveAdminUser(dbSuper *sql.DB) (string, error) {
	v, err := getDecryptedConfigValue(dbSuper, nextcloudAdminUserConfigKey)
	if err != nil || strings.TrimSpace(v) == "" {
		v, _, _, _, _ = dbpkg.GetConfigEntry(dbSuper, nextcloudAdminUserConfigKey)
	}
	return strings.TrimSpace(v), nil
}

func nextcloudResolveAdminSecret(dbSuper *sql.DB) (string, error) {
	v, err := getDecryptedConfigValue(dbSuper, nextcloudAdminSecretConfigKey)
	if err == nil && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v), nil
	}
	// fallback: allow env for VPS (no persist)
	if env := strings.TrimSpace(os.Getenv("NEXTCLOUD_ADMIN_SECRET")); env != "" {
		return env, nil
	}
	return strings.TrimSpace(v), nil
}

func nextcloudGeneratePassword() (string, error) {
	b := make([]byte, 18)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// sin caracteres raros para compatibilidad de UI
	return strings.TrimRight(base64.RawURLEncoding.EncodeToString(b), "="), nil
}

// NextcloudConfigHandler (super) gestiona activación y credenciales del servicio Nextcloud.
// Persistencia: tabla configuraciones en pcs_superadministrador.
func NextcloudConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			baseURL, _ := nextcloudResolveBaseURL(dbSuper)
			adminUser, _ := nextcloudResolveAdminUser(dbSuper)
			_, secretEncrypted, _, secretUpdated, _ := dbpkg.GetConfigEntry(dbSuper, nextcloudAdminSecretConfigKey)
			enabledRaw, _, _, enabledUpdated, _ := dbpkg.GetConfigEntry(dbSuper, nextcloudEnabledConfigKey)

			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true,
				"enabled": isNextcloudEnabled(dbSuper),
				"enabled_configured": strings.TrimSpace(enabledRaw) != "",
				"enabled_updated": enabledUpdated,
				"base_url": baseURL,
				"admin_user": strings.TrimSpace(adminUser),
				"admin_secret_set": strings.TrimSpace(secretUpdated) != "" || secretEncrypted,
				"admin_secret_encrypted": secretEncrypted,
				"admin_secret_updated": secretUpdated,
				"encryption_available": utils.EncryptionAvailable(),
			})
			return

		case http.MethodPut, http.MethodPost:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "test") {
				baseURL, _ := nextcloudResolveBaseURL(dbSuper)
				if baseURL == "" {
					http.Error(w, "nextcloud.base_url vacío", http.StatusBadRequest)
					return
				}
				client := &http.Client{Timeout: 12 * time.Second}
				req, _ := http.NewRequest(http.MethodGet, baseURL+"/status.php", nil)
				res, err := client.Do(req)
				if err != nil || res == nil {
					http.Error(w, "No se pudo conectar a Nextcloud: "+fmt.Sprint(err), http.StatusBadGateway)
					return
				}
				defer res.Body.Close()
				ok := res.StatusCode >= 200 && res.StatusCode <= 299
				writeJSON(w, http.StatusOK, map[string]any{
					"ok": ok,
					"status": res.StatusCode,
					"status_url": baseURL + "/status.php",
				})
				return
			}

			var payload struct {
				Enabled  *bool  `json:"enabled"`
				BaseURL  string `json:"base_url"`
				AdminUser string `json:"admin_user"`
				AdminSecret string `json:"admin_secret"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload inválido: "+err.Error(), http.StatusBadRequest)
				return
			}

			if payload.Enabled != nil {
				v := "0"
				if *payload.Enabled {
					v = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, nextcloudEnabledConfigKey, v, false); err != nil {
					http.Error(w, "No se pudo guardar nextcloud.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if strings.TrimSpace(payload.BaseURL) != "" {
				u := strings.TrimRight(strings.TrimSpace(payload.BaseURL), "/")
				if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
					http.Error(w, "base_url inválida: debe iniciar con http:// o https://", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, nextcloudBaseURLConfigKey, u, false); err != nil {
					http.Error(w, "No se pudo guardar nextcloud.base_url: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if strings.TrimSpace(payload.AdminUser) != "" {
				if err := dbpkg.SetConfigValue(dbSuper, nextcloudAdminUserConfigKey, strings.TrimSpace(payload.AdminUser), false); err != nil {
					http.Error(w, "No se pudo guardar nextcloud.admin_user: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			if strings.TrimSpace(payload.AdminSecret) != "" {
				if !utils.EncryptionAvailable() {
					http.Error(w, "Cifrado requerido: CONFIG_ENC_KEY no está disponible", http.StatusInternalServerError)
					return
				}
				encVal, err := utils.EncryptString(strings.TrimSpace(payload.AdminSecret))
				if err != nil {
					http.Error(w, "No se pudo cifrar admin_secret: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, nextcloudAdminSecretConfigKey, encVal, true); err != nil {
					http.Error(w, "No se pudo guardar nextcloud.admin_secret: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

