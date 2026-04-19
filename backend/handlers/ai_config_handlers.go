package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

const superAIEnabledConfigKey = "ai.global.enabled"

func isSuperAIEnabled(dbSuper *sql.DB) bool {
	if dbSuper == nil {
		return true
	}
	value, err := getDecryptedConfigValue(dbSuper, superAIEnabledConfigKey)
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

func superAIServiceStatus(dbSuper *sql.DB) map[string]interface{} {
	enabled := isSuperAIEnabled(dbSuper)
	message := "IA global habilitada para chats empresarial y super."
	if !enabled {
		message = "IA global desactivada desde configuracion avanzada."
	}
	return map[string]interface{}{
		"enabled": enabled,
		"message": message,
	}
}

func runSuperAITest(dbSuper *sql.DB) (int, map[string]interface{}) {
	if !isSuperAIEnabled(dbSuper) {
		return http.StatusServiceUnavailable, map[string]interface{}{
			"ok":             false,
			"code":           "ai_disabled",
			"error":          "La IA está desactivada desde configuración avanzada.",
			"service_status": superAIServiceStatus(dbSuper),
		}
	}

	defs := aiCredentialCatalogModels()
	if len(defs) == 0 {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_catalog_missing",
			"error":          "No hay modelo Gemini configurado en el catalogo.",
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	modelMap := empresaAIModelMap()
	model, ok := modelMap[defs[0].ModelID]
	if !ok {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_model_missing",
			"error":          "No se pudo resolver el modelo Gemini de prueba.",
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	apiKey, err := (&EmpresaAIChatController{dbSuper: dbSuper, client: &http.Client{Timeout: 20 * time.Second}}).resolveModelAPIKey(model)
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_api_key_missing",
			"error":          err.Error(),
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	endpoint := model.Endpoint
	separator := "?"
	if strings.Contains(endpoint, "?") {
		separator = "&"
	}
	endpoint = endpoint + separator + "key=" + url.QueryEscape(apiKey)

	payload, _ := json.Marshal(map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role":  "user",
				"parts": []map[string]string{{"text": "Responde solo OK_PANEL_TEST"}},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0,
			"maxOutputTokens": 32,
		},
	})
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_request_build_failed",
			"error":          "No se pudo construir la solicitud de prueba a Gemini.",
			"detail":         err.Error(),
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_unreachable",
			"error":          "No se pudo contactar Google Gemini con la API key configurada.",
			"detail":         err.Error(),
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_error",
			"error":          "Google Gemini respondió con error durante la prueba.",
			"status":         resp.StatusCode,
			"raw":            truncateText(string(raw), 300),
			"service_status": superAIServiceStatus(dbSuper),
		}
	}

	var parsed struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "gemini_invalid_json",
			"error":          "La respuesta de Gemini no fue JSON valido.",
			"raw":            truncateText(string(raw), 300),
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	response := extractGeminiText(parsed.Candidates)

	return http.StatusOK, map[string]interface{}{
		"ok":             true,
		"provider":       model.Provider,
		"model_id":       model.ID,
		"upstream_model": model.UpstreamModel,
		"response":       strings.TrimSpace(response),
		"service_status": superAIServiceStatus(dbSuper),
	}
}

// AIModelsConfigHandler gestiona credenciales IA disponibles en el catalogo backend.
func AIModelsConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("action")), "test") {
				status, body := runSuperAITest(dbSuper)
				writeJSON(w, status, body)
				return
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			defs := aiCredentialCatalogModels()
			items := make([]map[string]interface{}, 0, len(defs))
			for _, def := range defs {
				value := ""
				enc := false
				updatedAt := ""
				updatedBy := ""
				configured := false
				masked := ""
				source := "local-service"
				providerEnabled := isAIProviderEnabled(dbSuper, def.Provider)

				if strings.TrimSpace(def.ConfigKey) == "" {
					configured = true
					source = "vps-local"
				} else {
					value, enc, _, updatedAt, _ = dbpkg.GetConfigEntry(dbSuper, def.ConfigKey)
					updatedBy, _, _, _, _ = dbpkg.GetConfigEntry(dbSuper, def.ConfigKey+".updated_by")
					configured = strings.TrimSpace(value) != ""
					masked = maskSecretValue(value, enc)
					source = "db"
					if !configured && strings.TrimSpace(def.ApiKeyEnv) != "" {
						envVal := strings.TrimSpace(os.Getenv(def.ApiKeyEnv))
						if envVal != "" {
							configured = true
							masked = maskSecretValue(envVal, false)
							source = "env"
						}
					}
				}
				items = append(items, map[string]interface{}{
					"model_id":         def.ModelID,
					"provider":         def.Provider,
					"enabled":          providerEnabled,
					"display_name":     def.DisplayName,
					"config_key":       def.ConfigKey,
					"api_key_env":      def.ApiKeyEnv,
					"free_plan_note":   def.FreePlanNote,
					"configured":       configured,
					"masked":           masked,
					"source":           source,
					"updated_at":       updatedAt,
					"updated_by":       strings.TrimSpace(updatedBy),
					"google_account":   adminEmail,
					"encryption_state": enc,
				})
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                   true,
				"google_account":       adminEmail,
				"encryption_available": utils.EncryptionAvailable(),
				"service_status":       superAIServiceStatus(dbSuper),
				"modelos":              items,
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				Credentials []struct {
					ModelID string `json:"model_id"`
					APIKey  string `json:"api_key"`
				} `json:"credentials"`
				Enabled         *bool           `json:"enabled"`
				ProviderEnabled map[string]bool `json:"provider_enabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Forzar cifrado obligatorio para todas las credenciales sensibles en configuración avanzada.
			if len(payload.Credentials) > 0 && !utils.EncryptionAvailable() {
				http.Error(w, "encryption required: CONFIG_ENC_KEY not set", http.StatusBadRequest)
				return
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			if adminEmail == "" || adminEmail == "sistema" {
				http.Error(w, "unauthenticated", http.StatusUnauthorized)
				return
			}

			if payload.Enabled != nil {
				enabledValue := "0"
				if *payload.Enabled {
					enabledValue = "1"
				}
				if err := dbpkg.SetConfigValue(dbSuper, superAIEnabledConfigKey, enabledValue, false); err != nil {
					http.Error(w, "failed to save "+superAIEnabledConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, superAIEnabledConfigKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "failed to save updated_by for "+superAIEnabledConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			for provider, enabled := range payload.ProviderEnabled {
				provider = strings.ToLower(strings.TrimSpace(provider))
				if provider == "" {
					continue
				}
				known := false
				for _, item := range uniqueAIProviders() {
					if item == provider {
						known = true
						break
					}
				}
				if !known {
					http.Error(w, "provider no soportado: "+provider, http.StatusBadRequest)
					return
				}
				providerValue := "0"
				if enabled {
					providerValue = "1"
				}
				providerEnabledKey := aiProviderEnabledConfigKey(provider)
				if providerEnabledKey == "" {
					continue
				}
				if err := dbpkg.SetConfigValue(dbSuper, providerEnabledKey, providerValue, false); err != nil {
					http.Error(w, "failed to save "+providerEnabledKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, providerEnabledKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "failed to save updated_by for "+providerEnabledKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			defs := aiCredentialByModelID()
			updated := make([]string, 0, len(payload.Credentials))
			for _, item := range payload.Credentials {
				modelID := strings.TrimSpace(item.ModelID)
				apiKey := strings.TrimSpace(item.APIKey)
				if modelID == "" || apiKey == "" {
					continue
				}
				def, ok := defs[modelID]
				if !ok {
					http.Error(w, "model_id no soportado: "+modelID, http.StatusBadRequest)
					return
				}
				if strings.TrimSpace(def.ConfigKey) == "" {
					http.Error(w, "el modelo "+modelID+" no requiere credencial en este panel", http.StatusBadRequest)
					return
				}

				// Siempre cifrar el valor antes de persistir
				encVal, err := utils.EncryptString(apiKey)
				if err != nil {
					http.Error(w, "encryption failed: "+err.Error(), http.StatusInternalServerError)
					return
				}

				if err := dbpkg.SetConfigValue(dbSuper, def.ConfigKey, encVal, true); err != nil {
					http.Error(w, "failed to save "+def.ConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, def.ConfigKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "failed to save updated_by for "+def.ConfigKey+": "+err.Error(), http.StatusInternalServerError)
					return
				}

				providerKey := aiProviderConfigKey(def.Provider)
				if providerKey != "" {
					if err := dbpkg.SetConfigValue(dbSuper, providerKey, encVal, true); err != nil {
						http.Error(w, "failed to save provider key "+providerKey+": "+err.Error(), http.StatusInternalServerError)
						return
					}
				}

				updated = append(updated, modelID)
			}

			if len(updated) == 0 && payload.Enabled == nil && len(payload.ProviderEnabled) == 0 {
				http.Error(w, "debe enviar al menos una credencial valida o un cambio de estado", http.StatusBadRequest)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"saved":          true,
				"updated_models": updated,
				"google_account": adminEmail,
				"service_status": superAIServiceStatus(dbSuper),
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func readDecryptedConfigOrEmpty(dbSuper *sql.DB, key string) (string, error) {
	v, enc, err := dbpkg.GetConfigValue(dbSuper, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	if !enc {
		return v, nil
	}
	dec, derr := utils.DecryptString(v)
	if derr != nil {
		return "", derr
	}
	return dec, nil
}

func maskSecretValue(raw string, _ bool) string {
	v := strings.TrimSpace(raw)
	if v == "" {
		return ""
	}
	return "********"
}
