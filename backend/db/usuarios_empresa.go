package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	empresaUsuariosAuthSchemaMu    sync.Mutex
	empresaUsuariosAuthSchemaReady bool
)

// EmpresaUsuario representa un usuario gestionado dentro del contexto de una empresa.
type EmpresaUsuario struct {
	ID                       int64  `json:"id"`
	EmpresaID                int64  `json:"empresa_id"`
	Email                    string `json:"email"`
	Nombre                   string `json:"nombre"`
	DocumentoIdentidad       string `json:"documento_identidad,omitempty"`
	PasswordHash             string `json:"-"`
	PasswordSalt             string `json:"-"`
	PasswordSet              int    `json:"password_set,omitempty"`
	PasswordActualizadaEn    string `json:"password_actualizada_en,omitempty"`
	LoginFailedAttempts      int    `json:"-"`
	LoginFailedLastAt        string `json:"-"`
	LoginLockedUntil         string `json:"-"`
	PasswordResetToken       string `json:"-"`
	PasswordResetExpira      string `json:"-"`
	PasswordResetRequestedEn string `json:"-"`
	AceptaContrato           int    `json:"acepta_contrato,omitempty"`
	ContratoVersionAceptada  int    `json:"contrato_version_aceptada,omitempty"`
	FechaAceptaContrato      string `json:"fecha_acepta_contrato,omitempty"`
	RolUsuarioID             int64  `json:"rol_usuario_id"`
	RolNombre                string `json:"rol_nombre,omitempty"`
	EmailConfirmado          int    `json:"email_confirmado"`
	EmailConfirmadoEn        string `json:"email_confirmado_en,omitempty"`
	FechaCreacion            string `json:"fecha_creacion,omitempty"`
	FechaActualizacion       string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador           string `json:"usuario_creador,omitempty"`
	Estado                   string `json:"estado,omitempty"`
	Observaciones            string `json:"observaciones,omitempty"`
}

func EnsureEmpresaUsuariosAuthSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return errors.New("db connection is required")
	}

	empresaUsuariosAuthSchemaMu.Lock()
	defer empresaUsuariosAuthSchemaMu.Unlock()

	if empresaUsuariosAuthSchemaReady {
		return nil
	}

	if isPostgresDialect() {
		if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			email TEXT UNIQUE,
			name TEXT,
			role TEXT DEFAULT 'administrador',
			empresa_id BIGINT,
			documento_identidad TEXT,
			rol_usuario_id BIGINT,
			email_confirmado INTEGER DEFAULT 0,
			email_confirm_token TEXT,
			email_confirm_expira TEXT,
			email_confirmado_en TEXT,
			password_hash TEXT,
			password_salt TEXT,
			password_set INTEGER DEFAULT 0,
			password_actualizada_en TEXT,
			login_failed_attempts INTEGER DEFAULT 0,
			login_failed_last_at TEXT,
			login_locked_until TEXT,
			password_reset_token TEXT,
			password_reset_expira TEXT,
			password_reset_requested_en TEXT,
			acepta_contrato INTEGER DEFAULT 0,
			contrato_version_aceptada INTEGER DEFAULT 0,
			fecha_acepta_contrato TEXT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`); err != nil {
			return err
		}
		if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_users_lower_email_empresa ON users ((lower(email)), empresa_id)`); err != nil {
			return err
		}
		if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_users_email_confirm_token ON users (email_confirm_token)`); err != nil {
			return err
		}
	} else {
		if _, err := execSQLCompat(dbConn, `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			email TEXT UNIQUE,
			name TEXT,
			role TEXT DEFAULT 'administrador',
			empresa_id INTEGER,
			documento_identidad TEXT,
			rol_usuario_id INTEGER,
			email_confirmado INTEGER DEFAULT 0,
			email_confirm_token TEXT,
			email_confirm_expira TEXT,
			email_confirmado_en TEXT,
			password_hash TEXT,
			password_salt TEXT,
			password_set INTEGER DEFAULT 0,
			password_actualizada_en TEXT,
			login_failed_attempts INTEGER DEFAULT 0,
			login_failed_last_at TEXT,
			login_locked_until TEXT,
			password_reset_token TEXT,
			password_reset_expira TEXT,
			password_reset_requested_en TEXT,
			acepta_contrato INTEGER DEFAULT 0,
			contrato_version_aceptada INTEGER DEFAULT 0,
			fecha_acepta_contrato TEXT,
			fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
			fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`); err != nil {
			return err
		}
		if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_users_lower_email_empresa ON users (lower(email), empresa_id)`); err != nil {
			return err
		}
		if _, err := execSQLCompat(dbConn, `CREATE INDEX IF NOT EXISTS ix_users_email_confirm_token ON users (email_confirm_token)`); err != nil {
			return err
		}
	}

	columns := []struct {
		name string
		def  string
	}{
		{name: "documento_identidad", def: "TEXT"},
		{name: "rol_usuario_id", def: "INTEGER"},
		{name: "email_confirmado", def: "INTEGER DEFAULT 0"},
		{name: "email_confirm_token", def: "TEXT"},
		{name: "email_confirm_expira", def: "TEXT"},
		{name: "email_confirmado_en", def: "TEXT"},
		{name: "password_hash", def: "TEXT"},
		{name: "password_salt", def: "TEXT"},
		{name: "password_set", def: "INTEGER DEFAULT 0"},
		{name: "password_actualizada_en", def: "TEXT"},
		{name: "login_failed_attempts", def: "INTEGER DEFAULT 0"},
		{name: "login_failed_last_at", def: "TEXT"},
		{name: "login_locked_until", def: "TEXT"},
		{name: "password_reset_token", def: "TEXT"},
		{name: "password_reset_expira", def: "TEXT"},
		{name: "password_reset_requested_en", def: "TEXT"},
		{name: "acepta_contrato", def: "INTEGER DEFAULT 0"},
		{name: "contrato_version_aceptada", def: "INTEGER DEFAULT 0"},
		{name: "fecha_acepta_contrato", def: "TEXT"},
		{name: "fecha_creacion", def: "TEXT DEFAULT (datetime('now','localtime'))"},
		{name: "fecha_actualizacion", def: "TEXT DEFAULT (datetime('now','localtime'))"},
		{name: "usuario_creador", def: "TEXT"},
		{name: "estado", def: "TEXT DEFAULT 'activo'"},
		{name: "observaciones", def: "TEXT"},
	}

	for _, column := range columns {
		if err := ensureColumnIfMissing(dbConn, "users", column.name, column.def); err != nil {
			return err
		}
	}

	empresaUsuariosAuthSchemaReady = true
	return nil
}

// CreateEmpresaUsuario crea un usuario de empresa en estado pendiente de confirmación de correo.
func CreateEmpresaUsuario(
	dbConn *sql.DB,
	empresaID int64,
	email, nombre, documentoIdentidad string,
	rolUsuarioID int64,
	rolNombre, observaciones, usuarioCreador, confirmToken, confirmExpira string,
) (int64, error) {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return 0, err
	}
	id, err := insertSQLCompat(dbConn, `INSERT INTO users (
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
	return id, nil
}

// GetEmpresaUsuarios lista usuarios por empresa.
func GetEmpresaUsuarios(dbConn *sql.DB, empresaID int64, incluirInactivos bool) ([]EmpresaUsuario, error) {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return nil, err
	}
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
		COALESCE(acepta_contrato, 0),
		COALESCE(contrato_version_aceptada, 0),
		COALESCE(fecha_acepta_contrato, ''),
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
			&item.AceptaContrato,
			&item.ContratoVersionAceptada,
			&item.FechaAceptaContrato,
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
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return nil, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT
		id,
		empresa_id,
		email,
		COALESCE(name, ''),
		COALESCE(documento_identidad, ''),
		COALESCE(rol_usuario_id, 0),
		COALESCE(role, ''),
		COALESCE(email_confirmado, 0),
		COALESCE(email_confirmado_en, ''),
		COALESCE(acepta_contrato, 0),
		COALESCE(contrato_version_aceptada, 0),
		COALESCE(fecha_acepta_contrato, ''),
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
		&item.AceptaContrato,
		&item.ContratoVersionAceptada,
		&item.FechaAceptaContrato,
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
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return nil, err
	}
	query := `SELECT
		id,
		empresa_id,
		email,
		COALESCE(name, ''),
		COALESCE(documento_identidad, ''),
		COALESCE(password_hash, ''),
		COALESCE(password_salt, ''),
		COALESCE(password_set, 0),
		COALESCE(password_actualizada_en, ''),
		COALESCE(login_failed_attempts, 0),
		COALESCE(login_failed_last_at, ''),
		COALESCE(login_locked_until, ''),
		COALESCE(password_reset_token, ''),
		COALESCE(password_reset_expira, ''),
		COALESCE(password_reset_requested_en, ''),
		COALESCE(rol_usuario_id, 0),
		COALESCE(role, ''),
		COALESCE(email_confirmado, 0),
		COALESCE(email_confirmado_en, ''),
		COALESCE(acepta_contrato, 0),
		COALESCE(contrato_version_aceptada, 0),
		COALESCE(fecha_acepta_contrato, ''),
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

	row := queryRowSQLCompat(dbConn, query, args...)

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
		&item.PasswordActualizadaEn,
		&item.LoginFailedAttempts,
		&item.LoginFailedLastAt,
		&item.LoginLockedUntil,
		&item.PasswordResetToken,
		&item.PasswordResetExpira,
		&item.PasswordResetRequestedEn,
		&item.RolUsuarioID,
		&item.RolNombre,
		&item.EmailConfirmado,
		&item.EmailConfirmadoEn,
		&item.AceptaContrato,
		&item.ContratoVersionAceptada,
		&item.FechaAceptaContrato,
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
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users
		SET password_hash = ?,
			password_salt = ?,
			password_set = 1,
			password_actualizada_en = datetime('now','localtime'),
			password_reset_token = '',
			password_reset_expira = '',
			password_reset_requested_en = '',
			login_failed_attempts = 0,
			login_failed_last_at = '',
			login_locked_until = '',
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, passwordHash, passwordSalt, id, empresaID)
	return err
}

// SetEmpresaUsuarioPasswordResetToken registra un token temporal para recuperación de contraseña.
func SetEmpresaUsuarioPasswordResetToken(dbConn *sql.DB, empresaID, id int64, token, expira string) error {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users
		SET password_reset_token = ?,
			password_reset_expira = ?,
			password_reset_requested_en = datetime('now','localtime'),
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, token, expira, id, empresaID)
	return err
}

// ClearEmpresaUsuarioPasswordResetToken invalida el token de recuperación actual de un usuario.
func ClearEmpresaUsuarioPasswordResetToken(dbConn *sql.DB, empresaID, id int64) error {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users
		SET password_reset_token = '',
			password_reset_expira = '',
			password_reset_requested_en = '',
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, id, empresaID)
	return err
}

// RegisterEmpresaUsuarioLoginFailure incrementa intentos fallidos y aplica bloqueo temporal.
func RegisterEmpresaUsuarioLoginFailure(dbConn *sql.DB, empresaID, id int64, maxAttempts int, window, lockDuration time.Duration) (int, string, error) {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return 0, "", err
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	if window <= 0 {
		window = 15 * time.Minute
	}
	if lockDuration <= 0 {
		lockDuration = 15 * time.Minute
	}

	row := queryRowSQLCompat(dbConn, `SELECT
		COALESCE(login_failed_attempts, 0),
		COALESCE(login_failed_last_at, ''),
		COALESCE(login_locked_until, '')
	FROM users
	WHERE id = ? AND empresa_id = ?
	LIMIT 1`, id, empresaID)

	var currentAttempts int
	var lastFailedRaw string
	var lockedUntilRaw string
	if err := row.Scan(&currentAttempts, &lastFailedRaw, &lockedUntilRaw); err != nil {
		return 0, "", err
	}

	now := time.Now()
	attempts := currentAttempts
	lockedUntil := ""

	if lockAt, ok := parseDateTimeLocal(lockedUntilRaw); ok && now.Before(lockAt) {
		attempts = maxAttempts
		lockedUntil = lockAt.Format("2006-01-02 15:04:05")
	} else {
		if lastFailedAt, ok := parseDateTimeLocal(lastFailedRaw); !ok || now.Sub(lastFailedAt) > window {
			attempts = 0
		}
		attempts++
		if attempts >= maxAttempts {
			lockedUntil = now.Add(lockDuration).Format("2006-01-02 15:04:05")
		}
	}

	_, err := dbConn.Exec(`UPDATE users
		SET login_failed_attempts = ?,
			login_failed_last_at = ?,
			login_locked_until = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`,
		attempts,
		now.Format("2006-01-02 15:04:05"),
		lockedUntil,
		id,
		empresaID,
	)
	if err != nil {
		return 0, "", err
	}

	return attempts, lockedUntil, nil
}

// ClearEmpresaUsuarioLoginFailures limpia contador y bloqueo de intentos fallidos.
func ClearEmpresaUsuarioLoginFailures(dbConn *sql.DB, empresaID, id int64) error {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users
		SET login_failed_attempts = 0,
			login_failed_last_at = '',
			login_locked_until = '',
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, id, empresaID)
	return err
}

// IsEmpresaUsuarioLocked evalúa si un usuario está bloqueado por intentos fallidos.
func IsEmpresaUsuarioLocked(item *EmpresaUsuario, now time.Time) (bool, string) {
	if item == nil {
		return false, ""
	}
	if now.IsZero() {
		now = time.Now()
	}
	lockAt, ok := parseDateTimeLocal(item.LoginLockedUntil)
	if !ok {
		return false, ""
	}
	if now.Before(lockAt) {
		return true, lockAt.Format("2006-01-02 15:04:05")
	}
	return false, ""
}

func parseDateTimeLocal(raw string) (time.Time, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
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
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
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
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`DELETE FROM users WHERE id = ? AND empresa_id = ?`, id, empresaID)
	return err
}

// DeleteEmpresaUsuariosPreconfiguracion elimina usuarios guia creados por una preconfiguracion.
func DeleteEmpresaUsuariosPreconfiguracion(dbConn *sql.DB, empresaID int64, marker string) (int64, error) {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return 0, err
	}
	marker = strings.TrimSpace(marker)
	if marker == "" {
		return 0, nil
	}
	res, err := execSQLCompat(dbConn, `DELETE FROM users
		WHERE empresa_id = ?
		  AND COALESCE(observaciones, '') LIKE ?`,
		empresaID,
		"%"+marker+"%",
	)
	if err != nil {
		return 0, err
	}
	affected, _ := res.RowsAffected()
	return affected, nil
}

// SetEmpresaUsuarioEstado activa o desactiva un usuario de empresa.
func SetEmpresaUsuarioEstado(dbConn *sql.DB, empresaID, id int64, estado string) error {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users SET estado = ?, fecha_actualizacion = datetime('now','localtime') WHERE id = ? AND empresa_id = ?`, estado, id, empresaID)
	return err
}

// SetEmpresaUsuarioConfirmToken actualiza token de confirmación para reenvíos.
func SetEmpresaUsuarioConfirmToken(dbConn *sql.DB, empresaID, id int64, confirmToken, confirmExpira string) error {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users
		SET email_confirm_token = ?,
			email_confirm_expira = ?,
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, confirmToken, confirmExpira, id, empresaID)
	return err
}

// ConfirmEmpresaUsuarioByToken confirma el correo de un usuario usando su token.
func ConfirmEmpresaUsuarioByToken(dbConn *sql.DB, token string) (int64, error) {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return 0, err
	}
	row := queryRowSQLCompat(dbConn, `SELECT id, empresa_id, COALESCE(email_confirm_expira, '') FROM users WHERE email_confirm_token = ? LIMIT 1`, token)
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

func SetEmpresaUsuarioContratoAceptado(dbConn *sql.DB, empresaID, id int64, version int) error {
	if err := EnsureEmpresaUsuariosAuthSchema(dbConn); err != nil {
		return err
	}
	_, err := dbConn.Exec(`UPDATE users
		SET acepta_contrato = 1,
			contrato_version_aceptada = ?,
			fecha_acepta_contrato = datetime('now','localtime'),
			fecha_actualizacion = datetime('now','localtime')
		WHERE id = ? AND empresa_id = ?`, version, id, empresaID)
	return err
}
