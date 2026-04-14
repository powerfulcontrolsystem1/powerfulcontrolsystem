package db

import (
	"database/sql"
	"strings"
)

// UpsertUser inserta o actualiza un usuario en la base de datos de empresas (registro por empresa)
func UpsertUser(dbConn *sql.DB, email, name string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	upsertSQL := "INSERT INTO users (email, name) VALUES (?, ?) ON CONFLICT(email) DO UPDATE SET name = EXCLUDED.name"
	if _, err := execTxSQLCompat(tx, upsertSQL, email, name); err != nil {
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
	row := queryRowTxSQLCompat(tx, "SELECT id, empresa_id FROM users WHERE email = ?", email)
	if err := row.Scan(&userID, &empresaID); err != nil {
		return err
	}

	if empresaID.Valid {
		// ya tiene empresa asociada
		return tx.Commit()
	}

	var newEmpresaID int64
	if isPostgresDialect() {
		if err := queryRowTxSQLCompat(tx, "INSERT INTO empresas (nombre, usuario_creador) VALUES (?, ?) RETURNING id", empresaNombre, email).Scan(&newEmpresaID); err != nil {
			return err
		}
	} else {
		res, err := execTxSQLCompat(tx, "INSERT INTO empresas (nombre, usuario_creador) VALUES (?, ?)", empresaNombre, email)
		if err != nil {
			return err
		}
		newEmpresaID, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}

	if _, err := execTxSQLCompat(tx, "UPDATE users SET empresa_id = ? WHERE id = ?", newEmpresaID, userID); err != nil {
		return err
	}

	return tx.Commit()
}

// UpsertAdministrador inserta o actualiza un registro en la tabla administradores de la base superadministrador
// Si se inserta por primera vez, asigna el rol provisto (usualmente 'administrador').
// Ahora acepta un campo `photo` con la URL de la foto del perfil.
func UpsertAdministrador(dbConn *sql.DB, email, name, role, photo string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowExpr := sqlNowExpr()
	upsertSQL := "INSERT INTO administradores (email, name, role, photo, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, " + nowExpr + ", " + nowExpr + ", 'activo') ON CONFLICT(email) DO UPDATE SET name = EXCLUDED.name, role = EXCLUDED.role, photo = EXCLUDED.photo, fecha_actualizacion = " + nowExpr
	if _, err := execTxSQLCompat(tx, upsertSQL, email, name, role, photo); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateAdministrador actualiza el nombre y rol de un administrador por id
func UpdateAdministrador(dbConn *sql.DB, id int64, name, role string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET name = ?, role = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", name, role, id)
	return err
}

// DeleteAdministrador elimina un administrador por id
func DeleteAdministrador(dbConn *sql.DB, id int64) error {
	_, err := execSQLCompat(dbConn, "DELETE FROM administradores WHERE id = ?", id)
	return err
}

// SetAdministradorEstado activa/desactiva un administrador (estado: 'activo'/'inactivo')
func SetAdministradorEstado(dbConn *sql.DB, id int64, estado string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET estado = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", estado, id)
	return err
}

// SetAdministradorAceptaContrato marca si el administrador aceptó el contrato/registro.
func SetAdministradorAceptaContrato(dbConn *sql.DB, email string, acepta bool) error {
	v := 0
	if acepta {
		v = 1
	}
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET acepta_contrato = ?, fecha_actualizacion = "+nowExpr+" WHERE LOWER(COALESCE(email,'')) = LOWER(?)", v, strings.TrimSpace(email))
	return err
}

// CreateSession registra una sesión en la tabla sesiones de superadministrador
func CreateSession(dbConn *sql.DB, adminEmail, ip, userAgent, token string) error {
	nowExpr := sqlNowExpr()
	expiresExpr := sqlPlusHoursExpr(24)
	query := "INSERT INTO sesiones (admin_email, token, ip, user_agent, fecha_inicio, fecha_fin, activo, fecha_creacion) VALUES (?, ?, ?, ?, " + nowExpr + ", " + expiresExpr + ", 1, " + nowExpr + ")"
	_, err := execSQLCompat(dbConn, query, adminEmail, token, ip, userAgent)
	return err
}

// RevokeSessionByToken invalida una sesión activa por token.
func RevokeSessionByToken(dbConn *sql.DB, token string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE sesiones SET activo = 0, fecha_fin = "+nowExpr+" WHERE token = ? AND activo = 1", token)
	return err
}

// Admin representa un registro en la tabla administradores
type Admin struct {
	ID                 int64  `json:"id"`
	Email              string `json:"email"`
	Name               string `json:"name"`
	Role               string `json:"role"`
	Photo              string `json:"photo,omitempty"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	Estado             string `json:"estado"`
	AceptaContrato     int    `json:"acepta_contrato"`
}

// NOTE: tipos_de_licencia CRUD removed per project decision (frontend/page/link removed).

// Licencia representa una licencia asignada (nuevo CRUD)
type Licencia struct {
	ID            int64   `json:"id"`
	EmpresaID     int64   `json:"empresa_id"`
	TipoID        int64   `json:"tipo_id"`
	TipoNombre    string  `json:"tipo_nombre,omitempty"`
	Nombre        string  `json:"nombre"`
	Descripcion   string  `json:"descripcion"`
	Valor         float64 `json:"valor"`
	DuracionDias  int     `json:"duracion_dias"`
	ModulosHab    string  `json:"modulos_habilitados,omitempty"`
	SuperRol      int     `json:"super_rol_habilitado"`
	FechaCreacion string  `json:"fecha_creacion"`
	Activo        int     `json:"activo"`
}

// CreateLicencia inserta una nueva licencia en dbSuper
func CreateLicencia(dbConn *sql.DB, tipoID int64, nombre, descripcion string, valor float64, duracionDias int, modulosHabilitados string, superRolHabilitado int) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO licencias (tipo_id, nombre, descripcion, valor, duracion_dias, modulos_habilitados, super_rol_habilitado, fecha_creacion, fecha_actualizacion, activo, estado) VALUES (?, ?, ?, ?, ?, ?, ?, " + nowExpr + ", " + nowExpr + ", 1, 'activo')"
	return insertSQLCompat(dbConn, query, tipoID, nombre, descripcion, valor, duracionDias, strings.TrimSpace(modulosHabilitados), superRolHabilitado)
}

// GetLicencias obtiene todas las licencias (comportamiento legado sin filtros)
func GetLicencias(dbConn *sql.DB) ([]Licencia, error) {
	return GetLicenciasFiltered(dbConn, false, "", false)
}

// GetLicenciasFiltered obtiene licencias con filtros opcionales.
func GetLicenciasFiltered(dbConn *sql.DB, soloActivas bool, usuarioCreador string, conEmpresa bool) ([]Licencia, error) {
	q := `SELECT l.id, l.empresa_id, l.tipo_id, t.nombre, l.nombre, l.descripcion, l.valor, l.duracion_dias, COALESCE(l.modulos_habilitados, ''), COALESCE(l.super_rol_habilitado, 0), l.fecha_creacion, l.activo
		FROM licencias l LEFT JOIN tipos_de_empresas t ON l.tipo_id = t.id`

	usuarioCreador = strings.TrimSpace(usuarioCreador)
	canFilterByUsuarioCreador := false
	if usuarioCreador != "" {
		hasEmpresasTable, err := tableExists(dbConn, "empresas")
		if err != nil {
			return nil, err
		}
		if hasEmpresasTable {
			q += " LEFT JOIN empresas e ON e.id = l.empresa_id"
			canFilterByUsuarioCreador = true
		}
	}

	var where []string
	var args []interface{}
	if soloActivas {
		where = append(where, "l.activo = 1")
	}
	if conEmpresa {
		where = append(where, "l.empresa_id IS NOT NULL AND l.empresa_id > 0")
	}
	if usuarioCreador != "" && canFilterByUsuarioCreador {
		where = append(where, "LOWER(COALESCE(e.usuario_creador, '')) = LOWER(?)")
		args = append(args, usuarioCreador)
	}
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY l.id DESC"

	rows, err := querySQLCompat(dbConn, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Licencia
	for rows.Next() {
		var lic Licencia
		var empresaID sql.NullInt64
		var tipoNombre sql.NullString
		var descripcion sql.NullString
		var modulosHab sql.NullString
		var fechaCreacion sql.NullString
		if err := rows.Scan(&lic.ID, &empresaID, &lic.TipoID, &tipoNombre, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &modulosHab, &lic.SuperRol, &fechaCreacion, &lic.Activo); err != nil {
			return nil, err
		}
		if empresaID.Valid {
			lic.EmpresaID = empresaID.Int64
		}
		if tipoNombre.Valid {
			lic.TipoNombre = tipoNombre.String
		}
		if descripcion.Valid {
			lic.Descripcion = descripcion.String
		}
		if modulosHab.Valid {
			lic.ModulosHab = modulosHab.String
		}
		if fechaCreacion.Valid {
			lic.FechaCreacion = fechaCreacion.String
		}
		out = append(out, lic)
	}
	return out, nil
}

// GetLicenciaByID devuelve una licencia por id
func GetLicenciaByID(dbConn *sql.DB, id int64) (*Licencia, error) {
	q := `SELECT id, empresa_id, tipo_id, nombre, descripcion, valor, duracion_dias, COALESCE(modulos_habilitados, ''), COALESCE(super_rol_habilitado, 0), fecha_creacion, activo FROM licencias WHERE id = ? LIMIT 1`
	row := queryRowSQLCompat(dbConn, q, id)
	var lic Licencia
	var empresaID sql.NullInt64
	var descripcion sql.NullString
	var modulosHab sql.NullString
	var fechaCreacion sql.NullString
	if err := row.Scan(&lic.ID, &empresaID, &lic.TipoID, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &modulosHab, &lic.SuperRol, &fechaCreacion, &lic.Activo); err != nil {
		return nil, err
	}
	if empresaID.Valid {
		lic.EmpresaID = empresaID.Int64
	}
	if descripcion.Valid {
		lic.Descripcion = descripcion.String
	}
	if modulosHab.Valid {
		lic.ModulosHab = modulosHab.String
	}
	if fechaCreacion.Valid {
		lic.FechaCreacion = fechaCreacion.String
	}
	return &lic, nil
}

// UpdateLicencia actualiza campos editables de una licencia
func UpdateLicencia(dbConn *sql.DB, id, tipoID int64, nombre, descripcion string, valor float64, duracionDias int, modulosHabilitados string, superRolHabilitado int) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE licencias SET tipo_id = ?, nombre = ?, descripcion = ?, valor = ?, duracion_dias = ?, modulos_habilitados = ?, super_rol_habilitado = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", tipoID, nombre, descripcion, valor, duracionDias, strings.TrimSpace(modulosHabilitados), superRolHabilitado, id)
	return err
}

// LicenciaPermisoPolicy describe las capacidades de acceso habilitadas por licencia activa para una empresa.
type LicenciaPermisoPolicy struct {
	LicenciaID         int64
	Nombre             string
	ModulosHabilitados string
	SuperRolHabilitado bool
}

// GetLicenciaPermisoPolicyByEmpresa resuelve la licencia activa vigente para permisos de una empresa.
func GetLicenciaPermisoPolicyByEmpresa(dbConn *sql.DB, empresaID int64) (*LicenciaPermisoPolicy, error) {
	if dbConn == nil || empresaID <= 0 {
		return nil, nil
	}

	query := `SELECT id,
		COALESCE(nombre, ''),
		COALESCE(modulos_habilitados, ''),
		COALESCE(super_rol_habilitado, 0)
	FROM licencias
	WHERE empresa_id = ?
		AND COALESCE(activo, 1) = 1
		AND (COALESCE(fecha_inicio, '') = '' OR datetime(fecha_inicio) <= datetime('now','localtime'))
		AND (COALESCE(fecha_fin, '') = '' OR datetime(fecha_fin) >= datetime('now','localtime'))
	ORDER BY
		CASE WHEN COALESCE(fecha_fin, '') = '' THEN 1 ELSE 0 END DESC,
		datetime(COALESCE(fecha_fin, '9999-12-31 23:59:59')) DESC,
		id DESC
	LIMIT 1`
	if isPostgresDialect() {
		query = `SELECT id,
			COALESCE(nombre, ''),
			COALESCE(modulos_habilitados, ''),
			COALESCE(super_rol_habilitado, 0)
		FROM licencias
		WHERE empresa_id = ?
			AND COALESCE(activo, 1) = 1
			AND (COALESCE(CAST(fecha_inicio AS TEXT), '') = '' OR CAST(fecha_inicio AS TIMESTAMP) <= CURRENT_TIMESTAMP)
			AND (COALESCE(CAST(fecha_fin AS TEXT), '') = '' OR CAST(fecha_fin AS TIMESTAMP) >= CURRENT_TIMESTAMP)
		ORDER BY
			CASE WHEN COALESCE(CAST(fecha_fin AS TEXT), '') = '' THEN 1 ELSE 0 END DESC,
			COALESCE(CAST(fecha_fin AS TIMESTAMP), TIMESTAMP '9999-12-31 23:59:59') DESC,
			id DESC
		LIMIT 1`
	}
	row := queryRowSQLCompat(dbConn, query, empresaID)

	var item LicenciaPermisoPolicy
	var superRolRaw int
	if err := row.Scan(&item.LicenciaID, &item.Nombre, &item.ModulosHabilitados, &superRolRaw); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		if isMissingTableError(err) || isMissingColumnError(err) {
			return nil, nil
		}
		return nil, err
	}
	item.SuperRolHabilitado = superRolRaw == 1
	return &item, nil
}

// DeleteLicencia elimina una licencia por id
func DeleteLicencia(dbConn *sql.DB, id int64) error {
	_, err := execSQLCompat(dbConn, "DELETE FROM licencias WHERE id = ?", id)
	return err
}

// SetLicenciaActivo activa/desactiva una licencia (activo: 1 o 0)
func SetLicenciaActivo(dbConn *sql.DB, id int64, activo int) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE licencias SET activo = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", activo, id)
	return err
}

// Session representa una sesión del administrador registrada en la tabla sesiones
type Session struct {
	ID            int64  `json:"id"`
	AdminEmail    string `json:"admin_email"`
	Token         string `json:"token"`
	IP            string `json:"ip"`
	UserAgent     string `json:"user_agent"`
	FechaInicio   string `json:"fecha_inicio"`
	FechaFin      string `json:"fecha_fin,omitempty"`
	FechaCreacion string `json:"fecha_creacion"`
	Activo        int    `json:"activo"`
}

// GetSessionByToken devuelve una sesión activa por token
func GetSessionByToken(dbConn *sql.DB, token string) (*Session, error) {
	condition := sessionNotExpiredCondition("fecha_fin")
	query := "SELECT id, admin_email, token, ip, user_agent, fecha_inicio, fecha_fin, fecha_creacion, activo FROM sesiones WHERE token = ? AND activo = 1 AND " + condition + " LIMIT 1"
	row := queryRowSQLCompat(dbConn, query, token)
	var s Session
	var fechaInicio sql.NullString
	var fechaFin sql.NullString
	var fechaCreacion sql.NullString
	if err := row.Scan(&s.ID, &s.AdminEmail, &s.Token, &s.IP, &s.UserAgent, &fechaInicio, &fechaFin, &fechaCreacion, &s.Activo); err != nil {
		return nil, err
	}
	if fechaInicio.Valid {
		s.FechaInicio = fechaInicio.String
	}
	if fechaFin.Valid {
		s.FechaFin = fechaFin.String
	}
	if fechaCreacion.Valid {
		s.FechaCreacion = fechaCreacion.String
	}
	return &s, nil
}

// GetAdminByEmail devuelve el administrador por email
func GetAdminByEmail(dbConn *sql.DB, email string) (*Admin, error) {
	// Intentar obtener con la columna 'acepta_contrato' (para bases actualizadas).
	row := queryRowSQLCompat(dbConn, "SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado, COALESCE(acepta_contrato, 0) FROM administradores WHERE email = ? LIMIT 1", email)
	var a Admin
	var photo sql.NullString
	var acepta sql.NullInt64
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado, &acepta); err != nil {
		// Fallback: si la columna no existe en este esquema, consultar sin ella.
		if isMissingColumnError(err) {
			row2 := queryRowSQLCompat(dbConn, "SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado FROM administradores WHERE email = ? LIMIT 1", email)
			var photo2 sql.NullString
			if err2 := row2.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo2, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err2 != nil {
				return nil, err2
			}
			if photo2.Valid {
				a.Photo = photo2.String
			}
			a.AceptaContrato = 0
			return &a, nil
		}
		return nil, err
	}
	if photo.Valid {
		a.Photo = photo.String
	}
	if acepta.Valid {
		a.AceptaContrato = int(acepta.Int64)
	} else {
		a.AceptaContrato = 0
	}
	return &a, nil
}

// GetAdministradores lista todos los administradores
func GetAdministradores(dbConn *sql.DB) ([]Admin, error) {
	// Intentar consulta que incluya la columna acepta_contrato cuando exista
	rows, err := querySQLCompat(dbConn, "SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado, COALESCE(acepta_contrato, 0) FROM administradores ORDER BY id DESC")
	if err != nil {
		// Fallback si la columna no existe en esquemas antiguos
		if isMissingColumnError(err) {
			rows2, err2 := querySQLCompat(dbConn, "SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado FROM administradores ORDER BY id DESC")
			if err2 != nil {
				return nil, err2
			}
			defer rows2.Close()
			var out2 []Admin
			for rows2.Next() {
				var a Admin
				var photo sql.NullString
				if err := rows2.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err != nil {
					return nil, err
				}
				if photo.Valid {
					a.Photo = photo.String
				}
				a.AceptaContrato = 0
				out2 = append(out2, a)
			}
			return out2, nil
		}
		return nil, err
	}
	defer rows.Close()
	var out []Admin
	for rows.Next() {
		var a Admin
		var photo sql.NullString
		var acepta sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado, &acepta); err != nil {
			return nil, err
		}
		if photo.Valid {
			a.Photo = photo.String
		}
		if acepta.Valid {
			a.AceptaContrato = int(acepta.Int64)
		} else {
			a.AceptaContrato = 0
		}
		out = append(out, a)
	}
	return out, nil
}

// GetSesiones lista las sesiones registradas
func GetSesiones(dbConn *sql.DB) ([]Session, error) {
	rows, err := querySQLCompat(dbConn, "SELECT id, admin_email, token, ip, user_agent, fecha_inicio, fecha_fin, fecha_creacion, activo FROM sesiones ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		var fechaInicio sql.NullString
		var fechaFin sql.NullString
		var fechaCreacion sql.NullString
		if err := rows.Scan(&s.ID, &s.AdminEmail, &s.Token, &s.IP, &s.UserAgent, &fechaInicio, &fechaFin, &fechaCreacion, &s.Activo); err != nil {
			return nil, err
		}
		if fechaInicio.Valid {
			s.FechaInicio = fechaInicio.String
		}
		if fechaFin.Valid {
			s.FechaFin = fechaFin.String
		}
		if fechaCreacion.Valid {
			s.FechaCreacion = fechaCreacion.String
		}
		out = append(out, s)
	}
	return out, nil
}

// TipoEmpresa representa un tipo de empresa
type TipoEmpresa struct {
	ID            int64  `json:"id"`
	Nombre        string `json:"nombre"`
	Observaciones string `json:"observaciones"`
	FechaCreacion string `json:"fecha_creacion"`
	Estado        string `json:"estado"`
}

// CreateTipoEmpresa inserta un nuevo tipo de empresa
func CreateTipoEmpresa(dbConn *sql.DB, nombre, observaciones string) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO tipos_de_empresas (nombre, observaciones, fecha_creacion) VALUES (?, ?, " + nowExpr + ")"
	return insertSQLCompat(dbConn, query, nombre, observaciones)
}

// GetTiposEmpresas obtiene todos los tipos de empresa
func GetTiposEmpresas(dbConn *sql.DB) ([]TipoEmpresa, error) {
	rows, err := querySQLCompat(dbConn, "SELECT id, nombre, observaciones, fecha_creacion, estado FROM tipos_de_empresas ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TipoEmpresa
	for rows.Next() {
		var t TipoEmpresa
		if err := rows.Scan(&t.ID, &t.Nombre, &t.Observaciones, &t.FechaCreacion, &t.Estado); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

// UpdateTipoEmpresa actualiza un tipo de empresa por id
func UpdateTipoEmpresa(dbConn *sql.DB, id int64, nombre, observaciones string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE tipos_de_empresas SET nombre = ?, observaciones = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", nombre, observaciones, id)
	return err
}

// DeleteTipoEmpresa elimina un tipo de empresa por id
func DeleteTipoEmpresa(dbConn *sql.DB, id int64) error {
	_, err := execSQLCompat(dbConn, "DELETE FROM tipos_de_empresas WHERE id = ?", id)
	return err
}

// SetTipoEmpresaActivo activa/desactiva un tipo de empresa (activo: 'activo'/'inactivo' o 1/0)
func SetTipoEmpresaActivo(dbConn *sql.DB, id int64, estado string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE tipos_de_empresas SET estado = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", estado, id)
	return err
}

// Empresa representa una empresa registrada en empresas.db
type Empresa struct {
	ID                 int64  `json:"id"`
	EmpresaID          int64  `json:"empresa_id,omitempty"`
	Nombre             string `json:"nombre"`
	Nit                string `json:"nit,omitempty"`
	TipoID             int64  `json:"tipo_id,omitempty"`
	TipoNombre         string `json:"tipo_nombre,omitempty"`
	FechaCreacion      string `json:"fecha_creacion,omitempty"`
	FechaActualizacion string `json:"fecha_actualizacion,omitempty"`
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	Estado             string `json:"estado,omitempty"`
	Observaciones      string `json:"observaciones,omitempty"`
}

// CreateEmpresa inserta una nueva empresa en la base empresas.db
func CreateEmpresa(dbConn *sql.DB, tipoID int64, tipoNombre, nombre, nit, observaciones, usuarioCreador string) (int64, error) {
	tx, err := dbConn.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	nowExpr := sqlNowExpr()

	id, err := insertTxSQLCompat(tx, "INSERT INTO empresas (tipo_id, tipo_nombre, nombre, nit, observaciones, usuario_creador, fecha_creacion, estado) VALUES (?, ?, ?, ?, ?, ?, "+nowExpr+", 'activo')", tipoID, tipoNombre, nombre, nit, observaciones, usuarioCreador)
	if err != nil {
		return 0, err
	}
	if _, err := execTxSQLCompat(tx, "UPDATE empresas SET empresa_id = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ? AND (empresa_id IS NULL OR empresa_id <= 0)", id, id); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return id, nil
}

// GetEmpresas obtiene todas las empresas
func GetEmpresas(dbConn *sql.DB) ([]Empresa, error) {
	rows, err := querySQLCompat(dbConn, "SELECT id, COALESCE(empresa_id, id), nombre, nit, tipo_id, tipo_nombre, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresas ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Empresa
	for rows.Next() {
		var e Empresa
		var empresaID sql.NullInt64
		var nit sql.NullString
		var tipoID sql.NullInt64
		var tipoNombre sql.NullString
		var fechaCre sql.NullString
		var fechaAct sql.NullString
		var usuario sql.NullString
		var estado sql.NullString
		var obs sql.NullString
		if err := rows.Scan(&e.ID, &empresaID, &e.Nombre, &nit, &tipoID, &tipoNombre, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
			return nil, err
		}
		if empresaID.Valid {
			e.EmpresaID = empresaID.Int64
		} else {
			e.EmpresaID = e.ID
		}
		if nit.Valid {
			e.Nit = nit.String
		}
		if tipoID.Valid {
			e.TipoID = tipoID.Int64
		}
		if tipoNombre.Valid {
			e.TipoNombre = tipoNombre.String
		}
		if fechaCre.Valid {
			e.FechaCreacion = fechaCre.String
		}
		if fechaAct.Valid {
			e.FechaActualizacion = fechaAct.String
		}
		if usuario.Valid {
			e.UsuarioCreador = usuario.String
		}
		if estado.Valid {
			e.Estado = estado.String
		}
		if obs.Valid {
			e.Observaciones = obs.String
		}
		out = append(out, e)
	}
	return out, nil
}

// GetEmpresaByID devuelve una empresa por id
func GetEmpresaByID(dbConn *sql.DB, id int64) (*Empresa, error) {
	row := queryRowSQLCompat(dbConn, "SELECT id, COALESCE(empresa_id, id), nombre, nit, tipo_id, tipo_nombre, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresas WHERE id = ? LIMIT 1", id)
	var e Empresa
	var empresaID sql.NullInt64
	var nit sql.NullString
	var tipoID sql.NullInt64
	var tipoNombre sql.NullString
	var fechaCre sql.NullString
	var fechaAct sql.NullString
	var usuario sql.NullString
	var estado sql.NullString
	var obs sql.NullString
	if err := row.Scan(&e.ID, &empresaID, &e.Nombre, &nit, &tipoID, &tipoNombre, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
		return nil, err
	}
	if empresaID.Valid {
		e.EmpresaID = empresaID.Int64
	} else {
		e.EmpresaID = e.ID
	}
	if nit.Valid {
		e.Nit = nit.String
	}
	if tipoID.Valid {
		e.TipoID = tipoID.Int64
	}
	if tipoNombre.Valid {
		e.TipoNombre = tipoNombre.String
	}
	if fechaCre.Valid {
		e.FechaCreacion = fechaCre.String
	}
	if fechaAct.Valid {
		e.FechaActualizacion = fechaAct.String
	}
	if usuario.Valid {
		e.UsuarioCreador = usuario.String
	}
	if estado.Valid {
		e.Estado = estado.String
	}
	if obs.Valid {
		e.Observaciones = obs.String
	}
	return &e, nil
}

// UpdateEmpresa actualiza campos editables de una empresa
func UpdateEmpresa(dbConn *sql.DB, id, tipoID int64, tipoNombre, nombre, nit, observaciones string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE empresas SET tipo_id = ?, tipo_nombre = ?, nombre = ?, nit = ?, observaciones = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", tipoID, tipoNombre, nombre, nit, observaciones, id)
	return err
}

// DeleteEmpresa elimina una empresa por id
func DeleteEmpresa(dbConn *sql.DB, id int64) error {
	_, err := execSQLCompat(dbConn, "DELETE FROM empresas WHERE id = ?", id)
	return err
}

// SetEmpresaEstado activa/desactiva una empresa (estado: 'activo'/'inactivo')
func SetEmpresaEstado(dbConn *sql.DB, id int64, estado string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE empresas SET estado = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", estado, id)
	return err
}

// Metric representa una muestra de métricas del sistema
type Metric struct {
	ID            int64   `json:"id"`
	Timestamp     string  `json:"timestamp"`
	CPUPercent    float64 `json:"cpu_percent"`
	MemTotal      uint64  `json:"mem_total"`
	MemUsed       uint64  `json:"mem_used"`
	MemPercent    float64 `json:"mem_percent"`
	NetRecv       uint64  `json:"net_recv"`
	NetSent       uint64  `json:"net_sent"`
	FechaCreacion string  `json:"fecha_creacion"`
}

// InitMetricsTable crea la tabla metrics en la base de datos si no existe
func InitMetricsTable(dbConn *sql.DB) error {
	if isPostgresDialect() {
		create := `CREATE TABLE IF NOT EXISTS metrics (
			id BIGSERIAL PRIMARY KEY,
			timestamp TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			cpu_percent DOUBLE PRECISION,
			mem_total BIGINT,
			mem_used BIGINT,
			mem_percent DOUBLE PRECISION,
			net_recv BIGINT,
			net_sent BIGINT,
			fecha_creacion TEXT DEFAULT (CAST(CURRENT_TIMESTAMP AS TEXT)),
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`
		_, err := execSQLCompat(dbConn, create)
		return err
	}

	create := `CREATE TABLE IF NOT EXISTS metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT DEFAULT (datetime('now','localtime')),
		cpu_percent REAL,
		mem_total INTEGER,
		mem_used INTEGER,
		mem_percent REAL,
		net_recv INTEGER,
		net_sent INTEGER,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT,
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	);`
	_, err := execSQLCompat(dbConn, create)
	return err
}

// CreateWompiPaymentRecord registra una transacción inicial de Wompi en la tabla pagos_wompi.
func CreateWompiPaymentRecord(dbConn *sql.DB, licenciaID, empresaID int64, transactionID, reference, status, rawPayload, discountCode, asesorID string) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO pagos_wompi (licencia_id, empresa_id, transaction_id, reference, status, raw_payload, discount_code, asesor_id, fecha_creacion) VALUES (?, ?, ?, ?, ?, ?, ?, ?, " + nowExpr + ")"
	return insertSQLCompat(dbConn, query, licenciaID, empresaID, transactionID, reference, status, rawPayload, discountCode, asesorID)
}

// UpdateWompiPaymentRecordByTransaction actualiza una transacción de Wompi usando su transaction_id.
func UpdateWompiPaymentRecordByTransaction(dbConn *sql.DB, transactionID, status, rawPayload string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE pagos_wompi SET status = ?, raw_payload = ?, fecha_actualizacion = "+nowExpr+" WHERE transaction_id = ?", status, rawPayload, transactionID)
	if err == nil {
		return nil
	}
	// Compatibilidad con bases antiguas que aún no tienen la columna fecha_actualizacion.
	_, fallbackErr := execSQLCompat(dbConn, "UPDATE pagos_wompi SET status = ?, raw_payload = ? WHERE transaction_id = ?", status, rawPayload, transactionID)
	if fallbackErr == nil {
		return nil
	}
	return fallbackErr
}

// UpdateWompiPaymentRecordByReference actualiza una transaccion de Wompi usando su referencia.
func UpdateWompiPaymentRecordByReference(dbConn *sql.DB, reference, status, rawPayload string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE pagos_wompi SET status = ?, raw_payload = ?, fecha_actualizacion = "+nowExpr+" WHERE reference = ?", status, rawPayload, reference)
	if err == nil {
		return nil
	}
	_, fallbackErr := execSQLCompat(dbConn, "UPDATE pagos_wompi SET status = ?, raw_payload = ? WHERE reference = ?", status, rawPayload, reference)
	if fallbackErr == nil {
		return nil
	}
	return fallbackErr
}

// WompiPaymentRecord representa una fila de pagos_wompi
type WompiPaymentRecord struct {
	ID            int64
	LicenciaID    sql.NullInt64
	EmpresaID     sql.NullInt64
	TransactionID sql.NullString
	Reference     sql.NullString
	Status        sql.NullString
	RawPayload    sql.NullString
	DiscountCode  sql.NullString
	AsesorID      sql.NullString
	FechaCreacion sql.NullString
}

// GetWompiPaymentByTransaction obtiene una fila de pagos_wompi por transaction_id
func GetWompiPaymentByTransaction(dbConn *sql.DB, transactionID string) (*WompiPaymentRecord, error) {
	row := queryRowSQLCompat(dbConn, `SELECT id, licencia_id, empresa_id, transaction_id, reference, status, raw_payload, discount_code, asesor_id, fecha_creacion FROM pagos_wompi WHERE transaction_id = ? LIMIT 1`, transactionID)
	var r WompiPaymentRecord
	if err := row.Scan(&r.ID, &r.LicenciaID, &r.EmpresaID, &r.TransactionID, &r.Reference, &r.Status, &r.RawPayload, &r.DiscountCode, &r.AsesorID, &r.FechaCreacion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// GetWompiPaymentByReference obtiene una fila de pagos_wompi por reference
func GetWompiPaymentByReference(dbConn *sql.DB, reference string) (*WompiPaymentRecord, error) {
	row := queryRowSQLCompat(dbConn, `SELECT id, licencia_id, empresa_id, transaction_id, reference, status, raw_payload, discount_code, asesor_id, fecha_creacion FROM pagos_wompi WHERE reference = ? LIMIT 1`, reference)
	var r WompiPaymentRecord
	if err := row.Scan(&r.ID, &r.LicenciaID, &r.EmpresaID, &r.TransactionID, &r.Reference, &r.Status, &r.RawPayload, &r.DiscountCode, &r.AsesorID, &r.FechaCreacion); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &r, nil
}

// CRUD básicos para tabla asesores
func CreateAsesor(dbConn *sql.DB, email, nombre, rol, notas string) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO asesores (email, nombre, rol, notas, fecha_creacion) VALUES (?, ?, ?, ?, " + nowExpr + ")"
	return insertSQLCompat(dbConn, query, email, nombre, rol, notas)
}

type Asesor struct {
	ID                 int64  `json:"id"`
	Email              string `json:"email"`
	Nombre             string `json:"nombre"`
	Rol                string `json:"rol"`
	Notas              string `json:"notas"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	Estado             string `json:"estado"`
}

func ListAsesores(dbConn *sql.DB) ([]Asesor, error) {
	rows, err := querySQLCompat(dbConn, "SELECT id, email, nombre, rol, notas, fecha_creacion, fecha_actualizacion, estado FROM asesores WHERE estado IS NULL OR estado <> 'inactivo' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Asesor
	for rows.Next() {
		var a Asesor
		var fc, fa sql.NullString
		if err := rows.Scan(&a.ID, &a.Email, &a.Nombre, &a.Rol, &a.Notas, &fc, &fa, &a.Estado); err != nil {
			return nil, err
		}
		if fc.Valid {
			a.FechaCreacion = fc.String
		}
		if fa.Valid {
			a.FechaActualizacion = fa.String
		}
		out = append(out, a)
	}
	return out, nil
}

func UpdateAsesor(dbConn *sql.DB, id int64, email, nombre, rol, notas string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE asesores SET email = ?, nombre = ?, rol = ?, notas = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", email, nombre, rol, notas, id)
	return err
}

func DeleteAsesor(dbConn *sql.DB, id int64) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE asesores SET estado = 'inactivo', fecha_actualizacion = "+nowExpr+" WHERE id = ?", id)
	return err
}

// CRUD para planes de asesor comercial
type AsesorComercialPlan struct {
	ID               int64         `json:"id"`
	AsesorID         string        `json:"asesor_id"`
	AsesorEmail      string        `json:"asesor_email"`
	EmpresaID        sql.NullInt64 `json:"empresa_id"`
	ComisionVentaPct float64       `json:"comision_venta_pct"`
	ComisionPagoPct  float64       `json:"comision_pago_pct"`
	MesesRenovacion  int           `json:"meses_renovacion"`
	Notas            string        `json:"notas"`
	FechaCreacion    string        `json:"fecha_creacion"`
}

func CreateAsesorComercialPlan(dbConn *sql.DB, asesorID, asesorEmail string, empresaID int64, comisionVentaPct, comisionPagoPct float64, mesesRenovacion int, notas string) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO asesor_comercial (asesor_id, asesor_email, empresa_id, comision_venta_pct, comision_pago_pct, meses_renovacion, notas, fecha_creacion) VALUES (?, ?, ?, ?, ?, ?, ?, " + nowExpr + ")"
	return insertSQLCompat(dbConn, query, asesorID, asesorEmail, empresaID, comisionVentaPct, comisionPagoPct, mesesRenovacion, notas)
}

func ListAsesorComercialPlans(dbConn *sql.DB) ([]AsesorComercialPlan, error) {
	rows, err := querySQLCompat(dbConn, "SELECT id, asesor_id, asesor_email, empresa_id, comision_venta_pct, comision_pago_pct, meses_renovacion, notas, fecha_creacion FROM asesor_comercial WHERE estado IS NULL OR estado <> 'inactivo' ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []AsesorComercialPlan
	for rows.Next() {
		var p AsesorComercialPlan
		var empresaID sql.NullInt64
		var fc sql.NullString
		if err := rows.Scan(&p.ID, &p.AsesorID, &p.AsesorEmail, &empresaID, &p.ComisionVentaPct, &p.ComisionPagoPct, &p.MesesRenovacion, &p.Notas, &fc); err != nil {
			return nil, err
		}
		p.EmpresaID = empresaID
		if fc.Valid {
			p.FechaCreacion = fc.String
		}
		out = append(out, p)
	}
	return out, nil
}

func GetAsesorComercialPlanByAsesorID(dbConn *sql.DB, asesorID string) (*AsesorComercialPlan, error) {
	row := queryRowSQLCompat(dbConn, "SELECT id, asesor_id, asesor_email, empresa_id, comision_venta_pct, comision_pago_pct, meses_renovacion, notas, fecha_creacion FROM asesor_comercial WHERE asesor_id = ? AND (estado IS NULL OR estado <> 'inactivo') ORDER BY id DESC LIMIT 1", asesorID)
	var p AsesorComercialPlan
	var empresaID sql.NullInt64
	var fc sql.NullString
	if err := row.Scan(&p.ID, &p.AsesorID, &p.AsesorEmail, &empresaID, &p.ComisionVentaPct, &p.ComisionPagoPct, &p.MesesRenovacion, &p.Notas, &fc); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.EmpresaID = empresaID
	if fc.Valid {
		p.FechaCreacion = fc.String
	}
	return &p, nil
}

func UpdateAsesorComercialPlan(dbConn *sql.DB, id int64, comisionVentaPct, comisionPagoPct float64, mesesRenovacion int, notas string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE asesor_comercial SET comision_venta_pct = ?, comision_pago_pct = ?, meses_renovacion = ?, notas = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", comisionVentaPct, comisionPagoPct, mesesRenovacion, notas, id)
	return err
}

func DeleteAsesorComercialPlan(dbConn *sql.DB, id int64) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE asesor_comercial SET estado = 'inactivo', fecha_actualizacion = "+nowExpr+" WHERE id = ?", id)
	return err
}

// Registrar comisiones generadas por pagos/activaciones
func CreateAsesorComisionRecord(dbConn *sql.DB, asesorID string, empresaID, licenciaID, pagoID int64, transactionID string, montoTotal, porcentaje, montoComision float64, referencia, observaciones, programadoPara string, pagado int) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO asesor_comisiones (asesor_id, empresa_id, licencia_id, pago_id, transaction_id, monto_total, porcentaje, monto_comision, referencia, observaciones, programado_para, pagado, fecha_creacion) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " + nowExpr + ")"
	return insertSQLCompat(dbConn, query, asesorID, empresaID, licenciaID, pagoID, transactionID, montoTotal, porcentaje, montoComision, referencia, observaciones, programadoPara, pagado)
}

// GetWompiPaymentContext devuelve licencia_id y empresa_id para una transaccion/referencia Wompi.
func GetWompiPaymentContext(dbConn *sql.DB, transactionID, reference string) (int64, int64, bool, error) {
	row := queryRowSQLCompat(dbConn, `
		SELECT licencia_id, empresa_id
		FROM pagos_wompi
		WHERE (transaction_id = ? AND ? <> '') OR (reference = ? AND ? <> '')
		ORDER BY id DESC
		LIMIT 1
	`, transactionID, transactionID, reference, reference)

	var licenciaID sql.NullInt64
	var empresaID sql.NullInt64
	if err := row.Scan(&licenciaID, &empresaID); err != nil {
		if err == sql.ErrNoRows {
			return 0, 0, false, nil
		}
		return 0, 0, false, err
	}

	if !licenciaID.Valid || !empresaID.Valid {
		return 0, 0, false, nil
	}

	return licenciaID.Int64, empresaID.Int64, true, nil
}

// ActivateLicenciaForEmpresa asigna y activa una licencia para una empresa, estableciendo fechas de inicio y fin
func ActivateLicenciaForEmpresa(dbConn *sql.DB, licenciaID, empresaID int64, fechaInicio, fechaFin string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE licencias SET empresa_id = ?, activo = 1, fecha_inicio = ?, fecha_fin = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", empresaID, fechaInicio, fechaFin, licenciaID)
	if err == nil {
		return nil
	}
	// Compatibilidad con bases antiguas que no tienen fecha_actualizacion.
	_, fallbackErr := execSQLCompat(dbConn, "UPDATE licencias SET empresa_id = ?, activo = 1, fecha_inicio = ?, fecha_fin = ? WHERE id = ?", empresaID, fechaInicio, fechaFin, licenciaID)
	if fallbackErr == nil {
		return nil
	}
	return fallbackErr
}

// SetConfigValue inserta o actualiza una configuración en la tabla configuraciones
func SetConfigValue(dbConn *sql.DB, key, value string, encrypted bool) error {
	enc := 0
	if encrypted {
		enc = 1
	}
	// Preferimos mantener fecha_creacion en la fila original.
	// Si existe la clave hacemos UPDATE y seteamos fecha_actualizacion,
	// si no existe hacemos INSERT con fecha_creacion y fecha_actualizacion.
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	nowExpr := sqlNowExpr()

	var existing string
	err = queryRowTxSQLCompat(tx, "SELECT config_key FROM configuraciones WHERE config_key = ? LIMIT 1", key).Scan(&existing)
	if err != nil {
		if err == sql.ErrNoRows {
			_, err = execTxSQLCompat(tx, "INSERT INTO configuraciones (config_key, value, encrypted, fecha_creacion, fecha_actualizacion) VALUES (?, ?, ?, "+nowExpr+", "+nowExpr+")", key, value, enc)
			if err != nil {
				return err
			}
			return tx.Commit()
		}
		return err
	}

	_, err = execTxSQLCompat(tx, "UPDATE configuraciones SET value = ?, encrypted = ?, fecha_actualizacion = "+nowExpr+" WHERE config_key = ?", value, enc, key)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetConfigEntry devuelve el valor almacenado, si está cifrado, la fecha de creación y la fecha de última actualización.
// Si la clave no existe devuelve valores vacíos y nil error.
func GetConfigEntry(dbConn *sql.DB, key string) (string, bool, string, string, error) {
	row := queryRowSQLCompat(dbConn, "SELECT value, encrypted, fecha_creacion, fecha_actualizacion FROM configuraciones WHERE config_key = ? LIMIT 1", key)
	var val sql.NullString
	var enc sql.NullInt64
	var fechaCre sql.NullString
	var fechaAct sql.NullString
	if err := row.Scan(&val, &enc, &fechaCre, &fechaAct); err != nil {
		if err == sql.ErrNoRows {
			return "", false, "", "", nil
		}
		return "", false, "", "", err
	}
	v := ""
	if val.Valid {
		v = val.String
	}
	isEnc := false
	if enc.Valid && enc.Int64 == 1 {
		isEnc = true
	}
	fc := ""
	if fechaCre.Valid {
		fc = fechaCre.String
	}
	fa := ""
	if fechaAct.Valid {
		fa = fechaAct.String
	}
	return v, isEnc, fc, fa, nil
}

// GetConfigValue devuelve el valor almacenado y si estaba cifrado
func GetConfigValue(dbConn *sql.DB, key string) (string, bool, error) {
	row := queryRowSQLCompat(dbConn, "SELECT value, encrypted FROM configuraciones WHERE config_key = ? LIMIT 1", key)
	var val sql.NullString
	var enc sql.NullInt64
	if err := row.Scan(&val, &enc); err != nil {
		return "", false, err
	}
	v := ""
	if val.Valid {
		v = val.String
	}
	isEnc := false
	if enc.Valid && enc.Int64 == 1 {
		isEnc = true
	}
	return v, isEnc, nil
}

// InsertMetric inserta una muestra de métricas en la tabla metrics
func InsertMetric(dbConn *sql.DB, cpuPercent float64, memTotal, memUsed uint64, memPercent float64, netRecv, netSent uint64) error {
	_, err := execSQLCompat(dbConn, "INSERT INTO metrics (cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent) VALUES (?, ?, ?, ?, ?, ?)",
		cpuPercent, memTotal, memUsed, memPercent, netRecv, netSent)
	return err
}

// GetLatestMetric obtiene la última muestra registrada
func GetLatestMetric(dbConn *sql.DB) (*Metric, error) {
	row := queryRowSQLCompat(dbConn, "SELECT id, timestamp, cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent, fecha_creacion FROM metrics ORDER BY id DESC LIMIT 1")
	var m Metric
	var timestamp sql.NullString
	var fechaCre sql.NullString
	if err := row.Scan(&m.ID, &timestamp, &m.CPUPercent, &m.MemTotal, &m.MemUsed, &m.MemPercent, &m.NetRecv, &m.NetSent, &fechaCre); err != nil {
		return nil, err
	}
	if timestamp.Valid {
		m.Timestamp = timestamp.String
	}
	if fechaCre.Valid {
		m.FechaCreacion = fechaCre.String
	}
	return &m, nil
}

// GetMetricsHistory devuelve las últimas 'limit' muestras (ordenadas de más antiguo a más reciente)
func GetMetricsHistory(dbConn *sql.DB, limit int) ([]Metric, error) {
	q := "SELECT id, timestamp, cpu_percent, mem_total, mem_used, mem_percent, net_recv, net_sent, fecha_creacion FROM metrics ORDER BY id DESC LIMIT ?"
	rows, err := querySQLCompat(dbConn, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Metric
	for rows.Next() {
		var m Metric
		var timestamp sql.NullString
		var fechaCre sql.NullString
		if err := rows.Scan(&m.ID, &timestamp, &m.CPUPercent, &m.MemTotal, &m.MemUsed, &m.MemPercent, &m.NetRecv, &m.NetSent, &fechaCre); err != nil {
			return nil, err
		}
		if timestamp.Valid {
			m.Timestamp = timestamp.String
		}
		if fechaCre.Valid {
			m.FechaCreacion = fechaCre.String
		}
		out = append(out, m)
	}
	// invertir slice para devolver de más antiguo a más reciente
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}
