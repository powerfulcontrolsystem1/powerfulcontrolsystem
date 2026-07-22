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
	superChatIAEmpresaEnabledKey            = "ai.chat.empresa.enabled"
	superChatIASuperEnabledKey              = "ai.chat.super.enabled"
	superChatIAPortalEnabledKey             = "ai.chat.portal.enabled"
	superChatIAEmpresaMaxConsultasKey       = "ai.chat.empresa.max_consultas_dia"
	superChatIASuperMaxConsultasKey         = "ai.chat.super.max_consultas_dia"
	superChatIAEmpresaMaxGPT55ConsultasKey  = "ai.chat.empresa.max_gpt55_consultas_dia"
	superChatIAEmpresaStreamingEnabledKey   = "ai.chat.empresa.streaming_enabled"
	superChatIASuperStreamingEnabledKey     = "ai.chat.super.streaming_enabled"
	superChatIAEmpresaDBQueryEnabledKey     = "ai.chat.empresa.db_query_enabled"
	superChatIAEmpresaDBQueryMaxTablesKey   = "ai.chat.empresa.db_query_max_tables"
	superChatIAEmpresaDBQueryRowsKey        = "ai.chat.empresa.db_query_rows"
	superChatIASuperContextoAmplioKey       = "ai.chat.super.contexto_amplio"
	superChatIAEmpresaSoloLecturaKey        = "ai.chat.super.empresa_solo_lectura"
	superChatIAEmpresaModeloOperacionKey    = "ai.chat.empresa.modelo_operacion"
	superChatIAEmpresaModeloAdjuntosKey     = "ai.chat.empresa.modelo_adjuntos"
	superChatIAEmpresaModelosHabilitadosKey = "ai.chat.empresa.modelos_habilitados"
	superChatIAEmpresaModelosEsfuerzoKey    = "ai.chat.empresa.modelos_esfuerzo"

	superChatIALogicaUpdatedBySuffix = ".updated_by"

	defaultChatIAEmpresaEnabled           = true
	defaultChatIASuperEnabled             = true
	defaultChatIAPortalEnabled            = true
	defaultChatIAEmpresaMaxConsultas      = int64(10)
	defaultChatIASuperMaxConsultas        = int64(30)
	defaultChatIAEmpresaMaxGPT55Consultas = int64(2)
	defaultChatIAEmpresaStreamingEnabled  = false
	defaultChatIASuperStreamingEnabled    = false
	defaultChatIAEmpresaDBQueryEnabled    = true
	defaultChatIAEmpresaDBQueryMaxTables  = int64(25)
	defaultChatIAEmpresaDBQueryRows       = int64(8)
	defaultChatIASuperContextoAmplio      = true
	defaultChatIAEmpresaSoloLectura       = true
	defaultChatIAEmpresaModeloOperacion   = "openai:gpt-5.4-mini"
	defaultChatIAEmpresaModeloAdjuntos    = "openai:gpt-5.5"
)

func defaultChatIAEmpresaModelosEsfuerzo() map[string]string {
	return map[string]string{
		"openai:gpt-5.4-mini":  "none",
		"openai:gpt-5.5":       "medium",
		"openai:gpt-5.6-sol":   "high",
		"openai:gpt-5.6-terra": "medium",
		"openai:gpt-5.6-luna":  "low",
	}
}

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

func getChatIAPortalEnabled(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAPortalEnabledKey)
	if err != nil {
		return defaultChatIAPortalEnabled, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIAPortalEnabled), updatedAt, updatedBy, nil
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

func getChatIAEmpresaMaxGPT55ConsultasDia(dbSuper *sql.DB) (int64, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaMaxGPT55ConsultasKey)
	if err != nil {
		return defaultChatIAEmpresaMaxGPT55Consultas, "", "", err
	}
	if raw == "" {
		return defaultChatIAEmpresaMaxGPT55Consultas, updatedAt, updatedBy, nil
	}
	return parseConfigNonNegativeInt64WithDefault(raw, defaultChatIAEmpresaMaxGPT55Consultas), updatedAt, updatedBy, nil
}

func getChatIAEmpresaStreamingEnabled(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaStreamingEnabledKey)
	if err != nil {
		return defaultChatIAEmpresaStreamingEnabled, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIAEmpresaStreamingEnabled), updatedAt, updatedBy, nil
}

func getChatIAEmpresaDBQueryEnabled(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaDBQueryEnabledKey)
	if err != nil {
		return defaultChatIAEmpresaDBQueryEnabled, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIAEmpresaDBQueryEnabled), updatedAt, updatedBy, nil
}

func getChatIAEmpresaDBQueryMaxTables(dbSuper *sql.DB) (int64, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaDBQueryMaxTablesKey)
	if err != nil {
		return defaultChatIAEmpresaDBQueryMaxTables, "", "", err
	}
	if raw == "" {
		return defaultChatIAEmpresaDBQueryMaxTables, updatedAt, updatedBy, nil
	}
	value := parseConfigNonNegativeInt64WithDefault(raw, defaultChatIAEmpresaDBQueryMaxTables)
	if value <= 0 {
		value = defaultChatIAEmpresaDBQueryMaxTables
	}
	if value > 100 {
		value = 100
	}
	return value, updatedAt, updatedBy, nil
}

func getChatIAEmpresaDBQueryRows(dbSuper *sql.DB) (int64, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaDBQueryRowsKey)
	if err != nil {
		return defaultChatIAEmpresaDBQueryRows, "", "", err
	}
	if raw == "" {
		return defaultChatIAEmpresaDBQueryRows, updatedAt, updatedBy, nil
	}
	value := parseConfigNonNegativeInt64WithDefault(raw, defaultChatIAEmpresaDBQueryRows)
	if value <= 0 {
		value = defaultChatIAEmpresaDBQueryRows
	}
	if value > 30 {
		value = 30
	}
	return value, updatedAt, updatedBy, nil
}

func getChatIASuperStreamingEnabled(dbSuper *sql.DB) (bool, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIASuperStreamingEnabledKey)
	if err != nil {
		return defaultChatIASuperStreamingEnabled, "", "", err
	}
	return parseConfigBoolWithDefault(raw, defaultChatIASuperStreamingEnabled), updatedAt, updatedBy, nil
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

func getChatIAEmpresaModeloOperacion(dbSuper *sql.DB) (string, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaModeloOperacionKey)
	if err != nil {
		return defaultChatIAEmpresaModeloOperacion, "", "", err
	}
	if strings.TrimSpace(raw) == "" {
		return defaultChatIAEmpresaModeloOperacion, updatedAt, updatedBy, nil
	}
	return strings.TrimSpace(raw), updatedAt, updatedBy, nil
}

func getChatIAEmpresaModeloAdjuntos(dbSuper *sql.DB) (string, string, string, error) {
	raw, updatedAt, updatedBy, err := getSuperConfigString(dbSuper, superChatIAEmpresaModeloAdjuntosKey)
	if err != nil {
		return defaultChatIAEmpresaModeloAdjuntos, "", "", err
	}
	if strings.TrimSpace(raw) == "" {
		return defaultChatIAEmpresaModeloAdjuntos, updatedAt, updatedBy, nil
	}
	return strings.TrimSpace(raw), updatedAt, updatedBy, nil
}

func getChatIAEmpresaModelosHabilitados(dbSuper *sql.DB) (map[string]bool, error) {
	raw, _, _, err := getSuperConfigString(dbSuper, superChatIAEmpresaModelosHabilitadosKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil, err
	}
	out := map[string]bool{}
	for _, item := range strings.Split(raw, ",") {
		if id := strings.TrimSpace(item); id != "" {
			out[id] = true
		}
	}
	return out, nil
}

func getChatIAEmpresaModelosEsfuerzo(dbSuper *sql.DB) (map[string]string, error) {
	defaults := defaultChatIAEmpresaModelosEsfuerzo()
	raw, _, _, err := getSuperConfigString(dbSuper, superChatIAEmpresaModelosEsfuerzoKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return defaults, err
	}
	var stored map[string]string
	if json.Unmarshal([]byte(raw), &stored) != nil {
		return defaults, nil
	}
	for id, effort := range stored {
		if strings.TrimSpace(id) != "" && strings.TrimSpace(effort) != "" {
			defaults[id] = strings.TrimSpace(effort)
		}
	}
	return defaults, nil
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
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		switch r.Method {
		case http.MethodGet:
			empresaEnabled, empresaEnabledAt, empresaEnabledBy, err := getChatIAEmpresaEnabled(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			superEnabled, superEnabledAt, superEnabledBy, err := getChatIASuperEnabled(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			portalEnabled, portalEnabledAt, portalEnabledBy, err := getChatIAPortalEnabled(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			empresaMax, empresaMaxAt, empresaMaxBy, err := getChatIAEmpresaMaxConsultasDia(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			empresaMaxGPT55, empresaMaxGPT55At, empresaMaxGPT55By, err := getChatIAEmpresaMaxGPT55ConsultasDia(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			superMax, superMaxAt, superMaxBy, err := getChatIASuperMaxConsultasDia(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			empresaStreaming, empresaStreamingAt, empresaStreamingBy, err := getChatIAEmpresaStreamingEnabled(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			empresaDBQueryEnabled, empresaDBQueryEnabledAt, empresaDBQueryEnabledBy, err := getChatIAEmpresaDBQueryEnabled(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			empresaDBQueryMaxTables, empresaDBQueryMaxTablesAt, empresaDBQueryMaxTablesBy, err := getChatIAEmpresaDBQueryMaxTables(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			empresaDBQueryRows, empresaDBQueryRowsAt, empresaDBQueryRowsBy, err := getChatIAEmpresaDBQueryRows(dbSuper)
			if err != nil {
				http.Error(w, "No se pudo leer la configuracion del chat IA", http.StatusInternalServerError)
				return
			}
			superStreaming, superStreamingAt, superStreamingBy, err := getChatIASuperStreamingEnabled(dbSuper)
			if err != nil {
				http.Error(w, "error leyendo configuración: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_, superCtxAmplioAt, superCtxAmplioBy, err := getChatIASuperContextoAmplio(dbSuper)
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
				http.Error(w, "No se pudo leer el consumo de IA", http.StatusInternalServerError)
				return
			}
			superAllConsultasHoy, superAllTokensHoy, err := dbpkg.GetSuperAIUsoDiarioOpenAITokensGlobal(dbSuper, "", "openai", fechaUso)
			if err != nil {
				http.Error(w, "No se pudo leer el consumo de IA", http.StatusInternalServerError)
				return
			}
			empConsultasHoy, empTokensHoy, err := dbpkg.GetEmpresaAIUsoDiarioOpenAITokensGlobal(dbEmp, "openai", fechaUso)
			if err != nil {
				http.Error(w, "No se pudo leer el consumo de IA", http.StatusInternalServerError)
				return
			}

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true,
				"defaults": map[string]interface{}{
					"empresa_enabled":             defaultChatIAEmpresaEnabled,
					"empresa_max_consultas":       defaultChatIAEmpresaMaxConsultas,
					"empresa_max_gpt55_consultas": defaultChatIAEmpresaMaxGPT55Consultas,
					"empresa_streaming_enabled":   defaultChatIAEmpresaStreamingEnabled,
					"empresa_db_query_enabled":    defaultChatIAEmpresaDBQueryEnabled,
					"empresa_db_query_max_tables": defaultChatIAEmpresaDBQueryMaxTables,
					"empresa_db_query_rows":       defaultChatIAEmpresaDBQueryRows,
					"super_enabled":               defaultChatIASuperEnabled,
					"super_max_consultas":         defaultChatIASuperMaxConsultas,
					"super_streaming_enabled":     defaultChatIASuperStreamingEnabled,
					"portal_enabled":              defaultChatIAPortalEnabled,
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
					"portal_enabled": map[string]interface{}{
						"value":      portalEnabled,
						"config_key": superChatIAPortalEnabledKey,
						"updated_at": portalEnabledAt,
						"updated_by": portalEnabledBy,
					},
					"empresa_max_consultas": map[string]interface{}{
						"value":      empresaMax,
						"config_key": superChatIAEmpresaMaxConsultasKey,
						"updated_at": empresaMaxAt,
						"updated_by": empresaMaxBy,
					},
					"empresa_max_gpt55_consultas": map[string]interface{}{
						"value":      empresaMaxGPT55,
						"config_key": superChatIAEmpresaMaxGPT55ConsultasKey,
						"updated_at": empresaMaxGPT55At,
						"updated_by": empresaMaxGPT55By,
					},
					"empresa_streaming_enabled": map[string]interface{}{
						"value":      empresaStreaming,
						"config_key": superChatIAEmpresaStreamingEnabledKey,
						"updated_at": empresaStreamingAt,
						"updated_by": empresaStreamingBy,
					},
					"empresa_db_query_enabled": map[string]interface{}{
						"value":      empresaDBQueryEnabled,
						"config_key": superChatIAEmpresaDBQueryEnabledKey,
						"updated_at": empresaDBQueryEnabledAt,
						"updated_by": empresaDBQueryEnabledBy,
					},
					"empresa_db_query_max_tables": map[string]interface{}{
						"value":      empresaDBQueryMaxTables,
						"config_key": superChatIAEmpresaDBQueryMaxTablesKey,
						"updated_at": empresaDBQueryMaxTablesAt,
						"updated_by": empresaDBQueryMaxTablesBy,
					},
					"empresa_db_query_rows": map[string]interface{}{
						"value":      empresaDBQueryRows,
						"config_key": superChatIAEmpresaDBQueryRowsKey,
						"updated_at": empresaDBQueryRowsAt,
						"updated_by": empresaDBQueryRowsBy,
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
					"super_streaming_enabled": map[string]interface{}{
						"value":      superStreaming,
						"config_key": superChatIASuperStreamingEnabledKey,
						"updated_at": superStreamingAt,
						"updated_by": superStreamingBy,
					},
					"super_contexto_amplio": map[string]interface{}{
						"value":      true,
						"config_key": superChatIASuperContextoAmplioKey,
						"updated_at": superCtxAmplioAt,
						"updated_by": superCtxAmplioBy,
						"nota":       "El chat global siempre recibe metadatos completos del esquema super (conteos, columnas, roles de admin). La clave histórica en base de datos ya no desactiva este inventario.",
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
				EmpresaEnabled           *bool  `json:"empresa_enabled"`
				EmpresaMaxConsultas      *int64 `json:"empresa_max_consultas"`
				EmpresaMaxGPT55Consultas *int64 `json:"empresa_max_gpt55_consultas"`
				EmpresaStreamingEnabled  *bool  `json:"empresa_streaming_enabled"`
				EmpresaDBQueryEnabled    *bool  `json:"empresa_db_query_enabled"`
				EmpresaDBQueryMaxTables  *int64 `json:"empresa_db_query_max_tables"`
				EmpresaDBQueryRows       *int64 `json:"empresa_db_query_rows"`
				SuperEnabled             *bool  `json:"super_enabled"`
				SuperMaxConsultas        *int64 `json:"super_max_consultas"`
				SuperStreamingEnabled    *bool  `json:"super_streaming_enabled"`
				PortalEnabled            *bool  `json:"portal_enabled"`
				SuperContextoAmplio      *bool  `json:"super_contexto_amplio"`
				EmpresaSoloLectura       *bool  `json:"empresa_solo_lectura"`
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
			portalEnabled := defaultChatIAPortalEnabled
			if payload.PortalEnabled != nil {
				portalEnabled = *payload.PortalEnabled
			}
			empresaMax := defaultChatIAEmpresaMaxConsultas
			if payload.EmpresaMaxConsultas != nil {
				empresaMax = *payload.EmpresaMaxConsultas
			}
			if empresaMax < 0 {
				empresaMax = 0
			}
			empresaMaxGPT55 := defaultChatIAEmpresaMaxGPT55Consultas
			if payload.EmpresaMaxGPT55Consultas != nil {
				empresaMaxGPT55 = *payload.EmpresaMaxGPT55Consultas
			}
			if empresaMaxGPT55 < 0 {
				empresaMaxGPT55 = 0
			}
			empresaStreamingEnabled := defaultChatIAEmpresaStreamingEnabled
			if payload.EmpresaStreamingEnabled != nil {
				empresaStreamingEnabled = *payload.EmpresaStreamingEnabled
			}
			empresaDBQueryEnabled := defaultChatIAEmpresaDBQueryEnabled
			if payload.EmpresaDBQueryEnabled != nil {
				empresaDBQueryEnabled = *payload.EmpresaDBQueryEnabled
			}
			empresaDBQueryMaxTables := defaultChatIAEmpresaDBQueryMaxTables
			if payload.EmpresaDBQueryMaxTables != nil {
				empresaDBQueryMaxTables = *payload.EmpresaDBQueryMaxTables
			}
			if empresaDBQueryMaxTables <= 0 {
				empresaDBQueryMaxTables = defaultChatIAEmpresaDBQueryMaxTables
			}
			if empresaDBQueryMaxTables > 100 {
				empresaDBQueryMaxTables = 100
			}
			empresaDBQueryRows := defaultChatIAEmpresaDBQueryRows
			if payload.EmpresaDBQueryRows != nil {
				empresaDBQueryRows = *payload.EmpresaDBQueryRows
			}
			if empresaDBQueryRows <= 0 {
				empresaDBQueryRows = defaultChatIAEmpresaDBQueryRows
			}
			if empresaDBQueryRows > 30 {
				empresaDBQueryRows = 30
			}
			superMax := defaultChatIASuperMaxConsultas
			if payload.SuperMaxConsultas != nil {
				superMax = *payload.SuperMaxConsultas
			}
			if superMax < 0 {
				superMax = 0
			}
			superStreamingEnabled := defaultChatIASuperStreamingEnabled
			if payload.SuperStreamingEnabled != nil {
				superStreamingEnabled = *payload.SuperStreamingEnabled
			}
			// El chat global siempre inyecta inventario de la base super (conteos/columnas); la clave se mantiene en true.
			superCtxAmplio := true
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

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaMaxGPT55ConsultasKey, strconv.FormatInt(empresaMaxGPT55, 10), false); err != nil {
				http.Error(w, "error guardando empresa_max_gpt55_consultas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaMaxGPT55ConsultasKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaStreamingEnabledKey, strconv.FormatBool(empresaStreamingEnabled), false); err != nil {
				http.Error(w, "error guardando empresa_streaming_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaStreamingEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaDBQueryEnabledKey, strconv.FormatBool(empresaDBQueryEnabled), false); err != nil {
				http.Error(w, "error guardando empresa_db_query_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaDBQueryEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaDBQueryMaxTablesKey, strconv.FormatInt(empresaDBQueryMaxTables, 10), false); err != nil {
				http.Error(w, "error guardando empresa_db_query_max_tables: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaDBQueryMaxTablesKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaDBQueryRowsKey, strconv.FormatInt(empresaDBQueryRows, 10), false); err != nil {
				http.Error(w, "error guardando empresa_db_query_rows: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAEmpresaDBQueryRowsKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIASuperEnabledKey, strconv.FormatBool(superEnabled), false); err != nil {
				http.Error(w, "error guardando super_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIASuperEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIAPortalEnabledKey, strconv.FormatBool(portalEnabled), false); err != nil {
				http.Error(w, "error guardando portal_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIAPortalEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIASuperMaxConsultasKey, strconv.FormatInt(superMax, 10), false); err != nil {
				http.Error(w, "error guardando super_max_consultas: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIASuperMaxConsultasKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superChatIASuperStreamingEnabledKey, strconv.FormatBool(superStreamingEnabled), false); err != nil {
				http.Error(w, "error guardando super_streaming_enabled: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superChatIASuperStreamingEnabledKey+superChatIALogicaUpdatedBySuffix, adminEmail, false)

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
					"empresa_enabled":             empresaEnabled,
					"empresa_max_consultas":       empresaMax,
					"empresa_max_gpt55_consultas": empresaMaxGPT55,
					"empresa_streaming_enabled":   empresaStreamingEnabled,
					"empresa_db_query_enabled":    empresaDBQueryEnabled,
					"empresa_db_query_max_tables": empresaDBQueryMaxTables,
					"empresa_db_query_rows":       empresaDBQueryRows,
					"super_enabled":               superEnabled,
					"super_max_consultas":         superMax,
					"super_streaming_enabled":     superStreamingEnabled,
					"portal_enabled":              portalEnabled,
					"super_contexto_amplio":       superCtxAmplio,
					"empresa_solo_lectura":        empSoloLectura,
				},
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
