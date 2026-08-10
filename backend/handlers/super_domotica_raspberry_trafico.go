package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type superDomoticaRaspberryTrafficRow struct {
	dbpkg.EmpresaControlElectricoTraficoRaspberry
	EmpresaNombre string `json:"empresa_nombre"`
	Online        bool   `json:"online"`
	TotalHuman    string `json:"total_human"`
	TodayHuman    string `json:"today_human"`
}

func SuperDomoticaRaspberryTrafficHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper); !ok {
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := ensureDomoticaTunnelSchemaReady(dbEmp); err != nil {
			http.Error(w, "Metricas de tunel no disponibles", http.StatusServiceUnavailable)
			return
		}
		traffic, err := dbpkg.ListEmpresaControlElectricoTraficoRaspberry(dbEmp)
		if err != nil {
			http.Error(w, "No se pudo cargar el trafico de Raspberry Pi", http.StatusInternalServerError)
			return
		}
		empresas, _ := dbpkg.GetEmpresas(dbEmp)
		names := map[int64]string{}
		for _, empresa := range empresas {
			id := empresa.EmpresaID
			if id <= 0 {
				id = empresa.ID
			}
			names[id] = strings.TrimSpace(empresa.Nombre)
		}
		rows := make([]superDomoticaRaspberryTrafficRow, 0, len(traffic))
		var totalRX, totalTX, todayRX, todayTX int64
		online := 0
		for _, item := range traffic {
			isOnline := domoticaTunnelSeenRecently(item.LastSeen, 90*time.Second)
			if isOnline {
				online++
			}
			totalRX += item.BytesRx
			totalTX += item.BytesTx
			todayRX += item.TodayBytesRx
			todayTX += item.TodayBytesTx
			rows = append(rows, superDomoticaRaspberryTrafficRow{
				EmpresaControlElectricoTraficoRaspberry: item,
				EmpresaNombre:                           firstNonEmpty(names[item.EmpresaID], "Empresa "+strconv.FormatInt(item.EmpresaID, 10)),
				Online:                                  isOnline,
				TotalHuman:                              domoticaTunnelTrafficHuman(item.BytesRx + item.BytesTx),
				TodayHuman:                              domoticaTunnelTrafficHuman(item.TodayBytesRx + item.TodayBytesTx),
			})
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"ok":           true,
			"generated_at": time.Now().UTC().Format(time.RFC3339),
			"raspberries":  rows,
			"summary": map[string]interface{}{
				"devices":     len(rows),
				"online":      online,
				"bytes_rx":    totalRX,
				"bytes_tx":    totalTX,
				"today_rx":    todayRX,
				"today_tx":    todayTX,
				"total_human": domoticaTunnelTrafficHuman(totalRX + totalTX),
				"today_human": domoticaTunnelTrafficHuman(todayRX + todayTX),
			},
		})
	}
}

func domoticaTunnelSeenRecently(raw string, maxAge time.Duration) bool {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999-07:00", "2006-01-02 15:04:05-07:00", "2006-01-02 15:04:05.999999", "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return time.Since(parsed) >= 0 && time.Since(parsed) <= maxAge
		}
	}
	return false
}
