package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const (
	voiceStreamEnabledKey    = "voice_stream.enabled"
	voiceStreamBaseURLKey    = "voice_stream.base_url"
	voiceStreamProviderKey   = "voice_stream.provider"
	voiceStreamVoiceKey      = "voice_stream.voice"
	voiceStreamTimeoutMSKey  = "voice_stream.timeout_ms"
	voiceStreamAuthTokenKey  = "voice_stream.auth_token"
	voiceStreamAuthHeaderKey = "voice_stream.auth_header"
)

type voiceStreamConfig struct {
	Enabled             bool   `json:"enabled"`
	BaseURL             string `json:"base_url"`
	Provider            string `json:"provider"`
	Voice               string `json:"voice"`
	TimeoutMS           int    `json:"timeout_ms"`
	AuthHeader          string `json:"auth_header"`
	AuthConfigured      bool   `json:"auth_configured"`
	AuthEncrypted       bool   `json:"auth_encrypted"`
	AuthUpdated         string `json:"auth_updated,omitempty"`
	EncryptionAvailable bool   `json:"encryption_available"`
}

func defaultVoiceStreamConfig() voiceStreamConfig {
	baseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("VOICE_STREAM_BASE_URL")), "/")
	if baseURL == "" {
		baseURL = "http://127.0.0.1:8097"
	}
	provider := strings.TrimSpace(os.Getenv("VOICE_STREAM_PROVIDER"))
	if provider == "" {
		provider = "piper"
	}
	voice := strings.TrimSpace(os.Getenv("VOICE_STREAM_VOICE"))
	if voice == "" {
		voice = "es-CO"
	}
	authHeader := strings.TrimSpace(os.Getenv("VOICE_STREAM_AUTH_HEADER"))
	if authHeader == "" {
		authHeader = "X-PCS-Voice-Token"
	}
	return voiceStreamConfig{
		Enabled:             false,
		BaseURL:             baseURL,
		Provider:            provider,
		Voice:               voice,
		TimeoutMS:           12000,
		AuthHeader:          authHeader,
		EncryptionAvailable: utils.EncryptionAvailable(),
	}
}

func resolveVoiceStreamConfig(dbSuper *sql.DB) voiceStreamConfig {
	cfg := defaultVoiceStreamConfig()
	if dbSuper == nil {
		return cfg
	}

	if raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamEnabledKey); err == nil {
		switch strings.ToLower(strings.TrimSpace(raw)) {
		case "1", "true", "on", "activo", "enabled":
			cfg.Enabled = true
		case "0", "false", "off", "inactivo", "disabled":
			cfg.Enabled = false
		}
	}
	if raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamBaseURLKey); err == nil && strings.TrimSpace(raw) != "" {
		cfg.BaseURL = strings.TrimRight(strings.TrimSpace(raw), "/")
	}
	if raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamProviderKey); err == nil && strings.TrimSpace(raw) != "" {
		cfg.Provider = strings.TrimSpace(raw)
	}
	if raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamVoiceKey); err == nil && strings.TrimSpace(raw) != "" {
		cfg.Voice = strings.TrimSpace(raw)
	}
	if raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamTimeoutMSKey); err == nil {
		var parsed int
		if _, scanErr := fmt.Sscanf(strings.TrimSpace(raw), "%d", &parsed); scanErr == nil && parsed >= 1000 && parsed <= 60000 {
			cfg.TimeoutMS = parsed
		}
	}
	if raw, _, _, _, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamAuthHeaderKey); err == nil && strings.TrimSpace(raw) != "" {
		cfg.AuthHeader = strings.TrimSpace(raw)
	}
	if _, encrypted, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, voiceStreamAuthTokenKey); err == nil {
		cfg.AuthEncrypted = encrypted
		cfg.AuthUpdated = strings.TrimSpace(updatedAt)
	}
	if token, _, err := voiceStreamResolveAuthToken(dbSuper); err == nil && strings.TrimSpace(token) != "" {
		cfg.AuthConfigured = true
	}
	return cfg
}

func validateVoiceStreamAuthHeader(raw string) (string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return "X-PCS-Voice-Token", nil
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			continue
		}
		return "", fmt.Errorf("auth_header solo permite letras, numeros y guiones")
	}
	return value, nil
}

func voiceStreamGenerateAuthToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func voiceStreamResolveAuthToken(dbSuper *sql.DB) (string, bool, error) {
	env := strings.TrimSpace(os.Getenv("VOICE_STREAM_AUTH_TOKEN"))
	if env != "" {
		if dbSuper != nil && utils.EncryptionAvailable() {
			dbVal, err := getDecryptedConfigValue(dbSuper, voiceStreamAuthTokenKey)
			if err != nil || strings.TrimSpace(dbVal) == "" || strings.TrimSpace(dbVal) != env {
				if encVal, encErr := utils.EncryptString(env); encErr == nil {
					_ = dbpkg.SetConfigValue(dbSuper, voiceStreamAuthTokenKey, encVal, true)
				}
			}
		}
		return env, false, nil
	}
	if dbSuper != nil {
		v, err := getDecryptedConfigValue(dbSuper, voiceStreamAuthTokenKey)
		if err == nil && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v), true, nil
		}
		if err != nil {
			return "", false, err
		}
	}
	return "", false, nil
}

func voiceStreamEnsureAuthToken(dbSuper *sql.DB) (bool, error) {
	token, _, err := voiceStreamResolveAuthToken(dbSuper)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(token) != "" {
		return false, nil
	}
	if dbSuper == nil {
		return false, nil
	}
	if !utils.EncryptionAvailable() {
		return false, fmt.Errorf("cifrado requerido: CONFIG_ENC_KEY no esta disponible")
	}
	gen, err := voiceStreamGenerateAuthToken()
	if err != nil {
		return false, err
	}
	encVal, err := utils.EncryptString(gen)
	if err != nil {
		return false, err
	}
	if err := dbpkg.SetConfigValue(dbSuper, voiceStreamAuthTokenKey, encVal, true); err != nil {
		return false, err
	}
	_ = os.Setenv("VOICE_STREAM_AUTH_TOKEN", gen)
	return true, nil
}

func attachVoiceStreamAuthHeader(req *http.Request, dbSuper *sql.DB, cfg voiceStreamConfig) {
	if req == nil {
		return
	}
	token, _, err := voiceStreamResolveAuthToken(dbSuper)
	if err != nil || strings.TrimSpace(token) == "" {
		return
	}
	header, err := validateVoiceStreamAuthHeader(cfg.AuthHeader)
	if err != nil || strings.TrimSpace(header) == "" {
		header = "X-PCS-Voice-Token"
	}
	req.Header.Set(header, strings.TrimSpace(token))
}

func validateVoiceStreamBaseURL(raw string) (string, error) {
	value := strings.TrimRight(strings.TrimSpace(raw), "/")
	if value == "" {
		return "", nil
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		return "", fmt.Errorf("base_url debe iniciar con http:// o https://")
	}
	return value, nil
}

func voiceStreamRequireSession(w http.ResponseWriter, r *http.Request, dbSuper *sql.DB) bool {
	if dbSuper == nil {
		http.Error(w, "super db no disponible", http.StatusInternalServerError)
		return false
	}
	cookie, err := r.Cookie("session_token")
	if err != nil || strings.TrimSpace(cookie.Value) == "" {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return false
	}
	session, err := dbpkg.GetSessionByToken(dbSuper, strings.TrimSpace(cookie.Value))
	if err != nil || session == nil {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return false
	}
	return true
}

// SuperVoiceStreamConfigHandler administra el servicio abierto de voz natural para conversaciones con IA.
func SuperVoiceStreamConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}

		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, map[string]any{
				"ok":     true,
				"config": resolveVoiceStreamConfig(dbSuper),
			})
			return
		case http.MethodPut, http.MethodPost:
			action := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("action")))
			if action == "activate" || action == "activar" || action == "activate_test" || action == "activar_probar" {
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamEnabledKey, "1", false); err != nil {
					http.Error(w, "No se pudo activar voice_stream.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
				cfg := resolveVoiceStreamConfig(dbSuper)
				if strings.TrimSpace(cfg.BaseURL) == "" {
					cfg.BaseURL = defaultVoiceStreamConfig().BaseURL
					if err := dbpkg.SetConfigValue(dbSuper, voiceStreamBaseURLKey, cfg.BaseURL, false); err != nil {
						http.Error(w, "No se pudo guardar voice_stream.base_url: "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
				if _, err := voiceStreamEnsureAuthToken(dbSuper); err != nil {
					http.Error(w, "No se pudo registrar credencial de voz: "+err.Error(), http.StatusInternalServerError)
					return
				}
				cfg = resolveVoiceStreamConfig(dbSuper)
				status := probeVoiceStreamService(dbSuper, cfg)
				code := http.StatusOK
				if !status["ok"].(bool) {
					code = http.StatusBadGateway
				}
				writeJSON(w, code, map[string]any{
					"ok":     status["ok"],
					"status": status,
					"config": cfg,
				})
				return
			}
			if action == "test" {
				cfg := resolveVoiceStreamConfig(dbSuper)
				status := probeVoiceStreamService(dbSuper, cfg)
				code := http.StatusOK
				if !status["ok"].(bool) {
					code = http.StatusBadGateway
				}
				writeJSON(w, code, status)
				return
			}

			var payload struct {
				Enabled           *bool  `json:"enabled"`
				BaseURL           string `json:"base_url"`
				Provider          string `json:"provider"`
				Voice             string `json:"voice"`
				TimeoutMS         int    `json:"timeout_ms"`
				AuthHeader        string `json:"auth_header"`
				AuthToken         string `json:"auth_token"`
				GenerateAuthToken bool   `json:"generate_auth_token"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
				return
			}

			if payload.Enabled != nil {
				value := "0"
				if *payload.Enabled {
					value = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamEnabledKey, value, false); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.BaseURL) != "" {
				baseURL, err := validateVoiceStreamBaseURL(payload.BaseURL)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamBaseURLKey, baseURL, false); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.base_url: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.Provider) != "" {
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamProviderKey, strings.TrimSpace(payload.Provider), false); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.provider: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.Voice) != "" {
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamVoiceKey, strings.TrimSpace(payload.Voice), false); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.voice: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.TimeoutMS > 0 {
				if payload.TimeoutMS < 1000 || payload.TimeoutMS > 60000 {
					http.Error(w, "timeout_ms debe estar entre 1000 y 60000", http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamTimeoutMSKey, fmt.Sprintf("%d", payload.TimeoutMS), false); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.timeout_ms: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.AuthHeader) != "" {
				authHeader, err := validateVoiceStreamAuthHeader(payload.AuthHeader)
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadRequest)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamAuthHeaderKey, authHeader, false); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.auth_header: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.AuthToken) != "" || payload.GenerateAuthToken {
				if !utils.EncryptionAvailable() {
					http.Error(w, "Cifrado requerido: CONFIG_ENC_KEY no esta disponible", http.StatusInternalServerError)
					return
				}
				authToken := strings.TrimSpace(payload.AuthToken)
				if authToken == "" {
					var err error
					authToken, err = voiceStreamGenerateAuthToken()
					if err != nil {
						http.Error(w, "No se pudo generar auth_token: "+err.Error(), http.StatusInternalServerError)
						return
					}
				}
				encVal, err := utils.EncryptString(authToken)
				if err != nil {
					http.Error(w, "No se pudo cifrar auth_token: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, voiceStreamAuthTokenKey, encVal, true); err != nil {
					http.Error(w, "No se pudo guardar voice_stream.auth_token: "+err.Error(), http.StatusInternalServerError)
					return
				}
				_ = os.Setenv("VOICE_STREAM_AUTH_TOKEN", authToken)
			}

			writeJSON(w, http.StatusOK, map[string]any{
				"ok":     true,
				"config": resolveVoiceStreamConfig(dbSuper),
			})
			return
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

// VoiceStreamStatusHandler expone al chat solo el estado no sensible del servicio de voz.
func VoiceStreamStatusHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}
		cfg := resolveVoiceStreamConfig(dbSuper)
		status := map[string]any{
			"ok":         true,
			"enabled":    cfg.Enabled,
			"provider":   cfg.Provider,
			"voice":      cfg.Voice,
			"configured": strings.TrimSpace(cfg.BaseURL) != "",
		}
		if cfg.Enabled {
			probe := probeVoiceStreamService(dbSuper, cfg)
			status["service_ok"] = probe["ok"]
			status["service_status"] = probe["status"]
		}
		writeJSON(w, http.StatusOK, status)
	}
}

// VoiceStreamTTSProxyHandler convierte texto de respuesta IA a audio sin exponer la URL interna del VPS al navegador.
func VoiceStreamTTSProxyHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if !voiceStreamRequireSession(w, r, dbSuper) {
			return
		}
		cfg := resolveVoiceStreamConfig(dbSuper)
		if !cfg.Enabled {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "voice_stream_disabled"})
			return
		}
		if strings.TrimSpace(cfg.BaseURL) == "" {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": "voice_stream_unconfigured"})
			return
		}

		var payload struct {
			Text  string `json:"text"`
			Voice string `json:"voice"`
		}
		if err := json.NewDecoder(io.LimitReader(r.Body, 16*1024)).Decode(&payload); err != nil {
			http.Error(w, "payload invalido: "+err.Error(), http.StatusBadRequest)
			return
		}
		payload.Text = strings.TrimSpace(payload.Text)
		if payload.Text == "" {
			http.Error(w, "text requerido", http.StatusBadRequest)
			return
		}
		if len([]rune(payload.Text)) > 4000 {
			http.Error(w, "text demasiado largo", http.StatusRequestEntityTooLarge)
			return
		}
		if strings.TrimSpace(payload.Voice) == "" {
			payload.Voice = cfg.Voice
		}

		body, _ := json.Marshal(map[string]string{
			"text":  payload.Text,
			"voice": strings.TrimSpace(payload.Voice),
		})
		timeout := time.Duration(cfg.TimeoutMS) * time.Millisecond
		if timeout <= 0 {
			timeout = 12 * time.Second
		}
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.BaseURL+"/api/voice/tts", bytes.NewReader(body))
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": err.Error()})
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "audio/wav,audio/mpeg,audio/*;q=0.9,application/json;q=0.5")
		attachVoiceStreamAuthHeader(req, dbSuper, cfg)
		res, err := (&http.Client{Timeout: timeout}).Do(req)
		if err != nil || res == nil {
			writeJSON(w, http.StatusBadGateway, map[string]any{"ok": false, "error": fmt.Sprint(err)})
			return
		}
		defer res.Body.Close()

		if res.StatusCode < 200 || res.StatusCode > 299 {
			raw, _ := io.ReadAll(io.LimitReader(res.Body, 2048))
			writeJSON(w, http.StatusBadGateway, map[string]any{
				"ok":              false,
				"upstream_status": res.StatusCode,
				"error":           strings.TrimSpace(string(raw)),
			})
			return
		}

		contentType := strings.TrimSpace(res.Header.Get("Content-Type"))
		if contentType == "" {
			contentType = "audio/wav"
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, res.Body)
	}
}

func probeVoiceStreamService(dbSuper *sql.DB, cfg voiceStreamConfig) map[string]any {
	if !cfg.Enabled {
		return map[string]any{"ok": false, "status": "disabled"}
	}
	if strings.TrimSpace(cfg.BaseURL) == "" {
		return map[string]any{"ok": false, "status": "unconfigured"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2500*time.Millisecond)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, cfg.BaseURL+"/health", nil)
	attachVoiceStreamAuthHeader(req, dbSuper, cfg)
	res, err := (&http.Client{Timeout: 2500 * time.Millisecond}).Do(req)
	if err != nil || res == nil {
		return map[string]any{"ok": false, "status": "unreachable", "error": fmt.Sprint(err)}
	}
	defer res.Body.Close()
	return map[string]any{
		"ok":     res.StatusCode >= 200 && res.StatusCode <= 299,
		"status": res.StatusCode,
	}
}
