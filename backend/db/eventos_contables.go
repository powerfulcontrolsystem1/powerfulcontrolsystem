package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

// EmpresaEventoContable representa un evento de negocio listo para integracion contable.
type EmpresaEventoContable struct {
	ID                    int64   `json:"id"`
	EmpresaID             int64   `json:"empresa_id"`
	Modulo                string  `json:"modulo"`
	Evento                string  `json:"evento"`
	Entidad               string  `json:"entidad"`
	EntidadID             int64   `json:"entidad_id"`
	DocumentoTipo         string  `json:"documento_tipo"`
	DocumentoCodigo       string  `json:"documento_codigo"`
	PeriodoContable       string  `json:"periodo_contable"`
	MontoTotal            float64 `json:"monto_total"`
	Moneda                string  `json:"moneda"`
	PayloadJSON           string  `json:"payload_json"`
	Origen                string  `json:"origen"`
	FechaEvento           string  `json:"fecha_evento"`
	Procesado             bool    `json:"procesado"`
	FechaProcesado        string  `json:"fecha_procesado"`
	IntentosProcesamiento int64   `json:"intentos_procesamiento"`
	FechaUltimoIntento    string  `json:"fecha_ultimo_intento"`
	ErrorProcesamiento    string  `json:"error_procesamiento"`
	AsientoContableID     int64   `json:"asiento_contable_id"`
	FechaCreacion         string  `json:"fecha_creacion"`
	FechaActualizacion    string  `json:"fecha_actualizacion"`
	UsuarioCreador        string  `json:"usuario_creador"`
	Estado                string  `json:"estado"`
	Observaciones         string  `json:"observaciones"`
}

// EmpresaAsientoContableLinea representa una línea del asiento (débito/crédito).
type EmpresaAsientoContableLinea struct {
	Cuenta      string  `json:"cuenta"`
	Descripcion string  `json:"descripcion"`
	Debito      float64 `json:"debito"`
	Credito     float64 `json:"credito"`
}

// EmpresaAsientoContable representa un asiento canónico derivado de un evento contable.
type EmpresaAsientoContable struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	EventoContableID   int64   `json:"evento_contable_id"`
	Modulo             string  `json:"modulo"`
	Evento             string  `json:"evento"`
	FechaAsiento       string  `json:"fecha_asiento"`
	PeriodoContable    string  `json:"periodo_contable"`
	DocumentoTipo      string  `json:"documento_tipo"`
	DocumentoCodigo    string  `json:"documento_codigo"`
	Moneda             string  `json:"moneda"`
	TotalDebito        float64 `json:"total_debito"`
	TotalCredito       float64 `json:"total_credito"`
	Diferencia         float64 `json:"diferencia"`
	LineasJSON         string  `json:"lineas_json"`
	HashIdempotencia   string  `json:"hash_idempotencia"`
	PayloadOrigenJSON  string  `json:"payload_origen_json"`
	FechaProcesado     string  `json:"fecha_procesado"`
	ProcesadoPor       string  `json:"procesado_por"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
	UsuarioCreador     string  `json:"usuario_creador"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones"`
}

// EmpresaAsientoContableFilter permite consultar asientos por empresa.
type EmpresaAsientoContableFilter struct {
	Modulo          string
	Evento          string
	PeriodoContable string
	Desde           string
	Hasta           string
	IncludeInactive bool
	Limit           int
}

// EmpresaProcesoAsientosResultado resume la ejecución del procesamiento por lotes.
type EmpresaProcesoAsientosResultado struct {
	EmpresaID          int64    `json:"empresa_id"`
	EventosRevisados   int      `json:"eventos_revisados"`
	EventosProcesados  int      `json:"eventos_procesados"`
	AsientosCreados    int      `json:"asientos_creados"`
	AsientosExistentes int      `json:"asientos_existentes"`
	Fallidos           int      `json:"fallidos"`
	Errores            []string `json:"errores"`
}

// EmpresaEventoContableFilter permite consultar eventos contables por empresa.
type EmpresaEventoContableFilter struct {
	Modulo          string
	Evento          string
	PeriodoContable string
	IncludeInactive bool
	Limit           int
}

const (
	defaultAsientosWorkerInterval  = 15 * time.Minute
	defaultAsientosWorkerBatchSize = 100
	defaultAsientosWorkerRetries   = 5
	maxAsientosWorkerBatchSize     = 500
	maxAsientosWorkerRetries       = 50
)

// EmpresaAsientosWorkerResumen resume una corrida global del worker de asientos.
type EmpresaAsientosWorkerResumen struct {
	EmpresasConPendientes int      `json:"empresas_con_pendientes"`
	EmpresasProcesadas    int      `json:"empresas_procesadas"`
	EventosRevisados      int      `json:"eventos_revisados"`
	EventosProcesados     int      `json:"eventos_procesados"`
	AsientosCreados       int      `json:"asientos_creados"`
	AsientosExistentes    int      `json:"asientos_existentes"`
	Fallidos              int      `json:"fallidos"`
	Errores               []string `json:"errores"`
}

// EmpresaConciliacionContableFilter permite consultar conciliacion por periodo.
type EmpresaConciliacionContableFilter struct {
	Desde           string
	Hasta           string
	PeriodoContable string
	IncludeInactive bool
	Limit           int
}

// EmpresaConciliacionContablePeriodo representa un periodo conciliado entre eventos y asientos.
type EmpresaConciliacionContablePeriodo struct {
	PeriodoContable       string  `json:"periodo_contable"`
	EventosTotal          int64   `json:"eventos_total"`
	EventosProcesados     int64   `json:"eventos_procesados"`
	EventosPendientes     int64   `json:"eventos_pendientes"`
	EventosConError       int64   `json:"eventos_con_error"`
	EventosMontoProcesado float64 `json:"eventos_monto_procesado"`
	AsientosTotal         int64   `json:"asientos_total"`
	AsientosMontoTotal    float64 `json:"asientos_monto_total"`
	DesfaseEventosAsiento int64   `json:"desfase_eventos_asientos"`
	DesfaseMonto          float64 `json:"desfase_monto"`
	UltimoEvento          string  `json:"ultimo_evento"`
	UltimoAsiento         string  `json:"ultimo_asiento"`
	EstadoConciliacion    string  `json:"estado_conciliacion"`
}

// EmpresaConciliacionContableResumen consolida conciliacion por periodos para una empresa.
type EmpresaConciliacionContableResumen struct {
	EmpresaID              int64                                `json:"empresa_id"`
	Desde                  string                               `json:"desde"`
	Hasta                  string                               `json:"hasta"`
	PeriodoContable        string                               `json:"periodo_contable"`
	TotalPeriodos          int                                  `json:"total_periodos"`
	PeriodosConciliados    int                                  `json:"periodos_conciliados"`
	PeriodosConPendientes  int                                  `json:"periodos_con_pendientes"`
	PeriodosConDescuadre   int                                  `json:"periodos_con_descuadre"`
	PeriodosSinMovimientos int                                  `json:"periodos_sin_movimientos"`
	Filas                  []EmpresaConciliacionContablePeriodo `json:"filas"`
}

var empresaEventoContableContrato = map[string]map[string]struct{}{
	"ventas": {
		"venta_sesion_activada": {},
		"venta_activada":        {},
		"venta_suspendida":      {},
		"venta_cerrada":         {},
		"venta_reabierta":       {},
		"venta_pagada":          {},
	},
	"facturacion": {
		"factura_emitida":                       {},
		"factura_anulada":                       {},
		"nota_credito_emitida":                  {},
		"configuracion_facturacion_actualizada": {},
		"factura_integracion_enviada":           {},
		"factura_integracion_fallida":           {},
		"factura_contingencia_activada":         {},
	},
	"compras": {
		"orden_compra_creada":                {},
		"orden_compra_pendiente_aprobacion":  {},
		"orden_compra_emitida":               {},
		"compra_recepcionada":                {},
		"compra_contabilizada":               {},
		"devolucion_proveedor_contabilizada": {},
		"proveedor_registrado":               {},
		"proveedor_actualizado":              {},
		"proveedor_activado":                 {},
		"proveedor_desactivado":              {},
		"proveedor_eliminado":                {},
	},
	"finanzas": {
		"movimiento_ingreso_registrado": {},
		"movimiento_egreso_registrado":  {},
		"periodo_contable_cerrado":      {},
		"periodo_contable_reabierto":    {},
		"tarifa_por_minutos_calculada":  {},
	},
	"creditos": {
		"credito_abono_registrado": {},
	},
}

// EnsureEmpresaEventosContablesSchema crea y migra tabla de eventos contables empresariales.
func EnsureEmpresaEventosContablesSchema(dbConn *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_eventos_contables (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			evento TEXT NOT NULL,
			entidad TEXT NOT NULL,
			entidad_id INTEGER,
			documento_tipo TEXT,
			documento_codigo TEXT,
			periodo_contable TEXT,
			monto_total REAL DEFAULT 0,
			moneda TEXT DEFAULT 'COP',
			payload_json TEXT,
			origen TEXT DEFAULT 'backend',
			fecha_evento TEXT DEFAULT (datetime('now','localtime')),
			procesado INTEGER DEFAULT 0,
			fecha_procesado TEXT,
			intentos_procesamiento INTEGER DEFAULT 0,
			fecha_ultimo_intento TEXT,
			error_procesamiento TEXT,
			asiento_contable_id INTEGER,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE TABLE IF NOT EXISTS empresa_asientos_contables (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			empresa_id INTEGER NOT NULL,
			evento_contable_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			evento TEXT NOT NULL,
			fecha_asiento TEXT DEFAULT (datetime('now','localtime')),
			periodo_contable TEXT,
			documento_tipo TEXT,
			documento_codigo TEXT,
			moneda TEXT DEFAULT 'COP',
			total_debito REAL DEFAULT 0,
			total_credito REAL DEFAULT 0,
			diferencia REAL DEFAULT 0,
			lineas_json TEXT,
			hash_idempotencia TEXT NOT NULL,
			payload_origen_json TEXT,
			fecha_procesado TEXT,
			procesado_por TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, evento_contable_id),
			UNIQUE(empresa_id, hash_idempotencia)
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_empresa_fecha ON empresa_eventos_contables(empresa_id, fecha_evento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_empresa_modulo_evento ON empresa_eventos_contables(empresa_id, modulo, evento);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_pendientes ON empresa_eventos_contables(empresa_id, procesado, fecha_evento);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asientos_contables_empresa_fecha ON empresa_asientos_contables(empresa_id, fecha_asiento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asientos_contables_empresa_modulo_evento ON empresa_asientos_contables(empresa_id, modulo, evento);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_asientos_contables_empresa_periodo ON empresa_asientos_contables(empresa_id, periodo_contable, estado);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "entidad_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "documento_tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "documento_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "monto_total", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "payload_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "origen", "TEXT DEFAULT 'backend'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "fecha_evento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "procesado", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "fecha_procesado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "intentos_procesamiento", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "fecha_ultimo_intento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "error_procesamiento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "asiento_contable_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_eventos_contables", "observaciones", "TEXT"); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_empresa_periodo ON empresa_eventos_contables(empresa_id, periodo_contable, estado);`); err != nil {
		return err
	}
	if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_reintentos ON empresa_eventos_contables(empresa_id, procesado, intentos_procesamiento, fecha_evento);`); err != nil {
		return err
	}

	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "modulo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "evento", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "fecha_asiento", "TEXT DEFAULT (datetime('now','localtime'))"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "periodo_contable", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "documento_tipo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "documento_codigo", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "moneda", "TEXT DEFAULT 'COP'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "total_debito", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "total_credito", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "diferencia", "REAL DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "lineas_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "hash_idempotencia", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "payload_origen_json", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "fecha_procesado", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "procesado_por", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_asientos_contables", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// CreateEmpresaEventoContable registra un evento de contrato contable por empresa.
func CreateEmpresaEventoContable(dbConn *sql.DB, e EmpresaEventoContable) (int64, error) {
	if e.EmpresaID <= 0 {
		return 0, fmt.Errorf("empresa_id es obligatorio")
	}
	e.Modulo = normalizeEventoContableModulo(e.Modulo)
	e.Evento = normalizeEventoContableNombre(e.Evento)
	if !isEventoContableContratado(e.Modulo, e.Evento) {
		return 0, fmt.Errorf("evento contable no soportado para modulo=%s evento=%s", e.Modulo, e.Evento)
	}
	e.Entidad = strings.TrimSpace(strings.ToLower(e.Entidad))
	if e.Entidad == "" {
		e.Entidad = "documento"
	}
	e.DocumentoTipo = strings.TrimSpace(strings.ToLower(e.DocumentoTipo))
	if e.DocumentoTipo == "" {
		e.DocumentoTipo = e.Entidad
	}
	e.DocumentoCodigo = strings.TrimSpace(e.DocumentoCodigo)
	e.PeriodoContable = normalizePeriodoEventoContable(e.PeriodoContable)
	if e.PeriodoContable == "" {
		e.PeriodoContable = normalizePeriodoEventoContable(e.FechaEvento)
	}
	if e.PeriodoContable == "" {
		e.PeriodoContable = time.Now().Format("2006-01")
	}
	e.Moneda = strings.TrimSpace(strings.ToUpper(e.Moneda))
	if e.Moneda == "" {
		e.Moneda = "COP"
	}
	e.Origen = strings.TrimSpace(strings.ToLower(e.Origen))
	if e.Origen == "" {
		e.Origen = "backend"
	}
	e.UsuarioCreador = strings.TrimSpace(e.UsuarioCreador)
	e.Estado = strings.TrimSpace(strings.ToLower(e.Estado))
	if e.Estado == "" {
		e.Estado = "activo"
	}
	e.Observaciones = strings.TrimSpace(e.Observaciones)
	e.PayloadJSON = strings.TrimSpace(e.PayloadJSON)
	e.FechaEvento = strings.TrimSpace(e.FechaEvento)
	if e.FechaEvento == "" {
		e.FechaEvento = time.Now().Format("2006-01-02 15:04:05")
	}

	var entidadID interface{}
	if e.EntidadID > 0 {
		entidadID = e.EntidadID
	}

	query := `INSERT INTO empresa_eventos_contables (
		empresa_id,
		modulo,
		evento,
		entidad,
		entidad_id,
		documento_tipo,
		documento_codigo,
		periodo_contable,
		monto_total,
		moneda,
		payload_json,
		origen,
		fecha_evento,
		procesado,
		fecha_procesado,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NULL, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`

	id, err := insertSQLCompat(dbConn, query,
		e.EmpresaID,
		e.Modulo,
		e.Evento,
		e.Entidad,
		entidadID,
		e.DocumentoTipo,
		e.DocumentoCodigo,
		e.PeriodoContable,
		e.MontoTotal,
		e.Moneda,
		e.PayloadJSON,
		e.Origen,
		e.FechaEvento,
		e.UsuarioCreador,
		e.Estado,
		e.Observaciones,
	)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// ListEmpresaEventosContables lista eventos contables por empresa y filtros opcionales.
func ListEmpresaEventosContables(dbConn *sql.DB, empresaID int64, f EmpresaEventoContableFilter) ([]EmpresaEventoContable, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}
	query := `SELECT
		id,
		empresa_id,
		COALESCE(modulo, ''),
		COALESCE(evento, ''),
		COALESCE(entidad, ''),
		COALESCE(entidad_id, 0),
		COALESCE(documento_tipo, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(monto_total, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(payload_json, ''),
		COALESCE(origen, 'backend'),
		COALESCE(fecha_evento, ''),
		COALESCE(procesado, 0),
		COALESCE(fecha_procesado, ''),
		COALESCE(intentos_procesamiento, 0),
		COALESCE(fecha_ultimo_intento, ''),
		COALESCE(error_procesamiento, ''),
		COALESCE(asiento_contable_id, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_eventos_contables
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !f.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if modulo := normalizeEventoContableModulo(f.Modulo); modulo != "" {
		query += ` AND COALESCE(modulo, '') = ?`
		args = append(args, modulo)
	}
	if evento := normalizeEventoContableNombre(f.Evento); evento != "" {
		query += ` AND COALESCE(evento, '') = ?`
		args = append(args, evento)
	}
	if periodo := normalizePeriodoEventoContable(f.PeriodoContable); periodo != "" {
		query += ` AND COALESCE(periodo_contable, '') = ?`
		args = append(args, periodo)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	query += ` ORDER BY COALESCE(fecha_evento, '') DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaEventoContable, 0)
	for rows.Next() {
		var item EmpresaEventoContable
		var procesadoInt int64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Modulo,
			&item.Evento,
			&item.Entidad,
			&item.EntidadID,
			&item.DocumentoTipo,
			&item.DocumentoCodigo,
			&item.PeriodoContable,
			&item.MontoTotal,
			&item.Moneda,
			&item.PayloadJSON,
			&item.Origen,
			&item.FechaEvento,
			&procesadoInt,
			&item.FechaProcesado,
			&item.IntentosProcesamiento,
			&item.FechaUltimoIntento,
			&item.ErrorProcesamiento,
			&item.AsientoContableID,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Procesado = procesadoInt == 1
		out = append(out, item)
	}
	return out, nil
}

// ListEmpresaAsientosContables lista asientos canónicos por empresa con filtros opcionales.
func ListEmpresaAsientosContables(dbConn *sql.DB, empresaID int64, f EmpresaAsientoContableFilter) ([]EmpresaAsientoContable, error) {
	if empresaID <= 0 {
		return nil, fmt.Errorf("empresa_id es obligatorio")
	}

	query := `SELECT
		id,
		empresa_id,
		COALESCE(evento_contable_id, 0),
		COALESCE(modulo, ''),
		COALESCE(evento, ''),
		COALESCE(fecha_asiento, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(documento_tipo, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(moneda, 'COP'),
		COALESCE(total_debito, 0),
		COALESCE(total_credito, 0),
		COALESCE(diferencia, 0),
		COALESCE(lineas_json, ''),
		COALESCE(hash_idempotencia, ''),
		COALESCE(payload_origen_json, ''),
		COALESCE(fecha_procesado, ''),
		COALESCE(procesado_por, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_asientos_contables
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !f.IncludeInactive {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if modulo := normalizeEventoContableModulo(f.Modulo); modulo != "" {
		query += ` AND COALESCE(modulo, '') = ?`
		args = append(args, modulo)
	}
	if evento := normalizeEventoContableNombre(f.Evento); evento != "" {
		query += ` AND COALESCE(evento, '') = ?`
		args = append(args, evento)
	}
	if periodo := normalizePeriodoEventoContable(f.PeriodoContable); periodo != "" {
		query += ` AND COALESCE(periodo_contable, '') = ?`
		args = append(args, periodo)
	}
	if desde := strings.TrimSpace(f.Desde); desde != "" {
		query += ` AND date(fecha_asiento) >= date(?)`
		args = append(args, desde)
	}
	if hasta := strings.TrimSpace(f.Hasta); hasta != "" {
		query += ` AND date(fecha_asiento) <= date(?)`
		args = append(args, hasta)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}
	query += ` ORDER BY COALESCE(fecha_asiento, '') DESC, id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaAsientoContable, 0)
	for rows.Next() {
		var item EmpresaAsientoContable
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.EventoContableID,
			&item.Modulo,
			&item.Evento,
			&item.FechaAsiento,
			&item.PeriodoContable,
			&item.DocumentoTipo,
			&item.DocumentoCodigo,
			&item.Moneda,
			&item.TotalDebito,
			&item.TotalCredito,
			&item.Diferencia,
			&item.LineasJSON,
			&item.HashIdempotencia,
			&item.PayloadOrigenJSON,
			&item.FechaProcesado,
			&item.ProcesadoPor,
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

// GetEmpresaConciliacionContablePorPeriodo resume eventos vs asientos por periodo contable.
func GetEmpresaConciliacionContablePorPeriodo(dbConn *sql.DB, empresaID int64, f EmpresaConciliacionContableFilter) (EmpresaConciliacionContableResumen, error) {
	resumen := EmpresaConciliacionContableResumen{
		EmpresaID:       empresaID,
		Desde:           strings.TrimSpace(f.Desde),
		Hasta:           strings.TrimSpace(f.Hasta),
		PeriodoContable: normalizePeriodoEventoContable(f.PeriodoContable),
		Filas:           make([]EmpresaConciliacionContablePeriodo, 0),
	}
	if empresaID <= 0 {
		return resumen, fmt.Errorf("empresa_id es obligatorio")
	}

	periodos := make(map[string]*EmpresaConciliacionContablePeriodo)

	periodoEventoExpr := `COALESCE(NULLIF(periodo_contable, ''), substr(COALESCE(fecha_evento, ''), 1, 7), 'sin_periodo')`
	queryEventos := `SELECT
		` + periodoEventoExpr + `,
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN COALESCE(procesado, 0) = 1 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(procesado, 0) = 0 THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(procesado, 0) = 0 AND COALESCE(error_procesamiento, '') <> '' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN COALESCE(procesado, 0) = 1 THEN COALESCE(monto_total, 0) ELSE 0 END), 0),
		COALESCE(MAX(fecha_evento), '')
	FROM empresa_eventos_contables
	WHERE empresa_id = ?`
	argsEventos := []interface{}{empresaID}
	if !f.IncludeInactive {
		queryEventos += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if resumen.Desde != "" {
		queryEventos += ` AND date(fecha_evento) >= date(?)`
		argsEventos = append(argsEventos, resumen.Desde)
	}
	if resumen.Hasta != "" {
		queryEventos += ` AND date(fecha_evento) <= date(?)`
		argsEventos = append(argsEventos, resumen.Hasta)
	}
	if resumen.PeriodoContable != "" {
		queryEventos += ` AND ` + periodoEventoExpr + ` = ?`
		argsEventos = append(argsEventos, resumen.PeriodoContable)
	}
	queryEventos += ` GROUP BY ` + periodoEventoExpr

	rowsEventos, err := querySQLCompat(dbConn, queryEventos, argsEventos...)
	if err != nil {
		return resumen, err
	}
	defer rowsEventos.Close()

	for rowsEventos.Next() {
		var periodo, ultimoEvento string
		var eventosTotal, eventosProcesados, eventosPendientes, eventosConError int64
		var eventosMontoProcesado float64
		if err := rowsEventos.Scan(
			&periodo,
			&eventosTotal,
			&eventosProcesados,
			&eventosPendientes,
			&eventosConError,
			&eventosMontoProcesado,
			&ultimoEvento,
		); err != nil {
			return resumen, err
		}
		item := getOrCreateEmpresaConciliacionPeriodo(periodos, periodo)
		item.EventosTotal = eventosTotal
		item.EventosProcesados = eventosProcesados
		item.EventosPendientes = eventosPendientes
		item.EventosConError = eventosConError
		item.EventosMontoProcesado = eventosMontoProcesado
		item.UltimoEvento = strings.TrimSpace(ultimoEvento)
	}
	if err := rowsEventos.Err(); err != nil {
		return resumen, err
	}

	periodoAsientoExpr := `COALESCE(NULLIF(periodo_contable, ''), substr(COALESCE(fecha_asiento, ''), 1, 7), 'sin_periodo')`
	queryAsientos := `SELECT
		` + periodoAsientoExpr + `,
		COALESCE(COUNT(1), 0),
		COALESCE(SUM(CASE WHEN COALESCE(total_debito, 0) > 0 THEN COALESCE(total_debito, 0) ELSE COALESCE(total_credito, 0) END), 0),
		COALESCE(MAX(fecha_asiento), '')
	FROM empresa_asientos_contables
	WHERE empresa_id = ?`
	argsAsientos := []interface{}{empresaID}
	if !f.IncludeInactive {
		queryAsientos += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	if resumen.Desde != "" {
		queryAsientos += ` AND date(fecha_asiento) >= date(?)`
		argsAsientos = append(argsAsientos, resumen.Desde)
	}
	if resumen.Hasta != "" {
		queryAsientos += ` AND date(fecha_asiento) <= date(?)`
		argsAsientos = append(argsAsientos, resumen.Hasta)
	}
	if resumen.PeriodoContable != "" {
		queryAsientos += ` AND ` + periodoAsientoExpr + ` = ?`
		argsAsientos = append(argsAsientos, resumen.PeriodoContable)
	}
	queryAsientos += ` GROUP BY ` + periodoAsientoExpr

	rowsAsientos, err := querySQLCompat(dbConn, queryAsientos, argsAsientos...)
	if err != nil {
		return resumen, err
	}
	defer rowsAsientos.Close()

	for rowsAsientos.Next() {
		var periodo, ultimoAsiento string
		var asientosTotal int64
		var asientosMontoTotal float64
		if err := rowsAsientos.Scan(&periodo, &asientosTotal, &asientosMontoTotal, &ultimoAsiento); err != nil {
			return resumen, err
		}
		item := getOrCreateEmpresaConciliacionPeriodo(periodos, periodo)
		item.AsientosTotal = asientosTotal
		item.AsientosMontoTotal = asientosMontoTotal
		item.UltimoAsiento = strings.TrimSpace(ultimoAsiento)
	}
	if err := rowsAsientos.Err(); err != nil {
		return resumen, err
	}

	filas := make([]EmpresaConciliacionContablePeriodo, 0, len(periodos))
	for _, item := range periodos {
		item.EventosMontoProcesado = roundReportesMoney(item.EventosMontoProcesado)
		item.AsientosMontoTotal = roundReportesMoney(item.AsientosMontoTotal)
		item.DesfaseEventosAsiento = item.EventosProcesados - item.AsientosTotal
		item.DesfaseMonto = roundReportesMoney(item.EventosMontoProcesado - item.AsientosMontoTotal)

		switch {
		case item.EventosPendientes > 0 || item.EventosConError > 0:
			item.EstadoConciliacion = "con_pendientes"
		case item.DesfaseEventosAsiento != 0 || conciliacionAbsFloat64(item.DesfaseMonto) > 0.009:
			item.EstadoConciliacion = "con_descuadre"
		case item.EventosTotal == 0 && item.AsientosTotal == 0:
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
			if filas[i].UltimoEvento == filas[j].UltimoEvento {
				return filas[i].UltimoAsiento > filas[j].UltimoAsiento
			}
			return filas[i].UltimoEvento > filas[j].UltimoEvento
		}
		return keyI > keyJ
	})

	limit := normalizeConciliacionLimit(f.Limit)
	if len(filas) > limit {
		filas = filas[:limit]
	}

	resumen.TotalPeriodos = len(filas)
	for _, item := range filas {
		switch item.EstadoConciliacion {
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

// ProcessEmpresaEventosContablesPendientes procesa por lotes eventos pendientes y genera asientos idempotentes.
func ProcessEmpresaEventosContablesPendientes(dbConn *sql.DB, empresaID int64, procesadoPor string, limit int) (EmpresaProcesoAsientosResultado, error) {
	return ProcessEmpresaEventosContablesPendientesConPolitica(dbConn, empresaID, procesadoPor, limit, 0)
}

// ProcessEmpresaEventosContablesPendientesConPolitica procesa eventos pendientes aplicando limite de reintentos cuando se configura.
func ProcessEmpresaEventosContablesPendientesConPolitica(dbConn *sql.DB, empresaID int64, procesadoPor string, limit int, maxRetries int) (EmpresaProcesoAsientosResultado, error) {
	result := EmpresaProcesoAsientosResultado{
		EmpresaID: empresaID,
		Errores:   make([]string, 0),
	}
	if empresaID <= 0 {
		return result, fmt.Errorf("empresa_id es obligatorio")
	}
	limit = normalizeAsientosWorkerBatchSize(limit)
	procesadoPor = strings.TrimSpace(procesadoPor)
	if procesadoPor == "" {
		procesadoPor = "sistema"
	}
	maxRetries = normalizeAsientosWorkerRetriesForPolicy(maxRetries)

	eventos, err := listEmpresaEventosContablesPendientes(dbConn, empresaID, limit, maxRetries)
	if err != nil {
		return result, err
	}
	result.EventosRevisados = len(eventos)

	for _, evento := range eventos {
		asientoID, creado, err := ensureEmpresaAsientoContableFromEvento(dbConn, evento, procesadoPor)
		if err != nil {
			_ = markEmpresaEventoContableFailed(dbConn, evento.ID, err)
			result.Fallidos++
			if len(result.Errores) < 20 {
				result.Errores = append(result.Errores, fmt.Sprintf("evento_id=%d: %s", evento.ID, trimProcessingError(err)))
			}
			continue
		}

		if err := markEmpresaEventoContableProcessed(dbConn, evento.ID, asientoID); err != nil {
			_ = markEmpresaEventoContableFailed(dbConn, evento.ID, err)
			result.Fallidos++
			if len(result.Errores) < 20 {
				result.Errores = append(result.Errores, fmt.Sprintf("evento_id=%d: %s", evento.ID, trimProcessingError(err)))
			}
			continue
		}

		result.EventosProcesados++
		if creado {
			result.AsientosCreados++
		} else {
			result.AsientosExistentes++
		}
	}

	return result, nil
}

// RunEmpresaAsientosContablesWorkerCycle ejecuta una corrida global por empresas con eventos pendientes.
func RunEmpresaAsientosContablesWorkerCycle(dbConn *sql.DB, procesadoPor string, batchSize int, maxRetries int) (EmpresaAsientosWorkerResumen, error) {
	resumen := EmpresaAsientosWorkerResumen{Errores: make([]string, 0)}
	if dbConn == nil {
		return resumen, fmt.Errorf("conexion de base de datos no disponible")
	}

	batchSize = normalizeAsientosWorkerBatchSize(batchSize)
	maxRetries = normalizeAsientosWorkerRetries(maxRetries)
	procesadoPor = strings.TrimSpace(procesadoPor)
	if procesadoPor == "" {
		procesadoPor = "worker_asientos"
	}

	empresaIDs, err := listEmpresaIDsConEventosPendientes(dbConn, maxRetries, 500)
	if err != nil {
		return resumen, err
	}
	resumen.EmpresasConPendientes = len(empresaIDs)

	for _, empresaID := range empresaIDs {
		resultado, processErr := ProcessEmpresaEventosContablesPendientesConPolitica(dbConn, empresaID, procesadoPor, batchSize, maxRetries)
		if processErr != nil {
			if len(resumen.Errores) < 20 {
				resumen.Errores = append(resumen.Errores, fmt.Sprintf("empresa_id=%d: %s", empresaID, trimProcessingError(processErr)))
			}
			continue
		}

		if resultado.EventosRevisados > 0 {
			resumen.EmpresasProcesadas++
		}
		resumen.EventosRevisados += resultado.EventosRevisados
		resumen.EventosProcesados += resultado.EventosProcesados
		resumen.AsientosCreados += resultado.AsientosCreados
		resumen.AsientosExistentes += resultado.AsientosExistentes
		resumen.Fallidos += resultado.Fallidos

		for _, errMsg := range resultado.Errores {
			if len(resumen.Errores) >= 20 {
				break
			}
			resumen.Errores = append(resumen.Errores, fmt.Sprintf("empresa_id=%d: %s", empresaID, errMsg))
		}
	}

	return resumen, nil
}

// StartEmpresaAsientosContablesWorker ejecuta procesamiento automático por lotes en intervalo fijo.
func StartEmpresaAsientosContablesWorker(dbConn *sql.DB, interval time.Duration, batchSize int, maxRetries int, stop <-chan struct{}) {
	if dbConn == nil {
		return
	}

	interval = normalizeAsientosWorkerInterval(interval)
	batchSize = normalizeAsientosWorkerBatchSize(batchSize)
	maxRetries = normalizeAsientosWorkerRetries(maxRetries)

	runCycle := func(origin string) {
		resumen, err := RunEmpresaAsientosContablesWorkerCycle(dbConn, "worker_asientos", batchSize, maxRetries)
		if err != nil {
			log.Printf("[asientos_worker] ciclo=%s error=%v", origin, err)
			return
		}
		if resumen.EventosRevisados > 0 || resumen.Fallidos > 0 {
			log.Printf("[asientos_worker] ciclo=%s empresas=%d revisados=%d procesados=%d creados=%d existentes=%d fallidos=%d", origin, resumen.EmpresasProcesadas, resumen.EventosRevisados, resumen.EventosProcesados, resumen.AsientosCreados, resumen.AsientosExistentes, resumen.Fallidos)
		}
	}

	runCycle("startup")
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			runCycle("ticker")
		case <-stop:
			return
		}
	}
}

func listEmpresaEventosContablesPendientes(dbConn *sql.DB, empresaID int64, limit int, maxRetries int) ([]EmpresaEventoContable, error) {
	query := `SELECT
		id,
		empresa_id,
		COALESCE(modulo, ''),
		COALESCE(evento, ''),
		COALESCE(entidad, ''),
		COALESCE(entidad_id, 0),
		COALESCE(documento_tipo, ''),
		COALESCE(documento_codigo, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(monto_total, 0),
		COALESCE(moneda, 'COP'),
		COALESCE(payload_json, ''),
		COALESCE(origen, 'backend'),
		COALESCE(fecha_evento, ''),
		COALESCE(procesado, 0),
		COALESCE(fecha_procesado, ''),
		COALESCE(intentos_procesamiento, 0),
		COALESCE(fecha_ultimo_intento, ''),
		COALESCE(error_procesamiento, ''),
		COALESCE(asiento_contable_id, 0),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM empresa_eventos_contables
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
		AND COALESCE(procesado, 0) = 0`
	args := []interface{}{empresaID}
	if maxRetries > 0 {
		query += ` AND COALESCE(intentos_procesamiento, 0) < ?`
		args = append(args, maxRetries)
	}
	query += ` ORDER BY COALESCE(fecha_evento, '') ASC, id ASC LIMIT ?`
	args = append(args, limit)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaEventoContable, 0)
	for rows.Next() {
		var item EmpresaEventoContable
		var procesadoInt int64
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Modulo,
			&item.Evento,
			&item.Entidad,
			&item.EntidadID,
			&item.DocumentoTipo,
			&item.DocumentoCodigo,
			&item.PeriodoContable,
			&item.MontoTotal,
			&item.Moneda,
			&item.PayloadJSON,
			&item.Origen,
			&item.FechaEvento,
			&procesadoInt,
			&item.FechaProcesado,
			&item.IntentosProcesamiento,
			&item.FechaUltimoIntento,
			&item.ErrorProcesamiento,
			&item.AsientoContableID,
			&item.FechaCreacion,
			&item.FechaActualizacion,
			&item.UsuarioCreador,
			&item.Estado,
			&item.Observaciones,
		); err != nil {
			return nil, err
		}
		item.Procesado = procesadoInt == 1
		out = append(out, item)
	}
	return out, rows.Err()
}

func listEmpresaIDsConEventosPendientes(dbConn *sql.DB, maxRetries int, limitEmpresas int) ([]int64, error) {
	if limitEmpresas <= 0 {
		limitEmpresas = 500
	}

	query := `SELECT DISTINCT empresa_id
	FROM empresa_eventos_contables
	WHERE COALESCE(estado, 'activo') = 'activo'
		AND COALESCE(procesado, 0) = 0`
	args := make([]interface{}, 0)
	if maxRetries > 0 {
		query += ` AND COALESCE(intentos_procesamiento, 0) < ?`
		args = append(args, maxRetries)
	}
	query += ` ORDER BY empresa_id ASC LIMIT ?`
	args = append(args, limitEmpresas)

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]int64, 0)
	for rows.Next() {
		var empresaID int64
		if err := rows.Scan(&empresaID); err != nil {
			return nil, err
		}
		if empresaID > 0 {
			out = append(out, empresaID)
		}
	}
	return out, rows.Err()
}

func normalizeConciliacionLimit(limit int) int {
	if limit <= 0 {
		return 24
	}
	if limit > 120 {
		return 120
	}
	return limit
}

func conciliacionPeriodoSortKey(periodo string) string {
	periodo = normalizePeriodoEventoContable(periodo)
	if periodo == "" {
		return "0000-00"
	}
	return periodo
}

func conciliacionAbsFloat64(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func getOrCreateEmpresaConciliacionPeriodo(periodos map[string]*EmpresaConciliacionContablePeriodo, periodo string) *EmpresaConciliacionContablePeriodo {
	periodo = strings.TrimSpace(periodo)
	if norm := normalizePeriodoEventoContable(periodo); norm != "" {
		periodo = norm
	}
	if periodo == "" {
		periodo = "sin_periodo"
	}
	item, ok := periodos[periodo]
	if ok {
		return item
	}
	item = &EmpresaConciliacionContablePeriodo{PeriodoContable: periodo}
	periodos[periodo] = item
	return item
}

func normalizeAsientosWorkerInterval(interval time.Duration) time.Duration {
	if interval <= 0 {
		return defaultAsientosWorkerInterval
	}
	if interval < time.Minute {
		return time.Minute
	}
	return interval
}

func normalizeAsientosWorkerBatchSize(limit int) int {
	if limit <= 0 {
		return defaultAsientosWorkerBatchSize
	}
	if limit > maxAsientosWorkerBatchSize {
		return maxAsientosWorkerBatchSize
	}
	return limit
}

func normalizeAsientosWorkerRetries(maxRetries int) int {
	if maxRetries <= 0 {
		return defaultAsientosWorkerRetries
	}
	if maxRetries > maxAsientosWorkerRetries {
		return maxAsientosWorkerRetries
	}
	return maxRetries
}

func normalizeAsientosWorkerRetriesForPolicy(maxRetries int) int {
	if maxRetries <= 0 {
		return 0
	}
	if maxRetries > maxAsientosWorkerRetries {
		return maxAsientosWorkerRetries
	}
	return maxRetries
}

func ensureEmpresaAsientoContableFromEvento(dbConn *sql.DB, evento EmpresaEventoContable, procesadoPor string) (int64, bool, error) {
	if evento.EmpresaID <= 0 || evento.ID <= 0 {
		return 0, false, fmt.Errorf("evento contable invalido para generar asiento")
	}

	asientoID, err := findEmpresaAsientoByEventoOrHash(dbConn, evento.EmpresaID, evento.ID, "")
	if err != nil {
		return 0, false, err
	}
	if asientoID > 0 {
		return asientoID, false, nil
	}

	cfg, err := GetEmpresaFinanzasConfiguracion(dbConn, evento.EmpresaID)
	if err != nil {
		return 0, false, err
	}
	if cfg == nil {
		cfg = defaultEmpresaFinanzasConfiguracion(evento.EmpresaID)
	}

	lineas := buildEmpresaAsientoContableLineas(evento, cfg)
	lineasJSONBytes, err := json.Marshal(lineas)
	if err != nil {
		return 0, false, err
	}

	totalDebito := 0.0
	totalCredito := 0.0
	for _, ln := range lineas {
		totalDebito += maxFloat64(ln.Debito, 0)
		totalCredito += maxFloat64(ln.Credito, 0)
	}
	diferencia := totalDebito - totalCredito

	fechaAsiento := strings.TrimSpace(evento.FechaEvento)
	if fechaAsiento == "" {
		fechaAsiento = time.Now().Format("2006-01-02 15:04:05")
	}
	periodo := normalizePeriodoEventoContable(evento.PeriodoContable)
	if periodo == "" {
		periodo = normalizePeriodoEventoContable(fechaAsiento)
	}
	if periodo == "" {
		periodo = time.Now().Format("2006-01")
	}
	moneda := strings.TrimSpace(strings.ToUpper(evento.Moneda))
	if moneda == "" {
		moneda = "COP"
	}

	hash := buildAsientoIdempotenciaHash(evento)
	nowExpr := sqlNowExpr()
	query := `INSERT INTO empresa_asientos_contables (
		empresa_id,
		evento_contable_id,
		modulo,
		evento,
		fecha_asiento,
		periodo_contable,
		documento_tipo,
		documento_codigo,
		moneda,
		total_debito,
		total_credito,
		diferencia,
		lineas_json,
		hash_idempotencia,
		payload_origen_json,
		fecha_procesado,
		procesado_por,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (
		?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ` + nowExpr + `, ?, ` + nowExpr + `, ` + nowExpr + `, ?, 'activo', ?
	)`
	id, err := insertSQLCompat(dbConn, query,
		evento.EmpresaID,
		evento.ID,
		evento.Modulo,
		evento.Evento,
		fechaAsiento,
		periodo,
		evento.DocumentoTipo,
		evento.DocumentoCodigo,
		moneda,
		totalDebito,
		totalCredito,
		diferencia,
		string(lineasJSONBytes),
		hash,
		strings.TrimSpace(evento.PayloadJSON),
		procesadoPor,
		strings.TrimSpace(evento.UsuarioCreador),
		strings.TrimSpace(evento.Observaciones),
	)
	if err != nil {
		if isUniqueConstraintError(err) {
			existingID, findErr := findEmpresaAsientoByEventoOrHash(dbConn, evento.EmpresaID, evento.ID, hash)
			if findErr != nil {
				return 0, false, findErr
			}
			if existingID > 0 {
				return existingID, false, nil
			}
		}
		return 0, false, err
	}
	return id, true, nil
}

func markEmpresaEventoContableProcessed(dbConn *sql.DB, eventoID, asientoID int64) error {
	if eventoID <= 0 {
		return fmt.Errorf("evento_id invalido")
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE empresa_eventos_contables
	SET procesado = 1,
		fecha_procesado = `+nowExpr+`,
		fecha_ultimo_intento = `+nowExpr+`,
		intentos_procesamiento = COALESCE(intentos_procesamiento, 0) + 1,
		error_procesamiento = '',
		asiento_contable_id = ?,
		fecha_actualizacion = `+nowExpr+`
	WHERE id = ?`, asientoID, eventoID)
	return err
}

func markEmpresaEventoContableFailed(dbConn *sql.DB, eventoID int64, processErr error) error {
	if eventoID <= 0 {
		return fmt.Errorf("evento_id invalido")
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, `UPDATE empresa_eventos_contables
	SET procesado = 0,
		fecha_ultimo_intento = `+nowExpr+`,
		intentos_procesamiento = COALESCE(intentos_procesamiento, 0) + 1,
		error_procesamiento = ?,
		fecha_actualizacion = `+nowExpr+`
	WHERE id = ?`, trimProcessingError(processErr), eventoID)
	return err
}

func findEmpresaAsientoByEventoOrHash(dbConn *sql.DB, empresaID, eventoID int64, hash string) (int64, error) {
	if empresaID <= 0 {
		return 0, fmt.Errorf("empresa_id invalido")
	}
	if eventoID > 0 {
		var id int64
		err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_asientos_contables WHERE empresa_id = ? AND evento_contable_id = ? LIMIT 1`, empresaID, eventoID).Scan(&id)
		if err == nil {
			return id, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}
	hash = strings.TrimSpace(hash)
	if hash != "" {
		var id int64
		err := queryRowSQLCompat(dbConn, `SELECT id FROM empresa_asientos_contables WHERE empresa_id = ? AND hash_idempotencia = ? LIMIT 1`, empresaID, hash).Scan(&id)
		if err == nil {
			return id, nil
		}
		if err != sql.ErrNoRows {
			return 0, err
		}
	}
	return 0, nil
}

func buildEmpresaAsientoContableLineas(evento EmpresaEventoContable, cfg *EmpresaFinanzasConfiguracion) []EmpresaAsientoContableLinea {
	payload := parseEventoPayload(evento.PayloadJSON)
	monto := resolveMontoEventoContable(evento, payload)
	if monto <= 0 {
		return []EmpresaAsientoContableLinea{}
	}

	if cfg == nil {
		cfg = defaultEmpresaFinanzasConfiguracion(evento.EmpresaID)
	}
	categoria := payloadString(payload, "categoria")
	tipoMovimiento := payloadString(payload, "tipo_movimiento")

	cuentaCaja := sanitizeContableAccount(cfg.CuentaCajaBancos)
	if cuentaCaja == "" {
		cuentaCaja = "110505"
	}
	cuentaIngresos := resolveEmpresaCuentaPorCategoria(cfg.CuentasIngresoCategoria, categoria, cfg.CuentaIngresos, "413595")
	cuentaGastos := resolveEmpresaCuentaPorCategoria(cfg.CuentasEgresoCategoria, categoria, cfg.CuentaGastos, "519595")

	nuevoIngreso := func() []EmpresaAsientoContableLinea {
		return []EmpresaAsientoContableLinea{
			{Cuenta: cuentaCaja, Descripcion: "Caja y bancos", Debito: monto, Credito: 0},
			{Cuenta: cuentaIngresos, Descripcion: "Ingresos operacionales", Debito: 0, Credito: monto},
		}
	}
	nuevoEgreso := func() []EmpresaAsientoContableLinea {
		return []EmpresaAsientoContableLinea{
			{Cuenta: cuentaGastos, Descripcion: "Gastos operacionales", Debito: monto, Credito: 0},
			{Cuenta: cuentaCaja, Descripcion: "Caja y bancos", Debito: 0, Credito: monto},
		}
	}

	switch normalizeEventoContableModulo(evento.Modulo) {
	case "finanzas":
		switch normalizeEventoContableNombre(evento.Evento) {
		case "movimiento_ingreso_registrado":
			return nuevoIngreso()
		case "movimiento_egreso_registrado":
			return nuevoEgreso()
		}
		if strings.EqualFold(tipoMovimiento, "ingreso") {
			return nuevoIngreso()
		}
		if strings.EqualFold(tipoMovimiento, "egreso") {
			return nuevoEgreso()
		}

	case "ventas":
		switch normalizeEventoContableNombre(evento.Evento) {
		case "venta_pagada", "venta_cerrada":
			return nuevoIngreso()
		}

	case "facturacion":
		switch normalizeEventoContableNombre(evento.Evento) {
		case "factura_emitida":
			return nuevoIngreso()
		case "factura_anulada", "nota_credito_emitida":
			return []EmpresaAsientoContableLinea{
				{Cuenta: cuentaIngresos, Descripcion: "Reversion de ingresos", Debito: monto, Credito: 0},
				{Cuenta: cuentaCaja, Descripcion: "Reversion caja y bancos", Debito: 0, Credito: monto},
			}
		}

	case "compras":
		switch normalizeEventoContableNombre(evento.Evento) {
		case "orden_compra_creada", "orden_compra_pendiente_aprobacion":
			// Se registran como hitos precontables para trazabilidad, sin asiento.
			return []EmpresaAsientoContableLinea{}
		case "devolucion_proveedor_contabilizada":
			return nuevoIngreso()
		case "orden_compra_emitida", "compra_recepcionada", "compra_contabilizada":
			return nuevoEgreso()
		}

	case "creditos":
		switch normalizeEventoContableNombre(evento.Evento) {
		case "credito_abono_registrado":
			capitalAplicado, interesAplicado, moraAplicada := splitCreditoAbonoAsientoComponentes(payload, monto)

			cuentaCartera := sanitizeContableAccount(payloadString(payload, "cuenta_cartera", "cuenta_creditos_cartera"))
			if cuentaCartera == "" {
				cuentaCartera = "130505"
			}
			cuentaInteres := resolveEmpresaCuentaPorCategoria(cfg.CuentasIngresoCategoria, payloadString(payload, "categoria_interes", "categoria_intereses"), cfg.CuentaIngresos, "417595")
			cuentaMora := resolveEmpresaCuentaPorCategoria(cfg.CuentasIngresoCategoria, payloadString(payload, "categoria_mora"), cfg.CuentaIngresos, "421010")

			lineas := make([]EmpresaAsientoContableLinea, 0, 4)
			lineas = append(lineas, EmpresaAsientoContableLinea{Cuenta: cuentaCaja, Descripcion: "Caja y bancos", Debito: monto, Credito: 0})
			if capitalAplicado > 0 {
				lineas = append(lineas, EmpresaAsientoContableLinea{Cuenta: cuentaCartera, Descripcion: "Cartera de creditos", Debito: 0, Credito: capitalAplicado})
			}
			if interesAplicado > 0 {
				lineas = append(lineas, EmpresaAsientoContableLinea{Cuenta: cuentaInteres, Descripcion: "Ingresos por interes de credito", Debito: 0, Credito: interesAplicado})
			}
			if moraAplicada > 0 {
				lineas = append(lineas, EmpresaAsientoContableLinea{Cuenta: cuentaMora, Descripcion: "Ingresos por interes de mora", Debito: 0, Credito: moraAplicada})
			}

			return lineas
		}
	}

	return []EmpresaAsientoContableLinea{}
}

func resolveEmpresaCuentaPorCategoria(rawMap, categoria, fallback, fallbackDefault string) string {
	fallback = sanitizeContableAccount(fallback)
	if fallback == "" {
		fallback = fallbackDefault
	}
	cat := strings.TrimSpace(strings.ToLower(categoria))
	if cat == "" {
		return fallback
	}

	for _, line := range strings.Split(strings.TrimSpace(rawMap), "\n") {
		entry := strings.TrimSpace(line)
		if entry == "" {
			continue
		}
		idx := strings.IndexAny(entry, "=:")
		if idx <= 0 {
			continue
		}
		catLine := strings.TrimSpace(strings.ToLower(entry[:idx]))
		if catLine != cat {
			continue
		}
		cuenta := sanitizeContableAccount(strings.TrimSpace(entry[idx+1:]))
		if cuenta != "" {
			return cuenta
		}
	}
	return fallback
}

func parseEventoPayload(raw string) map[string]interface{} {
	out := map[string]interface{}{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return out
	}
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

func payloadString(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		v, ok := payload[key]
		if !ok {
			continue
		}
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

func payloadNumber(payload map[string]interface{}, keys ...string) float64 {
	for _, key := range keys {
		v, ok := payload[key]
		if !ok || v == nil {
			continue
		}
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s == "" {
			continue
		}
		var n float64
		if _, err := fmt.Sscanf(strings.ReplaceAll(s, ",", "."), "%f", &n); err == nil {
			if n > 0 {
				return n
			}
		}
	}
	return 0
}

func splitCreditoAbonoAsientoComponentes(payload map[string]interface{}, monto float64) (float64, float64, float64) {
	interes := roundReportesMoney(payloadNumber(payload, "interes_aplicado", "interes"))
	mora := roundReportesMoney(payloadNumber(payload, "mora_aplicada", "mora"))
	capital := roundReportesMoney(payloadNumber(payload, "capital_aplicado", "capital"))

	if interes < 0 {
		interes = 0
	}
	if mora < 0 {
		mora = 0
	}
	if interes+mora > monto {
		exceso := roundReportesMoney(interes + mora - monto)
		if mora >= exceso {
			mora = roundReportesMoney(mora - exceso)
			exceso = 0
		} else {
			exceso = roundReportesMoney(exceso - mora)
			mora = 0
		}
		if exceso > 0 {
			interes = roundReportesMoney(maxFloat64(interes-exceso, 0))
		}
	}

	if capital <= 0 {
		capital = roundReportesMoney(maxFloat64(monto-interes-mora, 0))
	}

	total := roundReportesMoney(capital + interes + mora)
	if total < monto {
		capital = roundReportesMoney(capital + (monto - total))
	}
	if total > monto {
		exceso := roundReportesMoney(total - monto)
		if capital >= exceso {
			capital = roundReportesMoney(capital - exceso)
			exceso = 0
		} else {
			exceso = roundReportesMoney(exceso - capital)
			capital = 0
		}
		if exceso > 0 {
			if interes >= exceso {
				interes = roundReportesMoney(interes - exceso)
				exceso = 0
			} else {
				exceso = roundReportesMoney(exceso - interes)
				interes = 0
			}
		}
		if exceso > 0 {
			mora = roundReportesMoney(maxFloat64(mora-exceso, 0))
		}
	}

	return roundReportesMoney(capital), roundReportesMoney(interes), roundReportesMoney(mora)
}

func resolveMontoEventoContable(evento EmpresaEventoContable, payload map[string]interface{}) float64 {
	if evento.MontoTotal > 0 {
		return evento.MontoTotal
	}
	monto := payloadNumber(payload, "total_neto", "total", "monto_total", "monto")
	if monto > 0 {
		return monto
	}
	return 0
}

func buildAsientoIdempotenciaHash(evento EmpresaEventoContable) string {
	base := fmt.Sprintf("%d|%d|%s|%s|%s|%s|%s|%.2f",
		evento.EmpresaID,
		evento.ID,
		normalizeEventoContableModulo(evento.Modulo),
		normalizeEventoContableNombre(evento.Evento),
		strings.TrimSpace(strings.ToLower(evento.DocumentoTipo)),
		strings.TrimSpace(evento.DocumentoCodigo),
		normalizePeriodoEventoContable(evento.PeriodoContable),
		evento.MontoTotal,
	)
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	if strings.Contains(msg, "unique constraint failed") {
		return true
	}
	return strings.Contains(msg, "duplicate key value violates unique constraint")
}

func trimProcessingError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "error de procesamiento"
	}
	const maxLen = 450
	if len(msg) > maxLen {
		return msg[:maxLen]
	}
	return msg
}

func isEventoContableContratado(modulo, evento string) bool {
	eventos, ok := empresaEventoContableContrato[modulo]
	if !ok {
		return false
	}
	_, ok = eventos[evento]
	return ok
}

func normalizeEventoContableModulo(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizeEventoContableNombre(v string) string {
	return strings.TrimSpace(strings.ToLower(v))
}

func normalizePeriodoEventoContable(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	if len(v) >= 7 && v[4] == '-' {
		return v[:7]
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, v); err == nil {
			return parsed.Format("2006-01")
		}
	}
	return ""
}
