package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

type superDomoticaRaspberryTrafficRow struct {
	dbpkg.EmpresaControlElectricoTraficoRaspberry
	EmpresaNombre string                                             `json:"empresa_nombre"`
	Online        bool                                               `json:"online"`
	TotalHuman    string                                             `json:"total_human"`
	TodayHuman    string                                             `json:"today_human"`
	MonthHuman    string                                             `json:"month_human"`
	Policy        dbpkg.EmpresaControlElectricoTunnelPolicy          `json:"policy"`
	Bandwidth     dbpkg.EmpresaControlElectricoTunnelBandwidthStatus `json:"bandwidth"`
}

type superDomoticaCompanyTrafficRow struct {
	EmpresaID     int64                                              `json:"empresa_id"`
	EmpresaNombre string                                             `json:"empresa_nombre"`
	Devices       int                                                `json:"devices"`
	Online        int                                                `json:"online"`
	TodayBytes    int64                                              `json:"today_bytes"`
	MonthBytes    int64                                              `json:"month_bytes"`
	TotalBytes    int64                                              `json:"total_bytes"`
	TodayHuman    string                                             `json:"today_human"`
	MonthHuman    string                                             `json:"month_human"`
	TotalHuman    string                                             `json:"total_human"`
	Policy        dbpkg.EmpresaControlElectricoTunnelPolicy          `json:"policy"`
	Bandwidth     dbpkg.EmpresaControlElectricoTunnelBandwidthStatus `json:"bandwidth"`
}

type superDomoticaTrafficAggregate struct {
	Rows      []superDomoticaRaspberryTrafficRow
	Companies []superDomoticaCompanyTrafficRow
	TotalRX   int64
	TotalTX   int64
	TodayRX   int64
	TodayTX   int64
	Online    int
	Alerts    int
}

func SuperDomoticaRaspberryTrafficHandler(dbSuper, dbEmp *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminEmail, ok := paginaPrincipalRequireSuperAdmin(w, r, dbSuper)
		if !ok {
			return
		}
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			http.Error(w, "Metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		if err := ensureDomoticaTunnelSchemaReady(dbEmp); err != nil {
			http.Error(w, "Metricas de tunel no disponibles", http.StatusServiceUnavailable)
			return
		}
		if r.Method == http.MethodPut {
			handleSuperDomoticaTunnelPolicyUpdate(w, r, dbEmp, adminEmail)
			return
		}
		handleSuperDomoticaTunnelTrafficList(w, dbEmp)
	}
}

func handleSuperDomoticaTunnelPolicyUpdate(w http.ResponseWriter, r *http.Request, dbEmp *sql.DB, adminEmail string) {
	var policy dbpkg.EmpresaControlElectricoTunnelPolicy
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&policy); err != nil {
		http.Error(w, "Datos de limite invalidos", http.StatusBadRequest)
		return
	}
	if policy.EmpresaID <= 0 {
		http.Error(w, "empresa_id es obligatorio", http.StatusBadRequest)
		return
	}
	if _, err := dbpkg.GetEmpresaByScopeID(dbEmp, policy.EmpresaID); err != nil {
		http.Error(w, "Empresa no encontrada", http.StatusNotFound)
		return
	}
	policy.UsuarioCreador = adminEmail
	saved, err := dbpkg.UpsertEmpresaControlElectricoTunnelPolicy(dbEmp, policy)
	if err != nil {
		http.Error(w, "No se pudo guardar el limite", http.StatusInternalServerError)
		return
	}
	status, err := dbpkg.EvaluateEmpresaControlElectricoTunnelBandwidth(dbEmp, policy.EmpresaID, time.Now().UTC())
	if err != nil {
		http.Error(w, "Limite guardado, pero no se pudo recalcular el consumo", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"ok": true, "policy": saved, "bandwidth": status})
}

func handleSuperDomoticaTunnelTrafficList(w http.ResponseWriter, dbEmp *sql.DB) {
	traffic, err := dbpkg.ListEmpresaControlElectricoTraficoRaspberry(dbEmp)
	if err != nil {
		http.Error(w, "No se pudo cargar el trafico de Raspberry Pi", http.StatusInternalServerError)
		return
	}
	aggregate, err := buildSuperDomoticaTrafficAggregate(dbEmp, traffic, superDomoticaEmpresaNames(dbEmp))
	if err != nil {
		http.Error(w, "No se pudo cargar el limite de la empresa", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"ok": true, "generated_at": time.Now().UTC().Format(time.RFC3339),
		"raspberries": aggregate.Rows, "companies": aggregate.Companies,
		"summary": map[string]interface{}{
			"devices": len(aggregate.Rows), "online": aggregate.Online,
			"companies": len(aggregate.Companies), "alerts": aggregate.Alerts,
			"bytes_rx": aggregate.TotalRX, "bytes_tx": aggregate.TotalTX,
			"today_rx": aggregate.TodayRX, "today_tx": aggregate.TodayTX,
			"total_human": domoticaTunnelTrafficHuman(aggregate.TotalRX + aggregate.TotalTX),
			"today_human": domoticaTunnelTrafficHuman(aggregate.TodayRX + aggregate.TodayTX),
		},
	})
}

func superDomoticaEmpresaNames(dbEmp *sql.DB) map[int64]string {
	empresas, _ := dbpkg.GetEmpresas(dbEmp)
	names := make(map[int64]string, len(empresas))
	for _, empresa := range empresas {
		id := empresa.EmpresaID
		if id <= 0 {
			id = empresa.ID
		}
		names[id] = strings.TrimSpace(empresa.Nombre)
	}
	return names
}

func buildSuperDomoticaTrafficAggregate(dbEmp *sql.DB, traffic []dbpkg.EmpresaControlElectricoTraficoRaspberry, names map[int64]string) (superDomoticaTrafficAggregate, error) {
	result := superDomoticaTrafficAggregate{Rows: make([]superDomoticaRaspberryTrafficRow, 0, len(traffic))}
	companyRows := map[int64]*superDomoticaCompanyTrafficRow{}
	companyPolicies := map[int64]dbpkg.EmpresaControlElectricoTunnelPolicy{}
	now := time.Now().UTC()
	for _, item := range traffic {
		policy, found := companyPolicies[item.EmpresaID]
		if !found {
			var err error
			policy, err = dbpkg.GetEmpresaControlElectricoTunnelPolicy(dbEmp, item.EmpresaID)
			if err != nil {
				return result, err
			}
			companyPolicies[item.EmpresaID] = policy
		}
		monthBytes := item.MonthBytesRx + item.MonthBytesTx
		isOnline := domoticaTunnelSeenRecently(item.LastSeen, 90*time.Second)
		name := firstNonEmpty(names[item.EmpresaID], "Empresa "+strconv.FormatInt(item.EmpresaID, 10))
		result.Rows = append(result.Rows, superDomoticaRaspberryTrafficRow{
			EmpresaControlElectricoTraficoRaspberry: item, EmpresaNombre: name, Online: isOnline,
			TotalHuman: domoticaTunnelTrafficHuman(item.BytesRx + item.BytesTx),
			TodayHuman: domoticaTunnelTrafficHuman(item.TodayBytesRx + item.TodayBytesTx),
			MonthHuman: domoticaTunnelTrafficHuman(monthBytes), Policy: policy,
		})
		company := companyRows[item.EmpresaID]
		if company == nil {
			company = &superDomoticaCompanyTrafficRow{EmpresaID: item.EmpresaID, EmpresaNombre: name, Policy: policy}
			companyRows[item.EmpresaID] = company
		}
		company.Devices++
		if isOnline {
			company.Online++
			result.Online++
		}
		company.TodayBytes += item.TodayBytesRx + item.TodayBytesTx
		company.MonthBytes += monthBytes
		company.TotalBytes += item.BytesRx + item.BytesTx
		result.TotalRX += item.BytesRx
		result.TotalTX += item.BytesTx
		result.TodayRX += item.TodayBytesRx
		result.TodayTX += item.TodayBytesTx
	}
	result.Companies = make([]superDomoticaCompanyTrafficRow, 0, len(companyRows))
	for _, company := range companyRows {
		company.TodayHuman = domoticaTunnelTrafficHuman(company.TodayBytes)
		company.MonthHuman = domoticaTunnelTrafficHuman(company.MonthBytes)
		company.TotalHuman = domoticaTunnelTrafficHuman(company.TotalBytes)
		company.Bandwidth = dbpkg.BuildEmpresaControlElectricoTunnelBandwidthStatus(company.Policy, company.MonthBytes, now)
		if company.Bandwidth.Nivel != "normal" {
			result.Alerts++
		}
		result.Companies = append(result.Companies, *company)
	}
	sort.Slice(result.Companies, func(i, j int) bool {
		if result.Companies[i].Bandwidth.Porcentaje == result.Companies[j].Bandwidth.Porcentaje {
			return result.Companies[i].EmpresaNombre < result.Companies[j].EmpresaNombre
		}
		return result.Companies[i].Bandwidth.Porcentaje > result.Companies[j].Bandwidth.Porcentaje
	})
	for i := range result.Rows {
		result.Rows[i].Bandwidth = companyRows[result.Rows[i].EmpresaID].Bandwidth
	}
	return result, nil
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
