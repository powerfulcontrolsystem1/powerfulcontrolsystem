package db

import (
	"database/sql"
	"encoding/json"
	"strings"
)

// EstacionColumnPreferences guarda las columnas visibles por empresa/usuario/rol
type EstacionColumnPreferences struct {
	ID                 int64           `json:"id"`
	EmpresaID          int64           `json:"empresa_id"`
	UsuarioEmail       string          `json:"usuario_email,omitempty"`
	RolID              int64           `json:"rol_id,omitempty"`
	ColumnasJSON       string          `json:"columnas_json,omitempty"`
	Columnas           map[string]bool `json:"columnas,omitempty"`
	FechaCreacion      string          `json:"fecha_creacion,omitempty"`
	FechaActualizacion string          `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string          `json:"usuario_creador,omitempty"`
	Estado             string          `json:"estado,omitempty"`
}

// EnsureEmpresaEstacionColumnPreferencesSchema crea/migra la tabla de preferencias de columnas por estacion
func EnsureEmpresaEstacionColumnPreferencesSchema(dbConn *sql.DB) error {
	if SchemaBootstrapDisabled() {
		return nil
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS estacion_column_preferences (
            id BIGSERIAL PRIMARY KEY,
            empresa_id INTEGER NOT NULL,
            usuario_email TEXT,
            rol_id INTEGER,
            columnas TEXT NOT NULL,
            fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
            fecha_actualizacion TEXT,
            usuario_creador TEXT,
            estado TEXT DEFAULT 'activo',
            observaciones TEXT
        );`,
		`CREATE UNIQUE INDEX IF NOT EXISTS ux_estacion_column_prefs_empresa_usuario_rol ON estacion_column_preferences(empresa_id, usuario_email, rol_id);`,
		`CREATE INDEX IF NOT EXISTS ix_estacion_column_prefs_empresa ON estacion_column_preferences(empresa_id);`,
	}

	for _, stmt := range stmts {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "empresa_id", "INTEGER NOT NULL"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "usuario_email", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "rol_id", "INTEGER"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "columnas", "TEXT NOT NULL DEFAULT '{}'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "fecha_creacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "fecha_actualizacion", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "usuario_creador", "TEXT"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "estado", "TEXT DEFAULT 'activo'"); err != nil {
		return err
	}
	if err := ensureColumnIfMissing(dbConn, "estacion_column_preferences", "observaciones", "TEXT"); err != nil {
		return err
	}

	return nil
}

// GetEstacionColumnPreferences obtiene las preferencias (intenta usuario -> company default)
func GetEstacionColumnPreferences(dbConn *sql.DB, empresaID int64, usuarioEmail string, rolID int64) (*EstacionColumnPreferences, error) {
	usuario := strings.TrimSpace(strings.ToLower(usuarioEmail))
	var p EstacionColumnPreferences
	var columnas string

	// Intentar match usuario + rol
	if usuario != "" {
		q := `SELECT id, empresa_id, usuario_email, rol_id, columnas, fecha_creacion, fecha_actualizacion, usuario_creador, estado
              FROM estacion_column_preferences
              WHERE empresa_id = ? AND lower(usuario_email) = ? AND (rol_id = ? OR rol_id IS NULL) AND estado = 'activo' LIMIT 1`
		row := dbConn.QueryRow(q, empresaID, usuario, rolID)
		if err := row.Scan(&p.ID, &p.EmpresaID, &p.UsuarioEmail, &p.RolID, &columnas, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado); err == nil {
			p.ColumnasJSON = columnas
			var mp map[string]bool
			_ = json.Unmarshal([]byte(columnas), &mp)
			p.Columnas = mp
			return &p, nil
		}
	}

	// Intentar company default (usuario_email IS NULL and rol_id IS NULL)
	q := `SELECT id, empresa_id, usuario_email, rol_id, columnas, fecha_creacion, fecha_actualizacion, usuario_creador, estado
          FROM estacion_column_preferences
          WHERE empresa_id = ? AND usuario_email IS NULL AND rol_id IS NULL AND estado = 'activo' LIMIT 1`
	row := dbConn.QueryRow(q, empresaID)
	if err := row.Scan(&p.ID, &p.EmpresaID, &p.UsuarioEmail, &p.RolID, &columnas, &p.FechaCreacion, &p.FechaActualizacion, &p.UsuarioCreador, &p.Estado); err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, err
	}
	p.ColumnasJSON = columnas
	var mp map[string]bool
	_ = json.Unmarshal([]byte(columnas), &mp)
	p.Columnas = mp
	return &p, nil
}

// UpsertEstacionColumnPreferences crea o actualiza preferencias para empresa/usuario/rol
func UpsertEstacionColumnPreferences(dbConn *sql.DB, p *EstacionColumnPreferences) (int64, error) {
	usuario := strings.TrimSpace(strings.ToLower(p.UsuarioEmail))

	// Buscar existente
	var existingID int64
	if usuario != "" {
		row := dbConn.QueryRow(`SELECT id FROM estacion_column_preferences WHERE empresa_id = ? AND lower(usuario_email) = ? AND (rol_id = ? OR rol_id IS NULL) LIMIT 1`, p.EmpresaID, usuario, p.RolID)
		if err := row.Scan(&existingID); err != nil && err != sql.ErrNoRows {
			return 0, err
		}
	} else if p.RolID > 0 {
		row := dbConn.QueryRow(`SELECT id FROM estacion_column_preferences WHERE empresa_id = ? AND usuario_email IS NULL AND rol_id = ? LIMIT 1`, p.EmpresaID, p.RolID)
		if err := row.Scan(&existingID); err != nil && err != sql.ErrNoRows {
			return 0, err
		}
	} else {
		row := dbConn.QueryRow(`SELECT id FROM estacion_column_preferences WHERE empresa_id = ? AND usuario_email IS NULL AND rol_id IS NULL LIMIT 1`, p.EmpresaID)
		if err := row.Scan(&existingID); err != nil && err != sql.ErrNoRows {
			return 0, err
		}
	}

	columnas := strings.TrimSpace(p.ColumnasJSON)
	if columnas == "" && p.Columnas != nil {
		by, _ := json.Marshal(p.Columnas)
		columnas = string(by)
	}
	if columnas == "" {
		columnas = "{}"
	}

	if existingID > 0 {
		_, err := dbConn.Exec(`UPDATE estacion_column_preferences SET columnas = ?, fecha_actualizacion = CURRENT_TIMESTAMP, usuario_creador = ?, estado = COALESCE(NULLIF(?, ''), 'activo') WHERE id = ?`, columnas, p.UsuarioCreador, p.Estado, existingID)
		if err != nil {
			return 0, err
		}
		return existingID, nil
	}

	// Insert
	res, err := dbConn.Exec(`INSERT INTO estacion_column_preferences (empresa_id, usuario_email, rol_id, columnas, fecha_creacion, fecha_actualizacion, usuario_creador, estado) VALUES (?, NULLIF(?, ''), NULLIF(?, 0), ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?, COALESCE(NULLIF(?, ''), 'activo'))`, p.EmpresaID, usuario, p.RolID, columnas, p.UsuarioCreador, p.Estado)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
