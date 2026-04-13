package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// EmpresaEstacionPref representa una preferencia por estación y empresa.
type EmpresaEstacionPref struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	EstacionID         int64  `json:"estacion_id"`
	Clave              string `json:"clave"`
	Valor              string `json:"valor"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EnsureEmpresaEstacionPrefsSchema crea/migra la tabla de preferencias por estacion.
func EnsureEmpresaEstacionPrefsSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS empresa_estacion_prefs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            empresa_id INTEGER NOT NULL,
            estacion_id INTEGER NOT NULL,
            clave TEXT NOT NULL,
            valor TEXT,
            fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
            fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
            usuario_creador TEXT,
            estado TEXT DEFAULT 'activo',
            observaciones TEXT
        );`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_empresa_estacion_prefs_empresa_estacion_clave ON empresa_estacion_prefs(empresa_id, estacion_id, clave);`,
		`CREATE INDEX IF NOT EXISTS ix_empresa_estacion_prefs_empresa_estacion ON empresa_estacion_prefs(empresa_id, estacion_id);`,
	}

	for _, s := range stmts {
		if _, err := dbConn.Exec(s); err != nil {
			return err
		}
	}

	// ensure common columns in case of incremental migrations
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "empresa_estacion_prefs", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// ListEmpresaEstacionPrefs lista preferencias por empresa y opcionalmente por estacion.
func ListEmpresaEstacionPrefs(dbConn *sql.DB, empresaID int64, estacionID int64, includeInactive bool) ([]EmpresaEstacionPref, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}

	where := "WHERE empresa_id = ?"
	args := []interface{}{empresaID}
	if estacionID > 0 {
		where += " AND estacion_id = ?"
		args = append(args, estacionID)
	}
	if !includeInactive {
		where += " AND LOWER(COALESCE(NULLIF(TRIM(estado), ''), 'activo')) = 'activo'"
	}

	q := "SELECT id, empresa_id, estacion_id, COALESCE(clave, ''), COALESCE(valor, ''), COALESCE(fecha_creacion, ''), COALESCE(fecha_actualizacion, ''), COALESCE(usuario_creador, ''), COALESCE(NULLIF(TRIM(estado), ''), 'activo'), COALESCE(observaciones, '') FROM empresa_estacion_prefs " + where + " ORDER BY estacion_id, clave"

	rows, err := dbConn.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaEstacionPref, 0)
	for rows.Next() {
		var p EmpresaEstacionPref
		if err := rows.Scan(&p.ID, &p.EmpresaID, &p.EstacionID, &p.Clave, &p.Valor, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, nil
}

// GetEmpresaEstacionPref obtiene una preferencia por clave (si existe).
func GetEmpresaEstacionPref(dbConn *sql.DB, empresaID int64, estacionID int64, clave string) (*EmpresaEstacionPref, error) {
	if dbConn == nil {
		return nil, errors.New("db connection is nil")
	}
	if empresaID <= 0 {
		return nil, errors.New("empresa_id invalido")
	}
	clave = strings.TrimSpace(clave)
	if clave == "" {
		return nil, errors.New("clave es obligatoria")
	}

	row := dbConn.QueryRow(`SELECT id, empresa_id, estacion_id, COALESCE(clave,''), COALESCE(valor,''), COALESCE(fecha_creacion,''), COALESCE(fecha_actualizacion,''), COALESCE(usuario_creador,''), COALESCE(NULLIF(TRIM(estado), ''),'activo'), COALESCE(observaciones,'') FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = ? AND clave = ? LIMIT 1`, empresaID, estacionID, clave)
	var p EmpresaEstacionPref
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.EstacionID, &p.Clave, &p.Valor, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado, &p.Observaciones); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

// UpsertEmpresaEstacionPref crea o actualiza una preferencia por clave (usa ON CONFLICT para SQLite).
func UpsertEmpresaEstacionPref(dbConn *sql.DB, p EmpresaEstacionPref) (int64, error) {
	if dbConn == nil {
		return 0, errors.New("db connection is nil")
	}
	if p.EmpresaID <= 0 {
		return 0, errors.New("empresa_id invalido")
	}
	if p.EstacionID < 0 {
		p.EstacionID = 0
	}
	p.Clave = strings.TrimSpace(p.Clave)
	if p.Clave == "" {
		return 0, errors.New("clave es obligatoria")
	}
	p.Estado = strings.ToLower(strings.TrimSpace(p.Estado))
	if p.Estado == "" {
		p.Estado = "activo"
	}

	// Use upsert via ON CONFLICT on unique index
	res, err := dbConn.Exec(`INSERT INTO empresa_estacion_prefs (
        empresa_id, estacion_id, clave, valor, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones
    ) VALUES (
        ?, ?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), ?, ?, ?
    ) ON CONFLICT(empresa_id, estacion_id, clave) DO UPDATE SET
        valor = excluded.valor,
        fecha_actualizacion = datetime('now','localtime'),
        usuario_creador = CASE WHEN trim(excluded.usuario_creador) <> '' THEN excluded.usuario_creador ELSE empresa_estacion_prefs.usuario_creador END,
		estado = COALESCE(NULLIF(TRIM(excluded.estado), ''), 'activo'),
        observaciones = excluded.observaciones`,
		p.EmpresaID, p.EstacionID, p.Clave, strings.TrimSpace(p.Valor), strings.TrimSpace(p.UsuarioCreador), p.Estado, strings.TrimSpace(p.Observaciones))
	if err != nil {
		return 0, fmt.Errorf("upsert error: %w", err)
	}

	id, err := res.LastInsertId()
	if err == nil && id > 0 {
		return id, nil
	}
	// If no new insert id, try to fetch existing id
	var existing int64
	if err := dbConn.QueryRow(`SELECT id FROM empresa_estacion_prefs WHERE empresa_id = ? AND estacion_id = ? AND clave = ? LIMIT 1`, p.EmpresaID, p.EstacionID, p.Clave).Scan(&existing); err != nil {
		return 0, err
	}
	return existing, nil
}
