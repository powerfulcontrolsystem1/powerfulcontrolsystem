package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type superLicenciaVentaMes struct {
	Mes      string  `json:"mes"`
	Cantidad int64   `json:"cantidad"`
	TotalCOP float64 `json:"total_cop"`
}

type superLicenciaVentaReciente struct {
	Fecha     string  `json:"fecha"`
	Provider  string  `json:"provider"`
	EmpresaID int64   `json:"empresa_id"`
	Licencia  string  `json:"licencia"`
	Estado    string  `json:"estado"`
	TotalCOP  float64 `json:"total_cop"`
}

func SuperLicenciasVentasResumenHandler(dbSuper *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		_ = adminEmail
		if !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if dbSuper == nil {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "meses": []superLicenciaVentaMes{}, "total_cop": 0, "cantidad": 0})
			return
		}
		months := 6
		if raw := strings.TrimSpace(r.URL.Query().Get("months")); raw != "" {
			if parsed, err := strconv.Atoi(raw); err == nil && parsed >= 1 && parsed <= 24 {
				months = parsed
			}
		}
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).AddDate(0, -(months - 1), 0)
		monthIndex := map[string]int{}
		series := make([]superLicenciaVentaMes, 0, months)
		for i := 0; i < months; i++ {
			month := start.AddDate(0, i, 0).Format("2006-01")
			monthIndex[month] = i
			series = append(series, superLicenciaVentaMes{Mes: month})
		}

		query := `
			SELECT provider, COALESCE(licencia_id,0), COALESCE(empresa_id,0), COALESCE(status,''), COALESCE(raw_payload,''), COALESCE(fecha_actualizacion,''), COALESCE(fecha_creacion,''), COALESCE(licencia_activation_status,''), COALESCE(l.nombre,''), COALESCE(l.valor,0)
			FROM (
				SELECT 'epayco' AS provider, licencia_id, empresa_id, status, raw_payload, fecha_actualizacion, fecha_creacion, licencia_activation_status FROM pagos_epayco
				UNION ALL
				SELECT 'wompi' AS provider, licencia_id, empresa_id, status, raw_payload, fecha_actualizacion, fecha_creacion, licencia_activation_status FROM pagos_wompi
			) p
			LEFT JOIN licencias l ON l.id = p.licencia_id
			WHERE LEFT(COALESCE(NULLIF(p.fecha_actualizacion,''), p.fecha_creacion), 10) >= $1
			ORDER BY COALESCE(NULLIF(p.fecha_actualizacion,''), p.fecha_creacion) DESC
			LIMIT 600`
		rows, err := dbSuper.Query(query, start.Format("2006-01-02"))
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "meses": series, "total_cop": 0, "cantidad": 0, "warning": "sin lectura de pagos de licencias"})
			return
		}
		defer rows.Close()

		var total float64
		var count int64
		recent := make([]superLicenciaVentaReciente, 0, 8)
		for rows.Next() {
			var provider, status, rawPayload, updated, created, activation, licencia string
			var licenciaID, empresaID int64
			var licenciaValor float64
			if err := rows.Scan(&provider, &licenciaID, &empresaID, &status, &rawPayload, &updated, &created, &activation, &licencia, &licenciaValor); err != nil {
				continue
			}
			if !isApprovedLicensePaymentStatus(status, activation) {
				continue
			}
			fecha := firstNonEmptyLicensePayment(updated, created)
			month := normalizeLicensePaymentMonth(fecha)
			idx, ok := monthIndex[month]
			if !ok {
				continue
			}
			amount := extractLicensePaymentAmountCOP(rawPayload)
			if amount <= 0 {
				amount = licenciaValor
			}
			if amount < 0 {
				amount = 0
			}
			series[idx].Cantidad++
			series[idx].TotalCOP += amount
			total += amount
			count++
			if len(recent) < 8 {
				recent = append(recent, superLicenciaVentaReciente{
					Fecha:     fecha,
					Provider:  provider,
					EmpresaID: empresaID,
					Licencia:  licencia,
					Estado:    firstNonEmptyLicensePayment(activation, status),
					TotalCOP:  amount,
				})
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":          true,
			"months":      months,
			"meses":       series,
			"total_cop":   total,
			"cantidad":    count,
			"recientes":   recent,
			"empresa":     "Powerful Control System",
			"fuentes":     []string{"pagos_epayco", "pagos_wompi", "licencias"},
			"moneda":      "COP",
			"actualizado": time.Now().Format(time.RFC3339),
		})
	}
}

func isApprovedLicensePaymentStatus(status, activation string) bool {
	value := strings.ToLower(strings.TrimSpace(status + " " + activation))
	for _, token := range []string{"approved", "aprob", "paid", "success", "succeeded", "captured", "activated", "activada", "ok"} {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func normalizeLicensePaymentMonth(fecha string) string {
	fecha = strings.TrimSpace(fecha)
	if len(fecha) >= 7 {
		return fecha[:7]
	}
	return ""
}

func extractLicensePaymentAmountCOP(raw string) float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	var payload any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return 0
	}
	return findLicensePaymentAmount(payload, "")
}

func findLicensePaymentAmount(value any, keyHint string) float64 {
	switch v := value.(type) {
	case map[string]any:
		preferred := []string{"valor_pagado", "total_value", "amount_paid", "x_amount", "amount", "value", "precio", "total"}
		for _, key := range preferred {
			if raw, ok := v[key]; ok {
				if amount := findLicensePaymentAmount(raw, key); amount > 0 {
					return amount
				}
			}
		}
		for key, raw := range v {
			if amount := findLicensePaymentAmount(raw, key); amount > 0 {
				return amount
			}
		}
	case []any:
		for _, raw := range v {
			if amount := findLicensePaymentAmount(raw, keyHint); amount > 0 {
				return amount
			}
		}
	case float64:
		if strings.Contains(strings.ToLower(keyHint), "cent") && v > 999 {
			return v / 100
		}
		return v
	case string:
		clean := strings.NewReplacer("$", "", "COP", "", "cop", "", " ", "").Replace(strings.TrimSpace(v))
		if strings.Contains(clean, ".") && strings.Contains(clean, ",") {
			clean = strings.ReplaceAll(clean, ".", "")
			clean = strings.ReplaceAll(clean, ",", ".")
		} else if strings.Contains(clean, ",") {
			clean = strings.ReplaceAll(clean, ",", ".")
		}
		if amount, err := strconv.ParseFloat(clean, 64); err == nil {
			if strings.Contains(strings.ToLower(keyHint), "cent") && amount > 999 {
				return amount / 100
			}
			return amount
		}
	}
	return 0
}

func firstNonEmptyLicensePayment(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
