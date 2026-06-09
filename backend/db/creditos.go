package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

// EmpresaCredito representa una linea de credito empresarial por cliente.
type EmpresaCredito struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Codigo                string  `json:"codigo"`
	ClienteID             int64   `json:"cliente_id"`
	ClienteNombre         string  `json:"cliente_nombre"`
	TipoCredito           string  `json:"tipo_credito"`
	MontoAprobado         float64 `json:"monto_aprobado"`
	CupoCredito           float64 `json:"cupo_credito"`
	SaldoActual           float64 `json:"saldo_actual"`
	SaldoDisponible       float64 `json:"saldo_disponible"`
	TasaInteres           float64 `json:"tasa_interes"`
	TasaMora              float64 `json:"tasa_mora"`
	PeriodicidadCuota     string  `json:"periodicidad_cuota"`
	ValorCuotaPactada     float64 `json:"valor_cuota_pactada"`
	OmitirDomingos        bool    `json:"omitir_domingos"`
	PlazoDias             int     `json:"plazo_dias"`
	PlazoCuotas           int     `json:"plazo_cuotas"`
	FechaInicio           string  `json:"fecha_inicio"`
	FechaVencimiento      string  `json:"fecha_vencimiento"`
	FechaUltimoPago       string  `json:"fecha_ultimo_pago,omitempty"`
	DiasMora              int     `json:"dias_mora"`
	CuotasPendientes      int     `json:"cuotas_pendientes,omitempty"`
	CuotasVencidas        int     `json:"cuotas_vencidas,omitempty"`
	DiasCuotasVencidas    int     `json:"dias_cuotas_vencidas,omitempty"`
	FechaProximaCuota     string  `json:"fecha_proxima_cuota,omitempty"`
	ClasificacionCartera  string  `json:"clasificacion_cartera"`
	BloqueoAutomaticoMora bool    `json:"bloqueo_automatico_mora"`
	VentaOrigenID         int64   `json:"venta_origen_id,omitempty"`
	DocumentoOrigen       string  `json:"documento_origen,omitempty"`
	EstadoCredito         string  `json:"estado_credito"`
	FechaCreacion         string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion    string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador        string  `json:"usuario_creador,omitempty"`
	Estado                string  `json:"estado,omitempty"`
	Observaciones         string  `json:"observaciones,omitempty"`
}

// EmpresaCreditoCuota representa una cuota de amortizacion del credito.
type EmpresaCreditoCuota struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	CreditoID          int64   `json:"credito_id"`
	NumeroCuota        int     `json:"numero_cuota"`
	FechaVencimiento   string  `json:"fecha_vencimiento"`
	ValorCuota         float64 `json:"valor_cuota"`
	CapitalCuota       float64 `json:"capital_cuota"`
	InteresCuota       float64 `json:"interes_cuota"`
	InteresMora        float64 `json:"interes_mora"`
	ValorPagado        float64 `json:"valor_pagado"`
	SaldoCuota         float64 `json:"saldo_cuota"`
	EstadoCuota        string  `json:"estado_cuota"`
	FechaUltimoPago    string  `json:"fecha_ultimo_pago,omitempty"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaCreditoMovimiento registra cobros, abonos y ajustes del credito.
type EmpresaCreditoMovimiento struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	CreditoID          int64   `json:"credito_id"`
	CuotaID            int64   `json:"cuota_id,omitempty"`
	TipoMovimiento     string  `json:"tipo_movimiento"`
	Monto              float64 `json:"monto"`
	CapitalAplicado    float64 `json:"capital_aplicado"`
	InteresAplicado    float64 `json:"interes_aplicado"`
	MoraAplicada       float64 `json:"mora_aplicada"`
	MetodoPago         string  `json:"metodo_pago,omitempty"`
	ReferenciaPago     string  `json:"referencia_pago,omitempty"`
	Comprobante        string  `json:"comprobante,omitempty"`
	AplicadoAutomatico bool    `json:"aplicado_automatico"`
	FechaMovimiento    string  `json:"fecha_movimiento"`
	FechaCreacion      string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string  `json:"usuario_creador,omitempty"`
	Estado             string  `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// EmpresaCreditoFilter permite filtrar consultas del modulo de creditos.
type EmpresaCreditoFilter struct {
	ClienteID       int64
	EstadoCredito   string
	Clasificacion   string
	SoloVencidos    bool
	IncludeInactive bool
	Desde           string
	Hasta           string
	Q               string
	Limit           int
	Offset          int
}

// EmpresaCreditoAbonoInput define entrada para aplicar un abono al credito.
type EmpresaCreditoAbonoInput struct {
	EmpresaID       int64
	CreditoID       int64
	Monto           float64
	MetodoPago      string
	ReferenciaPago  string
	Comprobante     string
	UsuarioCreador  string
	Observaciones   string
	FechaMovimiento string
}

// EmpresaCreditoCarteraResumen agrega informacion ejecutiva de cartera.
type EmpresaCreditoCarteraResumen struct {
	TotalCreditos    int64   `json:"total_creditos"`
	CreditosActivos  int64   `json:"creditos_activos"`
	CreditosVencidos int64   `json:"creditos_vencidos"`
	CreditosMora     int64   `json:"creditos_mora"`
	CreditosCerrados int64   `json:"creditos_cerrados"`
	MontoAprobado    float64 `json:"monto_aprobado"`
	SaldoTotal       float64 `json:"saldo_total"`
	CupoDisponible   float64 `json:"cupo_disponible"`
}

// EmpresaCreditosMoraDashboard concentra alertas y ranking de morosidad.
type EmpresaCreditosMoraDashboard struct {
	DiasProximos          int              `json:"dias_proximos"`
	Top                   int              `json:"top"`
	TotalProximosVencer   int64            `json:"total_proximos_vencer"`
	TotalVencidos         int64            `json:"total_vencidos"`
	TotalRankingMorosidad int64            `json:"total_ranking_morosidad"`
	MontoProximosVencer   float64          `json:"monto_proximos_vencer"`
	MontoVencidos         float64          `json:"monto_vencidos"`
	MontoRankingMorosidad float64          `json:"monto_ranking_morosidad"`
	ProximosVencer        []EmpresaCredito `json:"proximos_vencer"`
	Vencidos              []EmpresaCredito `json:"vencidos"`
	RankingMorosidad      []EmpresaCredito `json:"ranking_morosidad"`
	GeneradoEn            string           `json:"generado_en"`
}

// EmpresaCreditoWorkflow representa solicitudes de aprobacion para reversos/refinanciaciones.
type EmpresaCreditoWorkflow struct {
	ID                        int64  `json:"id"`
	EmpresaID                 int64  `json:"empresa_id"`
	CreditoID                 int64  `json:"credito_id"`
	WorkflowCodigo            string `json:"workflow_codigo"`
	TipoSolicitud             string `json:"tipo_solicitud"`
	EstadoSolicitud           string `json:"estado_solicitud"`
	MovimientoOrigenID        int64  `json:"movimiento_origen_id,omitempty"`
	MovimientoResultadoID     int64  `json:"movimiento_resultado_id,omitempty"`
	NivelAprobacionActual     int    `json:"nivel_aprobacion_actual"`
	NivelAprobacionRequerido  int    `json:"nivel_aprobacion_requerido"`
	AprobadoPor               string `json:"aprobado_por,omitempty"`
	CodigoAprobacion          string `json:"codigo_aprobacion,omitempty"`
	FechaAprobacionFinal      string `json:"fecha_aprobacion_final,omitempty"`
	EjecutadoPor              string `json:"ejecutado_por,omitempty"`
	FechaEjecucion            string `json:"fecha_ejecucion,omitempty"`
	MotivoSolicitud           string `json:"motivo_solicitud,omitempty"`
	MotivoRechazo             string `json:"motivo_rechazo,omitempty"`
	PayloadJSON               string `json:"payload_json,omitempty"`
	ResultadoJSON             string `json:"resultado_json,omitempty"`
	HistorialAprobacionesJSON string `json:"historial_aprobaciones_json,omitempty"`
	FechaCreacion             string `json:"fecha_creacion,omitempty"`
	FechaActualizacion        string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador            string `json:"usuario_creador,omitempty"`
	Estado                    string `json:"estado,omitempty"`
	Observaciones             string `json:"observaciones,omitempty"`
}

// EmpresaCreditoWorkflowFilter permite filtrar workflows de creditos.
type EmpresaCreditoWorkflowFilter struct {
	CreditoID       int64
	TipoSolicitud   string
	EstadoSolicitud string
	IncludeInactive bool
	Limit           int
	Offset          int
}

// EmpresaCreditoWorkflowSolicitudInput define datos para solicitar workflow de creditos.
type EmpresaCreditoWorkflowSolicitudInput struct {
	EmpresaID                int64
	CreditoID                int64
	TipoSolicitud            string
	MovimientoOrigenID       int64
	NivelAprobacionRequerido int
	MotivoSolicitud          string
	PayloadJSON              string
	UsuarioCreador           string
	Observaciones            string
}

// EmpresaCreditoWorkflowAprobacionInput define datos de aprobacion/rechazo del workflow.
type EmpresaCreditoWorkflowAprobacionInput struct {
	EmpresaID        int64
	WorkflowID       int64
	AprobadoPor      string
	CodigoAprobacion string
	MotivoAprobacion string
	MotivoRechazo    string
	EjecutadoPor     string
	UsuarioCreador   string
}

// EmpresaCreditoClienteLimite define reglas de cupo por cliente para controlar riesgo.
type EmpresaCreditoClienteLimite struct {
	ID                       int64   `json:"id"`
	EmpresaID                int64   `json:"empresa_id"`
	ClienteID                int64   `json:"cliente_id"`
	LimiteSaldoTotal         float64 `json:"limite_saldo_total"`
	MaxCreditosActivos       int     `json:"max_creditos_activos"`
	RequiereAprobacionExceso bool    `json:"requiere_aprobacion_exceso"`
	FechaCreacion            string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion       string  `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador           string  `json:"usuario_creador,omitempty"`
	Estado                   string  `json:"estado,omitempty"`
	Observaciones            string  `json:"observaciones,omitempty"`
}

// EmpresaCreditoClienteLimiteFilter permite listar limites por empresa/cliente.
type EmpresaCreditoClienteLimiteFilter struct {
	ClienteID       int64
	IncludeInactive bool
	Limit           int
	Offset          int
}

func creditoNormalizeTipo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "rotativo", "revolvente":
		return "rotativo"
	case "fijo":
		return "fijo"
	case "cuotas", "cuota":
		return "cuotas"
	default:
		return "cuotas"
	}
}

func creditoNormalizePeriodicidad(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	v = strings.ReplaceAll(v, "í", "i")
	switch v {
	case "diaria", "diario", "dia", "dias", "daily":
		return "diaria"
	case "semanal", "semana", "weekly":
		return "semanal"
	case "quincenal", "quincena":
		return "quincenal"
	case "mensual", "mes", "month", "monthly", "":
		return "mensual"
	default:
		return "mensual"
	}
}

func creditoMaxCuotas(periodicidad string) int {
	switch creditoNormalizePeriodicidad(periodicidad) {
	case "diaria":
		return 2400
	case "semanal":
		return 520
	default:
		return 600
	}
}

func creditoAddPeriodo(base time.Time, periodicidad string, step int) time.Time {
	if step <= 0 {
		step = 1
	}
	switch creditoNormalizePeriodicidad(periodicidad) {
	case "diaria":
		return base.AddDate(0, 0, step)
	case "semanal":
		return base.AddDate(0, 0, step*7)
	case "quincenal":
		return base.AddDate(0, 0, step*15)
	default:
		return base.AddDate(0, step, 0)
	}
}

func creditoNextFechaCuota(fechaInicio time.Time, periodicidad string, numeroCuota int, omitirDomingos bool) time.Time {
	periodicidad = creditoNormalizePeriodicidad(periodicidad)
	if numeroCuota <= 0 {
		numeroCuota = 1
	}
	if periodicidad != "diaria" || !omitirDomingos {
		return creditoAddPeriodo(fechaInicio, periodicidad, numeroCuota)
	}
	fecha := fechaInicio
	generadas := 0
	for generadas < numeroCuota {
		fecha = fecha.AddDate(0, 0, 1)
		if fecha.Weekday() == time.Sunday {
			continue
		}
		generadas++
	}
	return fecha
}

func creditoNormalizeEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "activo", "suspendido", "cerrado", "castigado":
		return v
	default:
		return "activo"
	}
}

func creditoNormalizeClasificacion(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "al_dia", "vencido", "castigado":
		return v
	default:
		return "al_dia"
	}
}

func creditoNormalizeRowEstado(raw string) string {
	if strings.EqualFold(strings.TrimSpace(raw), "inactivo") {
		return "inactivo"
	}
	return "activo"
}

func creditoNormalizeCuotaEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "pendiente", "parcial", "pagada", "vencida", "anulada":
		return v
	default:
		return "pendiente"
	}
}

func creditoNormalizeMovimiento(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "abono", "cargo_interes", "interes", "mora", "reverso", "ajuste", "refinanciacion":
		return v
	default:
		return "abono"
	}
}

func creditoWorkflowNormalizeTipo(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "reverso_abono", "reverso", "anulacion_abono", "anular_abono":
		return "reverso_abono"
	case "refinanciacion":
		return "refinanciacion"
	default:
		return ""
	}
}

func creditoWorkflowNormalizeEstado(raw string) string {
	v := strings.ToLower(strings.TrimSpace(raw))
	switch v {
	case "pendiente_aprobacion", "aprobada", "rechazada", "ejecutada", "cancelada":
		return v
	default:
		return "pendiente_aprobacion"
	}
}

func creditoWorkflowNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func creditoWorkflowDefaultCodigo(empresaID, id int64, tipo string) string {
	prefix := "WF"
	switch creditoWorkflowNormalizeTipo(tipo) {
	case "reverso_abono":
		prefix = "REV"
	case "refinanciacion":
		prefix = "REF"
	}
	return fmt.Sprintf("%s-%d-%d", prefix, empresaID, id)
}

func creditoParseJSONMap(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	if strings.TrimSpace(raw) == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func creditoMarshalJSON(v interface{}, fallback string) string {
	raw, err := json.Marshal(v)
	if err != nil {
		return fallback
	}
	return string(raw)
}

func creditoAppendObservacion(actual, extra string) string {
	actual = strings.TrimSpace(actual)
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return actual
	}
	if actual == "" {
		return extra
	}
	return actual + " | " + extra
}

func creditoNormalizeLimitOffset(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func creditoRound(v float64) float64 {
	return math.Round(v*100) / 100
}

func creditoMax(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func creditoMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func creditoParseDate(raw string) (time.Time, bool) {
	v := strings.TrimSpace(raw)
	if v == "" {
		return time.Time{}, false
	}
	formats := []string{"2006-01-02", "2006-01-02 15:04:05", time.RFC3339}
	for _, f := range formats {
		if dt, err := time.ParseInLocation(f, v, time.Local); err == nil {
			return dt, true
		}
	}
	return time.Time{}, false
}

func creditoFormatDate(dt time.Time) string {
	if dt.IsZero() {
		return ""
	}
	return dt.In(time.Local).Format("2006-01-02")
}

func creditoDaysMora(fechaVencimiento string, saldo float64) int {
	if saldo <= 0 {
		return 0
	}
	venc, ok := creditoParseDate(fechaVencimiento)
	if !ok {
		return 0
	}
	today := time.Now().In(time.Local)
	if today.Before(venc) {
		return 0
	}
	delta := int(today.Sub(venc).Hours() / 24)
	if delta < 0 {
		return 0
	}
	return delta
}

func creditoResolveClasificacion(estadoCredito, fechaVencimiento string, saldo float64) string {
	estado := creditoNormalizeEstado(estadoCredito)
	if estado == "castigado" {
		return "castigado"
	}
	if saldo <= 0 {
		return "al_dia"
	}
	if creditoDaysMora(fechaVencimiento, saldo) > 0 {
		return "vencido"
	}
	return "al_dia"
}

func creditoLikePattern(raw string) string {
	value := strings.TrimSpace(raw)
	value = strings.ReplaceAll(value, "!", "!!")
	value = strings.ReplaceAll(value, "%", "!%")
	value = strings.ReplaceAll(value, "_", "!_")
	return "%" + value + "%"
}

func creditoDefaultCodigo(empresaID int64, id int64) string {
	return "CRE-" + strconv.FormatInt(empresaID, 10) + "-" + strconv.FormatInt(id, 10)
}

func creditoClienteLimiteNormalize(payload *EmpresaCreditoClienteLimite) {
	if payload == nil {
		return
	}
	payload.LimiteSaldoTotal = creditoRound(creditoMax(payload.LimiteSaldoTotal, 0))
	if payload.MaxCreditosActivos < 0 {
		payload.MaxCreditosActivos = 0
	}
	payload.Estado = creditoNormalizeRowEstado(payload.Estado)
	if payload.Estado == "" {
		payload.Estado = "activo"
	}
	payload.UsuarioCreador = strings.TrimSpace(payload.UsuarioCreador)
	payload.Observaciones = strings.TrimSpace(payload.Observaciones)
}

func scanEmpresaCreditoClienteLimite(scanner interface {
	Scan(dest ...interface{}) error
}) (*EmpresaCreditoClienteLimite, error) {
	var row EmpresaCreditoClienteLimite
	var requiereAprobacion int
	if err := scanner.Scan(
		&row.ID,
		&row.EmpresaID,
		&row.ClienteID,
		&row.LimiteSaldoTotal,
		&row.MaxCreditosActivos,
		&requiereAprobacion,
		&row.FechaCreacion,
		&row.FechaActualizacion,
		&row.UsuarioCreador,
		&row.Estado,
		&row.Observaciones,
	); err != nil {
		return nil, err
	}
	row.RequiereAprobacionExceso = requiereAprobacion == 1
	creditoClienteLimiteNormalize(&row)
	return &row, nil
}

func creditoValidateClienteLimites(dbConn *sql.DB, empresaID, clienteID, excludeCreditoID int64, saldoProyectado float64) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return errors.New("empresa_id invalido")
	}
	if clienteID <= 0 {
		return nil
	}

	limite, err := GetEmpresaCreditoClienteLimite(dbConn, empresaID, clienteID, false)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	if limite == nil {
		return nil
	}
	if limite.LimiteSaldoTotal <= 0 && limite.MaxCreditosActivos <= 0 {
		return nil
	}

	query := `SELECT
		COUNT(1),
		COALESCE(SUM(COALESCE(saldo_actual, 0)), 0)
	FROM empresa_creditos
	WHERE empresa_id = ?
	  AND cliente_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND LOWER(COALESCE(estado_credito, 'activo')) IN ('activo', 'suspendido', 'castigado')
	  AND COALESCE(saldo_actual, 0) > 0`
	args := []interface{}{empresaID, clienteID}
	if excludeCreditoID > 0 {
		query += ` AND id <> ?`
		args = append(args, excludeCreditoID)
	}

	var creditosActivos int
	var saldoTotalActual float64
	if err := dbConn.QueryRow(query, args...).Scan(&creditosActivos, &saldoTotalActual); err != nil {
		return err
	}

	saldoNuevo := creditoRound(creditoMax(saldoProyectado, 0))
	totalCreditosProyectado := creditosActivos
	if saldoNuevo > 0 {
		totalCreditosProyectado++
	}
	saldoTotalProyectado := creditoRound(saldoTotalActual + saldoNuevo)

	if limite.MaxCreditosActivos > 0 && totalCreditosProyectado > limite.MaxCreditosActivos {
		if limite.RequiereAprobacionExceso {
			return fmt.Errorf("limite max_creditos_activos excedido para cliente_id=%d; requiere aprobacion adicional", clienteID)
		}
		return fmt.Errorf("limite max_creditos_activos excedido para cliente_id=%d", clienteID)
	}

	limiteSaldo := creditoRound(creditoMax(limite.LimiteSaldoTotal, 0))
	if limiteSaldo > 0 && saldoTotalProyectado > limiteSaldo {
		if limite.RequiereAprobacionExceso {
			return fmt.Errorf("limite_saldo_total excedido para cliente_id=%d; requiere aprobacion adicional", clienteID)
		}
		return fmt.Errorf("limite_saldo_total excedido para cliente_id=%d", clienteID)
	}

	return nil
}

// EnsureEmpresaCreditosSchema crea/migra las tablas del modulo de creditos.
func EnsureEmpresaCreditosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_creditos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			codigo TEXT,
			cliente_id INTEGER DEFAULT 0,
			cliente_nombre TEXT,
			tipo_credito TEXT DEFAULT 'cuotas',
			monto_aprobado REAL DEFAULT 0,
			cupo_credito REAL DEFAULT 0,
			saldo_actual REAL DEFAULT 0,
			saldo_disponible REAL DEFAULT 0,
			tasa_interes REAL DEFAULT 0,
			tasa_mora REAL DEFAULT 0,
			periodicidad_cuota TEXT DEFAULT 'mensual',
			valor_cuota_pactada REAL DEFAULT 0,
			omitir_domingos INTEGER DEFAULT 0,
			plazo_dias INTEGER DEFAULT 0,
			plazo_cuotas INTEGER DEFAULT 0,
			fecha_inicio TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_vencimiento TEXT,
			fecha_ultimo_pago TEXT,
			dias_mora INTEGER DEFAULT 0,
			clasificacion_cartera TEXT DEFAULT 'al_dia',
			bloqueo_automatico_mora INTEGER DEFAULT 1,
			venta_origen_id INTEGER DEFAULT 0,
			documento_origen TEXT,
			estado_credito TEXT DEFAULT 'activo',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_creditos_empresa_estado ON empresa_creditos(empresa_id, estado_credito, estado, id DESC);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_creditos_empresa_codigo ON empresa_creditos(empresa_id, codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_creditos_cliente ON empresa_creditos(empresa_id, cliente_id, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_creditos_clientes_limites (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			cliente_id INTEGER NOT NULL,
			limite_saldo_total REAL DEFAULT 0,
			max_creditos_activos INTEGER DEFAULT 0,
			requiere_aprobacion_exceso INTEGER DEFAULT 0,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_creditos_clientes_limites_empresa_cliente ON empresa_creditos_clientes_limites(empresa_id, cliente_id);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_creditos_clientes_limites_estado ON empresa_creditos_clientes_limites(empresa_id, estado, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_creditos_cuotas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			credito_id INTEGER NOT NULL,
			numero_cuota INTEGER DEFAULT 1,
			fecha_vencimiento TEXT,
			valor_cuota REAL DEFAULT 0,
			capital_cuota REAL DEFAULT 0,
			interes_cuota REAL DEFAULT 0,
			interes_mora REAL DEFAULT 0,
			valor_pagado REAL DEFAULT 0,
			saldo_cuota REAL DEFAULT 0,
			estado_cuota TEXT DEFAULT 'pendiente',
			fecha_ultimo_pago TEXT,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_creditos_cuotas_credito_numero ON empresa_creditos_cuotas(empresa_id, credito_id, numero_cuota);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_creditos_cuotas_estado ON empresa_creditos_cuotas(empresa_id, credito_id, estado_cuota, fecha_vencimiento);`,
		`CREATE TABLE IF NOT EXISTS empresa_creditos_movimientos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			credito_id INTEGER NOT NULL,
			cuota_id INTEGER DEFAULT 0,
			tipo_movimiento TEXT DEFAULT 'abono',
			monto REAL DEFAULT 0,
			capital_aplicado REAL DEFAULT 0,
			interes_aplicado REAL DEFAULT 0,
			mora_aplicada REAL DEFAULT 0,
			metodo_pago TEXT,
			referencia_pago TEXT,
			comprobante TEXT,
			aplicado_automatico INTEGER DEFAULT 0,
			fecha_movimiento TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_creditos_movimientos_credito ON empresa_creditos_movimientos(empresa_id, credito_id, fecha_movimiento DESC, id DESC);`,
		`CREATE TABLE IF NOT EXISTS empresa_creditos_workflow (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			credito_id INTEGER NOT NULL,
			workflow_codigo TEXT,
			tipo_solicitud TEXT NOT NULL,
			estado_solicitud TEXT DEFAULT 'pendiente_aprobacion',
			movimiento_origen_id INTEGER DEFAULT 0,
			movimiento_resultado_id INTEGER DEFAULT 0,
			nivel_aprobacion_actual INTEGER DEFAULT 0,
			nivel_aprobacion_requerido INTEGER DEFAULT 1,
			aprobado_por TEXT,
			codigo_aprobacion TEXT,
			fecha_aprobacion_final TEXT,
			ejecutado_por TEXT,
			fecha_ejecucion TEXT,
			motivo_solicitud TEXT,
			motivo_rechazo TEXT,
			payload_json TEXT DEFAULT '{}',
			resultado_json TEXT DEFAULT '{}',
			historial_aprobaciones_json TEXT DEFAULT '[]',
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_creditos_workflow_codigo ON empresa_creditos_workflow(empresa_id, workflow_codigo);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_creditos_workflow_estado ON empresa_creditos_workflow(empresa_id, credito_id, tipo_solicitud, estado_solicitud, id DESC);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "estado_credito", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "clasificacion_cartera", "TEXT DEFAULT 'al_dia'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "saldo_disponible", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "fecha_ultimo_pago", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "periodicidad_cuota", "TEXT DEFAULT 'mensual'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "valor_cuota_pactada", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "omitir_domingos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "documento_origen", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "limite_saldo_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "max_creditos_activos", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "requiere_aprobacion_exceso", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_clientes_limites", "observaciones", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_cuotas", "interes_mora", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_cuotas", "fecha_ultimo_pago", "TEXT"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_movimientos", "aplicado_automatico", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_movimientos", "fecha_movimiento", "TEXT DEFAULT (CURRENT_TIMESTAMP)"); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "workflow_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "estado_solicitud", "TEXT DEFAULT 'pendiente_aprobacion'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "movimiento_origen_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "movimiento_resultado_id", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "nivel_aprobacion_actual", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "nivel_aprobacion_requerido", "INTEGER DEFAULT 1"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "aprobado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "codigo_aprobacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "fecha_aprobacion_final", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "ejecutado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "fecha_ejecucion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "motivo_solicitud", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "motivo_rechazo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "payload_json", "TEXT DEFAULT '{}'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "resultado_json", "TEXT DEFAULT '{}'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_creditos_workflow", "historial_aprobaciones_json", "TEXT DEFAULT '[]'"); err != nil {
		return err
	}

	// Ensure related tables used by credit flows exist (clientes schema required by some handlers/tests)
	if err := EnsureEmpresaClientesSchema(dbConn); err != nil {
		return err
	}

	return nil
}

func creditoHydrate(row *EmpresaCredito) {
	if row == nil {
		return
	}
	row.TipoCredito = creditoNormalizeTipo(row.TipoCredito)
	row.PeriodicidadCuota = creditoNormalizePeriodicidad(row.PeriodicidadCuota)
	row.EstadoCredito = creditoNormalizeEstado(row.EstadoCredito)
	row.Estado = creditoNormalizeRowEstado(row.Estado)
	if row.CupoCredito <= 0 {
		row.CupoCredito = row.MontoAprobado
	}
	if row.CupoCredito < row.SaldoActual {
		row.CupoCredito = row.SaldoActual
	}
	row.SaldoActual = creditoRound(creditoMax(row.SaldoActual, 0))
	row.CupoCredito = creditoRound(creditoMax(row.CupoCredito, 0))
	row.SaldoDisponible = creditoRound(creditoMax(row.CupoCredito-row.SaldoActual, 0))
	row.ValorCuotaPactada = creditoRound(creditoMax(row.ValorCuotaPactada, 0))
	row.DiasMora = creditoDaysMora(row.FechaVencimiento, row.SaldoActual)
	row.ClasificacionCartera = creditoResolveClasificacion(row.EstadoCredito, row.FechaVencimiento, row.SaldoActual)
}

// CreateEmpresaCredito crea un credito y, cuando aplica, genera tabla de cuotas.
func CreateEmpresaCredito(dbConn *sql.DB, payload EmpresaCredito) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if payload.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	payload.ClienteNombre = strings.TrimSpace(payload.ClienteNombre)
	if payload.ClienteID <= 0 && payload.ClienteNombre == "" {
		return 0, errors.New("cliente_id o cliente_nombre es obligatorio")
	}

	payload.TipoCredito = creditoNormalizeTipo(payload.TipoCredito)
	payload.EstadoCredito = creditoNormalizeEstado(payload.EstadoCredito)
	payload.Estado = creditoNormalizeRowEstado(payload.Estado)
	if payload.EstadoCredito == "" {
		payload.EstadoCredito = "activo"
	}
	if payload.Estado == "" {
		payload.Estado = "activo"
	}

	payload.MontoAprobado = creditoRound(payload.MontoAprobado)
	if payload.MontoAprobado <= 0 {
		return 0, errors.New("monto_aprobado debe ser mayor a cero")
	}
	payload.CupoCredito = creditoRound(payload.CupoCredito)
	if payload.CupoCredito <= 0 {
		payload.CupoCredito = payload.MontoAprobado
	}
	if payload.CupoCredito < payload.MontoAprobado {
		payload.CupoCredito = payload.MontoAprobado
	}

	payload.SaldoActual = creditoRound(payload.SaldoActual)
	if payload.SaldoActual <= 0 {
		payload.SaldoActual = payload.MontoAprobado
	}
	if payload.SaldoActual > payload.CupoCredito {
		payload.SaldoActual = payload.CupoCredito
	}

	if payload.PlazoDias < 0 {
		payload.PlazoDias = 0
	}
	if payload.PlazoCuotas < 0 {
		payload.PlazoCuotas = 0
	}
	payload.PeriodicidadCuota = creditoNormalizePeriodicidad(payload.PeriodicidadCuota)
	payload.ValorCuotaPactada = creditoRound(creditoMax(payload.ValorCuotaPactada, 0))
	if payload.TipoCredito == "cuotas" && payload.ValorCuotaPactada > 0 && payload.PlazoCuotas <= 0 {
		payload.PlazoCuotas = int(math.Ceil(payload.MontoAprobado / payload.ValorCuotaPactada))
	}
	if payload.TipoCredito == "cuotas" && payload.PeriodicidadCuota == "diaria" && payload.PlazoDias > 0 && payload.PlazoCuotas <= 0 {
		payload.PlazoCuotas = payload.PlazoDias
	}
	if payload.TipoCredito == "cuotas" && payload.PlazoCuotas <= 0 {
		if payload.PlazoDias > 0 {
			payload.PlazoCuotas = int(math.Ceil(float64(payload.PlazoDias) / 30.0))
		}
		if payload.PlazoCuotas <= 0 {
			payload.PlazoCuotas = 12
		}
	}
	if payload.TipoCredito == "fijo" && payload.PlazoCuotas <= 0 {
		payload.PlazoCuotas = 1
	}
	if payload.TasaInteres < 0 {
		payload.TasaInteres = 0
	}
	if payload.TasaMora < 0 {
		payload.TasaMora = 0
	}

	fechaInicio := time.Now().In(time.Local)
	if dt, ok := creditoParseDate(payload.FechaInicio); ok {
		fechaInicio = dt
	}
	payload.FechaInicio = fechaInicio.Format("2006-01-02")

	if strings.TrimSpace(payload.FechaVencimiento) == "" {
		if payload.TipoCredito == "rotativo" {
			payload.FechaVencimiento = fechaInicio.AddDate(1, 0, 0).Format("2006-01-02")
		} else if payload.PlazoCuotas > 0 && payload.PeriodicidadCuota == "diaria" && payload.OmitirDomingos {
			payload.FechaVencimiento = creditoNextFechaCuota(fechaInicio, payload.PeriodicidadCuota, payload.PlazoCuotas, payload.OmitirDomingos).Format("2006-01-02")
		} else if payload.PlazoDias > 0 {
			payload.FechaVencimiento = fechaInicio.AddDate(0, 0, payload.PlazoDias).Format("2006-01-02")
		} else if payload.PlazoCuotas > 0 {
			payload.FechaVencimiento = creditoAddPeriodo(fechaInicio, payload.PeriodicidadCuota, payload.PlazoCuotas).Format("2006-01-02")
		}
	}
	if strings.TrimSpace(payload.FechaVencimiento) == "" {
		payload.FechaVencimiento = fechaInicio.AddDate(0, 6, 0).Format("2006-01-02")
	}

	payload.Codigo = strings.TrimSpace(payload.Codigo)

	if err := creditoValidateClienteLimites(dbConn, payload.EmpresaID, payload.ClienteID, 0, payload.SaldoActual); err != nil {
		return 0, err
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	nowExpr := sqlNowExpr()
	id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_creditos (
		empresa_id,
		codigo,
		cliente_id,
		cliente_nombre,
		tipo_credito,
		monto_aprobado,
		cupo_credito,
		saldo_actual,
		saldo_disponible,
		tasa_interes,
		tasa_mora,
		periodicidad_cuota,
		valor_cuota_pactada,
		omitir_domingos,
		plazo_dias,
		plazo_cuotas,
		fecha_inicio,
		fecha_vencimiento,
		dias_mora,
		clasificacion_cartera,
		bloqueo_automatico_mora,
		venta_origen_id,
		documento_origen,
		estado_credito,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, `+nowExpr+`, `+nowExpr+`, ?, ?, ?)`,
		payload.EmpresaID,
		payload.Codigo,
		payload.ClienteID,
		payload.ClienteNombre,
		payload.TipoCredito,
		payload.MontoAprobado,
		payload.CupoCredito,
		payload.SaldoActual,
		creditoRound(payload.CupoCredito-payload.SaldoActual),
		creditoRound(payload.TasaInteres),
		creditoRound(payload.TasaMora),
		payload.PeriodicidadCuota,
		payload.ValorCuotaPactada,
		boolToInt(payload.OmitirDomingos),
		payload.PlazoDias,
		payload.PlazoCuotas,
		payload.FechaInicio,
		payload.FechaVencimiento,
		creditoDaysMora(payload.FechaVencimiento, payload.SaldoActual),
		creditoResolveClasificacion(payload.EstadoCredito, payload.FechaVencimiento, payload.SaldoActual),
		boolToInt(payload.BloqueoAutomaticoMora),
		payload.VentaOrigenID,
		strings.TrimSpace(payload.DocumentoOrigen),
		payload.EstadoCredito,
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	if payload.Codigo == "" {
		codigo := creditoDefaultCodigo(payload.EmpresaID, id)
		if _, err = execTxSQLCompat(tx, `UPDATE empresa_creditos SET codigo = ?, fecha_actualizacion = `+nowExpr+` WHERE empresa_id = ? AND id = ?`, codigo, payload.EmpresaID, id); err != nil {
			return 0, err
		}
	}

	if payload.TipoCredito == "cuotas" || payload.TipoCredito == "fijo" {
		if err = creditoGenerateCuotasTx(tx, payload.EmpresaID, id, payload); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

func creditoGenerateCuotasTx(tx *sql.Tx, empresaID, creditoID int64, payload EmpresaCredito) error {
	return creditoGenerateCuotasTxWithStart(tx, empresaID, creditoID, payload, 1)
}

func creditoGenerateCuotasTxWithStart(tx *sql.Tx, empresaID, creditoID int64, payload EmpresaCredito, startNumero int) error {
	nCuotas := payload.PlazoCuotas
	if nCuotas <= 0 {
		nCuotas = 1
	}
	maxCuotas := creditoMaxCuotas(payload.PeriodicidadCuota)
	if nCuotas > maxCuotas {
		nCuotas = maxCuotas
	}
	if startNumero <= 0 {
		startNumero = 1
	}

	fechaInicio, ok := creditoParseDate(payload.FechaInicio)
	if !ok {
		fechaInicio = time.Now().In(time.Local)
	}
	fechaFin, ok := creditoParseDate(payload.FechaVencimiento)
	if !ok {
		fechaFin = fechaInicio.AddDate(0, nCuotas, 0)
	}

	interesTotal := creditoRound(payload.MontoAprobado * (payload.TasaInteres / 100.0))
	interesCuota := creditoRound(interesTotal / float64(nCuotas))
	capitalCuota := creditoRound(payload.MontoAprobado / float64(nCuotas))
	if payload.ValorCuotaPactada > 0 && payload.TasaInteres <= 0 {
		capitalCuota = creditoRound(payload.ValorCuotaPactada)
	}
	valorCuotaBase := creditoRound(capitalCuota + interesCuota)
	nowExpr := sqlNowExpr()

	for i := 1; i <= nCuotas; i++ {
		numeroCuota := startNumero + i - 1
		fechaCuota := fechaInicio
		if payload.TipoCredito == "fijo" {
			fechaCuota = fechaFin
		} else {
			fechaCuota = creditoNextFechaCuota(fechaInicio, payload.PeriodicidadCuota, i, payload.OmitirDomingos)
		}

		capital := capitalCuota
		interes := interesCuota
		valorCuota := valorCuotaBase

		// Ajuste de centavos en la ultima cuota.
		if i == nCuotas {
			subtotalCapital := creditoRound(capitalCuota * float64(nCuotas-1))
			capital = creditoRound(creditoMax(payload.MontoAprobado-subtotalCapital, 0))
			subtotalInteres := creditoRound(interesCuota * float64(nCuotas-1))
			interes = creditoRound(creditoMax(interesTotal-subtotalInteres, 0))
			valorCuota = creditoRound(capital + interes)
		}

		if _, err := execTxSQLCompat(tx, `INSERT INTO empresa_creditos_cuotas (
			empresa_id,
			credito_id,
			numero_cuota,
			fecha_vencimiento,
			valor_cuota,
			capital_cuota,
			interes_cuota,
			interes_mora,
			valor_pagado,
			saldo_cuota,
			estado_cuota,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0, 0, ?, 'pendiente', `+nowExpr+`, `+nowExpr+`, ?, 'activo', ?)`,
			empresaID,
			creditoID,
			numeroCuota,
			creditoFormatDate(fechaCuota),
			valorCuota,
			capital,
			interes,
			valorCuota,
			strings.TrimSpace(payload.UsuarioCreador),
			"generacion automatica de cuota",
		); err != nil {
			return err
		}
	}
	return nil
}

func scanEmpresaCredito(scanner interface {
	Scan(dest ...interface{}) error
}) (*EmpresaCredito, error) {
	var row EmpresaCredito
	var bloqueoMora int
	var omitirDomingos int
	if err := scanner.Scan(
		&row.ID,
		&row.EmpresaID,
		&row.Codigo,
		&row.ClienteID,
		&row.ClienteNombre,
		&row.TipoCredito,
		&row.MontoAprobado,
		&row.CupoCredito,
		&row.SaldoActual,
		&row.SaldoDisponible,
		&row.TasaInteres,
		&row.TasaMora,
		&row.PeriodicidadCuota,
		&row.ValorCuotaPactada,
		&omitirDomingos,
		&row.PlazoDias,
		&row.PlazoCuotas,
		&row.FechaInicio,
		&row.FechaVencimiento,
		&row.FechaUltimoPago,
		&row.DiasMora,
		&row.ClasificacionCartera,
		&bloqueoMora,
		&row.VentaOrigenID,
		&row.DocumentoOrigen,
		&row.EstadoCredito,
		&row.FechaCreacion,
		&row.FechaActualizacion,
		&row.UsuarioCreador,
		&row.Estado,
		&row.Observaciones,
	); err != nil {
		return nil, err
	}
	row.OmitirDomingos = omitirDomingos == 1
	row.BloqueoAutomaticoMora = bloqueoMora == 1
	creditoHydrate(&row)
	return &row, nil
}

func creditoHydrateCuotaStatus(dbConn *sql.DB, empresaID int64, row *EmpresaCredito) {
	if dbConn == nil || row == nil || empresaID <= 0 || row.ID <= 0 {
		return
	}
	var pendientes int
	var vencidas int
	var fechaMasAntigua string
	var fechaProxima string
	today := time.Now().In(time.Local).Format("2006-01-02")
	err := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida') AND COALESCE(saldo_cuota, 0) > 0 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN LOWER(COALESCE(estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida') AND COALESCE(saldo_cuota, 0) > 0 AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ? THEN 1 ELSE 0 END), 0),
		COALESCE(MIN(CASE WHEN LOWER(COALESCE(estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida') AND COALESCE(saldo_cuota, 0) > 0 AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ? THEN fecha_vencimiento ELSE NULL END), ''),
		COALESCE(MIN(CASE WHEN LOWER(COALESCE(estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida') AND COALESCE(saldo_cuota, 0) > 0 THEN fecha_vencimiento ELSE NULL END), '')
	FROM empresa_creditos_cuotas
	WHERE empresa_id = ?
	  AND credito_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, today, today, today, today, empresaID, row.ID).Scan(&pendientes, &vencidas, &fechaMasAntigua, &fechaProxima)
	if err != nil {
		return
	}
	row.CuotasPendientes = pendientes
	row.CuotasVencidas = vencidas
	row.FechaProximaCuota = strings.TrimSpace(fechaProxima)
	if dt, ok := creditoParseDate(fechaMasAntigua); ok && row.SaldoActual > 0 {
		today := time.Now().In(time.Local)
		days := int(today.Sub(dt).Hours() / 24)
		if days > 0 {
			row.DiasCuotasVencidas = days
		}
	}
	if row.CuotasVencidas > 0 && row.DiasMora <= 0 {
		row.ClasificacionCartera = "vencido"
	}
}

func creditoHydrateCuotaStatusRows(dbConn *sql.DB, empresaID int64, rows []EmpresaCredito) {
	for idx := range rows {
		creditoHydrateCuotaStatus(dbConn, empresaID, &rows[idx])
	}
}

func listEmpresaCreditosByWhere(dbConn *sql.DB, empresaID int64, whereSQL, orderSQL string, args []interface{}, limit int) ([]EmpresaCredito, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 2000 {
		limit = 2000
	}
	if strings.TrimSpace(orderSQL) == "" {
		orderSQL = "id DESC"
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(tipo_credito, 'cuotas'),
		COALESCE(monto_aprobado, 0),
		COALESCE(cupo_credito, 0),
		COALESCE(saldo_actual, 0),
		COALESCE(saldo_disponible, 0),
		COALESCE(tasa_interes, 0),
		COALESCE(tasa_mora, 0),
		COALESCE(periodicidad_cuota, 'mensual'),
		COALESCE(valor_cuota_pactada, 0),
		COALESCE(omitir_domingos, 0),
		COALESCE(plazo_dias, 0),
		COALESCE(plazo_cuotas, 0),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(fecha_ultimo_pago, ''),
		COALESCE(dias_mora, 0),
		COALESCE(clasificacion_cartera, 'al_dia'),
		COALESCE(bloqueo_automatico_mora, 1),
		COALESCE(venta_origen_id, 0),
		COALESCE(documento_origen, ''),
		COALESCE(estado_credito, 'activo'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos
	WHERE empresa_id = ?` + whereSQL + `
	ORDER BY ` + orderSQL + `
	LIMIT ?`

	queryArgs := make([]interface{}, 0, 2+len(args))
	queryArgs = append(queryArgs, empresaID)
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, limit)

	rows, err := dbConn.Query(query, queryArgs...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaCredito, 0)
	for rows.Next() {
		item, scanErr := scanEmpresaCredito(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	creditoHydrateCuotaStatusRows(dbConn, empresaID, out)
	return out, nil
}

// GetEmpresaCreditoByID obtiene un credito puntual por empresa.
func GetEmpresaCreditoByID(dbConn *sql.DB, empresaID, creditoID int64) (*EmpresaCredito, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || creditoID <= 0 {
		return nil, errors.New("empresa_id o credito_id invalido")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(tipo_credito, 'cuotas'),
		COALESCE(monto_aprobado, 0),
		COALESCE(cupo_credito, 0),
		COALESCE(saldo_actual, 0),
		COALESCE(saldo_disponible, 0),
		COALESCE(tasa_interes, 0),
		COALESCE(tasa_mora, 0),
		COALESCE(periodicidad_cuota, 'mensual'),
		COALESCE(valor_cuota_pactada, 0),
		COALESCE(omitir_domingos, 0),
		COALESCE(plazo_dias, 0),
		COALESCE(plazo_cuotas, 0),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(fecha_ultimo_pago, ''),
		COALESCE(dias_mora, 0),
		COALESCE(clasificacion_cartera, 'al_dia'),
		COALESCE(bloqueo_automatico_mora, 1),
		COALESCE(venta_origen_id, 0),
		COALESCE(documento_origen, ''),
		COALESCE(estado_credito, 'activo'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos
	WHERE empresa_id = ?
	  AND id = ?
	LIMIT 1`, empresaID, creditoID)

	credito, err := scanEmpresaCredito(row)
	if err != nil {
		return nil, err
	}
	creditoHydrateCuotaStatus(dbConn, empresaID, credito)
	return credito, nil
}

func creditoBuildWhere(filter EmpresaCreditoFilter) (string, []interface{}) {
	clauses := make([]string, 0, 8)
	args := make([]interface{}, 0, 8)

	if !filter.IncludeInactive {
		clauses = append(clauses, "LOWER(COALESCE(estado, 'activo')) = 'activo'")
	}
	if filter.ClienteID > 0 {
		clauses = append(clauses, "COALESCE(cliente_id, 0) = ?")
		args = append(args, filter.ClienteID)
	}
	if strings.TrimSpace(filter.EstadoCredito) != "" {
		clauses = append(clauses, "LOWER(COALESCE(estado_credito, 'activo')) = LOWER(?)")
		args = append(args, creditoNormalizeEstado(filter.EstadoCredito))
	}
	if strings.TrimSpace(filter.Clasificacion) != "" {
		clauses = append(clauses, "LOWER(COALESCE(clasificacion_cartera, 'al_dia')) = LOWER(?)")
		args = append(args, creditoNormalizeClasificacion(filter.Clasificacion))
	}
	if strings.TrimSpace(filter.Desde) != "" {
		clauses = append(clauses, "COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_inicio, fecha_creacion, '')), 1, 10), ''), '0000-00-00') >= ?")
		args = append(args, strings.TrimSpace(filter.Desde))
	}
	if strings.TrimSpace(filter.Hasta) != "" {
		clauses = append(clauses, "COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_inicio, fecha_creacion, '')), 1, 10), ''), '9999-12-31') <= ?")
		args = append(args, strings.TrimSpace(filter.Hasta))
	}
	if strings.TrimSpace(filter.Q) != "" {
		pattern := creditoLikePattern(filter.Q)
		clauses = append(clauses, "(LOWER(COALESCE(codigo, '')) LIKE LOWER(?) ESCAPE '!' OR LOWER(COALESCE(cliente_nombre, '')) LIKE LOWER(?) ESCAPE '!' OR LOWER(COALESCE(documento_origen, '')) LIKE LOWER(?) ESCAPE '!')")
		args = append(args, pattern, pattern, pattern)
	}
	if filter.SoloVencidos {
		today := time.Now().In(time.Local).Format("2006-01-02")
		clauses = append(clauses, `COALESCE(saldo_actual, 0) > 0 AND (
			COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ?
			OR EXISTS (
				SELECT 1
				FROM empresa_creditos_cuotas cc
				WHERE cc.empresa_id = empresa_creditos.empresa_id
				  AND cc.credito_id = empresa_creditos.id
				  AND LOWER(COALESCE(cc.estado, 'activo')) = 'activo'
				  AND LOWER(COALESCE(cc.estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
				  AND COALESCE(cc.saldo_cuota, 0) > 0
				  AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(cc.fecha_vencimiento, '')), 1, 10), ''), ?) < ?
			)
		)`)
		args = append(args, today, today, today, today)
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " AND " + strings.Join(clauses, " AND "), args
}

// ListEmpresaCreditos lista creditos por empresa con filtros y total.
func ListEmpresaCreditos(dbConn *sql.DB, empresaID int64, filter EmpresaCreditoFilter) ([]EmpresaCredito, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	whereSQL, whereArgs := creditoBuildWhere(filter)
	countQuery := "SELECT COUNT(1) FROM empresa_creditos WHERE empresa_id = ?" + whereSQL
	countArgs := append([]interface{}{empresaID}, whereArgs...)

	var total int64
	if err := dbConn.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := creditoNormalizeLimitOffset(filter.Limit, filter.Offset)
	query := `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(tipo_credito, 'cuotas'),
		COALESCE(monto_aprobado, 0),
		COALESCE(cupo_credito, 0),
		COALESCE(saldo_actual, 0),
		COALESCE(saldo_disponible, 0),
		COALESCE(tasa_interes, 0),
		COALESCE(tasa_mora, 0),
		COALESCE(periodicidad_cuota, 'mensual'),
		COALESCE(valor_cuota_pactada, 0),
		COALESCE(omitir_domingos, 0),
		COALESCE(plazo_dias, 0),
		COALESCE(plazo_cuotas, 0),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(fecha_ultimo_pago, ''),
		COALESCE(dias_mora, 0),
		COALESCE(clasificacion_cartera, 'al_dia'),
		COALESCE(bloqueo_automatico_mora, 1),
		COALESCE(venta_origen_id, 0),
		COALESCE(documento_origen, ''),
		COALESCE(estado_credito, 'activo'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos
	WHERE empresa_id = ?` + whereSQL + `
	ORDER BY id DESC
	LIMIT ? OFFSET ?`
	args := append([]interface{}{empresaID}, whereArgs...)
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaCredito, 0)
	for rows.Next() {
		item, err := scanEmpresaCredito(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	creditoHydrateCuotaStatusRows(dbConn, empresaID, out)
	return out, total, nil
}

// ListEmpresaCreditoCuotas lista cuotas de un credito ordenadas por numero.
func ListEmpresaCreditoCuotas(dbConn *sql.DB, empresaID, creditoID int64, includeInactive bool) ([]EmpresaCreditoCuota, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || creditoID <= 0 {
		return nil, errors.New("empresa_id o credito_id invalido")
	}

	query := `SELECT
		id,
		empresa_id,
		credito_id,
		numero_cuota,
		COALESCE(fecha_vencimiento, ''),
		COALESCE(valor_cuota, 0),
		COALESCE(capital_cuota, 0),
		COALESCE(interes_cuota, 0),
		COALESCE(interes_mora, 0),
		COALESCE(valor_pagado, 0),
		COALESCE(saldo_cuota, 0),
		COALESCE(estado_cuota, 'pendiente'),
		COALESCE(fecha_ultimo_pago, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_cuotas
	WHERE empresa_id = ?
	  AND credito_id = ?`
	args := []interface{}{empresaID, creditoID}
	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}
	query += ` ORDER BY numero_cuota ASC, id ASC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaCreditoCuota, 0)
	for rows.Next() {
		var item EmpresaCreditoCuota
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CreditoID,
			&item.NumeroCuota,
			&item.FechaVencimiento,
			&item.ValorCuota,
			&item.CapitalCuota,
			&item.InteresCuota,
			&item.InteresMora,
			&item.ValorPagado,
			&item.SaldoCuota,
			&item.EstadoCuota,
			&item.FechaUltimoPago,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.EstadoCuota = creditoNormalizeCuotaEstado(item.EstadoCuota)
		item.Estado = creditoNormalizeRowEstado(item.Estado)
		item.ValorCuota = creditoRound(item.ValorCuota)
		item.CapitalCuota = creditoRound(item.CapitalCuota)
		item.InteresCuota = creditoRound(item.InteresCuota)
		item.InteresMora = creditoRound(item.InteresMora)
		item.ValorPagado = creditoRound(item.ValorPagado)
		item.SaldoCuota = creditoRound(creditoMax(item.SaldoCuota, 0))
		if item.SaldoCuota <= 0 && item.EstadoCuota != "anulada" {
			item.EstadoCuota = "pagada"
		}
		if item.SaldoCuota > 0 && item.ValorPagado > 0 && item.EstadoCuota == "pendiente" {
			item.EstadoCuota = "parcial"
		}
		if item.SaldoCuota > 0 && creditoDaysMora(item.FechaVencimiento, item.SaldoCuota) > 0 && item.EstadoCuota != "pagada" && item.EstadoCuota != "anulada" {
			item.EstadoCuota = "vencida"
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListEmpresaCreditoMovimientos lista movimientos de un credito.
func ListEmpresaCreditoMovimientos(dbConn *sql.DB, empresaID, creditoID int64, includeInactive bool, limit int) ([]EmpresaCreditoMovimiento, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || creditoID <= 0 {
		return nil, errors.New("empresa_id o credito_id invalido")
	}
	if limit <= 0 {
		limit = 200
	}
	if limit > 2000 {
		limit = 2000
	}

	query := `SELECT
		id,
		empresa_id,
		credito_id,
		COALESCE(cuota_id, 0),
		COALESCE(tipo_movimiento, 'abono'),
		COALESCE(monto, 0),
		COALESCE(capital_aplicado, 0),
		COALESCE(interes_aplicado, 0),
		COALESCE(mora_aplicada, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(referencia_pago, ''),
		COALESCE(comprobante, ''),
		COALESCE(aplicado_automatico, 0),
		COALESCE(fecha_movimiento, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_movimientos
	WHERE empresa_id = ?
	  AND credito_id = ?`
	args := []interface{}{empresaID, creditoID}
	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}
	query += ` ORDER BY pcs_ts(COALESCE(fecha_movimiento, fecha_creacion)) DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaCreditoMovimiento, 0)
	for rows.Next() {
		var item EmpresaCreditoMovimiento
		var auto int
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.CreditoID,
			&item.CuotaID,
			&item.TipoMovimiento,
			&item.Monto,
			&item.CapitalAplicado,
			&item.InteresAplicado,
			&item.MoraAplicada,
			&item.MetodoPago,
			&item.ReferenciaPago,
			&item.Comprobante,
			&auto,
			&item.FechaMovimiento,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.TipoMovimiento = creditoNormalizeMovimiento(item.TipoMovimiento)
		item.AplicadoAutomatico = auto == 1
		item.Estado = creditoNormalizeRowEstado(item.Estado)
		item.Monto = creditoRound(item.Monto)
		item.CapitalAplicado = creditoRound(item.CapitalAplicado)
		item.InteresAplicado = creditoRound(item.InteresAplicado)
		item.MoraAplicada = creditoRound(item.MoraAplicada)
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// RegisterEmpresaCreditoAbono registra un abono y actualiza saldo/cuotas.
func RegisterEmpresaCreditoAbono(dbConn *sql.DB, input EmpresaCreditoAbonoInput) (int64, *EmpresaCredito, error) {
	if dbConn == nil {
		return 0, nil, errors.New("db connection is nil")
	}
	if input.EmpresaID <= 0 || input.CreditoID <= 0 {
		return 0, nil, errors.New("empresa_id o credito_id invalido")
	}
	input.Monto = creditoRound(input.Monto)
	if input.Monto <= 0 {
		return 0, nil, errors.New("monto debe ser mayor a cero")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	row := queryRowTxSQLCompat(tx, `SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(tipo_credito, 'cuotas'),
		COALESCE(monto_aprobado, 0),
		COALESCE(cupo_credito, 0),
		COALESCE(saldo_actual, 0),
		COALESCE(saldo_disponible, 0),
		COALESCE(tasa_interes, 0),
		COALESCE(tasa_mora, 0),
		COALESCE(periodicidad_cuota, 'mensual'),
		COALESCE(valor_cuota_pactada, 0),
		COALESCE(omitir_domingos, 0),
		COALESCE(plazo_dias, 0),
		COALESCE(plazo_cuotas, 0),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(fecha_ultimo_pago, ''),
		COALESCE(dias_mora, 0),
		COALESCE(clasificacion_cartera, 'al_dia'),
		COALESCE(bloqueo_automatico_mora, 1),
		COALESCE(venta_origen_id, 0),
		COALESCE(documento_origen, ''),
		COALESCE(estado_credito, 'activo'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos
	WHERE empresa_id = ?
	  AND id = ?
	LIMIT 1`, input.EmpresaID, input.CreditoID)
	credito, err := scanEmpresaCredito(row)
	if err != nil {
		return 0, nil, err
	}
	if credito.EstadoCredito == "cerrado" || credito.EstadoCredito == "castigado" {
		return 0, nil, fmt.Errorf("el credito no admite abonos en estado %s", credito.EstadoCredito)
	}
	if credito.SaldoActual <= 0 {
		return 0, nil, errors.New("el credito no tiene saldo pendiente")
	}

	montoAplicado := creditoMin(input.Monto, credito.SaldoActual)
	restante := montoAplicado
	capitalAplicado := 0.0
	interesAplicado := 0.0
	moraAplicada := 0.0

	cuotasRows, err := queryTxSQLCompat(tx, `SELECT
		id,
		COALESCE(valor_cuota, 0),
		COALESCE(capital_cuota, 0),
		COALESCE(interes_cuota, 0),
		COALESCE(interes_mora, 0),
		COALESCE(valor_pagado, 0),
		COALESCE(saldo_cuota, 0),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(estado_cuota, 'pendiente')
	FROM empresa_creditos_cuotas
	WHERE empresa_id = ?
	  AND credito_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND LOWER(COALESCE(estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
	ORDER BY numero_cuota ASC, id ASC`, input.EmpresaID, input.CreditoID)
	if err != nil {
		return 0, nil, err
	}
	type cuotaPendiente struct {
		id               int64
		valorCuota       float64
		capitalCuota     float64
		interesCuota     float64
		interesMora      float64
		valorPagado      float64
		saldoCuota       float64
		fechaVencimiento string
		estadoCuota      string
	}
	cuotasPendientes := make([]cuotaPendiente, 0)
	for cuotasRows.Next() {
		var cuota cuotaPendiente
		if err := cuotasRows.Scan(&cuota.id, &cuota.valorCuota, &cuota.capitalCuota, &cuota.interesCuota, &cuota.interesMora, &cuota.valorPagado, &cuota.saldoCuota, &cuota.fechaVencimiento, &cuota.estadoCuota); err != nil {
			_ = cuotasRows.Close()
			return 0, nil, err
		}
		cuotasPendientes = append(cuotasPendientes, cuota)
	}
	if err := cuotasRows.Err(); err != nil {
		_ = cuotasRows.Close()
		return 0, nil, err
	}
	if err := cuotasRows.Close(); err != nil {
		return 0, nil, err
	}

	for _, cuota := range cuotasPendientes {
		if restante <= 0 {
			break
		}
		cuotaID := cuota.id
		valorCuota := cuota.valorCuota
		capitalCuota := cuota.capitalCuota
		interesCuota := cuota.interesCuota
		interesMora := cuota.interesMora
		valorPagado := cuota.valorPagado
		saldoCuota := cuota.saldoCuota
		fechaVencimiento := cuota.fechaVencimiento
		saldoCuota = creditoRound(creditoMax(saldoCuota, 0))
		if saldoCuota <= 0 {
			continue
		}

		aplicar := creditoMin(restante, saldoCuota)
		nuevoPagado := creditoRound(valorPagado + aplicar)
		nuevoSaldo := creditoRound(creditoMax(valorCuota-nuevoPagado, 0))
		nuevoEstado := "parcial"
		if nuevoSaldo <= 0 {
			nuevoEstado = "pagada"
		} else if creditoDaysMora(fechaVencimiento, nuevoSaldo) > 0 {
			nuevoEstado = "vencida"
		}

		ratioCapital := 1.0
		ratioInteres := 0.0
		totalBase := capitalCuota + interesCuota
		if totalBase > 0 {
			ratioCapital = capitalCuota / totalBase
			ratioInteres = interesCuota / totalBase
		}
		capPart := creditoRound(aplicar * ratioCapital)
		intPart := creditoRound(aplicar * ratioInteres)
		if capPart+intPart > aplicar {
			intPart = creditoRound(aplicar - capPart)
		}

		capitalAplicado = creditoRound(capitalAplicado + capPart)
		interesAplicado = creditoRound(interesAplicado + intPart)
		if interesMora > 0 && nuevoEstado == "vencida" {
			moraAplicada = creditoRound(moraAplicada)
		}

		if _, err := execTxSQLCompat(tx, `UPDATE empresa_creditos_cuotas SET
			valor_pagado = ?,
			saldo_cuota = ?,
			estado_cuota = ?,
			fecha_ultimo_pago = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ?
		  AND credito_id = ?
		  AND id = ?`,
			nuevoPagado,
			nuevoSaldo,
			nuevoEstado,
			time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
			input.EmpresaID,
			input.CreditoID,
			cuotaID,
		); err != nil {
			return 0, nil, err
		}

		restante = creditoRound(restante - aplicar)
	}

	if restante > 0 {
		capitalAplicado = creditoRound(capitalAplicado + restante)
		restante = 0
	}

	saldoNuevo := creditoRound(creditoMax(credito.SaldoActual-montoAplicado, 0))
	estadoNuevo := credito.EstadoCredito
	if saldoNuevo <= 0 {
		estadoNuevo = "cerrado"
	}
	fechaPago := strings.TrimSpace(input.FechaMovimiento)
	if fechaPago == "" {
		fechaPago = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	}

	clasificacion := creditoResolveClasificacion(estadoNuevo, credito.FechaVencimiento, saldoNuevo)
	diasMora := creditoDaysMora(credito.FechaVencimiento, saldoNuevo)

	if _, err := execTxSQLCompat(tx, `UPDATE empresa_creditos SET
		saldo_actual = ?,
		saldo_disponible = ?,
		estado_credito = ?,
		clasificacion_cartera = ?,
		dias_mora = ?,
		fecha_ultimo_pago = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ?
	  AND id = ?`,
		saldoNuevo,
		creditoRound(creditoMax(credito.CupoCredito-saldoNuevo, 0)),
		estadoNuevo,
		clasificacion,
		diasMora,
		fechaPago,
		input.EmpresaID,
		input.CreditoID,
	); err != nil {
		return 0, nil, err
	}

	movID, err := insertTxSQLCompat(tx, `INSERT INTO empresa_creditos_movimientos (
		empresa_id,
		credito_id,
		cuota_id,
		tipo_movimiento,
		monto,
		capital_aplicado,
		interes_aplicado,
		mora_aplicada,
		metodo_pago,
		referencia_pago,
		comprobante,
		aplicado_automatico,
		fecha_movimiento,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, 0, 'abono', ?, ?, ?, ?, ?, ?, ?, 0, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?)`,
		input.EmpresaID,
		input.CreditoID,
		montoAplicado,
		capitalAplicado,
		interesAplicado,
		moraAplicada,
		strings.TrimSpace(input.MetodoPago),
		strings.TrimSpace(input.ReferenciaPago),
		strings.TrimSpace(input.Comprobante),
		fechaPago,
		strings.TrimSpace(input.UsuarioCreador),
		strings.TrimSpace(input.Observaciones),
	)
	if err != nil {
		return 0, nil, err
	}

	if err := tx.Commit(); err != nil {
		return 0, nil, err
	}

	updated, err := GetEmpresaCreditoByID(dbConn, input.EmpresaID, input.CreditoID)
	if err != nil {
		return 0, nil, err
	}
	return movID, updated, nil
}

// SetEmpresaCreditoEstado ajusta estado operativo del credito.
func SetEmpresaCreditoEstado(dbConn *sql.DB, empresaID, creditoID int64, estadoCredito string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || creditoID <= 0 {
		return errors.New("empresa_id o credito_id invalido")
	}
	estado := creditoNormalizeEstado(estadoCredito)
	_, err := dbConn.Exec(`UPDATE empresa_creditos SET estado_credito = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, estado, empresaID, creditoID)
	return err
}

// SetEmpresaCreditoRowEstado activa o desactiva el registro del credito.
func SetEmpresaCreditoRowEstado(dbConn *sql.DB, empresaID, creditoID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || creditoID <= 0 {
		return errors.New("empresa_id o credito_id invalido")
	}
	_, err := dbConn.Exec(`UPDATE empresa_creditos SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, creditoNormalizeRowEstado(estado), empresaID, creditoID)
	return err
}

// UpdateEmpresaCredito actualiza metadata operativa del credito sin afectar historico de movimientos.
func UpdateEmpresaCredito(dbConn *sql.DB, payload EmpresaCredito) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if payload.EmpresaID <= 0 || payload.ID <= 0 {
		return errors.New("empresa_id o id invalido")
	}

	payload.ClienteNombre = strings.TrimSpace(payload.ClienteNombre)
	payload.TipoCredito = creditoNormalizeTipo(payload.TipoCredito)
	payload.PeriodicidadCuota = creditoNormalizePeriodicidad(payload.PeriodicidadCuota)
	payload.ValorCuotaPactada = creditoRound(creditoMax(payload.ValorCuotaPactada, 0))
	payload.EstadoCredito = creditoNormalizeEstado(payload.EstadoCredito)
	payload.Estado = creditoNormalizeRowEstado(payload.Estado)
	payload.ClasificacionCartera = creditoNormalizeClasificacion(payload.ClasificacionCartera)

	payload.MontoAprobado = creditoRound(creditoMax(payload.MontoAprobado, 0))
	payload.CupoCredito = creditoRound(creditoMax(payload.CupoCredito, payload.MontoAprobado))
	payload.SaldoActual = creditoRound(creditoMax(payload.SaldoActual, 0))
	if payload.SaldoActual > payload.CupoCredito {
		payload.SaldoActual = payload.CupoCredito
	}
	payload.SaldoDisponible = creditoRound(creditoMax(payload.CupoCredito-payload.SaldoActual, 0))
	payload.DiasMora = creditoDaysMora(payload.FechaVencimiento, payload.SaldoActual)
	if payload.ClasificacionCartera == "" {
		payload.ClasificacionCartera = creditoResolveClasificacion(payload.EstadoCredito, payload.FechaVencimiento, payload.SaldoActual)
	}

	if payload.EstadoCredito == "cerrado" && payload.SaldoActual > 0 {
		return errors.New("no se puede cerrar credito con saldo pendiente")
	}
	if err := creditoValidateClienteLimites(dbConn, payload.EmpresaID, payload.ClienteID, payload.ID, payload.SaldoActual); err != nil {
		return err
	}

	_, err := dbConn.Exec(`UPDATE empresa_creditos SET
		codigo = ?,
		cliente_id = ?,
		cliente_nombre = ?,
		tipo_credito = ?,
		monto_aprobado = ?,
		cupo_credito = ?,
		saldo_actual = ?,
		saldo_disponible = ?,
		tasa_interes = ?,
		tasa_mora = ?,
		periodicidad_cuota = ?,
		valor_cuota_pactada = ?,
		omitir_domingos = ?,
		plazo_dias = ?,
		plazo_cuotas = ?,
		fecha_inicio = ?,
		fecha_vencimiento = ?,
		dias_mora = ?,
		clasificacion_cartera = ?,
		bloqueo_automatico_mora = ?,
		venta_origen_id = ?,
		documento_origen = ?,
		estado_credito = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ?
	  AND id = ?`,
		strings.TrimSpace(payload.Codigo),
		payload.ClienteID,
		payload.ClienteNombre,
		payload.TipoCredito,
		payload.MontoAprobado,
		payload.CupoCredito,
		payload.SaldoActual,
		payload.SaldoDisponible,
		creditoRound(payload.TasaInteres),
		creditoRound(payload.TasaMora),
		payload.PeriodicidadCuota,
		payload.ValorCuotaPactada,
		boolToInt(payload.OmitirDomingos),
		payload.PlazoDias,
		payload.PlazoCuotas,
		strings.TrimSpace(payload.FechaInicio),
		strings.TrimSpace(payload.FechaVencimiento),
		payload.DiasMora,
		payload.ClasificacionCartera,
		boolToInt(payload.BloqueoAutomaticoMora),
		payload.VentaOrigenID,
		strings.TrimSpace(payload.DocumentoOrigen),
		payload.EstadoCredito,
		strings.TrimSpace(payload.UsuarioCreador),
		payload.Estado,
		strings.TrimSpace(payload.Observaciones),
		payload.EmpresaID,
		payload.ID,
	)
	return err
}

// GetEmpresaCreditoClienteLimite obtiene limite puntual por cliente.
func GetEmpresaCreditoClienteLimite(dbConn *sql.DB, empresaID, clienteID int64, includeInactive bool) (*EmpresaCreditoClienteLimite, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || clienteID <= 0 {
		return nil, errors.New("empresa_id o cliente_id invalido")
	}

	query := `SELECT
		id,
		empresa_id,
		cliente_id,
		COALESCE(limite_saldo_total, 0),
		COALESCE(max_creditos_activos, 0),
		COALESCE(requiere_aprobacion_exceso, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_clientes_limites
	WHERE empresa_id = ? AND cliente_id = ?`
	args := []interface{}{empresaID, clienteID}
	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}
	query += ` LIMIT 1`

	return scanEmpresaCreditoClienteLimite(dbConn.QueryRow(query, args...))
}

// ListEmpresaCreditoClienteLimites lista limites por cliente con filtros.
func ListEmpresaCreditoClienteLimites(dbConn *sql.DB, empresaID int64, filter EmpresaCreditoClienteLimiteFilter) ([]EmpresaCreditoClienteLimite, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	clauses := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if filter.ClienteID > 0 {
		clauses = append(clauses, "cliente_id = ?")
		args = append(args, filter.ClienteID)
	}
	if !filter.IncludeInactive {
		clauses = append(clauses, "LOWER(COALESCE(estado, 'activo')) = 'activo'")
	}

	whereSQL := ""
	if len(clauses) > 0 {
		whereSQL = " AND " + strings.Join(clauses, " AND ")
	}

	countQuery := "SELECT COUNT(1) FROM empresa_creditos_clientes_limites WHERE empresa_id = ?" + whereSQL
	countArgs := append([]interface{}{empresaID}, args...)
	var total int64
	if err := dbConn.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := creditoNormalizeLimitOffset(filter.Limit, filter.Offset)
	query := `SELECT
		id,
		empresa_id,
		cliente_id,
		COALESCE(limite_saldo_total, 0),
		COALESCE(max_creditos_activos, 0),
		COALESCE(requiere_aprobacion_exceso, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_clientes_limites
	WHERE empresa_id = ?` + whereSQL + `
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	queryArgs := append([]interface{}{empresaID}, args...)
	queryArgs = append(queryArgs, limit, offset)

	rows, err := dbConn.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaCreditoClienteLimite, 0)
	for rows.Next() {
		item, scanErr := scanEmpresaCreditoClienteLimite(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

// UpsertEmpresaCreditoClienteLimite crea/actualiza limites de credito por cliente.
func UpsertEmpresaCreditoClienteLimite(dbConn *sql.DB, payload EmpresaCreditoClienteLimite) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if payload.EmpresaID <= 0 || payload.ClienteID <= 0 {
		return 0, errors.New("empresa_id o cliente_id invalido")
	}
	creditoClienteLimiteNormalize(&payload)

	var existingID int64
	err := dbConn.QueryRow(`SELECT id FROM empresa_creditos_clientes_limites WHERE empresa_id = ? AND cliente_id = ? LIMIT 1`, payload.EmpresaID, payload.ClienteID).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if err == sql.ErrNoRows {
		id, insertErr := insertSQLCompat(dbConn, `INSERT INTO empresa_creditos_clientes_limites (
			empresa_id,
			cliente_id,
			limite_saldo_total,
			max_creditos_activos,
			requiere_aprobacion_exceso,
			fecha_creacion,
			fecha_actualizacion,
			usuario_creador,
			estado,
			observaciones
		) VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, ?, ?)`,
			payload.EmpresaID,
			payload.ClienteID,
			payload.LimiteSaldoTotal,
			payload.MaxCreditosActivos,
			boolToInt(payload.RequiereAprobacionExceso),
			payload.UsuarioCreador,
			payload.Estado,
			payload.Observaciones,
		)
		if insertErr != nil {
			return 0, insertErr
		}
		return id, nil
	}

	_, err = dbConn.Exec(`UPDATE empresa_creditos_clientes_limites SET
		limite_saldo_total = ?,
		max_creditos_activos = ?,
		requiere_aprobacion_exceso = ?,
		usuario_creador = ?,
		estado = ?,
		observaciones = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND cliente_id = ?`,
		payload.LimiteSaldoTotal,
		payload.MaxCreditosActivos,
		boolToInt(payload.RequiereAprobacionExceso),
		payload.UsuarioCreador,
		payload.Estado,
		payload.Observaciones,
		payload.EmpresaID,
		payload.ClienteID,
	)
	if err != nil {
		return 0, err
	}

	return existingID, nil
}

// SetEmpresaCreditoClienteLimiteRowEstado activa o desactiva una regla de limite por cliente.
func SetEmpresaCreditoClienteLimiteRowEstado(dbConn *sql.DB, empresaID, clienteID int64, estado string) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	if empresaID <= 0 || clienteID <= 0 {
		return errors.New("empresa_id o cliente_id invalido")
	}
	_, err := dbConn.Exec(`UPDATE empresa_creditos_clientes_limites SET estado = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND cliente_id = ?`, creditoNormalizeRowEstado(estado), empresaID, clienteID)
	return err
}

// GetEmpresaCreditosCarteraResumen retorna agregados ejecutivos de cartera.
func GetEmpresaCreditosCarteraResumen(dbConn *sql.DB, empresaID int64, includeInactive bool) (*EmpresaCreditoCarteraResumen, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}

	today := time.Now().In(time.Local).Format("2006-01-02")
	query := `SELECT
		COUNT(1),
		SUM(CASE WHEN LOWER(COALESCE(estado_credito, 'activo')) = 'activo' THEN 1 ELSE 0 END),
		SUM(CASE WHEN COALESCE(saldo_actual, 0) > 0 AND (
			COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ?
			OR EXISTS (
				SELECT 1 FROM empresa_creditos_cuotas cc
				WHERE cc.empresa_id = empresa_creditos.empresa_id
				  AND cc.credito_id = empresa_creditos.id
				  AND LOWER(COALESCE(cc.estado, 'activo')) = 'activo'
				  AND LOWER(COALESCE(cc.estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
				  AND COALESCE(cc.saldo_cuota, 0) > 0
				  AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(cc.fecha_vencimiento, '')), 1, 10), ''), ?) < ?
			)
		) THEN 1 ELSE 0 END),
		SUM(CASE WHEN COALESCE(dias_mora, 0) > 0 OR EXISTS (
			SELECT 1 FROM empresa_creditos_cuotas cc
			WHERE cc.empresa_id = empresa_creditos.empresa_id
			  AND cc.credito_id = empresa_creditos.id
			  AND LOWER(COALESCE(cc.estado, 'activo')) = 'activo'
			  AND LOWER(COALESCE(cc.estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
			  AND COALESCE(cc.saldo_cuota, 0) > 0
			  AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(cc.fecha_vencimiento, '')), 1, 10), ''), ?) < ?
		) THEN 1 ELSE 0 END),
		SUM(CASE WHEN LOWER(COALESCE(estado_credito, 'activo')) = 'cerrado' THEN 1 ELSE 0 END),
		COALESCE(SUM(COALESCE(monto_aprobado, 0)), 0),
		COALESCE(SUM(COALESCE(saldo_actual, 0)), 0),
		COALESCE(SUM(COALESCE(saldo_disponible, 0)), 0)
	FROM empresa_creditos
	WHERE empresa_id = ?`
	args := []interface{}{today, today, today, today, today, today, empresaID}
	if !includeInactive {
		query += ` AND LOWER(COALESCE(estado, 'activo')) = 'activo'`
	}

	var out EmpresaCreditoCarteraResumen
	if err := queryRowSQLCompat(dbConn, query, args...).Scan(
		&out.TotalCreditos,
		&out.CreditosActivos,
		&out.CreditosVencidos,
		&out.CreditosMora,
		&out.CreditosCerrados,
		&out.MontoAprobado,
		&out.SaldoTotal,
		&out.CupoDisponible,
	); err != nil {
		return nil, err
	}
	out.MontoAprobado = creditoRound(out.MontoAprobado)
	out.SaldoTotal = creditoRound(out.SaldoTotal)
	out.CupoDisponible = creditoRound(out.CupoDisponible)
	return &out, nil
}

// GetEmpresaCreditosMoraDashboard retorna alertas proactivas y ranking de morosidad.
func GetEmpresaCreditosMoraDashboard(dbConn *sql.DB, empresaID int64, diasProximos, top int, includeInactive bool) (*EmpresaCreditosMoraDashboard, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	if diasProximos <= 0 {
		diasProximos = 7
	}
	if diasProximos > 365 {
		diasProximos = 365
	}
	if top <= 0 {
		top = 10
	}
	if top > 200 {
		top = 200
	}

	baseWhere := " AND COALESCE(saldo_actual, 0) > 0 AND LOWER(COALESCE(estado_credito, 'activo')) IN ('activo','suspendido','castigado')"
	if !includeInactive {
		baseWhere += " AND LOWER(COALESCE(estado, 'activo')) = 'activo'"
	}

	today := time.Now().In(time.Local).Format("2006-01-02")
	maxDate := time.Now().In(time.Local).AddDate(0, 0, diasProximos).Format("2006-01-02")

	proximosWhere := baseWhere + " AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) >= ? AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) <= ?"
	proximosRows, err := listEmpresaCreditosByWhere(
		dbConn,
		empresaID,
		proximosWhere,
		"COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), '9999-12-31') ASC, COALESCE(saldo_actual, 0) DESC, id DESC",
		[]interface{}{today, today, today, maxDate},
		top,
	)
	if err != nil {
		return nil, err
	}

	cuotaVencidaSQL := `EXISTS (
		SELECT 1
		FROM empresa_creditos_cuotas cc
		WHERE cc.empresa_id = empresa_creditos.empresa_id
		  AND cc.credito_id = empresa_creditos.id
		  AND LOWER(COALESCE(cc.estado, 'activo')) = 'activo'
		  AND LOWER(COALESCE(cc.estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
		  AND COALESCE(cc.saldo_cuota, 0) > 0
		  AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(cc.fecha_vencimiento, '')), 1, 10), ''), ?) < ?
	)`
	vencidosWhere := baseWhere + " AND (COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ? OR " + cuotaVencidaSQL + ")"
	vencidosRows, err := listEmpresaCreditosByWhere(
		dbConn,
		empresaID,
		vencidosWhere,
		"COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), '9999-12-31') ASC, COALESCE(saldo_actual, 0) DESC, id DESC",
		[]interface{}{today, today, today, today},
		top,
	)
	if err != nil {
		return nil, err
	}

	rankingWhere := baseWhere + " AND (COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ? OR " + cuotaVencidaSQL + ")"
	rankingRows, err := listEmpresaCreditosByWhere(
		dbConn,
		empresaID,
		rankingWhere,
		"COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), '0000-00-00') ASC, COALESCE(saldo_actual, 0) DESC, id DESC",
		[]interface{}{today, today, today, today},
		top,
	)
	if err != nil {
		return nil, err
	}

	countQuery := `SELECT
		COALESCE(SUM(CASE WHEN COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) >= ?
			AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) <= ? THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ? OR EXISTS (
			SELECT 1 FROM empresa_creditos_cuotas cc
			WHERE cc.empresa_id = empresa_creditos.empresa_id
			  AND cc.credito_id = empresa_creditos.id
			  AND LOWER(COALESCE(cc.estado, 'activo')) = 'activo'
			  AND LOWER(COALESCE(cc.estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
			  AND COALESCE(cc.saldo_cuota, 0) > 0
			  AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(cc.fecha_vencimiento, '')), 1, 10), ''), ?) < ?
		) THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) >= ?
			AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) <= ? THEN COALESCE(saldo_actual, 0) ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(fecha_vencimiento, '')), 1, 10), ''), ?) < ? OR EXISTS (
			SELECT 1 FROM empresa_creditos_cuotas cc
			WHERE cc.empresa_id = empresa_creditos.empresa_id
			  AND cc.credito_id = empresa_creditos.id
			  AND LOWER(COALESCE(cc.estado, 'activo')) = 'activo'
			  AND LOWER(COALESCE(cc.estado_cuota, 'pendiente')) IN ('pendiente','parcial','vencida')
			  AND COALESCE(cc.saldo_cuota, 0) > 0
			  AND COALESCE(NULLIF(SUBSTR(TRIM(COALESCE(cc.fecha_vencimiento, '')), 1, 10), ''), ?) < ?
		) THEN COALESCE(saldo_actual, 0) ELSE 0 END), 0)
	FROM empresa_creditos
	WHERE empresa_id = ?
	  AND COALESCE(saldo_actual, 0) > 0
	  AND LOWER(COALESCE(estado_credito, 'activo')) IN ('activo','suspendido','castigado')`
	countArgs := []interface{}{
		today, today, today, maxDate,
		today, today, today, today,
		today, today, today, maxDate,
		today, today, today, today,
		empresaID,
	}
	if !includeInactive {
		countQuery += " AND LOWER(COALESCE(estado, 'activo')) = 'activo'"
	}

	var totalProx int64
	var totalVenc int64
	var montoProx float64
	var montoVenc float64
	if err := dbConn.QueryRow(countQuery, countArgs...).Scan(&totalProx, &totalVenc, &montoProx, &montoVenc); err != nil {
		return nil, err
	}

	montoRanking := 0.0
	for _, item := range rankingRows {
		montoRanking += item.SaldoActual
	}

	out := &EmpresaCreditosMoraDashboard{
		DiasProximos:          diasProximos,
		Top:                   top,
		TotalProximosVencer:   totalProx,
		TotalVencidos:         totalVenc,
		TotalRankingMorosidad: int64(len(rankingRows)),
		MontoProximosVencer:   creditoRound(montoProx),
		MontoVencidos:         creditoRound(montoVenc),
		MontoRankingMorosidad: creditoRound(montoRanking),
		ProximosVencer:        proximosRows,
		Vencidos:              vencidosRows,
		RankingMorosidad:      rankingRows,
		GeneradoEn:            time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
	}

	return out, nil
}

func scanEmpresaCreditoWorkflow(scanner interface {
	Scan(dest ...interface{}) error
}) (*EmpresaCreditoWorkflow, error) {
	var row EmpresaCreditoWorkflow
	if err := scanner.Scan(
		&row.ID,
		&row.EmpresaID,
		&row.CreditoID,
		&row.WorkflowCodigo,
		&row.TipoSolicitud,
		&row.EstadoSolicitud,
		&row.MovimientoOrigenID,
		&row.MovimientoResultadoID,
		&row.NivelAprobacionActual,
		&row.NivelAprobacionRequerido,
		&row.AprobadoPor,
		&row.CodigoAprobacion,
		&row.FechaAprobacionFinal,
		&row.EjecutadoPor,
		&row.FechaEjecucion,
		&row.MotivoSolicitud,
		&row.MotivoRechazo,
		&row.PayloadJSON,
		&row.ResultadoJSON,
		&row.HistorialAprobacionesJSON,
		&row.FechaCreacion,
		&row.FechaActualizacion,
		&row.UsuarioCreador,
		&row.Estado,
		&row.Observaciones,
	); err != nil {
		return nil, err
	}
	row.TipoSolicitud = creditoWorkflowNormalizeTipo(row.TipoSolicitud)
	row.EstadoSolicitud = creditoWorkflowNormalizeEstado(row.EstadoSolicitud)
	row.Estado = creditoNormalizeRowEstado(row.Estado)
	if row.NivelAprobacionRequerido <= 0 {
		row.NivelAprobacionRequerido = 1
	}
	if row.NivelAprobacionActual < 0 {
		row.NivelAprobacionActual = 0
	}
	if row.NivelAprobacionActual > row.NivelAprobacionRequerido {
		row.NivelAprobacionActual = row.NivelAprobacionRequerido
	}
	if strings.TrimSpace(row.PayloadJSON) == "" {
		row.PayloadJSON = "{}"
	}
	if strings.TrimSpace(row.ResultadoJSON) == "" {
		row.ResultadoJSON = "{}"
	}
	if strings.TrimSpace(row.HistorialAprobacionesJSON) == "" {
		row.HistorialAprobacionesJSON = "[]"
	}
	return &row, nil
}

// GetEmpresaCreditoWorkflowByID obtiene una solicitud puntual de workflow de creditos.
func GetEmpresaCreditoWorkflowByID(dbConn *sql.DB, empresaID, workflowID int64) (*EmpresaCreditoWorkflow, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 || workflowID <= 0 {
		return nil, errors.New("empresa_id o workflow_id invalido")
	}

	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(credito_id, 0),
		COALESCE(workflow_codigo, ''),
		COALESCE(tipo_solicitud, ''),
		COALESCE(estado_solicitud, 'pendiente_aprobacion'),
		COALESCE(movimiento_origen_id, 0),
		COALESCE(movimiento_resultado_id, 0),
		COALESCE(nivel_aprobacion_actual, 0),
		COALESCE(nivel_aprobacion_requerido, 1),
		COALESCE(aprobado_por, ''),
		COALESCE(codigo_aprobacion, ''),
		COALESCE(fecha_aprobacion_final, ''),
		COALESCE(ejecutado_por, ''),
		COALESCE(fecha_ejecucion, ''),
		COALESCE(motivo_solicitud, ''),
		COALESCE(motivo_rechazo, ''),
		COALESCE(payload_json, '{}'),
		COALESCE(resultado_json, '{}'),
		COALESCE(historial_aprobaciones_json, '[]'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_workflow
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, workflowID)

	return scanEmpresaCreditoWorkflow(row)
}

func creditoWorkflowBuildWhere(filter EmpresaCreditoWorkflowFilter) (string, []interface{}) {
	clauses := make([]string, 0, 4)
	args := make([]interface{}, 0, 4)

	if !filter.IncludeInactive {
		clauses = append(clauses, "LOWER(COALESCE(estado, 'activo')) = 'activo'")
	}
	if filter.CreditoID > 0 {
		clauses = append(clauses, "COALESCE(credito_id, 0) = ?")
		args = append(args, filter.CreditoID)
	}
	if tipo := creditoWorkflowNormalizeTipo(filter.TipoSolicitud); tipo != "" {
		clauses = append(clauses, "LOWER(COALESCE(tipo_solicitud, '')) = ?")
		args = append(args, tipo)
	}
	if strings.TrimSpace(filter.EstadoSolicitud) != "" {
		estado := strings.ToLower(strings.TrimSpace(filter.EstadoSolicitud))
		clauses = append(clauses, "LOWER(COALESCE(estado_solicitud, 'pendiente_aprobacion')) = ?")
		args = append(args, estado)
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " AND " + strings.Join(clauses, " AND "), args
}

// ListEmpresaCreditoWorkflows lista solicitudes de workflow por empresa.
func ListEmpresaCreditoWorkflows(dbConn *sql.DB, empresaID int64, filter EmpresaCreditoWorkflowFilter) ([]EmpresaCreditoWorkflow, int64, error) {
	if dbConn == nil {
		return nil, 0, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, 0, errors.New("empresa_id invalido")
	}

	whereSQL, whereArgs := creditoWorkflowBuildWhere(filter)
	countQuery := "SELECT COUNT(1) FROM empresa_creditos_workflow WHERE empresa_id = ?" + whereSQL
	countArgs := append([]interface{}{empresaID}, whereArgs...)

	var total int64
	if err := dbConn.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit, offset := creditoWorkflowNormalizeLimitOffset(filter.Limit, filter.Offset)
	query := `SELECT
		id,
		empresa_id,
		COALESCE(credito_id, 0),
		COALESCE(workflow_codigo, ''),
		COALESCE(tipo_solicitud, ''),
		COALESCE(estado_solicitud, 'pendiente_aprobacion'),
		COALESCE(movimiento_origen_id, 0),
		COALESCE(movimiento_resultado_id, 0),
		COALESCE(nivel_aprobacion_actual, 0),
		COALESCE(nivel_aprobacion_requerido, 1),
		COALESCE(aprobado_por, ''),
		COALESCE(codigo_aprobacion, ''),
		COALESCE(fecha_aprobacion_final, ''),
		COALESCE(ejecutado_por, ''),
		COALESCE(fecha_ejecucion, ''),
		COALESCE(motivo_solicitud, ''),
		COALESCE(motivo_rechazo, ''),
		COALESCE(payload_json, '{}'),
		COALESCE(resultado_json, '{}'),
		COALESCE(historial_aprobaciones_json, '[]'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_workflow
	WHERE empresa_id = ?` + whereSQL + `
	ORDER BY id DESC
	LIMIT ? OFFSET ?`

	args := append([]interface{}{empresaID}, whereArgs...)
	args = append(args, limit, offset)

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]EmpresaCreditoWorkflow, 0)
	for rows.Next() {
		item, scanErr := scanEmpresaCreditoWorkflow(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		out = append(out, *item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return out, total, nil
}

func creditoValidateWorkflowSolicitudPrerequisitesTx(tx *sql.Tx, input EmpresaCreditoWorkflowSolicitudInput) error {
	if input.TipoSolicitud == "reverso_abono" {
		if input.MovimientoOrigenID <= 0 {
			return errors.New("movimiento_origen_id es obligatorio para reverso")
		}

		var movimientoID int64
		var tipoMovimiento string
		var estadoMovimiento string
		err := tx.QueryRow(`SELECT
			COALESCE(id, 0),
			COALESCE(tipo_movimiento, ''),
			COALESCE(estado, 'activo')
		FROM empresa_creditos_movimientos
		WHERE empresa_id = ? AND credito_id = ? AND id = ?
		LIMIT 1`, input.EmpresaID, input.CreditoID, input.MovimientoOrigenID).Scan(&movimientoID, &tipoMovimiento, &estadoMovimiento)
		if err != nil {
			if err == sql.ErrNoRows {
				return errors.New("movimiento_origen_id no encontrado para el credito")
			}
			return err
		}
		if creditoNormalizeMovimiento(tipoMovimiento) != "abono" {
			return errors.New("solo se permite reversar movimientos de tipo abono")
		}
		if creditoNormalizeRowEstado(estadoMovimiento) != "activo" {
			return errors.New("el movimiento origen esta inactivo y no se puede reversar")
		}

		var existentes int
		err = tx.QueryRow(`SELECT COUNT(1)
		FROM empresa_creditos_workflow
		WHERE empresa_id = ?
		  AND credito_id = ?
		  AND movimiento_origen_id = ?
		  AND LOWER(COALESCE(tipo_solicitud, '')) = 'reverso_abono'
		  AND LOWER(COALESCE(estado_solicitud, 'pendiente_aprobacion')) IN ('pendiente_aprobacion','aprobada','ejecutada')
		  AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, input.EmpresaID, input.CreditoID, input.MovimientoOrigenID).Scan(&existentes)
		if err != nil {
			return err
		}
		if existentes > 0 {
			return errors.New("ya existe una solicitud de reverso para el movimiento indicado")
		}
	}

	if input.TipoSolicitud == "refinanciacion" {
		credito, err := GetEmpresaCreditoByIDTx(tx, input.EmpresaID, input.CreditoID)
		if err != nil {
			return err
		}
		if credito.SaldoActual <= 0 {
			return errors.New("el credito no tiene saldo pendiente para refinanciacion")
		}
	}

	return nil
}

// CreateEmpresaCreditoWorkflowSolicitud registra solicitud de reverso/refinanciacion pendiente de aprobacion.
func CreateEmpresaCreditoWorkflowSolicitud(dbConn *sql.DB, input EmpresaCreditoWorkflowSolicitudInput) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if input.EmpresaID <= 0 || input.CreditoID <= 0 {
		return 0, errors.New("empresa_id o credito_id invalido")
	}

	input.TipoSolicitud = creditoWorkflowNormalizeTipo(input.TipoSolicitud)
	if input.TipoSolicitud == "" {
		return 0, errors.New("tipo_solicitud invalido")
	}
	if input.NivelAprobacionRequerido <= 0 {
		input.NivelAprobacionRequerido = 1
	}
	if input.NivelAprobacionRequerido > 10 {
		input.NivelAprobacionRequerido = 10
	}

	input.UsuarioCreador = strings.TrimSpace(input.UsuarioCreador)
	if input.UsuarioCreador == "" {
		input.UsuarioCreador = "sistema"
	}
	input.MotivoSolicitud = strings.TrimSpace(input.MotivoSolicitud)
	if input.MotivoSolicitud == "" {
		input.MotivoSolicitud = "solicitud operativa"
	}

	payloadJSON := strings.TrimSpace(input.PayloadJSON)
	if payloadJSON == "" {
		payloadJSON = "{}"
	}
	if !json.Valid([]byte(payloadJSON)) {
		return 0, errors.New("payload_json invalido")
	}

	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = GetEmpresaCreditoByIDTx(tx, input.EmpresaID, input.CreditoID); err != nil {
		if err == sql.ErrNoRows {
			return 0, errors.New("credito no encontrado")
		}
		return 0, err
	}
	if err = creditoValidateWorkflowSolicitudPrerequisitesTx(tx, input); err != nil {
		return 0, err
	}

	id, err := insertTxSQLCompat(tx, `INSERT INTO empresa_creditos_workflow (
		empresa_id,
		credito_id,
		workflow_codigo,
		tipo_solicitud,
		estado_solicitud,
		movimiento_origen_id,
		nivel_aprobacion_actual,
		nivel_aprobacion_requerido,
		motivo_solicitud,
		payload_json,
		resultado_json,
		historial_aprobaciones_json,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (
		?, ?, '', ?, 'pendiente_aprobacion', ?, 0, ?, ?, ?, '{}', '[]', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?
	)`,
		input.EmpresaID,
		input.CreditoID,
		input.TipoSolicitud,
		input.MovimientoOrigenID,
		input.NivelAprobacionRequerido,
		input.MotivoSolicitud,
		payloadJSON,
		input.UsuarioCreador,
		strings.TrimSpace(input.Observaciones),
	)
	if err != nil {
		return 0, err
	}
	codigo := creditoWorkflowDefaultCodigo(input.EmpresaID, id, input.TipoSolicitud)
	if _, err = tx.Exec(`UPDATE empresa_creditos_workflow SET workflow_codigo = ?, fecha_actualizacion = CURRENT_TIMESTAMP WHERE empresa_id = ? AND id = ?`, codigo, input.EmpresaID, id); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// GetEmpresaCreditoByIDTx obtiene un credito dentro de una transaccion.
func GetEmpresaCreditoByIDTx(tx *sql.Tx, empresaID, creditoID int64) (*EmpresaCredito, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	row := tx.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(codigo, ''),
		COALESCE(cliente_id, 0),
		COALESCE(cliente_nombre, ''),
		COALESCE(tipo_credito, 'cuotas'),
		COALESCE(monto_aprobado, 0),
		COALESCE(cupo_credito, 0),
		COALESCE(saldo_actual, 0),
		COALESCE(saldo_disponible, 0),
		COALESCE(tasa_interes, 0),
		COALESCE(tasa_mora, 0),
		COALESCE(periodicidad_cuota, 'mensual'),
		COALESCE(valor_cuota_pactada, 0),
		COALESCE(omitir_domingos, 0),
		COALESCE(plazo_dias, 0),
		COALESCE(plazo_cuotas, 0),
		COALESCE(fecha_inicio, ''),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(fecha_ultimo_pago, ''),
		COALESCE(dias_mora, 0),
		COALESCE(clasificacion_cartera, 'al_dia'),
		COALESCE(bloqueo_automatico_mora, 1),
		COALESCE(venta_origen_id, 0),
		COALESCE(documento_origen, ''),
		COALESCE(estado_credito, 'activo'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, creditoID)
	return scanEmpresaCredito(row)
}

func getEmpresaCreditoWorkflowByIDTx(tx *sql.Tx, empresaID, workflowID int64) (*EmpresaCreditoWorkflow, error) {
	if tx == nil {
		return nil, errors.New("tx is nil")
	}
	row := tx.QueryRow(`SELECT
		id,
		empresa_id,
		COALESCE(credito_id, 0),
		COALESCE(workflow_codigo, ''),
		COALESCE(tipo_solicitud, ''),
		COALESCE(estado_solicitud, 'pendiente_aprobacion'),
		COALESCE(movimiento_origen_id, 0),
		COALESCE(movimiento_resultado_id, 0),
		COALESCE(nivel_aprobacion_actual, 0),
		COALESCE(nivel_aprobacion_requerido, 1),
		COALESCE(aprobado_por, ''),
		COALESCE(codigo_aprobacion, ''),
		COALESCE(fecha_aprobacion_final, ''),
		COALESCE(ejecutado_por, ''),
		COALESCE(fecha_ejecucion, ''),
		COALESCE(motivo_solicitud, ''),
		COALESCE(motivo_rechazo, ''),
		COALESCE(payload_json, '{}'),
		COALESCE(resultado_json, '{}'),
		COALESCE(historial_aprobaciones_json, '[]'),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_creditos_workflow
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, workflowID)
	return scanEmpresaCreditoWorkflow(row)
}

func creditoParseHistorialAprobaciones(raw string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func creditoAppendHistorialEvento(raw string, item map[string]interface{}) string {
	historial := creditoParseHistorialAprobaciones(raw)
	historial = append(historial, item)
	return creditoMarshalJSON(historial, "[]")
}

func creditoResolveWorkflowActor(primary, fallback string) string {
	actor := strings.TrimSpace(primary)
	if actor == "" {
		actor = strings.TrimSpace(fallback)
	}
	if actor == "" {
		return "sistema"
	}
	return actor
}

func executeEmpresaCreditoWorkflowTx(tx *sql.Tx, workflow *EmpresaCreditoWorkflow, actor string) (int64, map[string]interface{}, error) {
	if workflow == nil {
		return 0, nil, errors.New("workflow invalido")
	}
	switch workflow.TipoSolicitud {
	case "reverso_abono":
		return executeEmpresaCreditoReversoWorkflowTx(tx, workflow, actor)
	case "refinanciacion":
		return executeEmpresaCreditoRefinanciacionWorkflowTx(tx, workflow, actor)
	default:
		return 0, nil, errors.New("tipo_solicitud no soportado")
	}
}

func executeEmpresaCreditoReversoWorkflowTx(tx *sql.Tx, workflow *EmpresaCreditoWorkflow, actor string) (int64, map[string]interface{}, error) {
	if workflow.MovimientoOrigenID <= 0 {
		return 0, nil, errors.New("movimiento_origen_id invalido para reverso")
	}

	var movID int64
	var tipoMov string
	var montoMov float64
	var capitalMov float64
	var interesMov float64
	var moraMov float64
	var metodoPago string
	var referenciaPago string
	var comprobante string
	var estadoMov string
	if err := tx.QueryRow(`SELECT
		COALESCE(id, 0),
		COALESCE(tipo_movimiento, ''),
		COALESCE(monto, 0),
		COALESCE(capital_aplicado, 0),
		COALESCE(interes_aplicado, 0),
		COALESCE(mora_aplicada, 0),
		COALESCE(metodo_pago, ''),
		COALESCE(referencia_pago, ''),
		COALESCE(comprobante, ''),
		COALESCE(estado, 'activo')
	FROM empresa_creditos_movimientos
	WHERE empresa_id = ? AND credito_id = ? AND id = ?
	LIMIT 1`, workflow.EmpresaID, workflow.CreditoID, workflow.MovimientoOrigenID).Scan(&movID, &tipoMov, &montoMov, &capitalMov, &interesMov, &moraMov, &metodoPago, &referenciaPago, &comprobante, &estadoMov); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, errors.New("movimiento origen no encontrado")
		}
		return 0, nil, err
	}
	if creditoNormalizeMovimiento(tipoMov) != "abono" {
		return 0, nil, errors.New("solo se puede reversar un movimiento tipo abono")
	}
	if creditoNormalizeRowEstado(estadoMov) != "activo" {
		return 0, nil, errors.New("movimiento origen inactivo")
	}

	var reversosExistentes int
	if err := tx.QueryRow(`SELECT COUNT(1)
	FROM empresa_creditos_movimientos
	WHERE empresa_id = ?
	  AND credito_id = ?
	  AND LOWER(COALESCE(tipo_movimiento, '')) = 'reverso'
	  AND LOWER(COALESCE(observaciones, '')) LIKE LOWER(?) ESCAPE '!'`, workflow.EmpresaID, workflow.CreditoID, creditoLikePattern(fmt.Sprintf("movimiento_origen=%d", workflow.MovimientoOrigenID))).Scan(&reversosExistentes); err != nil {
		return 0, nil, err
	}
	if reversosExistentes > 0 {
		return 0, nil, errors.New("el movimiento origen ya fue reversado")
	}

	credito, err := GetEmpresaCreditoByIDTx(tx, workflow.EmpresaID, workflow.CreditoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, errors.New("credito no encontrado")
		}
		return 0, nil, err
	}

	montoReverso := creditoRound(creditoMax(montoMov, 0))
	if montoReverso <= 0 {
		return 0, nil, errors.New("monto invalido para reverso")
	}

	restante := montoReverso
	cuotasAfectadas := 0
	rows, err := tx.Query(`SELECT
		id,
		COALESCE(valor_cuota, 0),
		COALESCE(valor_pagado, 0),
		COALESCE(fecha_vencimiento, ''),
		COALESCE(estado_cuota, 'pendiente')
	FROM empresa_creditos_cuotas
	WHERE empresa_id = ?
	  AND credito_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND COALESCE(valor_pagado, 0) > 0
	ORDER BY numero_cuota DESC, id DESC`, workflow.EmpresaID, workflow.CreditoID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()

	for rows.Next() && restante > 0 {
		var cuotaID int64
		var valorCuota float64
		var valorPagado float64
		var fechaVencimiento string
		var estadoCuota string
		if err := rows.Scan(&cuotaID, &valorCuota, &valorPagado, &fechaVencimiento, &estadoCuota); err != nil {
			return 0, nil, err
		}
		if valorPagado <= 0 {
			continue
		}
		revertir := creditoMin(restante, valorPagado)
		nuevoPagado := creditoRound(creditoMax(valorPagado-revertir, 0))
		nuevoSaldo := creditoRound(creditoMax(valorCuota-nuevoPagado, 0))
		nuevoEstado := "parcial"
		if nuevoSaldo <= 0 {
			nuevoEstado = "pagada"
		} else if nuevoPagado <= 0 {
			nuevoEstado = "pendiente"
		}
		if nuevoEstado != "pagada" && creditoDaysMora(fechaVencimiento, nuevoSaldo) > 0 {
			nuevoEstado = "vencida"
		}

		if _, err := tx.Exec(`UPDATE empresa_creditos_cuotas SET
			valor_pagado = ?,
			saldo_cuota = ?,
			estado_cuota = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND credito_id = ? AND id = ?`, nuevoPagado, nuevoSaldo, nuevoEstado, workflow.EmpresaID, workflow.CreditoID, cuotaID); err != nil {
			return 0, nil, err
		}

		restante = creditoRound(creditoMax(restante-revertir, 0))
		cuotasAfectadas++
	}
	if err := rows.Err(); err != nil {
		return 0, nil, err
	}

	saldoNuevo := creditoRound(creditoMin(credito.CupoCredito, credito.SaldoActual+montoReverso))
	estadoNuevo := credito.EstadoCredito
	if estadoNuevo == "cerrado" && saldoNuevo > 0 {
		estadoNuevo = "activo"
	}
	clasificacion := creditoResolveClasificacion(estadoNuevo, credito.FechaVencimiento, saldoNuevo)
	diasMora := creditoDaysMora(credito.FechaVencimiento, saldoNuevo)

	if _, err := tx.Exec(`UPDATE empresa_creditos SET
		saldo_actual = ?,
		saldo_disponible = ?,
		estado_credito = ?,
		clasificacion_cartera = ?,
		dias_mora = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`, saldoNuevo, creditoRound(creditoMax(credito.CupoCredito-saldoNuevo, 0)), estadoNuevo, clasificacion, diasMora, workflow.EmpresaID, workflow.CreditoID); err != nil {
		return 0, nil, err
	}

	ratio := 1.0
	if montoMov > 0 {
		ratio = creditoMin(1, montoReverso/montoMov)
	}
	capRev := -creditoRound(capitalMov * ratio)
	intRev := -creditoRound(interesMov * ratio)
	moraRev := -creditoRound(moraMov * ratio)

	nuevoMovimientoID, err := insertTxSQLCompat(tx, `INSERT INTO empresa_creditos_movimientos (
		empresa_id,
		credito_id,
		cuota_id,
		tipo_movimiento,
		monto,
		capital_aplicado,
		interes_aplicado,
		mora_aplicada,
		metodo_pago,
		referencia_pago,
		comprobante,
		aplicado_automatico,
		fecha_movimiento,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (
		?, ?, 0, 'reverso', ?, ?, ?, ?, ?, ?, ?, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?
	)`,
		workflow.EmpresaID,
		workflow.CreditoID,
		montoReverso,
		capRev,
		intRev,
		moraRev,
		metodoPago,
		"REVERSO-"+referenciaPago,
		comprobante,
		actor,
		fmt.Sprintf("workflow_id=%d movimiento_origen=%d", workflow.ID, workflow.MovimientoOrigenID),
	)
	if err != nil {
		return 0, nil, err
	}

	resultado := map[string]interface{}{
		"tipo_ejecucion":        "reverso_abono",
		"movimiento_origen_id":  workflow.MovimientoOrigenID,
		"movimiento_reverso_id": nuevoMovimientoID,
		"monto_reversado":       montoReverso,
		"cuotas_afectadas":      cuotasAfectadas,
		"saldo_credito_nuevo":   saldoNuevo,
	}
	return nuevoMovimientoID, resultado, nil
}

func executeEmpresaCreditoRefinanciacionWorkflowTx(tx *sql.Tx, workflow *EmpresaCreditoWorkflow, actor string) (int64, map[string]interface{}, error) {
	payload := creditoParseJSONMap(workflow.PayloadJSON)
	nuevoPlazo := int(payloadFloat(payload, "nuevo_plazo_cuotas", "plazo_cuotas"))
	if nuevoPlazo <= 0 {
		nuevoPlazo = 12
	}
	if nuevoPlazo > 600 {
		nuevoPlazo = 600
	}
	nuevaTasaInteres := creditoRound(payloadFloat(payload, "nueva_tasa_interes", "tasa_interes"))
	if nuevaTasaInteres < 0 {
		nuevaTasaInteres = 0
	}
	nuevaTasaMora := creditoRound(payloadFloat(payload, "nueva_tasa_mora", "tasa_mora"))
	if nuevaTasaMora < 0 {
		nuevaTasaMora = 0
	}

	nuevaFechaInicio := strings.TrimSpace(payloadString(payload, "nueva_fecha_inicio", "fecha_inicio"))
	nuevaFechaVencimiento := strings.TrimSpace(payloadString(payload, "nueva_fecha_vencimiento", "fecha_vencimiento"))
	nuevoTipoCredito := creditoNormalizeTipo(payloadString(payload, "nuevo_tipo_credito", "tipo_credito"))
	if nuevoTipoCredito == "" {
		nuevoTipoCredito = "cuotas"
	}

	credito, err := GetEmpresaCreditoByIDTx(tx, workflow.EmpresaID, workflow.CreditoID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil, errors.New("credito no encontrado")
		}
		return 0, nil, err
	}
	if credito.SaldoActual <= 0 {
		return 0, nil, errors.New("el credito no tiene saldo pendiente para refinanciar")
	}

	if strings.TrimSpace(nuevaFechaInicio) == "" {
		nuevaFechaInicio = time.Now().In(time.Local).Format("2006-01-02")
	}
	if strings.TrimSpace(nuevaFechaVencimiento) == "" {
		if dt, ok := creditoParseDate(nuevaFechaInicio); ok {
			nuevaFechaVencimiento = dt.AddDate(0, nuevoPlazo, 0).Format("2006-01-02")
		}
	}
	if strings.TrimSpace(nuevaFechaVencimiento) == "" {
		nuevaFechaVencimiento = time.Now().In(time.Local).AddDate(0, nuevoPlazo, 0).Format("2006-01-02")
	}

	resInactivar, err := tx.Exec(`UPDATE empresa_creditos_cuotas SET
		estado = 'inactivo',
		estado_cuota = 'anulada',
		fecha_actualizacion = CURRENT_TIMESTAMP,
		observaciones = CASE
			WHEN TRIM(COALESCE(observaciones,'')) = '' THEN ?
			ELSE observaciones || ' | ' || ?
		END
	WHERE empresa_id = ?
	  AND credito_id = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'
	  AND COALESCE(saldo_cuota, 0) > 0`,
		fmt.Sprintf("refinanciacion workflow_id=%d", workflow.ID),
		fmt.Sprintf("refinanciacion workflow_id=%d", workflow.ID),
		workflow.EmpresaID,
		workflow.CreditoID,
	)
	if err != nil {
		return 0, nil, err
	}
	cuotasInactivadas, _ := resInactivar.RowsAffected()

	ultimoNumeroCuota := 0
	if err := tx.QueryRow(`SELECT COALESCE(MAX(numero_cuota), 0)
	FROM empresa_creditos_cuotas
	WHERE empresa_id = ? AND credito_id = ?`, workflow.EmpresaID, workflow.CreditoID).Scan(&ultimoNumeroCuota); err != nil {
		return 0, nil, err
	}

	payloadCuotas := EmpresaCredito{
		EmpresaID:        workflow.EmpresaID,
		MontoAprobado:    credito.SaldoActual,
		TipoCredito:      nuevoTipoCredito,
		TasaInteres:      nuevaTasaInteres,
		PlazoCuotas:      nuevoPlazo,
		FechaInicio:      nuevaFechaInicio,
		FechaVencimiento: nuevaFechaVencimiento,
		UsuarioCreador:   actor,
	}
	if err := creditoGenerateCuotasTxWithStart(tx, workflow.EmpresaID, workflow.CreditoID, payloadCuotas, ultimoNumeroCuota+1); err != nil {
		return 0, nil, err
	}

	estadoNuevo := credito.EstadoCredito
	if estadoNuevo == "cerrado" || estadoNuevo == "castigado" {
		estadoNuevo = "activo"
	}
	clasificacion := creditoResolveClasificacion(estadoNuevo, nuevaFechaVencimiento, credito.SaldoActual)
	diasMora := creditoDaysMora(nuevaFechaVencimiento, credito.SaldoActual)

	if _, err := tx.Exec(`UPDATE empresa_creditos SET
		tipo_credito = ?,
		tasa_interes = ?,
		tasa_mora = ?,
		plazo_cuotas = ?,
		fecha_inicio = ?,
		fecha_vencimiento = ?,
		estado_credito = ?,
		clasificacion_cartera = ?,
		dias_mora = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`,
		nuevoTipoCredito,
		nuevaTasaInteres,
		nuevaTasaMora,
		nuevoPlazo,
		nuevaFechaInicio,
		nuevaFechaVencimiento,
		estadoNuevo,
		clasificacion,
		diasMora,
		workflow.EmpresaID,
		workflow.CreditoID,
	); err != nil {
		return 0, nil, err
	}

	nuevoMovimientoID, err := insertTxSQLCompat(tx, `INSERT INTO empresa_creditos_movimientos (
		empresa_id,
		credito_id,
		cuota_id,
		tipo_movimiento,
		monto,
		capital_aplicado,
		interes_aplicado,
		mora_aplicada,
		metodo_pago,
		referencia_pago,
		comprobante,
		aplicado_automatico,
		fecha_movimiento,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (
		?, ?, 0, 'refinanciacion', ?, 0, 0, 0, 'refinanciacion', ?, '', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, 'activo', ?
	)`,
		workflow.EmpresaID,
		workflow.CreditoID,
		credito.SaldoActual,
		fmt.Sprintf("WF-REF-%d", workflow.ID),
		actor,
		fmt.Sprintf("workflow_id=%d refinanciacion", workflow.ID),
	)
	if err != nil {
		return 0, nil, err
	}

	resultado := map[string]interface{}{
		"tipo_ejecucion":          "refinanciacion",
		"movimiento_refin_id":     nuevoMovimientoID,
		"saldo_refinanciado":      credito.SaldoActual,
		"cuotas_inactivadas":      cuotasInactivadas,
		"cuotas_nuevas_generadas": nuevoPlazo,
		"nueva_tasa_interes":      nuevaTasaInteres,
		"nueva_tasa_mora":         nuevaTasaMora,
		"nuevo_plazo_cuotas":      nuevoPlazo,
		"nueva_fecha_vencimiento": nuevaFechaVencimiento,
	}
	return nuevoMovimientoID, resultado, nil
}

// AprobarEmpresaCreditoWorkflow aprueba un nivel de workflow y ejecuta al completar niveles.
func AprobarEmpresaCreditoWorkflow(dbConn *sql.DB, input EmpresaCreditoWorkflowAprobacionInput) (*EmpresaCreditoWorkflow, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if input.EmpresaID <= 0 || input.WorkflowID <= 0 {
		return nil, errors.New("empresa_id o workflow_id invalido")
	}
	input.AprobadoPor = strings.TrimSpace(input.AprobadoPor)
	if input.AprobadoPor == "" {
		return nil, errors.New("aprobado_por es obligatorio")
	}
	input.CodigoAprobacion = strings.TrimSpace(input.CodigoAprobacion)
	if input.CodigoAprobacion == "" {
		return nil, errors.New("codigo_aprobacion es obligatorio")
	}
	actor := creditoResolveWorkflowActor(input.EjecutadoPor, input.UsuarioCreador)

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	workflow, err := getEmpresaCreditoWorkflowByIDTx(tx, input.EmpresaID, input.WorkflowID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("workflow no encontrado")
		}
		return nil, err
	}
	if workflow.EstadoSolicitud != "pendiente_aprobacion" {
		return nil, fmt.Errorf("workflow no aprobable en estado %s", workflow.EstadoSolicitud)
	}

	nivelNuevo := workflow.NivelAprobacionActual + 1
	if nivelNuevo > workflow.NivelAprobacionRequerido {
		nivelNuevo = workflow.NivelAprobacionRequerido
	}

	historialNuevo := creditoAppendHistorialEvento(workflow.HistorialAprobacionesJSON, map[string]interface{}{
		"accion":            "aprobar",
		"nivel_anterior":    workflow.NivelAprobacionActual,
		"nivel_nuevo":       nivelNuevo,
		"nivel_requerido":   workflow.NivelAprobacionRequerido,
		"aprobado_por":      input.AprobadoPor,
		"codigo_aprobacion": input.CodigoAprobacion,
		"motivo_aprobacion": strings.TrimSpace(input.MotivoAprobacion),
		"actor":             actor,
		"fecha":             time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
	})

	estadoSolicitud := "aprobada"
	movimientoResultadoID := int64(0)
	resultadoJSON := workflow.ResultadoJSON
	fechaAprobFinal := time.Now().In(time.Local).Format("2006-01-02 15:04:05")
	fechaEjecucion := ""
	ejecutadoPor := ""

	if nivelNuevo < workflow.NivelAprobacionRequerido {
		estadoSolicitud = "pendiente_aprobacion"
		fechaAprobFinal = ""
		if _, err := tx.Exec(`UPDATE empresa_creditos_workflow SET
			nivel_aprobacion_actual = ?,
			estado_solicitud = ?,
			aprobado_por = ?,
			codigo_aprobacion = ?,
			historial_aprobaciones_json = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`, nivelNuevo, estadoSolicitud, input.AprobadoPor, input.CodigoAprobacion, historialNuevo, input.EmpresaID, input.WorkflowID); err != nil {
			return nil, err
		}
	} else {
		movID, resultado, execErr := executeEmpresaCreditoWorkflowTx(tx, workflow, actor)
		if execErr != nil {
			return nil, execErr
		}
		movimientoResultadoID = movID
		estadoSolicitud = "ejecutada"
		resultadoJSON = creditoMarshalJSON(resultado, "{}")
		fechaEjecucion = time.Now().In(time.Local).Format("2006-01-02 15:04:05")
		ejecutadoPor = actor

		if _, err := tx.Exec(`UPDATE empresa_creditos_workflow SET
			nivel_aprobacion_actual = ?,
			estado_solicitud = ?,
			aprobado_por = ?,
			codigo_aprobacion = ?,
			fecha_aprobacion_final = ?,
			ejecutado_por = ?,
			fecha_ejecucion = ?,
			movimiento_resultado_id = ?,
			resultado_json = ?,
			historial_aprobaciones_json = ?,
			fecha_actualizacion = CURRENT_TIMESTAMP
		WHERE empresa_id = ? AND id = ?`, nivelNuevo, estadoSolicitud, input.AprobadoPor, input.CodigoAprobacion, fechaAprobFinal, ejecutadoPor, fechaEjecucion, movimientoResultadoID, resultadoJSON, historialNuevo, input.EmpresaID, input.WorkflowID); err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetEmpresaCreditoWorkflowByID(dbConn, input.EmpresaID, input.WorkflowID)
}

// RechazarEmpresaCreditoWorkflow marca solicitud como rechazada y conserva trazabilidad.
func RechazarEmpresaCreditoWorkflow(dbConn *sql.DB, input EmpresaCreditoWorkflowAprobacionInput) (*EmpresaCreditoWorkflow, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if input.EmpresaID <= 0 || input.WorkflowID <= 0 {
		return nil, errors.New("empresa_id o workflow_id invalido")
	}
	input.AprobadoPor = strings.TrimSpace(input.AprobadoPor)
	if input.AprobadoPor == "" {
		return nil, errors.New("aprobado_por es obligatorio")
	}
	input.CodigoAprobacion = strings.TrimSpace(input.CodigoAprobacion)
	if input.CodigoAprobacion == "" {
		return nil, errors.New("codigo_aprobacion es obligatorio")
	}
	if strings.TrimSpace(input.MotivoRechazo) == "" {
		return nil, errors.New("motivo_rechazo es obligatorio")
	}
	actor := creditoResolveWorkflowActor(input.EjecutadoPor, input.UsuarioCreador)

	tx, err := dbConn.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	workflow, err := getEmpresaCreditoWorkflowByIDTx(tx, input.EmpresaID, input.WorkflowID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("workflow no encontrado")
		}
		return nil, err
	}
	if workflow.EstadoSolicitud != "pendiente_aprobacion" && workflow.EstadoSolicitud != "aprobada" {
		return nil, fmt.Errorf("workflow no rechazable en estado %s", workflow.EstadoSolicitud)
	}

	historialNuevo := creditoAppendHistorialEvento(workflow.HistorialAprobacionesJSON, map[string]interface{}{
		"accion":            "rechazar",
		"nivel_anterior":    workflow.NivelAprobacionActual,
		"nivel_requerido":   workflow.NivelAprobacionRequerido,
		"aprobado_por":      input.AprobadoPor,
		"codigo_aprobacion": input.CodigoAprobacion,
		"motivo_rechazo":    strings.TrimSpace(input.MotivoRechazo),
		"actor":             actor,
		"fecha":             time.Now().In(time.Local).Format("2006-01-02 15:04:05"),
	})

	if _, err := tx.Exec(`UPDATE empresa_creditos_workflow SET
		estado_solicitud = 'rechazada',
		aprobado_por = ?,
		codigo_aprobacion = ?,
		motivo_rechazo = ?,
		historial_aprobaciones_json = ?,
		fecha_actualizacion = CURRENT_TIMESTAMP
	WHERE empresa_id = ? AND id = ?`, input.AprobadoPor, input.CodigoAprobacion, strings.TrimSpace(input.MotivoRechazo), historialNuevo, input.EmpresaID, input.WorkflowID); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return GetEmpresaCreditoWorkflowByID(dbConn, input.EmpresaID, input.WorkflowID)
}

func payloadFloat(payload map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		v, ok := payload[key]
		if !ok || v == nil {
			continue
		}
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s == "" {
			continue
		}
		s = strings.ReplaceAll(s, ",", "")
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return 0
}
