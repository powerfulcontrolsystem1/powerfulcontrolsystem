package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	superChatIAEmpresaEnabledKey     = "ai.chat.empresa.enabled"
	superChatIASuperEnabledKey       = "ai.chat.super.enabled"
	superChatIAEmpresaMaxConsultasKey = "ai.chat.empresa.max_consultas_dia"
	superChatIASuperMaxConsultasKey   = "ai.chat.super.max_consultas_dia"
	superChatIASuperContextoAmplioKey = "ai.chat.super.contexto_amplio"
	superChatIAEmpresaSoloLecturaKey  = "ai.chat.super.empresa_solo_lectura"

	superChatIALogicaUpdatedBySuffix = ".updated_by"

	defaultChatIAEmpresaEnabled      = true
	defaultChatIASuperEnabled        = true
	defaultChatIAEmpresaMaxConsultas = int64(10)
	defaultChatIASuperMaxConsultas   = int64(30)
	defaultChatIASuperContextoAmplio = false
	defaultChatIAEmpresaSoloLectura  = false
)

func parseConfigBoolWithDefault(raw string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(raw))
	if v == "" {
		return fallback
	}
	switch v {
	case "1", "true", "on", "enabled", "activo", "si", "sí":
		return true
	case "0", "false", "off", "disabled", "inactivo", "no":
		return false
	default:
		return fallback
	}
}

func parseConfigNonNegativeInt64WithDefault(raw string, fallback int64) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	if v < 0 {
		return 0
	}
	return v
}

func getSuperConfigString(dbSuper *sql.DB, key string) (string, string, string, error) {
	val, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil {
		return "", "", "", err
	}
	updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key+superChatIALogicaUpdatedBySuffix)
	return strings.TrimSpace(val), strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func getChatIAEmpresaEnabled(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaEnabledKey)
	if err != nil {
		return defaultChatIAEmpresaEnabled, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIAEmpresaEnabled), updatedAt, updatedBy, nil
}

func getChatIASuperEnabled(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIASuperEnabledKey)
	if err != nil {
		return defaultChatIASuperEnabled, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIASuperEnabled), updatedAt, updatedBy, nil
}

func getChatIAEmpresaMaxConsultasDia(dbSuper *sql.DB) (int64, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaMaxConsultasKey)
	if err != nil {
		return defaultChatIAEmpresaMaxConsultas, "", "", err
	}
	if raw == "" {
		return defaultChatIAEmpresaMaxConsultas, updatedAt, updatedBy, nil
	}
	return parseConfigNonNegativeInt64WithDefault(raw, defaultChatIAEmpresaMaxConsultas), updatedAt, updatedBy, nil
}

func getChatIASuperMaxConsultasDia(dbSuper *sql.DB) (int64, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIASuperMaxConsultasKey)
	if err != nil {
		return defaultChatIASuperMaxConsultas, "", "", err
	}
	if raw == "" {
		return defaultChatIASuperMaxConsultas, updatedAt, updatedBy, nil
	}
	return parseConfigNonNegativeInt64WithDefault(raw, defaultChatIASuperMaxConsultas), updatedAt, updatedBy, nil
}

func getChatIASuperContextoAmplio(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIASuperContextoAmplioKey)
	if err != nil {
		return defaultChatIASuperContextoAmplio, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIASuperContextoAmplio), updatedAt, updatedBy, nil
}

func getChatIAEmpresaSoloLectura(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaSoloLecturaKey)
	if err != nil {
		return defaultChatIAEmpresaSoloLectura, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIAEmpresaSoloLectura), updatedAt, updatedBy, nil
}

func effectiveDailyLimitBySuperConfig(maxConfigured int64, modelFreeDailyLimit int) int64 {
	effective := int64(modelFreeDailyLimit)
	if effective <= 0 {
		// Sin límite por modelo (o no informado): usar solo el configurado si existe.
		if maxConfigured < 0 {
			return 0
		}
		return maxConfigured
	}
	if maxConfigured < 0 {
		return effective
	}
	// maxConfigured==0 bloquea.
	if maxConfigured == 0 {
		return 0
	}
	if maxConfigured > 0 && maxConfigured < effective {
		return maxConfigured
	}
	return effective
}

// SuperChatIALogicaConfigHandler permite administrar restricciones lógicas del chat IA
// para empresas y para el chat global del super administrador.
// Persistencia: tabla configuraciones en pcs_superadministrador.
func SuperChatIALogicaConfigHandler(dbEmp, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			empresaEnabled, empresaEnabledAt, empresaEnabledBy, err := getChatIAEmpresaEnabled(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}
			superEnabled, superEnabledAt, superEnabledBy, err := getChatIASuperEnabled(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}
			empresaMax, empresaMaxAt, empresaMaxBy, err := getChatIAEmpresaMaxConsultasDia(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}
			superMax, superMaxAt, superMaxBy, err := getChatIASuperMaxConsultasDia(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}
			superCtxAmplio, superCtxAmplioAt, superCtxAmplioBy, err := getChatIASuperContextoAmplio(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}
			empSoloLectura, empSoloLecturaAt, empSoloLecturaBy, err := getChatIAEmpresaSoloLectura(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}

			fechaUso := time.Now().Format("2006-01-02")
			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			superConsultasHoy, superTokensHoy, err := dbpkg.GetSuperAIUsoDiarioOpenAITokensGlobal(dbSuper, adminEmail, "openai", fechaUso)
			if err != nil {
				http.Error(w, "error leyendo consumo openai super: "+err.Error(), http.StatusInternalServerError)
				return
			}
			superAllConsultasHoy, superAllTokensHoy, err := dbpkg.GetSuperAIUsoDiarioOpenAITokensGlobal(dbSuper, "", "openai", fechaUso)
			if err != nil {
				http.Error(w, "error leyendo consumo openai super global: "+err.Error(), http.StatusInternalServerError)
				return
			}
			empConsultasHoy, empTokensHoy, err := dbpkg.GetEmpresaAIUsoDiarioOpenAITokensGlobal(dbEmp, "openai", fechaUso)
			if err != nil {
				http.Error(w, "error leyendo consumo openai empresas: "+err.Error(), http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true,
				"defaults": map[string]interface{}{
					"empresa_enabled":             defaultChatIAEmpresaEnabled,
					"empresa_max_consultas":       defaultChatIAEmpresaMaxConsultas,
					"super_enabled":               defaultChatIASuperEnabled,
					"super_max_consultas":         defaultChatIASuperMaxConsultas,
					"super_contexto_amplio":       defaultChatIASuperContextoAmplio,
					"empresa_solo_lectura":        defaultChatIAEmpresaSoloLectura,
				},
				"values": map[string]interface{}{
					"empresa_enabled": map[string]interface{}{
						"value":      empresaEnabled,
						"config_key": superChatIAEmpresaEnabledKey,
						"updated_at": empresaEnabledAt,
						"updated_by": empresaEnabledBy,
					},
					"empresa_max_consultas": map[string]interface{}{
						"value":      empresaMax,
						"config_key": superChatIAEmpresaMaxConsultasKey,
						"updated_at": empresaMaxAt,
						"updated_by": empresaMaxBy,
					},
					"super_enabled": map[string]interface{}{
						"value":      superEnabled,
						"config_key": superChatIASuperEnabledKey,
						"updated_at": superEnabledAt,
						"updated_by": superEnabledBy,
					},
					"super_max_consultas": map[string]interface{}{
						"value":      superMax,
						"config_key": superChatIASuperMaxConsultasKey,
						"updated_at": superMaxAt,
						"updated_by": superMaxBy,
					},
					"super_contexto_amplio": map[string]interface{}{
						"value":      superCtxAmplio,
						"config_key": superChatIASuperContextoAmplioKey,
						"updated_at": superCtxAmplioAt,
						"updated_by": superCtxAmplioBy,
					},
					"empresa_solo_lectura": map[string]interface{}{
						"value":      empSoloLectura,
						"config_key": superChatIAEmpresaSoloLecturaKey,
						"updated_at": empSoloLecturaAt,
						"updated_by": empSoloLecturaBy,
					},
				},
				"openai_usage_today": map[string]interface{}{
					"fecha_uso": fechaUso,
					"super_admin": map[string]interface{}{
						"admin_email":     adminEmail,
						"consultas_total": superConsultasHoy,
						"tokens_total":    superTokensHoy,
					},
					"super_todos": map[string]interface{}{
						"consultas_total": superAllConsultasHoy,
						"tokens_total":    superAllTokensHoy,
					},
					"empresas_todas": map[string]interface{}{
						"consultas_total": empConsultasHoy,
						"tokens_total":    empTokensHoy,
					},
				},
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				EmpresaEnabled          *bool  `json:"empresa_enabled"`
				EmpresaMaxConsultas     *int64 `json:"empresa_max_consultas"`
				SuperEnabled            *bool  `json:"super_enabled"`
				SuperMaxConsultas       *int64 `json:"super_max_consultas"`
				SuperContextoAmplio     *bool  `json:"super_contexto_amplio"`
				EmpresaSoloLectura      *bool  `json:"empresa_solo_lectura"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			empresaEnabled := defaultChatIAEmpresaEnabled
			if payload.EmpresaEnabled != nil {
				empresaEnabled = *payload.EmpresaEnabled
			}
			superEnabled := defaultChatIASuperEnabled
			if payload.SuperEnabled != nil {
				superEnabled = *payload.SuperEnabled
			}
			empresaMax := defaultChatIAEmpresaMaxConsultas
			if payload.EmpresaMaxConsultas != nil {
				empresaMax = *payload.EmpresaMaxConsultas
			}
			if empresaMax < 0 {
				empresaMax = 0
			}
			superMax := defaultChatIASuperMaxConsultas
			if payload.SuperMaxConsultas != nil {
				superMax = *payload.SuperMaxConsultas
			}
			if superMax < 0 {
				superMax = 0
			}
			superCtxAmplio := defaultChatIASuperContextoAmplio
			if payload.SuperContextoAmplio != nil {
				superCtxAmplio = *payload.SuperContextoAmplio
			} else {
				if cur, _, _, err := getChatIASuperContextoAmplio(dbSuper); err == nil {
					superCtxAmplio = cur
				}
			}
			empSoloLectura := defaultChatIAEmpresaSoloLectura
			if payload.EmpresaSoloLectura != nil {
				empSoloLectura = *payload.EmpresaSoloLectura
			} else {
				if cur, _, _, err := getChatIAEmpresaSoloLectura(dbSuper); err == nil {
					empSoloLectura = cur
				}
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaEnabledKey, strconv.FormatBool(empresaEnabled), false); err != nil {
				http.Error(w, "error guardando empresa_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaMaxConsultasKey, strconv.FormatInt(empresaMax, 10), false); err != nil {
				http.Error(w, "error guardando empresa_max_consultas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaMaxConsultasKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIASuperEnabledKey, strconv.FormatBool(superEnabled), false); err != nil {
				http.Error(w, "error guardando super_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIASuperEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIASuperMaxConsultasKey, strconv.FormatInt(superMax, 10), false); err != nil {
				http.Error(w, "error guardando super_max_consultas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIASuperMaxConsultasKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIASuperContextoAmplioKey, strconv.FormatBool(superCtxAmplio), false); err != nil {
				http.Error(w, "error guardando super_contexto_amplio: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIASuperContextoAmplioKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaSoloLecturaKey, strconv.FormatBool(empSoloLectura), false); err != nil {
				http.Error(w, "error guardando empresa_solo_lectura: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaSoloLecturaKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true,
				"saved": map[string]interface{}{
					"empresa_enabled":        empresaEnabled,
					"empresa_max_consultas":  empresaMax,
					"super_enabled":          superEnabled,
					"super_max_consultas":    superMax,
					"super_contexto_amplio":  superCtxAmplio,
					"empresa_solo_lectura":   empSoloLectura,
				},
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

