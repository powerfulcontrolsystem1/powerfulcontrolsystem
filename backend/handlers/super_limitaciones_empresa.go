package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	dbpkg "github.com/you/pos-backend/db"
)

const (
	superEmpresaLimitRustDeskMinutesKey      = "empresa.limitaciones.rustdesk.max_minutos"
	superEmpresaLimitAIConsultasKey          = "empresa.limitaciones.ai.max_consultas"
	superEmpresaLimitGPSDispositivosKey      = "empresa.limitaciones.gps.max_dispositivos"
	superEmpresaLimitDBMaxGBKey              = "empresa.limitaciones.db.max_gb"
	superEmpresaLimitLegacyNextcloudMaxGBKey = "empresa.limitaciones.nextcloud.max_gb"
	superEmpresaLimitAPIRequestsPerMinuteKey = "empresa.limitaciones.api.max_requests_minuto"
	superEmpresaLimitDBQueriesPerMinuteKey   = "empresa.limitaciones.db.max_consultas_minuto"

	superEmpresaLimitUpdatedByKeySuffix = ".updated_by"

	defaultEmpresaRustDeskMaxMinutos   = int64(30)
	defaultEmpresaAIMaxConsultas       = int64(10)
	defaultEmpresaGPSMaxDispositivos   = int64(2)
	defaultEmpresaDBMaxGB              = int64(1)
	defaultEmpresaAPIRequestsPerMinute = int64(600)
	defaultEmpresaDBQueriesPerMinute   = int64(120)
)

func parsePositiveInt64OrDefault(raw string, fallback int64) int64 {
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

// MaxGPSDispositivosPorEmpresa devuelve el tope configurado de dispositivos GPS por empresa (pcs_superadministrador).
func MaxGPSDispositivosPorEmpresa(dbSuper *sql.DB) (int64, error) {
	v, _, _, err := getLimitacionInt64(dbSuper, superEmpresaLimitGPSDispositivosKey, defaultEmpresaGPSMaxDispositivos)
	return v, err
}

func getLimitacionInt64(dbSuper *sql.DB, key string, fallback int64) (int64, string, string, error) {
	val, _, _, updatedAt, err := dbpkg.GetConfigEntry(dbSuper, key)
	if err != nil {
		return fallback, "", "", err
	}
	updatedBy, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key+superEmpresaLimitUpdatedByKeySuffix)
	if strings.TrimSpace(val) == "" {
		return fallback, strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
	}
	return parsePositiveInt64OrDefault(val, fallback), strings.TrimSpace(updatedAt), strings.TrimSpace(updatedBy), nil
}

func getLimitacionInt64WithLegacy(dbSuper *sql.DB, key, legacyKey string, fallback int64) (int64, string, string, error) {
	val, updatedAt, updatedBy, err := getLimitacionInt64(dbSuper, key, fallback)
	if err != nil || strings.TrimSpace(legacyKey) == "" {
		return val, updatedAt, updatedBy, err
	}
	currentRaw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key)
	if strings.TrimSpace(currentRaw) != "" {
		return val, updatedAt, updatedBy, nil
	}
	return getLimitacionInt64(dbSuper, legacyKey, fallback)
}

// SuperEmpresaLimitacionesConfigHandler permite configurar límites por empresa desde super.
// Persistencia: tabla configuraciones en pcs_superadministrador.
func SuperEmpresaLimitacionesConfigHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			rustdeskMinutes, rustdeskUpdatedAt, rustdeskUpdatedBy, err := getLimitacionInt64(dbSuper, superEmpresaLimitRustDeskMinutesKey, defaultEmpresaRustDeskMaxMinutos)
			if err != nil {
				http.Error(w, "error leyendo limitaciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			aiConsultas, aiUpdatedAt, aiUpdatedBy, err := getLimitacionInt64(dbSuper, superEmpresaLimitAIConsultasKey, defaultEmpresaAIMaxConsultas)
			if err != nil {
				http.Error(w, "error leyendo limitaciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			gpsMax, gpsUpdatedAt, gpsUpdatedBy, err := getLimitacionInt64(dbSuper, superEmpresaLimitGPSDispositivosKey, defaultEmpresaGPSMaxDispositivos)
			if err != nil {
				http.Error(w, "error leyendo limitaciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			dbMaxGB, dbMaxGBUpdatedAt, dbMaxGBUpdatedBy, err := getLimitacionInt64WithLegacy(dbSuper, superEmpresaLimitDBMaxGBKey, superEmpresaLimitLegacyNextcloudMaxGBKey, defaultEmpresaDBMaxGB)
			if err != nil {
				http.Error(w, "error leyendo limitaciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			apiRequests, apiRequestsUpdatedAt, apiRequestsUpdatedBy, err := getLimitacionInt64(dbSuper, superEmpresaLimitAPIRequestsPerMinuteKey, defaultEmpresaAPIRequestsPerMinute)
			if err != nil {
				http.Error(w, "error leyendo limitaciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			dbQueries, dbQueriesUpdatedAt, dbQueriesUpdatedBy, err := getLimitacionInt64(dbSuper, superEmpresaLimitDBQueriesPerMinuteKey, defaultEmpresaDBQueriesPerMinute)
			if err != nil {
				http.Error(w, "error leyendo limitaciones: "+err.Error(), http.StatusInternalServerError)
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true,
				"defaults": map[string]int64{
					"rustdesk_max_minutos":    defaultEmpresaRustDeskMaxMinutos,
					"ai_max_consultas":        defaultEmpresaAIMaxConsultas,
					"gps_max_dispositivos":    defaultEmpresaGPSMaxDispositivos,
					"db_max_gb":               defaultEmpresaDBMaxGB,
					"api_max_requests_minuto": defaultEmpresaAPIRequestsPerMinute,
					"db_max_consultas_minuto": defaultEmpresaDBQueriesPerMinute,
				},
				"values": map[string]interface{}{
					"rustdesk_max_minutos": map[string]interface{}{
						"value":      rustdeskMinutes,
						"updated_at": rustdeskUpdatedAt,
						"updated_by": rustdeskUpdatedBy,
						"config_key": superEmpresaLimitRustDeskMinutesKey,
					},
					"ai_max_consultas": map[string]interface{}{
						"value":      aiConsultas,
						"updated_at": aiUpdatedAt,
						"updated_by": aiUpdatedBy,
						"config_key": superEmpresaLimitAIConsultasKey,
					},
					"gps_max_dispositivos": map[string]interface{}{
						"value":      gpsMax,
						"updated_at": gpsUpdatedAt,
						"updated_by": gpsUpdatedBy,
						"config_key": superEmpresaLimitGPSDispositivosKey,
					},
					"db_max_gb": map[string]interface{}{
						"value":      dbMaxGB,
						"updated_at": dbMaxGBUpdatedAt,
						"updated_by": dbMaxGBUpdatedBy,
						"config_key": superEmpresaLimitDBMaxGBKey,
					},
					"api_max_requests_minuto": map[string]interface{}{
						"value":      apiRequests,
						"updated_at": apiRequestsUpdatedAt,
						"updated_by": apiRequestsUpdatedBy,
						"config_key": superEmpresaLimitAPIRequestsPerMinuteKey,
					},
					"db_max_consultas_minuto": map[string]interface{}{
						"value":      dbQueries,
						"updated_at": dbQueriesUpdatedAt,
						"updated_by": dbQueriesUpdatedBy,
						"config_key": superEmpresaLimitDBQueriesPerMinuteKey,
					},
				},
			})
			return

		case http.MethodPut, http.MethodPost:
			var payload struct {
				RustDeskMaxMinutos   *int64 `json:"rustdesk_max_minutos"`
				AIMaxConsultas       *int64 `json:"ai_max_consultas"`
				GPSMaxDispositivos   *int64 `json:"gps_max_dispositivos"`
				DBMaxGB              *int64 `json:"db_max_gb"`
				NextcloudMaxGB       *int64 `json:"nextcloud_max_gb"`
				APIMaxRequestsMinuto *int64 `json:"api_max_requests_minuto"`
				DBMaxConsultasMinuto *int64 `json:"db_max_consultas_minuto"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload: "+err.Error(), http.StatusBadRequest)
				return
			}

			// Normalizar (no negativos)
			rustdesk := defaultEmpresaRustDeskMaxMinutos
			if payload.RustDeskMaxMinutos != nil {
				rustdesk = *payload.RustDeskMaxMinutos
			}
			if rustdesk < 0 {
				rustdesk = 0
			}
			ai := defaultEmpresaAIMaxConsultas
			if payload.AIMaxConsultas != nil {
				ai = *payload.AIMaxConsultas
			}
			if ai < 0 {
				ai = 0
			}
			gps := defaultEmpresaGPSMaxDispositivos
			if payload.GPSMaxDispositivos != nil {
				gps = *payload.GPSMaxDispositivos
			}
			if gps < 0 {
				gps = 0
			}
			dbMaxGB := defaultEmpresaDBMaxGB
			if payload.DBMaxGB != nil {
				dbMaxGB = *payload.DBMaxGB
			} else if payload.NextcloudMaxGB != nil {
				dbMaxGB = *payload.NextcloudMaxGB
			}
			if dbMaxGB < 0 {
				dbMaxGB = 0
			}
			apiRequests := defaultEmpresaAPIRequestsPerMinute
			if payload.APIMaxRequestsMinuto != nil {
				apiRequests = *payload.APIMaxRequestsMinuto
			}
			if apiRequests < 0 {
				apiRequests = 0
			}
			dbQueries := defaultEmpresaDBQueriesPerMinute
			if payload.DBMaxConsultasMinuto != nil {
				dbQueries = *payload.DBMaxConsultasMinuto
			}
			if dbQueries < 0 {
				dbQueries = 0
			}

			adminEmail := strings.TrimSpace(adminEmailFromRequest(r))
			if err := dbpkg.SetConfigValue(dbSuper, superEmpresaLimitRustDeskMinutesKey, strconv.FormatInt(rustdesk, 10), false); err != nil {
				http.Error(w, "error guardando rustdesk limit: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superEmpresaLimitRustDeskMinutesKey+superEmpresaLimitUpdatedByKeySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superEmpresaLimitAIConsultasKey, strconv.FormatInt(ai, 10), false); err != nil {
				http.Error(w, "error guardando ai limit: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superEmpresaLimitAIConsultasKey+superEmpresaLimitUpdatedByKeySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superEmpresaLimitGPSDispositivosKey, strconv.FormatInt(gps, 10), false); err != nil {
				http.Error(w, "error guardando limite gps: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superEmpresaLimitGPSDispositivosKey+superEmpresaLimitUpdatedByKeySuffix, adminEmail, false)

			if err := dbpkg.SetConfigValue(dbSuper, superEmpresaLimitDBMaxGBKey, strconv.FormatInt(dbMaxGB, 10), false); err != nil {
				http.Error(w, "error guardando limite maximo de base de datos: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superEmpresaLimitDBMaxGBKey+superEmpresaLimitUpdatedByKeySuffix, adminEmail, false)
			if err := dbpkg.SetConfigValue(dbSuper, superEmpresaLimitAPIRequestsPerMinuteKey, strconv.FormatInt(apiRequests, 10), false); err != nil {
				http.Error(w, "error guardando limite api: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superEmpresaLimitAPIRequestsPerMinuteKey+superEmpresaLimitUpdatedByKeySuffix, adminEmail, false)
			if err := dbpkg.SetConfigValue(dbSuper, superEmpresaLimitDBQueriesPerMinuteKey, strconv.FormatInt(dbQueries, 10), false); err != nil {
				http.Error(w, "error guardando limite consultas db: "+err.Error(), http.StatusInternalServerError)
				return
			}
			_ = dbpkg.SetConfigValue(dbSuper, superEmpresaLimitDBQueriesPerMinuteKey+superEmpresaLimitUpdatedByKeySuffix, adminEmail, false)

			writeJSON(w, http.StatusOK, map[string]interface{}{
				"ok": true,
				"saved": map[string]int64{
					"rustdesk_max_minutos":    rustdesk,
					"ai_max_consultas":        ai,
					"gps_max_dispositivos":    gps,
					"db_max_gb":               dbMaxGB,
					"api_max_requests_minuto": apiRequests,
					"db_max_consultas_minuto": dbQueries,
				},
			})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}
