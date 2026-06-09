package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// EmpresaFinanzasMovimientoBancario representa una linea de extracto bancario por empresa.
type EmpresaFinanzasMovimientoBancario struct {
	ID                   int64   `json:"id"`
	EmpresaID            int64   `json:"empresa_id"`
	PeriodoContable      string  `json:"periodo_contable"`
	FechaMovimiento      string  `json:"fecha_movimiento"`
	FechaValor           string  `json:"fecha_valor"`
	CuentaBancaria       string  `json:"cuenta_bancaria"`
	BancoNombre          string  `json:"banco_nombre"`
	TipoMovimiento       string  `json:"tipo_movimiento"`
	Descripcion          string  `json:"descripcion"`
	ReferenciaBancaria   string  `json:"referencia_bancaria"`
	DocumentoCodigo      string  `json:"documento_codigo"`
	Moneda               string  `json:"moneda"`
	Monto                float64 `json:"monto"`
	Total                float64 `json:"total"`
	MovimientoFinanzasID int64   `json:"movimiento_finanzas_id"`
	EstadoConciliacion   string  `json:"estado_conciliacion"`
	ConciliadoEn         string  `json:"conciliado_en"`
	ConciliadoPor        string  `json:"conciliado_por"`
	Origen               string  `json:"origen"`
	HashMovimiento       string  `json:"hash_movimiento"`
	FechaCreacion        string  `json:"fecha_creacion"`
	FechaActualizacion   string  `json:"fecha_actualizacion"`
	UsuarioCreador       string  `json:"usuario_creador"`
	Estado               string  `json:"estado"`
	Observaciones        string  `json:"observaciones"`
}

// EmpresaFinanzasMovimientoBancarioFilter permite filtrar extractos bancarios por empresa.
type EmpresaFinanzasMovimientoBancarioFilter struct {
	Desde              string
	Hasta              string
	PeriodoContable    string
	EstadoConciliacion string
	IncludeInactive    bool
	Limit              int
}

// EmpresaConciliacionBancariaFilter permite consultar conciliacion bancaria por periodo.
type EmpresaConciliacionBancariaFilter struct {
	Desde           string
	Hasta           string
	PeriodoContable string
	IncludeInactive bool
	Limit           int
}

// EmpresaConciliacionBancariaPeriodo resume conciliacion bancaria por periodo.
type EmpresaConciliacionBancariaPeriodo struct {
	PeriodoContable          string  `json:"periodo_contable"`
	ExtractosTotal           int64   `json:"extractos_total"`
	ExtractosConciliados     int64   `json:"extractos_conciliados"`
	ExtractosPendientes      int64   `json:"extractos_pendientes"`
	ExtractosConDesviacion   int64   `json:"extractos_con_desviacion"`
	ExtractosMontoTotal      float64 `json:"extractos_monto_total"`
	ExtractosMontoConciliado float64 `json:"extractos_monto_conciliado"`
	MovimientosInternosTotal int64   `json:"movimientos_internos_total"`
	MovimientosInternosMonto float64 `json:"movimientos_internos_monto"`
	DesfaseRegistros         int64   `json:"desfase_registros"`
	DesfaseMonto             float64 `json:"desfase_monto"`
	UltimoExtracto           string  `json:"ultimo_extracto"`
	UltimaConciliacion       string  `json:"ultima_conciliacion"`
	EstadoConciliacion       string  `json:"estado_conciliacion"`
}

// EmpresaConciliacionBancariaResumen consolida conciliacion bancaria por periodos.
type EmpresaConciliacionBancariaResumen struct {
	EmpresaID              int64                                `json:"empresa_id"`
	Desde                  string                               `json:"desde"`
	Hasta                  string                               `json:"hasta"`
	PeriodoContable        string                               `json:"periodo_contable"`
	TotalPeriodos          int                                  `json:"total_periodos"`
	PeriodosConciliados    int                                  `json:"periodos_conciliados"`
	PeriodosConPendientes  int                                  `json:"periodos_con_pendientes"`
	PeriodosConDescuadre   int                                  `json:"periodos_con_descuadre"`
	PeriodosSinMovimientos int                                  `json:"periodos_sin_movimientos"`
	Filas                  []EmpresaConciliacionBancariaPeriodo `json:"filas"`
}

// EmpresaConciliacionBancariaAutoConfig parametriza la conciliacion bancaria automatica.
type EmpresaConciliacionBancariaAutoConfig struct {
	Desde           string
	Hasta           string
	PeriodoContable string
	ToleranciaDias  int
	ToleranciaMonto float64
	Limit           int
	Usuario         string
}

// EmpresaConciliacionBancariaAutoResultado resume la corrida de conciliacion bancaria automatica.
type EmpresaConciliacionBancariaAutoResultado struct {
	EmpresaID       int64    `json:"empresa_id"`
	PeriodoContable string   `json:"periodo_contable"`
	Revisados       int      `json:"revisados"`
	Conciliados     int      `json:"conciliados"`
	Pendientes      int      `json:"pendientes"`
	ConDesviacion   int      `json:"con_desviacion"`
	MontoConciliado float64  `json:"monto_conciliado"`
	Errores         []string `json:"errores"`
}

// EmpresaImportacionMovimientosBancariosResultado resume el proceso de importacion de extractos.
type EmpresaImportacionMovimientosBancariosResultado struct {
	EmpresaID    int64    `json:"empresa_id"`
	Recibidos    int      `json:"recibidos"`
	Creados      int      `json:"creados"`
	Actualizados int      `json:"actualizados"`
	IDs          []int64  `json:"ids"`
	Errores      []string `json:"errores"`
}

func normalizeConciliacionBancariaLimit(limit int) int {
	if limit <= 0 {
		return 200
	}
	if limit > 2000 {
		return 2000
	}
	return limit
}

func normalizeConciliacionBancariaToleranciaDias(v int) int {
	if v <= 0 {
		return 2
	}
	if v > 15 {
		return 15
	}
	return v
}

func normalizeConciliacionBancariaToleranciaMonto(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v == 0 {
		return 5
	}
	return roundReportesMoney(v)
}

func normalizeEstadoConciliacionBancaria(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "conciliado":
		return "conciliado"
	case "con_desviacion", "desviado":
		return "con_desviacion"
	default:
		return "pendiente"
	}
}

func normalizeEmpresaFinanzasMovimientoBancario(item EmpresaFinanzasMovimientoBancario) (EmpresaFinanzasMovimientoBancario, error) {
	if item.EmpresaID <= 0 {
		return item, fmt.Errorf("empresa_id es obligatorio")
	}

	item.TipoMovimiento = normalizeTipoMovimiento(item.TipoMovimiento)
	if item.TipoMovimiento == "" {
		return item, fmt.Errorf("tipo_movimiento debe ser ingreso o egreso")
	}

	item.Monto = maxFloat64(item.Monto, 0)
	item.Total = maxFloat64(item.Total, 0)
	if item.Total <= 0 {
		item.Total = item.Monto
	}
	if item.Total <= 0 {
		return item, fmt.Errorf("total debe ser mayor que cero")
	}

	item.FechaMovimiento = strings.TrimSpace(item.FechaMovimiento)
	if item.FechaMovimiento == "" {
		item.FechaMovimiento = time.Now().Format("2006-01-02 15:04:05")
	}
	item.FechaValor = strings.TrimSpace(item.FechaValor)
	if item.FechaValor == "" {
		item.FechaValor = item.FechaMovimiento
	}

	item.PeriodoContable = normalizePeriodoContable(item.PeriodoContable)
	if item.PeriodoContable == "" {
		item.PeriodoContable = normalizePeriodoContable(item.FechaMovimiento)
	}
	if item.PeriodoContable == "" {
		item.PeriodoContable = time.Now().Format("2006-01")
	}

	item.CuentaBancaria = strings.TrimSpace(item.CuentaBancaria)
	item.BancoNombre = strings.TrimSpace(item.BancoNombre)
	item.Descripcion = strings.TrimSpace(item.Descripcion)
	item.ReferenciaBancaria = strings.TrimSpace(item.ReferenciaBancaria)
	item.DocumentoCodigo = strings.TrimSpace(item.DocumentoCodigo)

	item.Moneda = strings.ToUpper(strings.TrimSpace(item.Moneda))
	if item.Moneda == "" {
		item.Moneda = "COP"
	}

	item.Origen = strings.ToLower(strings.TrimSpace(item.Origen))
	if item.Origen == "" {
		item.Origen = "manual"
	}

	item.UsuarioCreador = strings.TrimSpace(item.UsuarioCreador)
	if item.UsuarioCreador == "" {
		item.UsuarioCreador = "sistema"
	}
	item.Estado = normalizeEstadoMovimiento(item.Estado)
	item.Observaciones = strings.TrimSpace(item.Observaciones)

	item.MovimientoFinanzasID = maxInt64Bancario(item.MovimientoFinanzasID, 0)
	item.EstadoConciliacion = normalizeEstadoConciliacionBancaria(item.EstadoConciliacion)
	if item.MovimientoFinanzasID > 0 {
		item.EstadoConciliacion = "conciliado"
		if strings.TrimSpace(item.ConciliadoEn) == "" {
			item.ConciliadoEn = time.Now().Format("2006-01-02 15:04:05")
		}
		if strings.TrimSpace(item.ConciliadoPor) == "" {
			item.ConciliadoPor = item.UsuarioCreador
		}
	} else {
		item.ConciliadoEn = ""
		item.ConciliadoPor = ""
	}

	item.HashMovimiento = strings.ToLower(strings.TrimSpace(item.HashMovimiento))
	if item.HashMovimiento == "" {
		item.HashMovimiento = buildEmpresaFinanzasMovimientoBancarioHash(item)
	}
	if item.HashMovimiento == "" {
		return item, fmt.Errorf("no se pudo generar hash_movimiento")
	}

	item.Monto = roundReportesMoney(item.Monto)
	item.Total = roundReportesMoney(item.Total)
	return item, nil
}

func buildEmpresaFinanzasMovimientoBancarioHash(item EmpresaFinanzasMovimientoBancario) string {
	montoHash := item.Total
	if montoHash <= 0 {
		montoHash = item.Monto
	}
	if montoHash <= 0 {
		return ""
	}
	parts := []string{
		strconv.FormatInt(item.EmpresaID, 10),
		strings.ToLower(strings.TrimSpace(item.TipoMovimiento)),
		strings.TrimSpace(item.PeriodoContable),
		normalizeDateOnly(item.FechaMovimiento),
		fmt.Sprintf("%.2f", roundReportesMoney(montoHash)),
		strings.ToUpper(strings.TrimSpace(item.Moneda)),
		strings.ToLower(strings.TrimSpace(item.CuentaBancaria)),
		strings.ToLower(strings.TrimSpace(item.ReferenciaBancaria)),
		strings.ToLower(strings.TrimSpace(item.DocumentoCodigo)),
		strings.ToLower(strings.TrimSpace(item.Descripcion)),
		strings.ToLower(strings.TrimSpace(item.Origen)),
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(sum[:])
}

func maxInt64Bancario(v, min int64) int64 {
	if v < min {
		return min
	}
	return v
}

func upsertEmpresaFinanzasMovimientoBancario(dbConn *sql.DB, item EmpresaFinanzasMovimientoBancario) (int64, error) {
	item, err := normalizeEmpresaFinanzasMovimientoBancario(item)
	if err != nil {
		return 0, err
	}

	_, err = dbConn.Exec(`INSERT INTO empresa_finanzas_bancos_movimientos (
		empresa_id,
		periodo_contable,
		fecha_movimiento,
		fecha_valor,
		cuenta_bancaria,
		banco_nombre,
		tipo_movimiento,
		descripcion,
		referencia_bancaria,
		documento_codigo,
		moneda,
		monto,
		total,
		movimiento_finanzas_id,
		estado_conciliacion,
		conciliado_en,
		conciliado_por,
		origen,
		hash_movimiento,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	ON CONFLICT(empresa_id, hash_movimiento) DO UPDATE SET
		periodo_contable = excluded.periodo_contable,
		fecha_movimiento = excluded.fecha_movimiento,
		fecha_valor = excluded.fecha_valor,
		cuenta_bancaria = excluded.cuenta_bancaria,
		banco_nombre = excluded.banco_nombre,
		tipo_movimiento = excluded.tipo_movimiento,
		descripcion = excluded.descripcion,
		referencia_bancaria = excluded.referencia_bancaria,
		documento_codigo = excluded.documento_codigo,
		moneda = excluded.moneda,
		monto = excluded.monto,
		total = excluded.total,
		movimiento_finanzas_id = CASE
			WHEN empresa_finanzas_bancos_movimientos.movimiento_finanzas_id > 0 THEN empresa_finanzas_bancos_movimientos.movimiento_finanzas_id
			ELSE excluded.movimiento_finanzas_id
		END,
		estado_conciliacion = CASE
			WHEN empresa_finanzas_bancos_movimientos.movimiento_finanzas_id > 0 THEN empresa_finanzas_bancos_movimientos.estado_conciliacion
			ELSE excluded.estado_conciliacion
		END,
		conciliado_en = CASE
			WHEN empresa_finanzas_bancos_movimientos.movimiento_finanzas_id > 0 THEN empresa_finanzas_bancos_movimientos.conciliado_en
			ELSE excluded.conciliado_en
		END,
		conciliado_por = CASE
			WHEN empresa_finanzas_bancos_movimientos.movimiento_finanzas_id > 0 THEN empresa_finanzas_bancos_movimientos.conciliado_por
			ELSE excluded.conciliado_por
		END,
		origen = excluded.origen,
		usuario_creador = excluded.usuario_creador,
		estado = excluded.estado,
		observaciones = excluded.observaciones,
		fecha_actualizacion = CURRENT_TIMESTAMP`,
		item.EmpresaID,
		item.PeriodoContable,
		item.FechaMovimiento,
		item.FechaValor,
		item.CuentaBancaria,
		item.BancoNombre,
		item.TipoMovimiento,
		item.Descripcion,
		item.ReferenciaBancaria,
		item.DocumentoCodigo,
		item.Moneda,
		item.Monto,
		item.Total,
		item.MovimientoFinanzasID,
		item.EstadoConciliacion,
		item.ConciliadoEn,
		item.ConciliadoPor,
		item.Origen,
		item.HashMovimiento,
		item.UsuarioCreador,
		item.Estado,
		item.Observaciones,
	)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_finanzas_bancos_movimientos WHERE empresa_id = ? AND hash_movimiento = ? LIMIT 1`, item.EmpresaID, item.HashMovimiento).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertEmpresaFinanzasMovimientosBancarios importa o actualiza extractos bancarios por hash idempotente.
func UpsertEmpresaFinanzasMovimientosBancarios(dbConn *sql.DB, empresaID int64, movimientos []EmpresaFinanzasMovimientoBancario) (EmpresaImportacionMovimientosBancariosResultado, error) {
	result := EmpresaImportacionMovimientosBancariosResultado{
		EmpresaID: empresaID,
		IDs:       make([]int64, 0, len(movimientos)),
		Errores:   make([]string, 0),
	}
	if empresaID <= 0 {
		return result, fmt.Errorf("empresa_id es obligatorio")
	}
	for _, item := range movimientos {
		if item.EmpresaID <= 0 {
			item.EmpresaID = empresaID
		}
		normalized, err := normalizeEmpresaFinanzasMovimientoBancario(item)
		if err != nil {
			result.Errores = append(result.Errores, err.Error())
			continue
		}
		result.Recibidos++

		var existingID int64
		err = dbConn.QueryRow(`SELECT id FROM empresa_finanzas_bancos_movimientos WHERE empresa_id = ? AND hash_movimiento = ? LIMIT 1`, normalized.EmpresaID, normalized.HashMovimiento).Scan(&existingID)
		exists := err == nil
		if err != nil && err != sql.ErrNoRows {
			result.Errores = append(result.Errores, err.Error())
			continue
		}

		id, upsertErr := upsertEmpresaFinanzasMovimientoBancario(dbConn, normalized)
		if upsertErr != nil {
			result.Errores = append(result.Errores, upsertErr.Error())
			continue
		}
		result.IDs = append(result.IDs, id)
		if exists {
			result.Actualizados++
		} else {
			result.Creados++
		}
	}
	return result, nil
}

// ListEmpresaFinanzasMovimientosBancarios lista extractos bancarios por empresa.
func ListEmpresaFinanzasMovimientosBancarios(dbConn *sql.DB, empresaID int64, f EmpresaFinanzasMovimientoBancarioFilter) ([]EmpresaFinanzasMovimientoBancario, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(periodo_contable, ''),
		COALESCE(fecha_movimiento, ''),
		COALESCE(fecha_valor, ''),
		COALESCE(cuenta_bancaria, ''),
		COALESCE(banco_nombre, ''),
		COALESCE(tipo_movimiento, ''),
		COALESCE(descripcion, ''),
		COALESCE(referencia_bancaria, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(monto, 0),
		COALESCE(total, 0),
		COALESCE(movimiento_finanzas_id, 0),
		COALESCE(estado_conciliacion, 'pendiente'),
		COALESCE(conciliado_en, ''),
		COALESCE(conciliado_por, ''),
		COALESCE(origen, 'manual'),
		COALESCE(hash_movimiento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_finanzas_bancos_movimientos
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !f.IncludeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}
	if strings.TrimSpace(f.Desde) != "" {
		query += ` AND date(fecha_movimiento) >= date(?)`
		args = append(args, strings.TrimSpace(f.Desde))
	}
	if strings.TrimSpace(f.Hasta) != "" {
		query += ` AND date(fecha_movimiento) <= date(?)`
		args = append(args, strings.TrimSpace(f.Hasta))
	}
	if periodo := normalizePeriodoContable(f.PeriodoContable); periodo != "" {
		query += ` AND COALESCE(periodo_contable, '') = ?`
		args = append(args, periodo)
	}
	if estadoConciliacion := normalizeEstadoConciliacionBancaria(f.EstadoConciliacion); strings.TrimSpace(f.EstadoConciliacion) != "" {
		query += ` AND LOWER(COALESCE(estado_conciliacion, 'pendiente')) = ?`
		args = append(args, estadoConciliacion)
	}

	query += ` ORDER BY pcs_ts(fecha_movimiento) DESC, id DESC`
	limit := normalizeConciliacionBancariaLimit(f.Limit)
	query += ` LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaFinanzasMovimientoBancario, 0)
	for rows.Next() {
		var item EmpresaFinanzasMovimientoBancario
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.PeriodoContable,
			&item.FechaMovimiento,
			&item.FechaValor,
			&item.CuentaBancaria,
			&item.BancoNombre,
			&item.TipoMovimiento,
			&item.Descripcion,
			&item.ReferenciaBancaria,
			&item.DocumentoCodigo,
			&item.Moneda,
			&item.Monto,
			&item.Total,
			&item.MovimientoFinanzasID,
			&item.EstadoConciliacion,
			&item.ConciliadoEn,
			&item.ConciliadoPor,
			&item.Origen,
			&item.HashMovimiento,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func findEmpresaMovimientoFinanzasBancarioMatch(dbConn *sql.DB, empresaID int64, item EmpresaFinanzasMovimientoBancario, cfg EmpresaConciliacionBancariaAutoConfig) (int64, error) {
	montoObjetivo := item.Total
	if montoObjetivo <= 0 {
		montoObjetivo = item.Monto
	}
	if montoObjetivo <= 0 {
		return 0, nil
	}
	fechaObjetivo := normalizeDateOnly(item.FechaMovimiento)
	if fechaObjetivo == "" {
		fechaObjetivo = time.Now().Format("2006-01-02")
	}
	referencia := strings.TrimSpace(item.ReferenciaBancaria)
	if referencia == "" {
		referencia = strings.TrimSpace(item.DocumentoCodigo)
	}

	query := `SELECT m.id
	FROM empresa_finanzas_movimientos m
	LEFT JOIN empresa_finanzas_bancos_movimientos b
		ON b.empresa_id = m.empresa_id
		AND b.movimiento_finanzas_id = m.id
		AND LOWER(COALESCE(b.estado, 'activo')) = 'activo'
	WHERE m.empresa_id = ?
		AND LOWER(COALESCE(m.estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(m.tipo_movimiento, '')) = ?
		AND b.id IS NULL
		AND ABS(COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0) - ?) <= ?`
	args := []interface{}{empresaID, item.TipoMovimiento, montoObjetivo, cfg.ToleranciaMonto}
	fechaDistExpr := `ABS(pcs_julian_day(pcs_date(COALESCE(m.fecha_movimiento, ''))) - pcs_julian_day(pcs_date(?)))`
	if isPostgresDialect() {
		fechaDistExpr = `ABS((COALESCE(NULLIF(m.fecha_movimiento, ''), CURRENT_DATE::text)::date) - (?::date))`
	}

	if cfg.PeriodoContable != "" {
		query += ` AND COALESCE(m.periodo_contable, '') = ?`
		args = append(args, cfg.PeriodoContable)
	}
	if fechaObjetivo != "" {
		query += ` AND ` + fechaDistExpr + ` <= ?`
		args = append(args, fechaObjetivo, cfg.ToleranciaDias)
	}

	query += ` ORDER BY
		CASE WHEN ? <> '' AND (
			LOWER(COALESCE(m.referencia_externa, '')) = LOWER(?) OR
			LOWER(COALESCE(m.numero_comprobante, '')) = LOWER(?) OR
			LOWER(COALESCE(m.codigo, '')) = LOWER(?)
		) THEN 0 ELSE 1 END,
		` + fechaDistExpr + ` ASC,
		ABS(COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0) - ?) ASC,
		m.id DESC
	LIMIT 1`
	args = append(args, referencia, referencia, referencia, referencia, fechaObjetivo, montoObjetivo)

	var movimientoID int64
	err := dbConn.QueryRow(query, args...).Scan(&movimientoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	return movimientoID, nil
}

func setEmpresaMovimientoBancarioConciliado(dbConn *sql.DB, itemID, movimientoID int64, usuario string) error {
	_, err := dbConn.Exec(`UPDATE empresa_finanzas_bancos_movimientos
	SET movimiento_finanzas_id = ?,
		estado_conciliacion = 'conciliado',
		conciliado_en = CURRENT_TIMESTAMP,
		conciliado_por = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE id = ?`, movimientoID, usuario, itemID)
	return err
}

func setEmpresaMovimientoBancarioConDesviacion(dbConn *sql.DB, itemID int64) error {
	_, err := dbConn.Exec(`UPDATE empresa_finanzas_bancos_movimientos
	SET movimiento_finanzas_id = 0,
		estado_conciliacion = 'con_desviacion',
		conciliado_en = '',
		conciliado_por = '',
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE id = ?`, itemID)
	return err
}

// ConciliarEmpresaMovimientosBancariosAutomatico ejecuta conciliacion bancaria automatica contra movimientos financieros internos.
func ConciliarEmpresaMovimientosBancariosAutomatico(dbConn *sql.DB, empresaID int64, cfg EmpresaConciliacionBancariaAutoConfig) (EmpresaConciliacionBancariaAutoResultado, error) {
	result := EmpresaConciliacionBancariaAutoResultado{
		EmpresaID: empresaID,
		Errores:   make([]string, 0),
	}
	if empresaID <= 0 {
		return result, fmt.Errorf("empresa_id es obligatorio")
	}

	cfg.PeriodoContable = normalizePeriodoContable(cfg.PeriodoContable)
	cfg.ToleranciaDias = normalizeConciliacionBancariaToleranciaDias(cfg.ToleranciaDias)
	cfg.ToleranciaMonto = normalizeConciliacionBancariaToleranciaMonto(cfg.ToleranciaMonto)
	cfg.Limit = normalizeConciliacionBancariaLimit(cfg.Limit)
	cfg.Usuario = strings.TrimSpace(cfg.Usuario)
	if cfg.Usuario == "" {
		cfg.Usuario = "sistema"
	}
	result.PeriodoContable = cfg.PeriodoContable

	query := `SELECT
		id,
		empresa_id,
		COALESCE(periodo_contable, ''),
		COALESCE(fecha_movimiento, ''),
		COALESCE(fecha_valor, ''),
		COALESCE(cuenta_bancaria, ''),
		COALESCE(banco_nombre, ''),
		COALESCE(tipo_movimiento, ''),
		COALESCE(descripcion, ''),
		COALESCE(referencia_bancaria, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(monto, 0),
		COALESCE(total, 0),
		COALESCE(movimiento_finanzas_id, 0),
		COALESCE(estado_conciliacion, 'pendiente'),
		COALESCE(conciliado_en, ''),
		COALESCE(conciliado_por, ''),
		COALESCE(origen, 'manual'),
		COALESCE(hash_movimiento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_finanzas_bancos_movimientos
	WHERE empresa_id = ?
		AND LOWER(COALESCE(estado, 'activo')) = 'activo'
		AND LOWER(COALESCE(estado_conciliacion, 'pendiente')) <> 'conciliado'`
	args := []interface{}{empresaID}

	if strings.TrimSpace(cfg.Desde) != "" {
		query += ` AND date(fecha_movimiento) >= date(?)`
		args = append(args, strings.TrimSpace(cfg.Desde))
	}
	if strings.TrimSpace(cfg.Hasta) != "" {
		query += ` AND date(fecha_movimiento) <= date(?)`
		args = append(args, strings.TrimSpace(cfg.Hasta))
	}
	if cfg.PeriodoContable != "" {
		query += ` AND COALESCE(periodo_contable, '') = ?`
		args = append(args, cfg.PeriodoContable)
	}

	query += ` ORDER BY pcs_ts(fecha_movimiento) ASC, id ASC LIMIT ?`
	args = append(args, cfg.Limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return result, err
	}
	defer rows.Close()

	items := make([]EmpresaFinanzasMovimientoBancario, 0)
	for rows.Next() {
		var item EmpresaFinanzasMovimientoBancario
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.PeriodoContable,
			&item.FechaMovimiento,
			&item.FechaValor,
			&item.CuentaBancaria,
			&item.BancoNombre,
			&item.TipoMovimiento,
			&item.Descripcion,
			&item.ReferenciaBancaria,
			&item.DocumentoCodigo,
			&item.Moneda,
			&item.Monto,
			&item.Total,
			&item.MovimientoFinanzasID,
			&item.EstadoConciliacion,
			&item.ConciliadoEn,
			&item.ConciliadoPor,
			&item.Origen,
			&item.HashMovimiento,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return result, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return result, err
	}

	result.Revisados = len(items)
	for _, item := range items {
		movimientoID, matchErr := findEmpresaMovimientoFinanzasBancarioMatch(dbConn, empresaID, item, cfg)
		if matchErr != nil {
			if len(result.Errores) < 20 {
				result.Errores = append(result.Errores, fmt.Sprintf("extracto_id=%d: %s", item.ID, matchErr.Error()))
			}
			continue
		}

		if movimientoID > 0 {
			if err := setEmpresaMovimientoBancarioConciliado(dbConn, item.ID, movimientoID, cfg.Usuario); err != nil {
				if len(result.Errores) < 20 {
					result.Errores = append(result.Errores, fmt.Sprintf("extracto_id=%d: %s", item.ID, err.Error()))
				}
				continue
			}
			result.Conciliados++
			monto := item.Total
			if monto <= 0 {
				monto = item.Monto
			}
			result.MontoConciliado += monto
			continue
		}

		if err := setEmpresaMovimientoBancarioConDesviacion(dbConn, item.ID); err != nil {
			if len(result.Errores) < 20 {
				result.Errores = append(result.Errores, fmt.Sprintf("extracto_id=%d: %s", item.ID, err.Error()))
			}
			continue
		}
		result.ConDesviacion++
	}

	result.Pendientes = result.Revisados - result.Conciliados
	if result.Pendientes < 0 {
		result.Pendientes = 0
	}
	result.MontoConciliado = roundReportesMoney(result.MontoConciliado)
	return result, nil
}

func getOrCreateEmpresaConciliacionBancariaPeriodo(periodos map[string]*EmpresaConciliacionBancariaPeriodo, periodo string) *EmpresaConciliacionBancariaPeriodo {
	periodo = strings.TrimSpace(periodo)
	if norm := normalizePeriodoContable(periodo); norm != "" {
		periodo = norm
	}
	if periodo == "" {
		periodo = "sin_periodo"
	}
	if item, ok := periodos[periodo]; ok {
		return item
	}
	item := &EmpresaConciliacionBancariaPeriodo{PeriodoContable: periodo, EstadoConciliacion: "sin_movimientos"}
	periodos[periodo] = item
	return item
}

// GetEmpresaConciliacionBancariaPorPeriodo construye el tablero de desviaciones financieras y conciliacion bancaria.
func GetEmpresaConciliacionBancariaPorPeriodo(dbConn *sql.DB, empresaID int64, f EmpresaConciliacionBancariaFilter) (EmpresaConciliacionBancariaResumen, error) {
	resumen := EmpresaConciliacionBancariaResumen{
		EmpresaID:       empresaID,
		Desde:           strings.TrimSpace(f.Desde),
		Hasta:           strings.TrimSpace(f.Hasta),
		PeriodoContable: normalizePeriodoContable(f.PeriodoContable),
		Filas:           make([]EmpresaConciliacionBancariaPeriodo, 0),
	}
	if empresaID <= 0 {
		return resumen, fmt.Errorf("empresa_id es obligatorio")
	}

	periodos := make(map[string]*EmpresaConciliacionBancariaPeriodo)

	periodoExtractoExpr := `COALESCE(NULLIF(periodo_contable, ''), substr(COALESCE(fecha_movimiento, ''), 1, 7), 'sin_periodo')`
	queryExtractos := `SELECT
		` + periodoExtractoExpr + `,
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_conciliacion, 'pendiente')) = 'conciliado' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_conciliacion, 'pendiente')) = 'con_desviacion' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_conciliacion, 'pendiente')) <> 'conciliado' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(COALESCE(NULLIF(total, 0), monto, 0)), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_conciliacion, 'pendiente')) = 'conciliado' THEN COALESCE(NULLIF(total, 0), monto, 0) ELSE 0 END), 0),
		COALESCE(MAX(fecha_movimiento), ''),
		COALESCE(MAX(conciliado_en), '')
	FROM empresa_finanzas_bancos_movimientos
	WHERE empresa_id = ?`
	argsExtractos := []interface{}{empresaID}
	if !f.IncludeInactive {
		queryExtractos += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if resumen.Desde != "" {
		queryExtractos += ` AND date(fecha_movimiento) >= date(?)`
		argsExtractos = append(argsExtractos, resumen.Desde)
	}
	if resumen.Hasta != "" {
		queryExtractos += ` AND date(fecha_movimiento) <= date(?)`
		argsExtractos = append(argsExtractos, resumen.Hasta)
	}
	if resumen.PeriodoContable != "" {
		queryExtractos += ` AND ` + periodoExtractoExpr + ` = ?`
		argsExtractos = append(argsExtractos, resumen.PeriodoContable)
	}
	queryExtractos += ` GROUP BY ` + periodoExtractoExpr

	rowsExtractos, err := dbConn.Query(queryExtractos, argsExtractos...)
	if err != nil {
		return resumen, err
	}
	defer rowsExtractos.Close()

	for rowsExtractos.Next() {
		var periodo, ultimoExtracto, ultimaConciliacion string
		var total, conciliados, conDesviacion, pendientes int64
		var montoTotal, montoConciliado float64
		if err := rowsExtractos.Scan(
			&periodo,
			&total,
			&conciliados,
			&conDesviacion,
			&pendientes,
			&montoTotal,
			&montoConciliado,
			&ultimoExtracto,
			&ultimaConciliacion,
		); err != nil {
			return resumen, err
		}
		item := getOrCreateEmpresaConciliacionBancariaPeriodo(periodos, periodo)
		item.ExtractosTotal = total
		item.ExtractosConciliados = conciliados
		item.ExtractosConDesviacion = conDesviacion
		item.ExtractosPendientes = pendientes
		item.ExtractosMontoTotal = roundReportesMoney(montoTotal)
		item.ExtractosMontoConciliado = roundReportesMoney(montoConciliado)
		item.UltimoExtracto = strings.TrimSpace(ultimoExtracto)
		item.UltimaConciliacion = strings.TrimSpace(ultimaConciliacion)
	}
	if err := rowsExtractos.Err(); err != nil {
		return resumen, err
	}

	periodoInternoExpr := `COALESCE(NULLIF(m.periodo_contable, ''), substr(COALESCE(m.fecha_movimiento, ''), 1, 7), 'sin_periodo')`
	queryInternos := `SELECT
		` + periodoInternoExpr + `,
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(COALESCE(NULLIF(m.total_neto, 0), NULLIF(m.total, 0), m.monto, 0)), 0)
	FROM empresa_finanzas_movimientos m
	WHERE m.empresa_id = ?
		AND COALESCE(m.estado, 'activo') = 'activo'
		AND LOWER(COALESCE(m.tipo_movimiento, '')) IN ('ingreso', 'egreso')
		AND (
			LOWER(COALESCE(m.metodo_pago, '')) IN ('transferencia', 'transferencia_bancaria', 'tarjeta_credito', 'tarjeta_debito', 'pasarela', 'nequi', 'daviplata')
			OR TRIM(COALESCE(m.referencia_externa, '')) <> ''
		)`
	argsInternos := []interface{}{empresaID}
	if resumen.Desde != "" {
		queryInternos += ` AND date(m.fecha_movimiento) >= date(?)`
		argsInternos = append(argsInternos, resumen.Desde)
	}
	if resumen.Hasta != "" {
		queryInternos += ` AND date(m.fecha_movimiento) <= date(?)`
		argsInternos = append(argsInternos, resumen.Hasta)
	}
	if resumen.PeriodoContable != "" {
		queryInternos += ` AND ` + periodoInternoExpr + ` = ?`
		argsInternos = append(argsInternos, resumen.PeriodoContable)
	}
	queryInternos += ` GROUP BY ` + periodoInternoExpr

	rowsInternos, err := dbConn.Query(queryInternos, argsInternos...)
	if err != nil {
		return resumen, err
	}
	defer rowsInternos.Close()

	for rowsInternos.Next() {
		var periodo string
		var total int64
		var monto float64
		if err := rowsInternos.Scan(&periodo, &total, &monto); err != nil {
			return resumen, err
		}
		item := getOrCreateEmpresaConciliacionBancariaPeriodo(periodos, periodo)
		item.MovimientosInternosTotal = total
		item.MovimientosInternosMonto = roundReportesMoney(monto)
	}
	if err := rowsInternos.Err(); err != nil {
		return resumen, err
	}

	filas := make([]EmpresaConciliacionBancariaPeriodo, 0, len(periodos))
	for _, item := range periodos {
		item.DesfaseRegistros = item.MovimientosInternosTotal - item.ExtractosConciliados
		item.DesfaseMonto = roundReportesMoney(item.MovimientosInternosMonto - item.ExtractosMontoConciliado)

		switch {
		case item.ExtractosPendientes > 0:
			item.EstadoConciliacion = "con_pendientes"
		case item.ExtractosConDesviacion > 0 || item.DesfaseRegistros != 0 || conciliacionAbsFloat64(item.DesfaseMonto) > 0.009:
			item.EstadoConciliacion = "con_descuadre"
		case item.ExtractosTotal == 0 && item.MovimientosInternosTotal == 0:
			item.EstadoConciliacion = "sin_movimientos"
		default:
			item.EstadoConciliacion = "conciliado"
		}
		filas = append(filas, *item)
	}

	sort.Slice(filas, func(i, j int) bool {
		keyI := conciliacionPeriodoSortKey(filas[i].PeriodoContable)
		keyJ := conciliacionPeriodoSortKey(filas[j].PeriodoContable)
		if keyI == keyJ {
			return filas[i].UltimoExtracto > filas[j].UltimoExtracto
		}
		return keyI > keyJ
	})

	limit := normalizeConciliacionLimit(f.Limit)
	if len(filas) > limit {
		filas = filas[:limit]
	}

	resumen.TotalPeriodos = len(filas)
	for _, row := range filas {
		switch row.EstadoConciliacion {
		case "conciliado":
			resumen.PeriodosConciliados++
		case "con_pendientes":
			resumen.PeriodosConPendientes++
		case "con_descuadre":
			resumen.PeriodosConDescuadre++
		case "sin_movimientos":
			resumen.PeriodosSinMovimientos++
		}
	}
	resumen.Filas = filas

	return resumen, nil
}
