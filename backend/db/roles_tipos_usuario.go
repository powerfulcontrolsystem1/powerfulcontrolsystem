package db

import "database/sql"

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

// CreateRolDeUsuario crea un rol de usuario para un tipo de empresa.
func CreateRolDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, nombre, descripcion, usuarioCreador string) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO roles_de_usuario (
		tipo_empresa_id, nombre, descripcion, usuario_creador, estado, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, 'activo', datetime('now','localtime'), datetime('now','localtime'))`, tipoEmpresaID, nombre, descripcion, usuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetRolesDeUsuario obtiene roles de usuario, con filtro opcional por tipo de empresa.
func GetRolesDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, incluirInactivos bool) ([]RolDeUsuario, error) {
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

	rows, err := dbConn.Query(query, args...)
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
	_, err := dbConn.Exec(`UPDATE roles_de_usuario
		SET tipo_empresa_id = ?, nombre = ?, descripcion = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ?`, tipoEmpresaID, nombre, descripcion, id)
	return err
}

// DeleteRolDeUsuario elimina un rol de usuario.
func DeleteRolDeUsuario(dbConn *sql.DB, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM roles_de_usuario WHERE id = ?`, id)
	return err
}

// SetRolDeUsuarioEstado activa/desactiva un rol de usuario.
func SetRolDeUsuarioEstado(dbConn *sql.DB, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE roles_de_usuario SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?`, estado, id)
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
