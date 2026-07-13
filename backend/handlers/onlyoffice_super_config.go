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
	onlyOfficeEnabledConfigKey = "onlyoffice.enabled"
)

func isOnlyOfficeEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return true
	}
	value, err := getDecryptedConfigValue(dbSuper, onlyOfficeEnabledConfigKey)
	if err != nil {
		return true
	}
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "1", "true", "on", "activo", "enabled":
		return true
	case "0", "false", "off", "inactivo", "disabled":
		return false
	default:
		return true
	}
}

func onlyOfficeResolveJWTSecret(dbSuper *sql.DB) (string, bool, error) {
	// Prioridad:
	// 1) ENV ONLYOFFICE_JWT_SECRET (para automatizar despliegue VPS)
	// 2) DB (cifrado)
	// Si viene de ENV y encryption está disponible, se persiste/actualiza en DB para evitar desalineación (DS vs backend).
	env := strings.TrimSpace(os.Getenv("ONLYOFFICE_JWT_SECRET"))
	if env != "" {
		if dbSuper != nil && utils.EncryptionAvailable() {
			dbVal, err := getDecryptedConfigValue(dbSuper, onlyOfficeConfigKeyJWT)
			if err != nil || strings.TrimSpace(dbVal) == "" || strings.TrimSpace(dbVal) != env {
				if encVal, encErr := utils.EncryptString(env); encErr == nil {
					_ = dbpkg.SetConfigValue(dbSuper, onlyOfficeConfigKeyJWT, encVal, true)
				}
			}
		}
		return env, false, nil
	}

	if dbSuper != nil {
		v, err := getDecryptedConfigValue(dbSuper, onlyOfficeConfigKeyJWT)
		if err == nil && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v), true, nil
		}
	}

	return "", false, nil
}

func onlyOfficeResolveDocumentServerURL(dbSuper *sql.DB) (string, bool, error) {
	if dbSuper != nil {
		v, err := getDecryptedConfigValue(dbSuper, onlyOfficeConfigKeyDSURL)
		if err == nil && strings.TrimSpace(v) != "" {
			return strings.TrimRight(strings.TrimSpace(v), "/"), true, nil
		}
	}
	env := strings.TrimSpace(os.Getenv("ONLYOFFICE_DOCUMENT_SERVER_URL"))
	if env == "" {
		return "", false, nil
	}
	env = strings.TrimRight(env, "/")
	if dbSuper != nil {
		_ = dbpkg.SetConfigValue(dbSuper, onlyOfficeConfigKeyDSURL, env, false)
	}
	return env, false, nil
}

func onlyOfficeGenerateJWTSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func onlyOfficeEnsureJWTSecret(dbSuper *sql.DB) (bool, error) {
	secret, _, err := onlyOfficeResolveJWTSecret(dbSuper)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(secret) != "" {
		return false, nil
	}
	// Si no existe ni en DB ni en ENV, generamos uno nuevo y lo guardamos cifrado en DB.
	if dbSuper == nil {
		return false, nil
	}
	if !utils.EncryptionAvailable() {
		return false, fmt.Errorf("cifrado requerido: CONFIG_ENC_KEY no está disponible")
	}
	gen, err := onlyOfficeGenerateJWTSecret()
	if err != nil {
		return false, err
	}
	encVal, err := utils.EncryptString(gen)
	if err != nil {
		return false, err
	}
	if err := dbpkg.SetConfigValue(dbSuper, onlyOfficeConfigKeyJWT, encVal, true); err != nil {
		return false, err
	}
	// También dejamos el ENV en memoria del proceso (no persistente) para que el resto del runtime lo use.
	_ = os.Setenv("ONLYOFFICE_JWT_SECRET", gen)
	return true, nil
}

// OnlyOfficeConfigHandler gestiona el switch y credenciales globales de OnlyOffice (GET/PUT).
// - onlyoffice.enabled: 1/0 (por defecto true si no existe)
// - onlyoffice.document_server_url: URL base del Document Server (https://onlyoffice.dominio)
// - onlyoffice.jwt_secret: secreto HS256 (debe estar cifrado en DB)
func OnlyOfficeConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = onlyOfficeEnsureJWTSecret(dbSuper)
			_, _, _ = onlyOfficeResolveDocumentServerURL(dbSuper)
			enabledRaw, _, _, enabledUpdated, _ := dbpkg.GetConfigEntry(dbSuper, onlyOfficeEnabledConfigKey)
			dsURL, _, _, dsUpdated, _ := dbpkg.GetConfigEntry(dbSuper, onlyOfficeConfigKeyDSURL)
			secretRaw, secretEncrypted, _, secretUpdated, _ := dbpkg.GetConfigEntry(dbSuper, onlyOfficeConfigKeyJWT)

			enabled := isOnlyOfficeEnabled(dbSuper)
			enabledConfigured := strings.TrimSpace(enabledRaw) != ""

			w.Header().Set("Content-Type", "application/json")
			encodeJSONResponse(w, map[string]interface{}{
				"enabled":                 enabled,
				"enabled_configured":      enabledConfigured,
				"enabled_updated":         enabledUpdated,
				"document_server_url":     strings.TrimSpace(dsURL),
				"document_server_url_set": strings.TrimSpace(dsURL) != "",
				"document_server_updated": dsUpdated,
				"jwt_secret_set":          strings.TrimSpace(secretRaw) != "",
				"jwt_secret_encrypted":    secretEncrypted,
				"jwt_secret_updated":      secretUpdated,
				"encryption_available":    utils.EncryptionAvailable(),
			})
			return

		case http.MethodPut, http.MethodPost:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "test") {
				// Probar /healthcheck del Document Server.
				dsURL, err := getDecryptedConfigValue(dbSuper, onlyOfficeConfigKeyDSURL)
				if err != nil {
					http.Error(w, "No se pudo leer onlyoffice.document_server_url", http.StatusInternalServerError)
					return
				}
				dsURL = strings.TrimRight(strings.TrimSpace(dsURL), "/")
				if dsURL == "" {
					http.Error(w, "OnlyOffice no configurado (document_server_url vacío)", http.StatusBadRequest)
					return
				}
				client := &http.Client{Timeout: 12 * time.Second}
				req, _ := http.NewRequest(http.MethodGet, dsURL+"/healthcheck", nil)
				res, err := client.Do(req)
				if err != nil || res == nil {
					http.Error(w, "No se pudo conectar a OnlyOffice: "+fmt.Sprint(err), http.StatusBadGateway)
					return
				}
				defer res.Body.Close()
				ok := res.StatusCode >= 200 && res.StatusCode <= 299
				w.Header().Set("Content-Type", "application/json")
				if !ok {
					w.WriteHeader(http.StatusBadGateway)
				}
				encodeJSONResponse(w, map[string]interface{}{
					"ok":         ok,
					"status":     res.StatusCode,
					"health_url": dsURL + "/healthcheck",
				})
				return
			}

			var payload struct {
				Enabled           *bool  `json:"enabled"`
				DocumentServerURL string `json:"document_server_url"`
				JWTSecret         string `json:"jwt_secret"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload inválido: "+err.Error(), http.StatusBadRequest)
				return
			}

			payload.DocumentServerURL = strings.TrimSpace(payload.DocumentServerURL)
			payload.JWTSecret = strings.TrimSpace(payload.JWTSecret)

			if payload.Enabled == nil && payload.DocumentServerURL == "" && payload.JWTSecret == "" {
				http.Error(w, "Debes enviar enabled o document_server_url o jwt_secret", http.StatusBadRequest)
				return
			}

			if payload.Enabled != nil {
				v := "0"
				if *payload.Enabled {
					v = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, onlyOfficeEnabledConfigKey, v, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion de OnlyOffice", http.StatusInternalServerError)
					return
				}
			}

			if payload.DocumentServerURL != "" {
				// Normalizar sin slash final.
				payload.DocumentServerURL = strings.TrimRight(payload.DocumentServerURL, "/")
				if !strings.HasPrefix(payload.DocumentServerURL, "http://") && !strings.HasPrefix(payload.DocumentServerURL, "https://") {
					http.Error(w, "document_server_url inválida: debe iniciar con http:// o https://", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, onlyOfficeConfigKeyDSURL, payload.DocumentServerURL, false); err != nil {
					http.Error(w, "No se pudo guardar la configuracion de OnlyOffice", http.StatusInternalServerError)
					return
				}
			}

			if payload.JWTSecret != "" {
				if !utils.EncryptionAvailable() {
					http.Error(w, "El cifrado no esta disponible para guardar la clave de OnlyOffice", http.StatusInternalServerError)
					return
				}
				encVal, err := utils.EncryptString(payload.JWTSecret)
				if err != nil {
					http.Error(w, "No se pudo cifrar la clave de OnlyOffice", http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, onlyOfficeConfigKeyJWT, encVal, true); err != nil {
					http.Error(w, "No se pudo guardar la clave de OnlyOffice", http.StatusInternalServerError)
					return
				}
			}

			// Validación básica: si enabled=true, permitir guardar aunque falten credenciales; el UI mostrará pendientes.
			w.Header().Set("Content-Type", "application/json")
			encodeJSONResponse(w, map[string]interface{}{"ok": true})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
