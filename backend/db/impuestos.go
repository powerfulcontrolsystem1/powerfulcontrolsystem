package db

import (
	"database/sql"
	"fmt"
	"strings"
)

// EmpresaImpuestoConfig representa un impuesto configurable por empresa.
// Nota: es un catálogo operativo (habilitar/deshabilitar + tasa por defecto). No reemplaza el cálculo por item/factura.
type EmpresaImpuestoConfig struct {
	ID               int64   `json:"id"`
	EmpresaID         int64   `json:"empresa_id"`
	PaisCodigo        string  `json:"pais_codigo"`
	Codigo            string  `json:"codigo"`
	Nombre            string  `json:"nombre"`
	Tipo              string  `json:"tipo"` // impuesto | retencion
	TasaPorcentaje    float64 `json:"tasa_porcentaje"`
	Habilitado        int     `json:"habilitado"`
	AplicaEn          string  `json:"aplica_en,omitempty"` // ventas | compras | ambos
	FechaCreacion     string  `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador    string  `json:"usuario_creador,omitempty"`
	Estado            string  `json:"estado,omitempty"`
	Observaciones     string  `json:"observaciones,omitempty"`
}

func normalizeImpuestoTipo(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if v != "retencion" {
		return "impuesto"
	}
	return v
}

func normalizeAplicaEn(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "ventas", "compras", "ambos":
		return v
	default:
		return "ventas"
	}
}

func EnsureEmpresaImpuestosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return fmt.Errorf("db nil")
	}
	query := `
	CREATE TABLE IF NOT EXISTS empresa_impuestos_config (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		pais_codigo TEXT DEFAULT 'CO',
		codigo TEXT NOT NULL,
		nombre TEXT NOT NULL,
		tipo TEXT DEFAULT 'impuesto',
		tasa_porcentaje REAL DEFAULT 0,
		habilitado INTEGER DEFAULT 1,
		aplica_en TEXT DEFAULT 'ventas',
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT,
		UNIQUE(empresa_id, codigo)
	);`
	if shouldUsePostgresCompat(dbConn) {
		query = `
		CREATE TABLE IF NOT EXISTS empresa_impuestos_config (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT NOT NULL,
			pais_codigo TEXT DEFAULT 'CO',
			codigo TEXT NOT NULL,
			nombre TEXT NOT NULL,
			tipo TEXT DEFAULT 'impuesto',
			tasa_porcentaje DOUBLE PRECISION DEFAULT 0,
			habilitado INTEGER DEFAULT 1,
			aplica_en TEXT DEFAULT 'ventas',
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT,
			UNIQUE(empresa_id, codigo)
		);`
	}
	if _, err := execSQLCompat(dbConn, query); err != nil {
		return err
	}
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "pais_codigo", "TEXT DEFAULT 'CO'")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "tipo", "TEXT DEFAULT 'impuesto'")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "tasa_porcentaje", "DOUBLE PRECISION DEFAULT 0")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "habilitado", "INTEGER DEFAULT 1")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "aplica_en", "TEXT DEFAULT 'ventas'")
	_ = ensureColumnIfMissing(dbConn, "empresa_impuestos_config", "fecha_actualizacion", "TEXT")
	return nil
}

func ListEmpresaImpuestos(dbConn *sql.DB, empresaID int64) ([]EmpresaImpuestoConfig, error) {
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return nil, err
	}
	rows, err := querySQLCompat(dbConn, `SELECT
		id, empresa_id,
		COALESCE(pais_codigo,'CO'),
		COALESCE(codigo,''),
		COALESCE(nombre,''),
		COALESCE(tipo,'impuesto'),
		COALESCE(tasa_porcentaje,0),
		COALESCE(habilitado,1),
		COALESCE(aplica_en,'ventas'),
		COALESCE(fecha_creacion,''),
		COALESCE(fecha_actualizacion,''),
		COALESCE(usuario_creador,''),
		COALESCE(estado,'activo'),
		COALESCE(observaciones,'')
	FROM empresa_impuestos_config
	WHERE empresa_id = ?
	ORDER BY tipo ASC, codigo ASC`, empresaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]EmpresaImpuestoConfig, 0)
	for rows.Next() {
		var it EmpresaImpuestoConfig
		if err := rows.Scan(
			&it.ID, &it.EmpresaID, &it.PaisCodigo, &it.Codigo, &it.Nombre, &it.Tipo, &it.TasaPorcentaje,
			&it.Habilitado, &it.AplicaEn, &it.FechaCreacion, &it.FechaActualizacion, &it.UsuarioCreador, &it.Estado, &it.Observaciones,
		); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, nil
}

func UpsertEmpresaImpuesto(dbConn *sql.DB, payload EmpresaImpuestoConfig) (int64, error) {
	if err := EnsureEmpresaImpuestosSchema(dbConn); err != nil {
		return 0, err
	}
	payload.Codigo = strings.ToUpper(strings.TrimSpace(payload.Codigo))
	payload.Nombre = strings.TrimSpace(payload.Nombre)
	payload.PaisCodigo = strings.ToUpper(strings.TrimSpace(payload.PaisCodigo))
	payload.Tipo = normalizeImpuestoTipo(payload.Tipo)
	payload.AplicaEn = normalizeAplicaEn(payload.AplicaEn)
	if payload.Habilitado != 1 {
		payload.Habilitado = 0
	}
	if payload.EmpresaID <= 0 || payload.Codigo == "" || payload.Nombre == "" {
		return 0, fmt.Errorf("empresa_id, codigo y nombre son obligatorios")
	}

	// Upsert por UNIQUE(empresa_id, codigo).
	id, err := insertSQLCompat(dbConn, `INSERT INTO empresa_impuestos_config (
		empresa_id, pais_codigo, codigo, nombre, tipo, tasa_porcentaje, habilitado, aplica_en,
		usuario_creador, estado, observaciones, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'))
	ON CONFLICT (empresa_id, codigo) DO UPDATE SET
		pais_codigo = excluded.pais_codigo,
		nombre = excluded.nombre,
		tipo = excluded.tipo,
		tasa_porcentaje = excluded.tasa_porcentaje,
		habilitado = excluded.habilitado,
		aplica_en = excluded.aplica_en,
		observaciones = excluded.observaciones,
		fecha_actualizacion = datetime('now','localtime')
	RETURNING id`,
		payload.EmpresaID, payload.PaisCodigo, payload.Codigo, payload.Nombre, payload.Tipo, payload.TasaPorcentaje, payload.Habilitado, payload.AplicaEn,
		strings.TrimSpace(payload.UsuarioCreador), normalizeChatEstado(payload.Estado), strings.TrimSpace(payload.Observaciones),
	)
	if err != nil {
		// Algunas versiones de insertSQLCompat ya agregan RETURNING; si el query ya lo trae, se respeta.
		return 0, err
	}
	return id, nil
}

