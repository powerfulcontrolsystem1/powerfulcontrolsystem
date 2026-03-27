package db

import (
	"database/sql"
)

// UpsertUser inserta o actualiza un usuario en la base de datos de empresas (registro por empresa)
func UpsertUser(dbConn *sql.DB, email, name string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO users (email, name) VALUES (?, ?)", email, name); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE users SET name = ? WHERE email = ?", name, email); err != nil {
		return err
	}
	return tx.Commit()
}

// EnsureUserEmpresa crea una empresa por defecto para el usuario si no tiene una asociada
func EnsureUserEmpresa(dbConn *sql.DB, email, empresaNombre string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var userID int64
	var empresaID sql.NullInt64
	row := tx.QueryRow("SELECT id, empresa_id FROM users WHERE email = ?", email)
	if err := row.Scan(&userID, &empresaID); err != nil {
		return err
	}

	if empresaID.Valid {
		// ya tiene empresa asociada
		return tx.Commit()
	}

	res, err := tx.Exec("INSERT INTO empresas (nombre, usuario_creador) VALUES (?, ?)", empresaNombre, email)
	if err != nil {
		return err
	}
	newEmpresaID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if _, err := tx.Exec("UPDATE users SET empresa_id = ? WHERE id = ?", newEmpresaID, userID); err != nil {
		return err
	}

	return tx.Commit()
}

// UpsertAdministrador inserta o actualiza un registro en la tabla administradores de la base superadministrador
// Si se inserta por primera vez, asigna el rol provisto (usualmente 'administrador')
func UpsertAdministrador(dbConn *sql.DB, email, name, role string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("INSERT OR IGNORE INTO administradores (email, name, role, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, datetime('now','localtime'), datetime('now','localtime'), 'activo')", email, name, role); err != nil {
		return err
	}
	if _, err := tx.Exec("UPDATE administradores SET name = ?, role = ?, fecha_actualizacion = datetime('now','localtime') WHERE email = ?", name, role, email); err != nil {
		return err
	}

	return tx.Commit()
}

// CreateSession registra una sesión en la tabla sesiones de superadministrador
func CreateSession(dbConn *sql.DB, adminEmail, ip, userAgent, token string) error {
	_, err := dbConn.Exec("INSERT INTO sesiones (admin_email, token, ip, user_agent, fecha_inicio, activo, fecha_creacion) VALUES (?, ?, ?, ?, datetime('now','localtime'), 1, datetime('now','localtime'))", adminEmail, token, ip, userAgent)
	return err
}

// Admin representa un registro en la tabla administradores
type Admin struct {
	ID                 int64  `json:"id"`
	Email              string `json:"email"`
	Name               string `json:"name"`
	Role               string `json:"role"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	Estado             string `json:"estado"`
}

// Session representa un registro en la tabla sesiones
type Session struct {
	ID            int64  `json:"id"`
	AdminEmail    string `json:"admin_email"`
	Token         string `json:"token"`
	IP            string `json:"ip"`
	UserAgent     string `json:"user_agent"`
	FechaInicio   string `json:"fecha_inicio"`
	FechaCreacion string `json:"fecha_creacion"`
	Activo        int    `json:"activo"`
}

// GetAdministradores devuelve la lista de administradores desde la BD superadministrador
func GetAdministradores(dbConn *sql.DB) ([]Admin, error) {
	rows, err := dbConn.Query("SELECT id, email, name, role, fecha_creacion, fecha_actualizacion, estado FROM administradores ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Admin
	for rows.Next() {
		var a Admin
		if err := rows.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

// GetSesiones devuelve la lista de sesiones desde la BD superadministrador
func GetSesiones(dbConn *sql.DB) ([]Session, error) {
	rows, err := dbConn.Query("SELECT id, admin_email, token, ip, user_agent, fecha_inicio, fecha_creacion, activo FROM sesiones ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Session
	for rows.Next() {
		var s Session
		var fechaInicio sql.NullString
		var fechaCreacion sql.NullString
		if err := rows.Scan(&s.ID, &s.AdminEmail, &s.Token, &s.IP, &s.UserAgent, &fechaInicio, &fechaCreacion, &s.Activo); err != nil {
			return nil, err
		}
		if fechaInicio.Valid {
			s.FechaInicio = fechaInicio.String
		}
		if fechaCreacion.Valid {
			s.FechaCreacion = fechaCreacion.String
		}
		out = append(out, s)
	}
	return out, nil
}

// GetSessionByToken devuelve una sesión activa por token
func GetSessionByToken(dbConn *sql.DB, token string) (*Session, error) {
	row := dbConn.QueryRow("SELECT id, admin_email, token, ip, user_agent, fecha_inicio, fecha_creacion, activo FROM sesiones WHERE token = ? AND activo = 1 LIMIT 1", token)
	var s Session
	var fechaInicio sql.NullString
	var fechaCreacion sql.NullString
	if err := row.Scan(&s.ID, &s.AdminEmail, &s.Token, &s.IP, &s.UserAgent, &fechaInicio, &fechaCreacion, &s.Activo); err != nil {
		return nil, err
	}
	if fechaInicio.Valid {
		s.FechaInicio = fechaInicio.String
	}
	if fechaCreacion.Valid {
		s.FechaCreacion = fechaCreacion.String
	}
	return &s, nil
}

// GetAdminByEmail devuelve el administrador por email
func GetAdminByEmail(dbConn *sql.DB, email string) (*Admin, error) {
	row := dbConn.QueryRow("SELECT id, email, name, role, fecha_creacion, fecha_actualizacion, estado FROM administradores WHERE email = ? LIMIT 1", email)
	var a Admin
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err != nil {
		return nil, err
	}
	return &a, nil
}
