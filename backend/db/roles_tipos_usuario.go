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

// TipoDeUsuario define tipos de usuario asociados a un rol y tipo de empresa.
type TipoDeUsuario struct {
	ID                 int64  `json:"id"`
	TipoEmpresaID      int64  `json:"tipo_empresa_id"`
	TipoEmpresaNombre  string `json:"tipo_empresa_nombre,omitempty"`
	RolID              int64  `json:"rol_id"`
	RolNombre          string `json:"rol_nombre,omitempty"`
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

// CreateTipoDeUsuario crea un tipo de usuario ligado a un rol y tipo de empresa.
func CreateTipoDeUsuario(dbConn *sql.DB, tipoEmpresaID, rolID int64, nombre, descripcion, usuarioCreador string) (int64, error) {
	if err := validateRolPorTipoEmpresa(dbConn, tipoEmpresaID, rolID); err != nil {
		return 0, err
	}
	res, err := dbConn.Exec(`INSERT INTO tipos_de_usuario (
		tipo_empresa_id, rol_id, nombre, descripcion, usuario_creador, estado, fecha_creacion, fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, 'activo', datetime('now','localtime'), datetime('now','localtime'))`, tipoEmpresaID, rolID, nombre, descripcion, usuarioCreador)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetTiposDeUsuario obtiene tipos de usuario, con filtro opcional por tipo de empresa.
func GetTiposDeUsuario(dbConn *sql.DB, tipoEmpresaID int64, incluirInactivos bool) ([]TipoDeUsuario, error) {
	query := `SELECT
		tu.id,
		tu.tipo_empresa_id,
		COALESCE(te.nombre, ''),
		tu.rol_id,
		COALESCE(ru.nombre, ''),
		tu.nombre,
		COALESCE(tu.descripcion, ''),
		COALESCE(tu.fecha_creacion, ''),
		COALESCE(tu.fecha_actualizacion, ''),
		COALESCE(tu.usuario_creador, ''),
		COALESCE(tu.estado, 'activo'),
		COALESCE(tu.observaciones, '')
	FROM tipos_de_usuario tu
	LEFT JOIN tipos_de_empresas te ON te.id = tu.tipo_empresa_id
	LEFT JOIN roles_de_usuario ru ON ru.id = tu.rol_id
	WHERE 1 = 1`
	args := make([]interface{}, 0)

	if tipoEmpresaID > 0 {
		query += ` AND tu.tipo_empresa_id = ?`
		args = append(args, tipoEmpresaID)
	}
	if !incluirInactivos {
		query += ` AND COALESCE(tu.estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY tu.id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]TipoDeUsuario, 0)
	for rows.Next() {
		var item TipoDeUsuario
		if err := rows.Scan(
			&item.ID,
			&item.TipoEmpresaID,
			&item.TipoEmpresaNombre,
			&item.RolID,
			&item.RolNombre,
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

// UpdateTipoDeUsuario actualiza un tipo de usuario.
func UpdateTipoDeUsuario(dbConn *sql.DB, id, tipoEmpresaID, rolID int64, nombre, descripcion string) error {
	if err := validateRolPorTipoEmpresa(dbConn, tipoEmpresaID, rolID); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE tipos_de_usuario
		SET tipo_empresa_id = ?, rol_id = ?, nombre = ?, descripcion = ?, fecha_actualizacion = datetime('now','localtime')
		WHERE id = ?`, tipoEmpresaID, rolID, nombre, descripcion, id)
	return err
}

// DeleteTipoDeUsuario elimina un tipo de usuario.
func DeleteTipoDeUsuario(dbConn *sql.DB, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM tipos_de_usuario WHERE id = ?`, id)
	return err
}

// SetTipoDeUsuarioEstado activa/desactiva un tipo de usuario.
func SetTipoDeUsuarioEstado(dbConn *sql.DB, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE tipos_de_usuario SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ?`, estado, id)
	return err
}

func validateRolPorTipoEmpresa(dbConn *sql.DB, tipoEmpresaID, rolID int64) error {
	row := dbConn.QueryRow(`SELECT id FROM roles_de_usuario WHERE id = ? AND tipo_empresa_id = ? LIMIT 1`, rolID, tipoEmpresaID)
	var id int64
	if err := row.Scan(&id); err != nil {
		return err
	}
	return nil
}
