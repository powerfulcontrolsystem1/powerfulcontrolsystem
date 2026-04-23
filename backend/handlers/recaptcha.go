package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	superRecaptchaEnabledConfigKey   = "security.recaptcha.enabled"
	superRecaptchaSiteKeyConfigKey   = "security.recaptcha.site_key"
	superRecaptchaSecretKeyConfigKey = "security.recaptcha.secret_key"
	superRecaptchaProviderConfigKey  = "security.recaptcha.provider"
)

type recaptchaValidationError struct {
	Status  int
	Message string
	Detail  string
}

func (e recaptchaValidationError) Error() string {
	if strings.TrimSpace(e.Message) != "" {
		return e.Message
	}
	return "La verificacion de seguridad no se pudo completar."
}

type recaptchaSiteVerifyResponse struct {
	Success    bool     `json:"success"`
	Hostname   string   `json:"hostname"`
	ErrorCodes []string `json:"error-codes"`
}

func parseTruthyConfigValue(raw string, defaultValue bool) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "on", "activo", "enabled", "si", "yes":
		return true
	case "0", "false", "off", "inactivo", "disabled", "no":
		return false
	default:
		return defaultValue
	}
}

func recaptchaEnvValue(keys ...string) string {
	for _, key := range keys {
		value := strings.TrimSpace(os.Getenv(key))
		if value != "" {
			return value
		}
	}
	return ""
}

func recaptchaConfigValue(dbSuper *sql.DB, configKey string) string {
	if dbSuper == nil {
		return ""
	}
	value, err := getDecryptedConfigValue(dbSuper, configKey)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func recaptchaSiteKey(dbSuper *sql.DB) string {
	if value := recaptchaConfigValue(dbSuper, superRecaptchaSiteKeyConfigKey); value != "" {
		return value
	}
	return recaptchaEnvValue("GOOGLE_RECAPTCHA_SITE_KEY", "RECAPTCHA_SITE_KEY")
}

func recaptchaSecretKey(dbSuper *sql.DB) string {
	if value := recaptchaConfigValue(dbSuper, superRecaptchaSecretKeyConfigKey); value != "" {
		return value
	}
	return recaptchaEnvValue("GOOGLE_RECAPTCHA_SECRET_KEY", "RECAPTCHA_SECRET_KEY")
}

func normalizeRecaptchaProvider(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "v3", "google-recaptcha-v3", "google_recaptcha_v3", "recaptcha-v3", "recaptcha_v3":
		return "google-recaptcha-v3"
	case "enterprise", "google-recaptcha-enterprise", "google_recaptcha_enterprise", "recaptcha-enterprise", "recaptcha_enterprise":
		return "google-recaptcha-enterprise"
	case "", "v2", "google-recaptcha-v2", "google_recaptcha_v2", "recaptcha-v2", "recaptcha_v2":
		return "google-recaptcha-v2"
	default:
		// Mantener compatibilidad: ante valores desconocidos, caer a v2.
		return "google-recaptcha-v2"
	}
}

func recaptchaProvider(dbSuper *sql.DB) string {
	if value := recaptchaConfigValue(dbSuper, superRecaptchaProviderConfigKey); value != "" {
		return normalizeRecaptchaProvider(value)
	}
	if value := recaptchaEnvValue("RECAPTCHA_PROVIDER", "GOOGLE_RECAPTCHA_PROVIDER"); value != "" {
		return normalizeRecaptchaProvider(value)
	}
	return "google-recaptcha-v2"
}

func recaptchaDevBypassEnabled() bool {
	return parseTruthyConfigValue(recaptchaEnvValue("RECAPTCHA_DEV_BYPASS"), false)
}

func isRecaptchaFeatureEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return false
	}
	value, err := getDecryptedConfigValue(dbSuper, superRecaptchaEnabledConfigKey)
	if err != nil {
		return false
	}
	return parseTruthyConfigValue(value, false)
}

func isRecaptchaConfigured(dbSuper *sql.DB) bool {
	return recaptchaSiteKey(dbSuper) != "" && recaptchaSecretKey(dbSuper) != ""
}

func recaptchaStoredSecretKeyPresent(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return false
	}
	value, encrypted, err := dbpkg.GetConfigValue(dbSuper, superRecaptchaSecretKeyConfigKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false
		}
		return false
	}
	return strings.TrimSpace(value) != "" || encrypted
}

func publicRecaptchaConfig(dbSuper *sql.DB) map[string]interface{} {
	requestedEnabled := isRecaptchaFeatureEnabled(dbSuper)
	configured := isRecaptchaConfigured(dbSuper)
	siteKey := recaptchaSiteKey(dbSuper)
	secretKey := recaptchaSecretKey(dbSuper)
	enabled := requestedEnabled && configured
	provider := recaptchaProvider(dbSuper)

	// Timestamps (si existen) para trazabilidad en configuración avanzada.
	siteUpdatedAt := ""
	secretUpdatedAt := ""
	providerUpdatedAt := ""
	if dbSuper != nil {
		if _, _, _, fa, err := dbpkg.GetConfigEntry(dbSuper, superRecaptchaSiteKeyConfigKey); err == nil {
			siteUpdatedAt = strings.TrimSpace(fa)
		}
		if _, _, _, fa, err := dbpkg.GetConfigEntry(dbSuper, superRecaptchaSecretKeyConfigKey); err == nil {
			secretUpdatedAt = strings.TrimSpace(fa)
		}
		if _, _, _, fa, err := dbpkg.GetConfigEntry(dbSuper, superRecaptchaProviderConfigKey); err == nil {
			providerUpdatedAt = strings.TrimSpace(fa)
		}
	}

	message := "reCAPTCHA desactivado desde configuracion avanzada."
	if requestedEnabled && !configured {
		message = "reCAPTCHA esta activado, pero faltan credenciales validas (site key o secret key)."
	} else if enabled {
		message = "reCAPTCHA activo para formularios publicos protegidos."
	}
	return map[string]interface{}{
		"provider":           provider,
		"enabled":            enabled,
		"requested_enabled":  requestedEnabled,
		"configured":         configured,
		"site_key":           siteKey,
		"site_key_updated_at":   siteUpdatedAt,
		"secret_key_updated_at": secretUpdatedAt,
		"provider_updated_at":   providerUpdatedAt,
		"provider_present":   strings.TrimSpace(recaptchaConfigValue(dbSuper, superRecaptchaProviderConfigKey)) != "" || strings.TrimSpace(recaptchaEnvValue("RECAPTCHA_PROVIDER", "GOOGLE_RECAPTCHA_PROVIDER")) != "",
		"stored_site_key":    recaptchaConfigValue(dbSuper, superRecaptchaSiteKeyConfigKey),
		"site_key_present":   siteKey != "",
		"secret_key_present": secretKey != "",
		"stored_secret_key":  recaptchaStoredSecretKeyPresent(dbSuper),
		"dev_bypass":         recaptchaDevBypassEnabled(),
		"message":            message,
	}
}

func PublicConfigJSHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := publicRecaptchaConfig(dbSuper)
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		fmt.Fprintf(
			w,
			"window.RECAPTCHA_SITE_KEY = %q; window.RECAPTCHA_ENABLED = %t; window.RECAPTCHA_REQUESTED_ENABLED = %t; window.RECAPTCHA_CONFIGURED = %t; window.RECAPTCHA_DEV_BYPASS = %t; window.RECAPTCHA_PROVIDER = %q;",
			strings.TrimSpace(fmt.Sprint(cfg["site_key"])),
			cfg["enabled"].(bool),
			cfg["requested_enabled"].(bool),
			cfg["configured"].(bool),
			cfg["dev_bypass"].(bool),
			cfg["provider"].(string),
		)
	}
}

func RecaptchaConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":      true,
				"service": publicRecaptchaConfig(dbSuper),
			})
			return

		case http.MethodPut, http.MethodPost:
			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			if adminEmail == "" || adminEmail == "sistema" {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}

			var payload struct {
				Enabled   *bool   `json:"enabled"`
				Provider  *string `json:"provider"`
				SiteKey   *string `json:"site_key"`
				SecretKey *string `json:"secret_key"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}
			if payload.Enabled == nil && payload.Provider == nil && payload.SiteKey == nil && payload.SecretKey == nil {
				http.Error(w, "debe enviar enabled, provider, site_key o secret_key", http.StatusBadRequest)
				return
			}

			if payload.Enabled != nil {
				value := "0"
				if *payload.Enabled {
					value = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, superRecaptchaEnabledConfigKey, value, false); err != nil {
					http.Error(w, "failed to save "+superRecaptchaEnabledConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.SiteKey != nil {
				siteKey := strings.TrimSpace(*payload.SiteKey)
				if err := dbpkg.SetConfigValue(dbSuper, superRecaptchaSiteKeyConfigKey, siteKey, false); err != nil {
					http.Error(w, "failed to save "+superRecaptchaSiteKeyConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.Provider != nil {
				provider := normalizeRecaptchaProvider(strings.TrimSpace(*payload.Provider))
				if err := dbpkg.SetConfigValue(dbSuper, superRecaptchaProviderConfigKey, provider, false); err != nil {
					http.Error(w, "failed to save "+superRecaptchaProviderConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.SecretKey != nil {
				secretKey := strings.TrimSpace(*payload.SecretKey)
				if secretKey != "" {
					encryptedValue, err := utils.EncryptString(secretKey)
					if err != nil {
						http.Error(w, "failed to encrypt "+superRecaptchaSecretKeyConfigKey+": "+err.Error(), http.StatusInternalServerError)
						return
					}
					if err := dbpkg.SetConfigValue(dbSuper, superRecaptchaSecretKeyConfigKey, encryptedValue, true); err != nil {
						http.Error(w, "failed to save "+superRecaptchaSecretKeyConfigKey+": "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
			}
			for _, key := range []string{superRecaptchaEnabledConfigKey, superRecaptchaProviderConfigKey, superRecaptchaSiteKeyConfigKey, superRecaptchaSecretKeyConfigKey} {
				if err := dbpkg.SetConfigValue(dbSuper, key+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "failed to save updated_by for "+key+": "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":      true,
				"service": publicRecaptchaConfig(dbSuper),
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func validateRecaptchaToken(dbSuper *sql.DB, r *http.Request, token string) error {
	if !isRecaptchaFeatureEnabled(dbSuper) {
		return nil
	}
	if recaptchaDevBypassEnabled() {
		return nil
	}
	// Si la funcionalidad fue activada pero faltan credenciales, no bloquear el acceso:
	// el frontend no tendrá un widget operativo (RECAPTCHA_ENABLED=false) y, si aquí
	// exigiéramos token, romperíamos el login/reset/registro. La configuración avanzada
	// ya expone que está "requested_enabled" pero no "configured".
	if !isRecaptchaConfigured(dbSuper) {
		return nil
	}
	if strings.TrimSpace(token) == "" {
		return recaptchaValidationError{Status: http.StatusBadRequest, Message: "Completa la verificacion de seguridad para continuar."}
	}

	form := url.Values{}
	form.Set("secret", recaptchaSecretKey(dbSuper))
	form.Set("response", strings.TrimSpace(token))
	if host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil && strings.TrimSpace(host) != "" {
		form.Set("remoteip", host)
	} else if strings.TrimSpace(r.RemoteAddr) != "" {
		form.Set("remoteip", strings.TrimSpace(r.RemoteAddr))
	}

	req, err := http.NewRequest(http.MethodPost, "https://www.google.com/recaptcha/api/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return recaptchaValidationError{Status: http.StatusBadGateway, Message: "No se pudo construir la verificacion de seguridad.", Detail: err.Error()}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return recaptchaValidationError{Status: http.StatusBadGateway, Message: "No se pudo validar la verificacion de seguridad con Google.", Detail: err.Error()}
	}
	defer resp.Body.Close()

	var payload recaptchaSiteVerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return recaptchaValidationError{Status: http.StatusBadGateway, Message: "Google devolvio una respuesta invalida durante la verificacion.", Detail: err.Error()}
	}
	if !payload.Success {
		message := "La verificacion de seguridad no fue valida. Intenta nuevamente."
		if len(payload.ErrorCodes) > 0 {
			switch strings.ToLower(strings.TrimSpace(payload.ErrorCodes[0])) {
			case "timeout-or-duplicate":
				message = "La verificacion de seguridad expiro o ya fue usada. Intenta nuevamente."
			case "missing-input-response":
				message = "Completa la verificacion de seguridad para continuar."
			}
		}
		return recaptchaValidationError{Status: http.StatusForbidden, Message: message, Detail: strings.Join(payload.ErrorCodes, ",")}
	}
	return nil
}

func writeRecaptchaValidationError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if recaptchaErr, ok := err.(recaptchaValidationError); ok {
		writeJSON(w, recaptchaErr.Status, map[string]interface{}{"ok": false, "message": recaptchaErr.Message})
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]interface{}{"ok": false, "message": "No se pudo validar la verificacion de seguridad."})
}
