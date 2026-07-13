package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
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
			"code":           "ai_catalog_missing",
			"error":          "No hay modelo IA configurado en el catalogo.",
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	modelMap := empresaAIModelMap()
	model, ok := modelMap[defs[0].ModelID]
	if !ok {
		return http.StatusBadGateway, map[string]interface{}{
			"ok":             false,
			"code":           "ai_model_missing",
			"error":          "No se pudo resolver el modelo IA de prueba.",
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	apiKey, err := (&EmpresaAIChatController{dbSuper: dbSuper, client: &http.Client{Timeout: 20 * time.Second}}).resolveModelAPIKey(model)
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"ok":             false,
			"code":           "ai_api_key_missing",
			"error":          "No hay una credencial de IA disponible para la prueba.",
			"service_status": superAIServiceStatus(dbSuper),
		}
	}
	testController := &EmpresaAIChatController{dbSuper: dbSuper, client: &http.Client{Timeout: 20 * time.Second}}
	_ = apiKey

	respuesta, promptTokens, completionTokens, err := testController.generateResponseWithSystemPrompt(
		model,
		"Responde solo OK_PANEL_TEST",
		nil,
		"Eres un asistente de prueba. Responde solo OK_PANEL_TEST.",
	)
	if err != nil {
		status := http.StatusBadGateway
		code := "ai_test_failed"
		publicErr := "No se pudo completar la prueba con el proveedor de IA. Intenta de nuevo o revisa la configuracion."
		providerStatus := 0
		if perr := (*aiProviderHTTPError)(nil); errors.As(err, &perr) && perr != nil {
			providerStatus = perr.Status
			// Mapear errores de proveedor a códigos no-5xx para que el panel vea el detalle.
			switch perr.Status {
			case http.StatusUnauthorized, http.StatusForbidden:
				status = perr.Status
				code = "ai_api_key_invalid"
				publicErr = "API key de OpenAI inválida o sin permisos. Usa Editar y vuelve a guardar la credencial."
			case http.StatusTooManyRequests:
				status = perr.Status
				code = "ai_rate_limited"
				publicErr = "OpenAI respondió límite de uso (429). Intenta de nuevo en unos minutos."
			case http.StatusBadRequest:
				status = perr.Status
				code = "ai_bad_request"
				publicErr = "El proveedor de IA rechazó la solicitud de prueba. Revisa la configuración del modelo."
			default:
				// 4xx distintos: devolverlos como 400 para no esconder el mensaje por middleware.
				if perr.Status >= 400 && perr.Status < 500 {
					status = http.StatusBadRequest
					code = "ai_provider_rejected"
					publicErr = "El proveedor de IA rechazó la solicitud de prueba. Revisa la configuración del modelo."
				}
			}
		}

		return status, map[string]interface{}{
			"ok":             false,
			"code":           code,
			"error":          publicErr,
			"provider":       model.Provider,
			"model_id":       model.ID,
			"upstream_model": model.UpstreamModel,
			"provider_http":  providerStatus,
			"service_status": superAIServiceStatus(dbSuper),
		}
	}

	return http.StatusOK, map[string]interface{}{
		"ok":                true,
		"provider":          model.Provider,
		"model_id":          model.ID,
		"upstream_model":    model.UpstreamModel,
		"response":          strings.TrimSpace(respuesta),
		"prompt_tokens":     promptTokens,
		"completion_tokens": completionTokens,
		"service_status":    superAIServiceStatus(dbSuper),
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
					"usage_today":      nil,
				})
			}

			// Enriquecer consumo del día (si se puede) por modelo.
			fechaUso := time.Now().Format("2006-01-02")
			for idx := range items {
				modelID, _ := items[idx]["model_id"].(string)
				provider, _ := items[idx]["provider"].(string)
				if strings.TrimSpace(modelID) == "" || strings.TrimSpace(provider) == "" || strings.TrimSpace(adminEmail) == "" {
					continue
				}
				uso, err := dbpkg.GetSuperAIUsoDiario(dbSuper, adminEmail, provider, modelID, fechaUso)
				if err != nil {
					continue
				}
				items[idx]["usage_today"] = map[string]interface{}{
					"fecha_uso":       fechaUso,
					"consultas_total": uso.Consultas,
					"tokens_total":    uso.TokensTotal,
					"plan_actual":     strings.TrimSpace(uso.PlanActual),
				}
			}

			modeloOperacion, _, _, _ := getChatIAEmpresaModeloOperacion(dbSuper)
			modeloAdjuntos, _, _, _ := getChatIAEmpresaModeloAdjuntos(dbSuper)
			modelosHabilitados, _ := getChatIAEmpresaModelosHabilitados(dbSuper)
			if len(modelosHabilitados) == 0 {
				modelosHabilitados = map[string]bool{}
				for _, model := range empresaAIModelCatalog() {
					modelosHabilitados[model.ID] = true
				}
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok":                   true,
				"google_account":       adminEmail,
				"encryption_available": utils.EncryptionAvailable(),
				"service_status":       superAIServiceStatus(dbSuper),
				"modelos":              items,
				"chat_model_policy": map[string]interface{}{
					"operation_model_id":  modeloOperacion,
					"attachment_model_id": modeloAdjuntos,
					"enabled_model_ids":   modelosHabilitados,
				},
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				Credentials []struct {
					ModelID string `json:"model_id"`
					APIKey  string `json:"api_key"`
				} `json:"credentials"`
				Enabled           *bool           `json:"enabled"`
				ProviderEnabled   map[string]bool `json:"provider_enabled"`
				OperationModelID  string          `json:"operation_model_id"`
				AttachmentModelID string          `json:"attachment_model_id"`
				EnabledModelIDs   []string        `json:"enabled_model_ids"`
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
					http.Error(w, "no se pudo guardar la configuracion de IA", http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, superAIEnabledConfigKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "no se pudo guardar la configuracion de IA", http.StatusInternalServerError)
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
					http.Error(w, "no se pudo guardar la configuracion de IA", http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, providerEnabledKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "no se pudo guardar la configuracion de IA", http.StatusInternalServerError)
					return
				}
			}

			if payload.OperationModelID != "" || payload.AttachmentModelID != "" || payload.EnabledModelIDs != nil {
				catalog := empresaAIModelMap()
				enabled := make([]string, 0, len(payload.EnabledModelIDs))
				seen := map[string]bool{}
				for _, raw := range payload.EnabledModelIDs {
					id := strings.TrimSpace(raw)
					if _, ok := catalog[id]; !ok {
						http.Error(w, "modelo no soportado", http.StatusBadRequest)
						return
					}
					if !seen[id] {
						enabled = append(enabled, id)
						seen[id] = true
					}
				}
				if payload.EnabledModelIDs != nil && len(enabled) == 0 {
					http.Error(w, "debe permanecer al menos un modelo habilitado", http.StatusBadRequest)
					return
				}
				operation := strings.TrimSpace(payload.OperationModelID)
				attachment := strings.TrimSpace(payload.AttachmentModelID)
				if operation == "" {
					operation, _, _, _ = getChatIAEmpresaModeloOperacion(dbSuper)
				}
				if attachment == "" {
					attachment, _, _, _ = getChatIAEmpresaModeloAdjuntos(dbSuper)
				}
				if _, ok := catalog[operation]; !ok {
					http.Error(w, "modelo de operaciones no soportado", http.StatusBadRequest)
					return
				}
				if _, ok := catalog[attachment]; !ok {
					http.Error(w, "modelo de adjuntos no soportado", http.StatusBadRequest)
					return
				}
				if payload.EnabledModelIDs != nil && (!seen[operation] || !seen[attachment]) {
					http.Error(w, "los modelos elegidos deben permanecer habilitados", http.StatusBadRequest)
					return
				}
				entries := map[string]string{
					superChatIAEmpresaModeloOperacionKey: operation,
					superChatIAEmpresaModeloAdjuntosKey:  attachment,
				}
				if payload.EnabledModelIDs != nil {
					entries[superChatIAEmpresaModelosHabilitadosKey] = strings.Join(enabled, ",")
				}
				for key, value := range entries {
					if err := dbpkg.SetConfigValue(dbSuper, key, value, false); err != nil {
						http.Error(w, "no se pudo guardar la politica de modelos", http.StatusInternalServerError)
						return
					}
					if err := dbpkg.SetConfigValue(dbSuper, key+superChatIALogicaUpdatedBySuffix, adminEmail, false); err != nil {
						http.Error(w, "no se pudo guardar la politica de modelos", http.StatusInternalServerError)
						return
					}
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
					http.Error(w, "no se pudo cifrar la credencial de IA", http.StatusInternalServerError)
					return
				}

				if err := dbpkg.SetConfigValue(dbSuper, def.ConfigKey, encVal, true); err != nil {
					http.Error(w, "no se pudo guardar la credencial de IA", http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, def.ConfigKey+".updated_by", adminEmail, false); err != nil {
					http.Error(w, "no se pudo guardar la credencial de IA", http.StatusInternalServerError)
					return
				}

				providerKey := aiProviderConfigKey(def.Provider)
				if providerKey != "" {
					if err := dbpkg.SetConfigValue(dbSuper, providerKey, encVal, true); err != nil {
						http.Error(w, "no se pudo guardar la credencial de IA", http.StatusInternalServerError)
						return
					}
				}

				updated = append(updated, modelID)
			}

			if len(updated) == 0 && payload.Enabled == nil && len(payload.ProviderEnabled) == 0 && payload.EnabledModelIDs == nil && payload.OperationModelID == "" && payload.AttachmentModelID == "" {
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
