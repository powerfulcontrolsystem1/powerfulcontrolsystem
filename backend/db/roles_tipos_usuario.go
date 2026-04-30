package db

import (
	"database/sql"
	"errors"
	"strings"
)

// RolDeUsuario define un rol configurable por tipo de empresa.
type RolDeUsuario struct {
	ID                 int64  `json:"id"`
	TipoEmpresaID      int64  `json:"tipo_empresa_id"`
	TipoEmpresaNombre  string `json:"tipo_empresa_nombre,omitempty"`
	Nombre             string `json:"nombre"`
	Descripcion        string `json:"descripcion,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// EnsureRolesDeUsuarioSchema crea/migra la tabla base de roles por tipo de empresa.
func EnsureRolesDeUsuarioSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is nil")
	}
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS roles_de_usuario (
			id BIGSERIAL PRIMARY KEY,
			tipo_empresa_id BIGINT NOT NULL,
			nombre TEXT NOT NULL,
			descripcion TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_tipo ON roles_de_usuario(tipo_empresa_id);`,
		`CREATE INDEX IF NOT EXISTS ix_roles_de_usuario_tipo_nombre ON roles_de_usuario(tipo_empresa_id, nombre);`,
	}
	for _, stmt := range stmts {
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	for _, col := range []struct {
		name string
		def  string
	}{
		{"tipo_empresa_id", "BIGINT DEFAULT 0"},
		{"nombre", "TEXT"},
		{"descripcion", "TEXT"},
		{"fecha_creacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"fecha_actualizacion", "TEXT DEFAULT (datetime('now','localtime'))"},
		{"usuario_creador", "TEXT"},
		{"estado", "TEXT DEFAULT 'activo'"},
		{"observaciones", "TEXT"},
	} {
		if err := ensureColumnIfMissing(dbConn, "roles_de_usuario", col.name, col.def); err != nil {
			return err
		}
	}
	return nil
}

// CreateRolDeUsuario crea un rol de usuario para un tipo de empresa.
func CreateRolDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, nombre, descripcion, usuarioCreador string) (int64, error) {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return 0, err
	}
	nombre = strings.TrimSpace(nombre)
	if tipoEmpresaID <= 0 || nombre == "" {
		return 0, errors.New("tipo_empresa_id y nombre son obligatorios")
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO roles_de_usuario (
		tipo_empresa_id, nombre, descripcion, usuario_creador, estado, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, 'activo', datetime('now','localtime'), datetime('now','localtime'))`, tipoEmpresaID, nombre, descripcion, usuarioCreador)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// UpsertRolDeUsuarioByTipoNombre crea o reactiva un rol por tipo de empresa y nombre.
func UpsertRolDeUsuarioByTipoNombre(dbConn *sql.DB, tipoEmpresaID int64, nombre, descripcion, usuarioCreador string) (int64, bool, error) {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return 0, false, err
	}
	nombre = strings.TrimSpace(nombre)
	descripcion = strings.TrimSpace(descripcion)
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if usuarioCreador == "" {
		usuarioCreador = "sistema.roles"
	}
	if tipoEmpresaID <= 0 || nombre == "" {
		return 0, false, errors.New("tipo_empresa_id y nombre son obligatorios")
	}
	var id int64
	err := queryRowSQLCompat(dbConn, `SELECT id
		FROM roles_de_usuario
		WHERE tipo_empresa_id = ? AND lower(trim(nombre)) = lower(trim(?))
		ORDER BY id DESC
		LIMIT 1`, tipoEmpresaID, nombre).Scan(&id)
	if err == nil {
		_, err = execSQLCompat(dbConn, `UPDATE roles_de_usuario
			SET descripcion = COALESCE(NULLIF(?, ''), descripcion),
				estado = 'activo',
				usuario_creador = COALESCE(NULLIF(?, ''), usuario_creador),
				fecha_actualizacion = datetime('now','localtime')
			WHERE id = ?`, descripcion, usuarioCreador, id)
		return id, false, err
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, false, err
	}
	id, err = CreateRolDeUsuario(dbConn, tipoEmpresaID, nombre, descripcion, usuarioCreador)
	return id, true, err
}

// GetRolesDeUsuario obtiene roles de usuario, con filtro opcional por tipo de empresa.
func GetRolesDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, incluirInactivos bool) ([]RolDeUsuario, error) {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT
		r.id,
		r.tipo_empresa_id,
		COALESCE(t.nombre, ''),
		r.nombre,
		COALESCE(r.descripcion, ''),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM roles_de_usuario r
	LEFT JOIN tipos_de_empresas t ON t.id = r.tipo_empresa_id
	WHERE 1 = 1`
	args := make([]interface{}, 0)

	if tipoEmpresaID > 0 {
		query += ` AND r.tipo_empresa_id = ?`
		args = append(args, tipoEmpresaID)
	}
	if !incluirInactivos {
		query += ` AND COALESCE(r.estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY r.id DESC`

	rows, err := querySQLCompat(dbConn, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]RolDeUsuario, 0)
	for rows.Next() {
		var item RolDeUsuario
		if err := rows.Scan(
			&item.ID,
			&item.TipoEmpresaID,
			&item.TipoEmpresaNombre,
			&item.Nombre,
			&item.Descripcion,
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
	return out, nil
}

// UpdateRolDeUsuario actualiza un rol de usuario.
func UpdateRolDeUsuario(dbConn *sql.DB, id, tipoEmpresaID int64, nombre, descripcion string) error {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `UPDATE roles_de_usuario
		SET tipo_empresa_id = ?, nombre = ?, descripcion = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ?`, tipoEmpresaID, nombre, descripcion, id)
	return err
}

// DeleteRolDeUsuario elimina un rol de usuario.
func DeleteRolDeUsuario(dbConn *sql.DB, id int64) error {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `DELETE FROM roles_de_usuario WHERE id = ?`, id)
	return err
}

// SetRolDeUsuarioEstado activa/desactiva un rol de usuario.
func SetRolDeUsuarioEstado(dbConn *sql.DB, id int64, estado string) error {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return err
	}
	_, err := execSQLCompat(dbConn, `UPDATE roles_de_usuario SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?`, estado, id)
	return err
}

// DropTiposDeUsuarioTable elimina la tabla legada `tipos_de_usuario` (modulo retirado del producto).
func DropTiposDeUsuarioTable(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	_, err := execSQLCompat(dbConn, `DROP TABLE IF EXISTS tipos_de_usuario`)
	return err
}
