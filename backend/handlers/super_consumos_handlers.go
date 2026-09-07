package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
	"github.com/you/pos-backend/utils"
)

// Config keys (super) para consumos externos / costos aproximados.
const (
	// OpenAI (estimación monetaria opcional)
	openAIProviderKey           = "openai"
	openAICostPer1MTokensUSDKey = "ai.openai.cost_per_1m_tokens_usd" // #nosec G101 -- nombre de metrica, no token secreto.

	// Hostinger (manual/API pendiente)
	hostingerEnabledKey          = "hostinger.enabled"
	hostingerAPITokenKey         = "hostinger.api_token" // cifrado, opcional (tarjeta de configuración)
	hostingerBandwidthUsedGBKey  = "hostinger.bandwidth.used_gb"
	hostingerBandwidthLimitGBKey = "hostinger.bandwidth.limit_gb"
	hostingerDiskUsedGBKey       = "hostinger.disk.used_gb"
	hostingerDiskLimitGBKey      = "hostinger.disk.limit_gb"
	hostingerCostUSDMonthKey     = "hostinger.cost.usd_month"

	// Cursor (manual/API pendiente)
	cursorEnabledKey      = "cursor.enabled"
	cursorAPIKeyKey       = "cursor.api_key" // #nosec G101 -- ruta de configuracion; el valor se almacena cifrado.
	cursorCostUSDMonthKey = "cursor.cost.usd_month"
	cursorNotesKey        = "cursor.notes"
)

func parseFloatConfig(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil || !isFiniteFloat(f) {
		return 0
	}
	return f
}

func isFiniteFloat(f float64) bool {
	return !((f != f) || (f > 1e308) || (f < -1e308))
}

func boolToConfig(v bool) string {
	if v {
		return "1"
	}
	return "0"
}

func SuperConsumosHandler(dbEmpresas *sql.DB, dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if dbSuper == nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "db_super no disponible"})
			return
		}

		switch r.Method {
		case http.MethodGet:
			fecha := time.Now().Format("2006-01-02")

			// 1) OpenAI (tokens/consultas) desde tabla de uso diario (super y empresas)
			superConsultas, superTokens, _ := dbpkg.GetSuperAIUsoDiarioOpenAITokensGlobal(dbSuper, adminEmail, openAIProviderKey, fecha)
			empConsultas, empTokens, _ := dbpkg.GetEmpresaAIUsoDiarioOpenAITokensGlobal(dbEmpresas, openAIProviderKey, fecha)
			desdeMes := time.Now().AddDate(0, 0, -29).Format("2006-01-02")
			superMes, _ := dbpkg.GetSuperAIUsoDiarioOpenAITokensPorRango(dbSuper, adminEmail, openAIProviderKey, desdeMes, fecha)
			empMes, _ := dbpkg.GetEmpresaAIUsoDiarioOpenAITokensPorRango(dbEmpresas, openAIProviderKey, desdeMes, fecha)

			costPer1M := parseFloatConfig(mustGetConfigPlain(dbSuper, openAICostPer1MTokensUSDKey))
			var superCostUSD float64
			var empCostUSD float64
			if costPer1M > 0 {
				superCostUSD = (float64(superTokens) / 1_000_000.0) * costPer1M
				empCostUSD = (float64(empTokens) / 1_000_000.0) * costPer1M
			}

			// 2) Hostinger (valores manuales)
			hostEnabled := strings.TrimSpace(mustGetConfigPlain(dbSuper, hostingerEnabledKey))
			host := map[string]any{
				"enabled":              strings.ToLower(hostEnabled) == "1" || strings.ToLower(hostEnabled) == "true" || strings.ToLower(hostEnabled) == "on",
				"bandwidth_used_gb":    parseFloatConfig(mustGetConfigPlain(dbSuper, hostingerBandwidthUsedGBKey)),
				"bandwidth_limit_gb":   parseFloatConfig(mustGetConfigPlain(dbSuper, hostingerBandwidthLimitGBKey)),
				"disk_used_gb":         parseFloatConfig(mustGetConfigPlain(dbSuper, hostingerDiskUsedGBKey)),
				"disk_limit_gb":        parseFloatConfig(mustGetConfigPlain(dbSuper, hostingerDiskLimitGBKey)),
				"cost_usd_month":       parseFloatConfig(mustGetConfigPlain(dbSuper, hostingerCostUSDMonthKey)),
				"api_token_set":        isEncryptedValueSet(dbSuper, hostingerAPITokenKey),
				"encryption_available": utils.EncryptionAvailable(),
				"note":                 "API automática de Hostinger pendiente; por ahora se soporta captura manual (y se deja tarjeta para token).",
			}

			// 3) Cursor (valores manuales)
			cursorEnabled := strings.TrimSpace(mustGetConfigPlain(dbSuper, cursorEnabledKey))
			cur := map[string]any{
				"enabled":              strings.ToLower(cursorEnabled) == "1" || strings.ToLower(cursorEnabled) == "true" || strings.ToLower(cursorEnabled) == "on",
				"cost_usd_month":       parseFloatConfig(mustGetConfigPlain(dbSuper, cursorCostUSDMonthKey)),
				"notes":                strings.TrimSpace(mustGetConfigPlain(dbSuper, cursorNotesKey)),
				"api_key_set":          isEncryptedValueSet(dbSuper, cursorAPIKeyKey),
				"encryption_available": utils.EncryptionAvailable(),
				"note":                 "Cursor usage API no implementada; se soporta captura manual (y tarjeta para API key si se habilita en el futuro).",
			}

			// 4) Contador de errores del sistema (solo número)
			totalErrores, _ := countSuperErrores(dbSuper)
			whatsappCfg := getWhatsAppNotificationsConfig(dbSuper)
			whatsappUsage := buildWhatsAppUsageSeries(dbSuper, desdeMes, fecha)

			writeJSON(w, http.StatusOK, map[string]any{
				"ok": true,
				"openai": map[string]any{
					"fecha": fecha,
					"rango_mes": map[string]any{
						"desde": desdeMes,
						"hasta": fecha,
						"dias":  mergeOpenAIUsageSeries(superMes, empMes),
					},
					"provider": openAIProviderKey,
					"cost_per_1m_tokens_usd": func() any {
						if costPer1M > 0 {
							return costPer1M
						}
						return nil
					}(),
					"super": map[string]any{
						"consultas": superConsultas,
						"tokens":    superTokens,
						"cost_usd_est": func() any {
							if costPer1M > 0 {
								return superConsumosRound2(superCostUSD)
							}
							return nil
						}(),
					},
					"empresas": map[string]any{
						"consultas": empConsultas,
						"tokens":    empTokens,
						"cost_usd_est": func() any {
							if costPer1M > 0 {
								return superConsumosRound2(empCostUSD)
							}
							return nil
						}(),
					},
				},
				"whatsapp": map[string]any{
					"enabled":                 whatsappCfg.Enabled,
					"provider":                whatsappCfg.Provider,
					"phone_number_configured": strings.TrimSpace(whatsappCfg.PhoneNumberID) != "",
					"access_token_configured": whatsappCfg.AccessTokenConfigured,
					"test_mode":               whatsappCfg.TestMode,
					"fecha":                   fecha,
					"rango_mes": map[string]any{
						"desde": desdeMes,
						"hasta": fecha,
						"dias":  whatsappUsage,
					},
				},
				"hostinger":     host,
				"cursor":        cur,
				"errores_total": totalErrores,
			})
			return

		case http.MethodPut, http.MethodPost:
			// Guardado de valores manuales y tokens (sin exponer secretos)
			var payload struct {
				OpenAICostPer1MTokensUSD *float64 `json:"openai_cost_per_1m_tokens_usd"`

				Hostinger struct {
					Enabled          *bool    `json:"enabled"`
					APIToken         string   `json:"api_token"`
					BandwidthUsedGB  *float64 `json:"bandwidth_used_gb"`
					BandwidthLimitGB *float64 `json:"bandwidth_limit_gb"`
					DiskUsedGB       *float64 `json:"disk_used_gb"`
					DiskLimitGB      *float64 `json:"disk_limit_gb"`
					CostUSDMonth     *float64 `json:"cost_usd_month"`
				} `json:"hostinger"`

				Cursor struct {
					Enabled      *bool    `json:"enabled"`
					APIKey       string   `json:"api_key"`
					CostUSDMonth *float64 `json:"cost_usd_month"`
					Notes        string   `json:"notes"`
				} `json:"cursor"`
			}

			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "payload inválido: "+err.Error(), http.StatusBadRequest)
				return
			}

			// OpenAI pricing (opcional)
			if payload.OpenAICostPer1MTokensUSD != nil {
				v := fmt.Sprintf("%.6f", *payload.OpenAICostPer1MTokensUSD)
				v = strings.TrimRight(strings.TrimRight(v, "0"), ".")
				if err := dbpkg.SetConfigValue(dbSuper, openAICostPer1MTokensUSDKey, v, false); err != nil {
					http.Error(w, "No se pudo guardar costo OpenAI: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}

			// Hostinger
			if payload.Hostinger.Enabled != nil {
				if err := dbpkg.SetConfigValue(dbSuper, hostingerEnabledKey, boolToConfig(*payload.Hostinger.Enabled), false); err != nil {
					http.Error(w, "No se pudo guardar hostinger.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.Hostinger.APIToken) != "" {
				if !utils.EncryptionAvailable() {
					http.Error(w, "Cifrado requerido: CONFIG_ENC_KEY no está disponible", http.StatusInternalServerError)
					return
				}
				encVal, err := utils.EncryptString(strings.TrimSpace(payload.Hostinger.APIToken))
				if err != nil {
					http.Error(w, "No se pudo cifrar hostinger.api_token: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, hostingerAPITokenKey, encVal, true); err != nil {
					http.Error(w, "No se pudo guardar hostinger.api_token: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.Hostinger.BandwidthUsedGB != nil {
				_ = dbpkg.SetConfigValue(dbSuper, hostingerBandwidthUsedGBKey, fmt.Sprintf("%.3f", *payload.Hostinger.BandwidthUsedGB), false)
			}
			if payload.Hostinger.BandwidthLimitGB != nil {
				_ = dbpkg.SetConfigValue(dbSuper, hostingerBandwidthLimitGBKey, fmt.Sprintf("%.3f", *payload.Hostinger.BandwidthLimitGB), false)
			}
			if payload.Hostinger.DiskUsedGB != nil {
				_ = dbpkg.SetConfigValue(dbSuper, hostingerDiskUsedGBKey, fmt.Sprintf("%.3f", *payload.Hostinger.DiskUsedGB), false)
			}
			if payload.Hostinger.DiskLimitGB != nil {
				_ = dbpkg.SetConfigValue(dbSuper, hostingerDiskLimitGBKey, fmt.Sprintf("%.3f", *payload.Hostinger.DiskLimitGB), false)
			}
			if payload.Hostinger.CostUSDMonth != nil {
				_ = dbpkg.SetConfigValue(dbSuper, hostingerCostUSDMonthKey, fmt.Sprintf("%.2f", *payload.Hostinger.CostUSDMonth), false)
			}

			// Cursor
			if payload.Cursor.Enabled != nil {
				if err := dbpkg.SetConfigValue(dbSuper, cursorEnabledKey, boolToConfig(*payload.Cursor.Enabled), false); err != nil {
					http.Error(w, "No se pudo guardar cursor.enabled: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if strings.TrimSpace(payload.Cursor.APIKey) != "" {
				if !utils.EncryptionAvailable() {
					http.Error(w, "Cifrado requerido: CONFIG_ENC_KEY no está disponible", http.StatusInternalServerError)
					return
				}
				encVal, err := utils.EncryptString(strings.TrimSpace(payload.Cursor.APIKey))
				if err != nil {
					http.Error(w, "No se pudo cifrar cursor.api_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
				if err := dbpkg.SetConfigValue(dbSuper, cursorAPIKeyKey, encVal, true); err != nil {
					http.Error(w, "No se pudo guardar cursor.api_key: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
			if payload.Cursor.CostUSDMonth != nil {
				_ = dbpkg.SetConfigValue(dbSuper, cursorCostUSDMonthKey, fmt.Sprintf("%.2f", *payload.Cursor.CostUSDMonth), false)
			}
			if strings.TrimSpace(payload.Cursor.Notes) != "" || payload.Cursor.Notes == "" {
				_ = dbpkg.SetConfigValue(dbSuper, cursorNotesKey, strings.TrimSpace(payload.Cursor.Notes), false)
			}

			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
	}
}

func mustGetConfigPlain(dbSuper *sql.DB, key string) string {
	if dbSuper == nil {
		return ""
	}
	v, err := getDecryptedConfigValue(dbSuper, key)
	if err == nil && strings.TrimSpace(v) != "" {
		return v
	}
	raw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key)
	return strings.TrimSpace(raw)
}

func isEncryptedValueSet(dbSuper *sql.DB, key string) bool {
	if dbSuper == nil {
		return false
	}
	_, encrypted, _, updated, _ := dbpkg.GetConfigEntry(dbSuper, key)
	if encrypted && strings.TrimSpace(updated) != "" {
		return true
	}
	// Si se guardó sin cifrar por alguna razón, igual cuenta como "set".
	raw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key)
	return strings.TrimSpace(raw) != ""
}

func countSuperErrores(dbSuper *sql.DB) (int64, error) {
	filter := dbpkg.SuperErrorSistemaFiltro{
		EmpresaID: 0,
		Nivel:     "",
		TipoError: "",
		Desde:     "",
		Hasta:     "",
		Search:    "",
		Limit:     1,
		Offset:    0,
	}
	_, total, _, err := dbpkg.ListSuperErroresSistema(dbSuper, filter)
	if err != nil {
		return 0, err
	}
	return total, nil
}

func buildWhatsAppUsageSeries(dbSuper *sql.DB, desde, hasta string) []map[string]any {
	out := make([]map[string]any, 0, 30)
	start, err := time.Parse("2006-01-02", strings.TrimSpace(desde))
	if err != nil {
		start = time.Now().AddDate(0, 0, -29)
	}
	end, err := time.Parse("2006-01-02", strings.TrimSpace(hasta))
	if err != nil {
		end = time.Now()
	}
	if start.After(end) {
		start = end.AddDate(0, 0, -29)
	}
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		fecha := d.Format("2006-01-02")
		out = append(out, map[string]any{
			"fecha":    fecha,
			"total":    int64Config(dbSuper, whatsAppUsageCounterKey(fecha, "total")),
			"sent":     int64Config(dbSuper, whatsAppUsageCounterKey(fecha, "sent")),
			"captured": int64Config(dbSuper, whatsAppUsageCounterKey(fecha, "captured")),
			"errors":   int64Config(dbSuper, whatsAppUsageCounterKey(fecha, "errors")),
			"disabled": int64Config(dbSuper, whatsAppUsageCounterKey(fecha, "disabled")),
		})
	}
	return out
}

func int64Config(dbSuper *sql.DB, key string) int64 {
	raw, _, _, _, _ := dbpkg.GetConfigEntry(dbSuper, key)
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func superConsumosRound2(v float64) float64 {
	s := fmt.Sprintf("%.2f", v)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func mergeOpenAIUsageSeries(superRows, empresaRows []dbpkg.AIUsoDiarioResumen) []map[string]any {
	type agg struct {
		SuperConsultas   int64
		SuperTokens      int64
		EmpresaConsultas int64
		EmpresaTokens    int64
	}
	byDate := map[string]*agg{}
	add := func(rows []dbpkg.AIUsoDiarioResumen, super bool) {
		for _, row := range rows {
			fecha := strings.TrimSpace(row.Fecha)
			if fecha == "" {
				continue
			}
			item := byDate[fecha]
			if item == nil {
				item = &agg{}
				byDate[fecha] = item
			}
			if super {
				item.SuperConsultas += row.Consultas
				item.SuperTokens += row.Tokens
			} else {
				item.EmpresaConsultas += row.Consultas
				item.EmpresaTokens += row.Tokens
			}
		}
	}
	add(superRows, true)
	add(empresaRows, false)
	out := make([]map[string]any, 0, 30)
	start := time.Now().AddDate(0, 0, -29)
	for i := 0; i < 30; i++ {
		fecha := start.AddDate(0, 0, i).Format("2006-01-02")
		item := byDate[fecha]
		if item == nil {
			item = &agg{}
		}
		out = append(out, map[string]any{
			"fecha":             fecha,
			"super_consultas":   item.SuperConsultas,
			"super_tokens":      item.SuperTokens,
			"empresa_consultas": item.EmpresaConsultas,
			"empresa_tokens":    item.EmpresaTokens,
			"consultas_total":   item.SuperConsultas + item.EmpresaConsultas,
			"tokens_total":      item.SuperTokens + item.EmpresaTokens,
		})
	}
	return out
}
