package db

import (
	"database/sql"
	"fmt"
	"time"
)

// EmpresaUsuario representa un usuario gestionado dentro del contexto de una empresa.
type EmpresaUsuario struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id"`
	Email              string `json:"email"`
	Nombre             string `json:"nombre"`
	DocumentoIdentidad string `json:"documento_identidad,omitempty"`
	PasswordHash       string `json:"-"`
	PasswordSalt       string `json:"-"`
	PasswordSet        int    `json:"password_set,omitempty"`
	RolUsuarioID       int64  `json:"rol_usuario_id"`
	RolNombre          string `json:"rol_nombre,omitempty"`
	EmailConfirmado    int    `json:"email_confirmado"`
	EmailConfirmadoEn  string `json:"email_confirmado_en,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// CreateEmpresaUsuario crea un usuario de empresa en estado pendiente de confirmación de correo.
func CreateEmpresaUsuario(
	dbConn *sql.DB,
	empresaID int64,
	email, nombre, documentoIdentidad string,
	rolUsuarioID int64,
	rolNombre, observaciones, usuarioCreador, confirmToken, confirmExpira string,
) (int64, error) {
	res, err := dbConn.Exec(`INSERT INTO users (
		email,
		name,
		role,
		empresa_id,
		documento_identidad,
		rol_usuario_id,
		email_confirmado,
		email_confirm_token,
		email_confirm_expira,
		usuario_creador,
		estado,
		observaciones,
		fecha_creacion,
		fecha_actualizacion
	) VALUES (?, ?, ?, ?, ?, ?, 0, ?, ?, ?, 'inactivo', ?, datetime('now','localtime'), datetime('now','localtime'))`,
		email,
		nombre,
		rolNombre,
		empresaID,
		documentoIdentidad,
		rolUsuarioID,
		confirmToken,
		confirmExpira,
		usuarioCreador,
		observaciones,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// GetEmpresaUsuarios lista usuarios por empresa.
func GetEmpresaUsuarios(dbConn *sql.DB, empresaID int64, incluirInactivos bool) ([]EmpresaUsuario, error) {
	query := `SELECT
		id,
		empresa_id,
		email,
		COALESCE(name, ''),
		COALESCE(documento_identidad, ''),
		COALESCE(rol_usuario_id, 0),
		COALESCE(role, ''),
		COALESCE(email_confirmado, 0),
		COALESCE(email_confirmado_en, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM users
	WHERE empresa_id = ?`
	args := []interface{}{empresaID}

	if !incluirInactivos {
		query += ` AND COALESCE(estado, 'activo') = 'activo'`
	}
	query += ` ORDER BY id DESC`

	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaUsuario, 0)
	for rows.Next() {
		var item EmpresaUsuario
		if err := rows.Scan(
			&item.ID,
			&item.EmpresaID,
			&item.Email,
			&item.Nombre,
			&item.DocumentoIdentidad,
			&item.RolUsuarioID,
			&item.RolNombre,
			&item.EmailConfirmado,
			&item.EmailConfirmadoEn,
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

// GetEmpresaUsuarioByID obtiene un usuario por id dentro de una empresa.
func GetEmpresaUsuarioByID(dbConn *sql.DB, empresaID, id int64) (*EmpresaUsuario, error) {
	row := dbConn.QueryRow(`SELECT
		id,
		empresa_id,
		email,
		COALESCE(name, ''),
		COALESCE(documento_identidad, ''),
		COALESCE(rol_usuario_id, 0),
		COALESCE(role, ''),
		COALESCE(email_confirmado, 0),
		COALESCE(email_confirmado_en, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM users
	WHERE empresa_id = ? AND id = ?
	LIMIT 1`, empresaID, id)

	var item EmpresaUsuario
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Email,
		&item.Nombre,
		&item.DocumentoIdentidad,
		&item.RolUsuarioID,
		&item.RolNombre,
		&item.EmailConfirmado,
		&item.EmailConfirmadoEn,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

// GetEmpresaUsuarioByEmailScoped obtiene un usuario por correo con alcance opcional por empresa.
func GetEmpresaUsuarioByEmailScoped(dbConn *sql.DB, email string, empresaID int64) (*EmpresaUsuario, error) {
	query := `SELECT
		id,
		empresa_id,
		email,
		COALESCE(name, ''),
		COALESCE(documento_identidad, ''),
		COALESCE(password_hash, ''),
		COALESCE(password_salt, ''),
		COALESCE(password_set, 0),
		COALESCE(rol_usuario_id, 0),
		COALESCE(role, ''),
		COALESCE(email_confirmado, 0),
		COALESCE(email_confirmado_en, ''),
		COALESCE(fecha_creacion, ''),
		COALESCE(fecha_actualizacion, ''),
		COALESCE(usuario_creador, ''),
		COALESCE(estado, 'activo'),
		COALESCE(observaciones, '')
	FROM users
	WHERE lower(email) = lower(?)`
	args := []interface{}{email}
	if empresaID > 0 {
		query += " AND empresa_id = ?"
		args = append(args, empresaID)
	}
	query += " LIMIT 1"

	row := dbConn.QueryRow(query, args...)

	var item EmpresaUsuario
	if err := row.Scan(
		&item.ID,
		&item.EmpresaID,
		&item.Email,
		&item.Nombre,
		&item.DocumentoIdentidad,
		&item.PasswordHash,
		&item.PasswordSalt,
		&item.PasswordSet,
		&item.RolUsuarioID,
		&item.RolNombre,
		&item.EmailConfirmado,
		&item.EmailConfirmadoEn,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	); err != nil {
		return nil, err
	}
	return &item, nil
}

// GetEmpresaUsuarioByEmail obtiene un usuario por correo (case-insensitive).
func GetEmpresaUsuarioByEmail(dbConn *sql.DB, email string) (*EmpresaUsuario, error) {
	return GetEmpresaUsuarioByEmailScoped(dbConn, email, 0)
}

// SetEmpresaUsuarioPassword define la contraseña de acceso para un usuario de empresa.
func SetEmpresaUsuarioPassword(dbConn *sql.DB, empresaID, id int64, passwordHash, passwordSalt string) error {
	_, err := dbConn.Exec(`UPDATE users
		SET password_hash = ?,
			password_salt = ?,
			password_set = 1,
			password_actualizada_en = datetime('now','localtime'),
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, passwordHash, passwordSalt, id, empresaID)
	return err
}

// UpdateEmpresaUsuario actualiza los datos de un usuario de empresa.
func UpdateEmpresaUsuario(
	dbConn *sql.DB,
	id, empresaID int64,
	email, nombre, documentoIdentidad string,
	rolUsuarioID int64,
	rolNombre, observaciones string,
	resetConfirmacion bool,
	confirmToken, confirmExpira string,
) error {
	if resetConfirmacion {
		_, err := dbConn.Exec(`UPDATE users
			SET email = ?,
				name = ?,
				documento_identidad = ?,
				rol_usuario_id = ?,
				role = ?,
				observaciones = ?,
				email_confirmado = 0,
				email_confirmado_en = '',
				estado = 'inactivo',
				email_confirm_token = ?,
				email_confirm_expira = ?,
				fecha_actualizacion = datetime('now','localtime')
			WHERE id = ? AND empresa_id = ?`,
			email,
			nombre,
			documentoIdentidad,
			rolUsuarioID,
			rolNombre,
			observaciones,
			confirmToken,
			confirmExpira,
			id,
			empresaID,
		)
		return err
	}

	_, err := dbConn.Exec(`UPDATE users
		SET email = ?,
			name = ?,
			documento_identidad = ?,
			rol_usuario_id = ?,
			role = ?,
			observaciones = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		email,
		nombre,
		documentoIdentidad,
		rolUsuarioID,
		rolNombre,
		observaciones,
		id,
		empresaID,
	)
	return err
}

// DeleteEmpresaUsuario elimina un usuario de empresa.
func DeleteEmpresaUsuario(dbConn *sql.DB, empresaID, id int64) error {
	_, err := dbConn.Exec(`DELETE FROM users WHERE id = ? AND empresa_id = ?`, id, empresaID)
	return err
}

// SetEmpresaUsuarioEstado activa o desactiva un usuario de empresa.
func SetEmpresaUsuarioEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	_, err := dbConn.Exec(`UPDATE users SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, estado, id, empresaID)
	return err
}

// SetEmpresaUsuarioConfirmToken actualiza token de confirmación para reenvíos.
func SetEmpresaUsuarioConfirmToken(dbConn *sql.DB, empresaID, id int64, confirmToken, confirmExpira string) error {
	_, err := dbConn.Exec(`UPDATE users
		SET email_confirm_token = ?,
			email_confirm_expira = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, confirmToken, confirmExpira, id, empresaID)
	return err
}

// ConfirmEmpresaUsuarioByToken confirma el correo de un usuario usando su token.
func ConfirmEmpresaUsuarioByToken(dbConn *sql.DB, token string) (int64, error) {
	row := dbConn.QueryRow(`SELECT id, empresa_id, COALESCE(email_confirm_expira, '') FROM users WHERE email_confirm_token = ? LIMIT 1`, token)
	var id int64
	var empresaID int64
	var expiraRaw string
	if err := row.Scan(&id, &empresaID, &expiraRaw); err != nil {
		return 0, err
	}

	if expiraRaw != "" {
		expiraAt, err := time.ParseInLocation("2006-01-02 15:04:05", expiraRaw, time.Local)
		if err == nil && time.Now().After(expiraAt) {
			return 0, fmt.Errorf("token de confirmacion expirado")
		}
	}

	_, err := dbConn.Exec(`UPDATE users
		SET email_confirmado = 1,
			email_confirmado_en = datetime('now','localtime'),
			estado = 'activo',
			email_confirm_token = '',
			email_confirm_expira = '',
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return empresaID, nil
}
