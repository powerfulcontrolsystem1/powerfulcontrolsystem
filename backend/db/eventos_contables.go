package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// EmpresaEventoContable representa un evento de negocio listo para integracion contable.
type EmpresaEventoContable struct {
	ID                 int64   `json:"id"`
	EmpresaID          int64   `json:"empresa_id"`
	Modulo             string  `json:"modulo"`
	Evento             string  `json:"evento"`
	Entidad            string  `json:"entidad"`
	EntidadID          int64   `json:"entidad_id"`
	DocumentoTipo      string  `json:"documento_tipo"`
	DocumentoCodigo    string  `json:"documento_codigo"`
	PeriodoContable    string  `json:"periodo_contable"`
	MontoTotal         float64 `json:"monto_total"`
	Moneda             string  `json:"moneda"`
	PayloadJSON        string  `json:"payload_json"`
	Origen             string  `json:"origen"`
	FechaEvento        string  `json:"fecha_evento"`
	Procesado          bool    `json:"procesado"`
	FechaProcesado     string  `json:"fecha_procesado"`
	FechaCreacion      string  `json:"fecha_creacion"`
	FechaActualizacion string  `json:"fecha_actualizacion"`
	UsuarioCreador     string  `json:"usuario_creador"`
	Estado             string  `json:"estado"`
	Observaciones      string  `json:"observaciones"`
}

// EmpresaEventoContableFilter permite consultar eventos contables por empresa.
type EmpresaEventoContableFilter struct {
	Modulo          string
	Evento          string
	PeriodoContable string
	IncludeInactive bool
	Limit           int
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
	},
	"compras": {
		"orden_compra_emitida":  {},
		"compra_recepcionada":   {},
		"compra_contabilizada":  {},
		"proveedor_registrado":  {},
		"proveedor_actualizado": {},
		"proveedor_activado":    {},
		"proveedor_desactivado": {},
		"proveedor_eliminado":   {},
	},
	"finanzas": {
		"movimiento_ingreso_registrado": {},
		"movimiento_egreso_registrado":  {},
		"periodo_contable_cerrado":      {},
		"periodo_contable_reabierto":    {},
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
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_empresa_fecha ON empresa_eventos_contables(empresa_id, fecha_evento DESC, id DESC);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_empresa_modulo_evento ON empresa_eventos_contables(empresa_id, modulo, evento);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_pendientes ON empresa_eventos_contables(empresa_id, procesado, fecha_evento);`,
	}
	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
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
	if _, err := dbConn.Exec(`CREATE INDEX IF NOT EXISTS ix_empresa_eventos_contables_empresa_periodo ON empresa_eventos_contables(empresa_id, periodo_contable, estado);`); err != nil {
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

	res, err := dbConn.Exec(`INSERT INTO empresa_eventos_contables (
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
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, NULL, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?)`,
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
	return res.LastInsertId()
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

	rows, err := dbConn.Query(query, args...)
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
