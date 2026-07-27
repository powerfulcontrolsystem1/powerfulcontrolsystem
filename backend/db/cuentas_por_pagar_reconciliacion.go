package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EmpresaCxPReconciliacionItem is a read-only comparison of one documentary
// reference. It deliberately never proposes an automatic correction.
type EmpresaCxPReconciliacionItem struct {
	Clave             string  `json:"clave"`
	Estado            string  `json:"estado"`
	Proveedor         string  `json:"proveedor"`
	Documento         string  `json:"documento"`
	Moneda            string  `json:"moneda"`
	CanonicoOriginal  float64 `json:"canonico_valor_original"`
	HistoricoOriginal float64 `json:"historico_valor_original"`
	CanonicoSaldo     float64 `json:"canonico_saldo"`
	HistoricoSaldo    float64 `json:"historico_saldo"`
}

type EmpresaCxPReconciliacion struct {
	EmpresaID              int64                          `json:"empresa_id"`
	FuenteCanonica         string                         `json:"fuente_canonica"`
	FuenteHistorica        string                         `json:"fuente_historica"`
	RegistrosCanonicos     int                            `json:"registros_canonicos"`
	RegistrosHistoricos    int                            `json:"registros_historicos"`
	SoloCanonica           int                            `json:"solo_canonica"`
	SoloHistorica          int                            `json:"solo_historica"`
	ConDiferencias         int                            `json:"con_diferencias"`
	ValorCanonico          float64                        `json:"valor_canonico"`
	ValorHistorico         float64                        `json:"valor_historico"`
	SaldoCanonico          float64                        `json:"saldo_canonico"`
	SaldoHistorico         float64                        `json:"saldo_historico"`
	DiferenciaValor        float64                        `json:"diferencia_valor"`
	DiferenciaSaldo        float64                        `json:"diferencia_saldo"`
	Items                  []EmpresaCxPReconciliacionItem `json:"items"`
	SoloLectura            bool                           `json:"solo_lectura"`
	RequiereRevisionHumana bool                           `json:"requiere_revision_humana"`
}

type empresaCxPReconciliacionRow struct {
	Proveedor string
	Documento string
	Moneda    string
	Original  float64
	Saldo     float64
}

// BuildEmpresaCxPReconciliacion compares the canonical CxP source against the
// frozen historical ledger for one tenant. It does not create, update, migrate
// or delete any record.
func BuildEmpresaCxPReconciliacion(dbConn *sql.DB, empresaID int64) (EmpresaCxPReconciliacion, error) {
	report := EmpresaCxPReconciliacion{
		EmpresaID: empresaID, FuenteCanonica: "empresa_cuentas_por_pagar", FuenteHistorica: "empresa_contabilidad_cartera_cxp",
		Items: []EmpresaCxPReconciliacionItem{}, SoloLectura: true,
	}
	if dbConn == nil || empresaID <= 0 {
		return report, fmt.Errorf("empresa_id es obligatorio")
	}
	rows, err := ExecQueryCompat(dbConn, `SELECT COALESCE(proveedor_nombre,''), COALESCE(documento_codigo,''), COALESCE(moneda,'COP'), COALESCE(valor_original,0), COALESCE(saldo,0)
		FROM empresa_cuentas_por_pagar WHERE empresa_id=? AND COALESCE(estado,'activo')='activo' ORDER BY id`, empresaID)
	if err != nil {
		return report, err
	}
	defer rows.Close()
	canonical := []empresaCxPReconciliacionRow{}
	for rows.Next() {
		var row empresaCxPReconciliacionRow
		if err := rows.Scan(&row.Proveedor, &row.Documento, &row.Moneda, &row.Original, &row.Saldo); err != nil {
			return report, err
		}
		canonical = append(canonical, row)
	}
	if err := rows.Err(); err != nil {
		return report, err
	}
	historicalRows, err := ListEmpresaCarteraCXP(dbConn, empresaID, "cxp", "")
	if err != nil {
		return report, err
	}
	historical := make([]empresaCxPReconciliacionRow, 0, len(historicalRows))
	for _, row := range historicalRows {
		historical = append(historical, empresaCxPReconciliacionRow{Proveedor: row.TerceroNombre, Documento: row.Documento, Moneda: "COP", Original: row.ValorOriginal, Saldo: row.Saldo})
	}
	return reconcileEmpresaCxPRows(empresaID, canonical, historical), nil
}

func reconcileEmpresaCxPRows(empresaID int64, canonical, historical []empresaCxPReconciliacionRow) EmpresaCxPReconciliacion {
	report := EmpresaCxPReconciliacion{EmpresaID: empresaID, FuenteCanonica: "empresa_cuentas_por_pagar", FuenteHistorica: "empresa_contabilidad_cartera_cxp", RegistrosCanonicos: len(canonical), RegistrosHistoricos: len(historical), Items: []EmpresaCxPReconciliacionItem{}, SoloLectura: true}
	canon := aggregateEmpresaCxPRows(canonical)
	historic := aggregateEmpresaCxPRows(historical)
	for _, row := range canonical {
		report.ValorCanonico += row.Original
		report.SaldoCanonico += row.Saldo
	}
	for _, row := range historical {
		report.ValorHistorico += row.Original
		report.SaldoHistorico += row.Saldo
	}
	keys := map[string]bool{}
	for key := range canon {
		keys[key] = true
	}
	for key := range historic {
		keys[key] = true
	}
	for key := range keys {
		c, hasC := canon[key]
		h, hasH := historic[key]
		item := EmpresaCxPReconciliacionItem{Clave: key, Proveedor: firstNonEmpty(c.Proveedor, h.Proveedor), Documento: firstNonEmpty(c.Documento, h.Documento), Moneda: firstNonEmpty(c.Moneda, h.Moneda), CanonicoOriginal: c.Original, HistoricoOriginal: h.Original, CanonicoSaldo: c.Saldo, HistoricoSaldo: h.Saldo}
		switch {
		case !hasH:
			item.Estado = "solo_canonica"
			report.SoloCanonica++
		case !hasC:
			item.Estado = "solo_historica"
			report.SoloHistorica++
		case !empresaCxPAmountsEqual(c.Original, h.Original) || !empresaCxPAmountsEqual(c.Saldo, h.Saldo) || !strings.EqualFold(c.Moneda, h.Moneda):
			item.Estado = "diferencia"
			report.ConDiferencias++
		default:
			continue
		}
		if len(report.Items) < 300 {
			report.Items = append(report.Items, item)
		}
	}
	report.ValorCanonico = roundReportesMoney(report.ValorCanonico)
	report.ValorHistorico = roundReportesMoney(report.ValorHistorico)
	report.SaldoCanonico = roundReportesMoney(report.SaldoCanonico)
	report.SaldoHistorico = roundReportesMoney(report.SaldoHistorico)
	report.DiferenciaValor = roundReportesMoney(report.ValorCanonico - report.ValorHistorico)
	report.DiferenciaSaldo = roundReportesMoney(report.SaldoCanonico - report.SaldoHistorico)
	report.RequiereRevisionHumana = report.SoloCanonica > 0 || report.SoloHistorica > 0 || report.ConDiferencias > 0
	return report
}

func aggregateEmpresaCxPRows(rows []empresaCxPReconciliacionRow) map[string]empresaCxPReconciliacionRow {
	out := map[string]empresaCxPReconciliacionRow{}
	for _, row := range rows {
		key := empresaCxPReconciliacionKey(row)
		current := out[key]
		current.Proveedor = firstNonEmpty(current.Proveedor, row.Proveedor)
		current.Documento = firstNonEmpty(current.Documento, row.Documento)
		current.Moneda = firstNonEmpty(current.Moneda, strings.ToUpper(strings.TrimSpace(row.Moneda)))
		current.Original += row.Original
		current.Saldo += row.Saldo
		out[key] = current
	}
	return out
}
func empresaCxPReconciliacionKey(row empresaCxPReconciliacionRow) string {
	return strings.ToUpper(strings.TrimSpace(row.Documento)) + "|" + strings.ToUpper(strings.TrimSpace(row.Proveedor))
}
func empresaCxPAmountsEqual(a, b float64) bool { return roundReportesMoney(a) == roundReportesMoney(b) }
