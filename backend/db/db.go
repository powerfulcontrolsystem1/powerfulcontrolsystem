package db

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"
	"time"
)

var (
	licenciaPermisoPolicyCacheMu  sync.Mutex
	licenciaPermisoPolicyCache    = map[int64]cachedLicenciaPermisoPolicy{}
	licenciaPermisoPolicyCacheTTL = 60 * time.Second

	adminByEmailCacheMu  sync.Mutex
	adminByEmailCache    = map[string]cachedAdminByEmail{}
	adminByEmailCacheTTL = 60 * time.Second

	empresaByScopeCacheMu  sync.Mutex
	empresaByScopeCache    = map[int64]cachedEmpresaByScope{}
	empresaByScopeCacheTTL = 60 * time.Second
)

type cachedLicenciaPermisoPolicy struct {
	Policy   *LicenciaPermisoPolicy
	LoadedAt time.Time
}

// InvalidateLicenciaPermisoPolicyCacheForEmpresa descarta el permiso efectivo
// calculado desde licencias para que activaciones/desactivaciones se reflejen
// inmediatamente en menus, roles y protecciones de API.
func InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID int64) {
	if empresaID <= 0 {
		return
	}
	licenciaPermisoPolicyCacheMu.Lock()
	delete(licenciaPermisoPolicyCache, empresaID)
	licenciaPermisoPolicyCacheMu.Unlock()
}

func invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn *sql.DB, licenciaID int64) {
	if dbConn == nil || licenciaID <= 0 {
		return
	}
	empresaIDs := map[int64]struct{}{}

	rows, err := querySQLCompat(dbConn, `SELECT empresa_id FROM licencias WHERE id = ? AND COALESCE(empresa_id, 0) > 0`, licenciaID)
	if err == nil {
		for rows.Next() {
			var empresaID sql.NullInt64
			if scanErr := rows.Scan(&empresaID); scanErr == nil && empresaID.Valid && empresaID.Int64 > 0 {
				empresaIDs[empresaID.Int64] = struct{}{}
			}
		}
		_ = rows.Close()
	}

	addonRows, addonErr := querySQLCompat(dbConn, `SELECT empresa_id FROM empresa_licencias_adicionales WHERE licencia_id = ? AND COALESCE(empresa_id, 0) > 0`, licenciaID)
	if addonErr == nil {
		for addonRows.Next() {
			var empresaID sql.NullInt64
			if scanErr := addonRows.Scan(&empresaID); scanErr == nil && empresaID.Valid && empresaID.Int64 > 0 {
				empresaIDs[empresaID.Int64] = struct{}{}
			}
		}
		_ = addonRows.Close()
	}

	for empresaID := range empresaIDs {
		InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	}
}

type cachedAdminByEmail struct {
	Admin    *Admin
	LoadedAt time.Time
}

type cachedEmpresaByScope struct {
	Empresa  *Empresa
	LoadedAt time.Time
}

// EnsureAdministradoresAuthSchema regulariza las columnas operativas y de seguridad
// usadas por el flujo administrativo en PostgreSQL.
func EnsureAdministradoresAuthSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}

	statements := []string{
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS photo TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS acepta_contrato INTEGER DEFAULT 0`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS telefono TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS pais TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS ciudad TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS email_confirm_token TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS email_confirm_expira TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS email_confirmado INTEGER DEFAULT 0`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS email_confirmado_en TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS password_hash TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS password_salt TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS password_set INTEGER DEFAULT 0`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS password_reset_token TEXT`,
		`ALTER TABLE administradores ADD COLUMN IF NOT EXISTS password_reset_expira TEXT`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// EnsurePaymentGatewaySchema prepara las tablas de checkout de licencias en PostgreSQL.
func EnsurePaymentGatewaySchema(dbConn *sql.DB) error {
	if dbConn == nil || !isPostgresDialect() {
		return nil
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS pagos_wompi (
			id BIGSERIAL PRIMARY KEY,
			licencia_id BIGINT,
			empresa_id BIGINT,
			transaction_id TEXT,
			reference TEXT,
			status TEXT,
			raw_payload TEXT,
			discount_code TEXT,
			asesor_id TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS discount_code TEXT`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS asesor_id TEXT`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activo'`,
		`ALTER TABLE pagos_wompi ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`CREATE INDEX IF NOT EXISTS ix_pagos_wompi_transaction_id ON pagos_wompi (transaction_id)`,
		`CREATE INDEX IF NOT EXISTS ix_pagos_wompi_reference ON pagos_wompi (reference)`,
		`CREATE TABLE IF NOT EXISTS pagos_epayco (
			id BIGSERIAL PRIMARY KEY,
			licencia_id BIGINT,
			empresa_id BIGINT,
			transaction_id TEXT,
			reference TEXT,
			status TEXT,
			raw_payload TEXT,
			discount_code TEXT,
			asesor_id TEXT,
			fecha_creacion TEXT DEFAULT CURRENT_TIMESTAMP::text,
			fecha_actualizacion TEXT,
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS discount_code TEXT`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS asesor_id TEXT`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activo'`,
		`ALTER TABLE pagos_epayco ADD COLUMN IF NOT EXISTS observaciones TEXT`,
		`CREATE INDEX IF NOT EXISTS ix_pagos_epayco_transaction_id ON pagos_epayco (transaction_id)`,
		`CREATE INDEX IF NOT EXISTS ix_pagos_epayco_reference ON pagos_epayco (reference)`,
	}

	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	return nil
}

// EnsureLicenciasSchema regulariza la tabla licencias (PostgreSQL-only).
func EnsureLicenciasSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}

	statements := []string{
		`CREATE TABLE IF NOT EXISTS licencias (
			id BIGSERIAL PRIMARY KEY,
			empresa_id BIGINT,
			tipo_id BIGINT,
			pais_codigo TEXT DEFAULT 'CO',
			nombre TEXT,
			descripcion TEXT,
			valor DOUBLE PRECISION DEFAULT 0,
			duracion_dias INTEGER DEFAULT 0,
			modulos_habilitados TEXT,
			es_adicional INTEGER DEFAULT 0,
			codigo_funcion TEXT,
			super_rol_habilitado INTEGER DEFAULT 0,
			fecha_inicio TEXT,
			fecha_fin TEXT,
			activo INTEGER DEFAULT 1,
			fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		)`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS empresa_id BIGINT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS tipo_id BIGINT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS pais_codigo TEXT DEFAULT 'CO'`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS nombre TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS descripcion TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS valor DOUBLE PRECISION DEFAULT 0`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS duracion_dias INTEGER DEFAULT 0`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS modulos_habilitados TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS es_adicional INTEGER DEFAULT 0`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS codigo_funcion TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS super_rol_habilitado INTEGER DEFAULT 0`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS fecha_inicio TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS fecha_fin TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS activo INTEGER DEFAULT 1`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS fecha_creacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS fecha_actualizacion TEXT DEFAULT CAST(CURRENT_TIMESTAMP AS TEXT)`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS usuario_creador TEXT`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS estado TEXT DEFAULT 'activo'`,
		`ALTER TABLE licencias ADD COLUMN IF NOT EXISTS observaciones TEXT`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}

	// Backfill: licencias ya creadas deben quedar en Colombia por defecto.
	// Es idempotente y no afecta registros ya definidos.
	_, _ = dbConn.Exec(`UPDATE licencias SET pais_codigo = 'CO' WHERE COALESCE(TRIM(pais_codigo),'') = ''`)
	return nil
}

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
	return UpsertAdministradorConCreador(dbConn, email, name, role, photo, "")
}

// UpsertAdministradorConCreador inserta o actualiza un administrador conservando el administrador principal.
func UpsertAdministradorConCreador(dbConn *sql.DB, email, name, role, photo, usuarioCreador string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	nowExpr := sqlNowExpr()
	upsertSQL := "INSERT INTO administradores (email, name, role, photo, usuario_creador, fecha_creacion, fecha_actualizacion, estado) VALUES (?, ?, ?, ?, ?, " + nowExpr + ", " + nowExpr + ", 'activo') ON CONFLICT(email) DO UPDATE SET name = EXCLUDED.name, role = EXCLUDED.role, photo = EXCLUDED.photo, usuario_creador = CASE WHEN TRIM(COALESCE(administradores.usuario_creador, '')) <> '' THEN administradores.usuario_creador ELSE EXCLUDED.usuario_creador END, fecha_actualizacion = " + nowExpr
	if _, err := execTxSQLCompat(tx, upsertSQL, email, name, role, photo, strings.TrimSpace(usuarioCreador)); err != nil {
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
	UsuarioCreador     string `json:"usuario_creador,omitempty"`
	FechaCreacion      string `json:"fecha_creacion"`
	FechaActualizacion string `json:"fecha_actualizacion"`
	Estado             string `json:"estado"`
	AceptaContrato     int    `json:"acepta_contrato"`
	Telefono           string `json:"telefono,omitempty"`
	Pais               string `json:"pais,omitempty"`
	Ciudad             string `json:"ciudad,omitempty"`
	// Campos de seguridad y confirmación
	EmailConfirmado     int    `json:"email_confirmado,omitempty"`
	EmailConfirmToken   string `json:"-"`
	EmailConfirmExpira  string `json:"-"`
	EmailConfirmadoEn   string `json:"email_confirmado_en,omitempty"`
	PasswordSet         int    `json:"password_set,omitempty"`
	PasswordHash        string `json:"-"`
	PasswordSalt        string `json:"-"`
	PasswordResetToken  string `json:"-"`
	PasswordResetExpira string `json:"-"`
}

// NOTE: tipos_de_licencia CRUD removed per project decision (frontend/page/link removed).

// Licencia representa una licencia asignada (nuevo CRUD)
type Licencia struct {
	ID            int64   `json:"id"`
	EmpresaID     int64   `json:"empresa_id"`
	EmpresaNombre string  `json:"empresa_nombre,omitempty"`
	TipoID        int64   `json:"tipo_id"`
	TipoNombre    string  `json:"tipo_nombre,omitempty"`
	PaisCodigo    string  `json:"pais_codigo,omitempty"`
	Nombre        string  `json:"nombre"`
	Descripcion   string  `json:"descripcion"`
	Valor         float64 `json:"valor"`
	DuracionDias  int     `json:"duracion_dias"`
	ModulosHab    string  `json:"modulos_habilitados,omitempty"`
	EsAdicional   int     `json:"es_adicional"`
	CodigoFuncion string  `json:"codigo_funcion,omitempty"`
	SuperRol      int     `json:"super_rol_habilitado"`
	FechaInicio   string  `json:"fecha_inicio,omitempty"`
	FechaFin      string  `json:"fecha_fin,omitempty"`
	FechaCreacion string  `json:"fecha_creacion"`
	Activo        int     `json:"activo"`
}

// CreateLicencia inserta una nueva licencia en dbSuper
func CreateLicencia(dbConn *sql.DB, tipoID int64, paisCodigo, nombre, descripcion string, valor float64, duracionDias int, modulosHabilitados string, superRolHabilitado int) (int64, error) {
	return CreateLicenciaAdvanced(dbConn, tipoID, paisCodigo, nombre, descripcion, valor, duracionDias, modulosHabilitados, 0, "", superRolHabilitado)
}

func CreateLicenciaAdvanced(dbConn *sql.DB, tipoID int64, paisCodigo, nombre, descripcion string, valor float64, duracionDias int, modulosHabilitados string, esAdicional int, codigoFuncion string, superRolHabilitado int) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO licencias (tipo_id, pais_codigo, nombre, descripcion, valor, duracion_dias, modulos_habilitados, es_adicional, codigo_funcion, super_rol_habilitado, fecha_creacion, fecha_actualizacion, activo, estado) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, " + nowExpr + ", " + nowExpr + ", 1, 'activo')"
	id, err := insertSQLCompat(dbConn, query, tipoID, strings.TrimSpace(paisCodigo), nombre, descripcion, valor, duracionDias, strings.TrimSpace(modulosHabilitados), esAdicional, strings.TrimSpace(codigoFuncion), superRolHabilitado)
	if err == nil {
		return id, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return 0, err
	}
	if schemaErr := EnsureLicenciasSchema(dbConn); schemaErr != nil {
		return 0, err
	}
	return insertSQLCompat(dbConn, query, tipoID, strings.TrimSpace(paisCodigo), nombre, descripcion, valor, duracionDias, strings.TrimSpace(modulosHabilitados), esAdicional, strings.TrimSpace(codigoFuncion), superRolHabilitado)
}

// GetLicencias obtiene todas las licencias (comportamiento legado sin filtros)
func GetLicencias(dbConn *sql.DB) ([]Licencia, error) {
	return GetLicenciasFiltered(dbConn, false, "", false)
}

// GetLicenciasFilteredByPais filtra licencias por país (best-effort).
// Si paisCodigo está vacío, retorna todas las licencias.
func GetLicenciasFilteredByPais(dbConn *sql.DB, soloActivas bool, usuarioCreador string, conEmpresa bool, paisCodigo string) ([]Licencia, error) {
	rows, err := GetLicenciasFiltered(dbConn, soloActivas, usuarioCreador, conEmpresa)
	if err != nil {
		return nil, err
	}
	pais := strings.ToUpper(strings.TrimSpace(paisCodigo))
	if pais == "" {
		return rows, nil
	}
	out := make([]Licencia, 0, len(rows))
	for _, item := range rows {
		code := strings.ToUpper(strings.TrimSpace(item.PaisCodigo))
		if code == "" {
			code = "CO"
		}
		if code == pais {
			out = append(out, item)
		}
	}
	return out, nil
}

// GetLicenciasFiltered obtiene licencias con filtros opcionales.
func GetLicenciasFiltered(dbConn *sql.DB, soloActivas bool, usuarioCreador string, conEmpresa bool) ([]Licencia, error) {
	q := `SELECT l.id, l.empresa_id, l.tipo_id, t.nombre, COALESCE(l.pais_codigo,'CO'), l.nombre, l.descripcion, l.valor, l.duracion_dias, COALESCE(l.modulos_habilitados, ''), COALESCE(l.es_adicional, 0), COALESCE(l.codigo_funcion, ''), COALESCE(l.super_rol_habilitado, 0), COALESCE(l.fecha_inicio, ''), COALESCE(l.fecha_fin, ''), l.fecha_creacion, l.activo`
	baseFrom := `
		FROM licencias l LEFT JOIN tipos_de_empresas t ON l.tipo_id = t.id`
	q += baseFrom

	usuarioCreador = strings.TrimSpace(usuarioCreador)
	hasEmpresasTable, err := tableExists(dbConn, "empresas")
	if err != nil {
		return nil, err
	}
	if hasEmpresasTable {
		q += " LEFT JOIN empresas e ON e.id = l.empresa_id"
		q = strings.Replace(q, "SELECT l.id, l.empresa_id", "SELECT l.id, l.empresa_id, COALESCE(e.nombre, '')", 1)
	} else {
		q = strings.Replace(q, "SELECT l.id, l.empresa_id", "SELECT l.id, l.empresa_id, ''", 1)
	}
	canFilterByUsuarioCreador := false
	if usuarioCreador != "" {
		if hasEmpresasTable {
			canFilterByUsuarioCreador = true
		}
	}

	var where []string
	var args []interface{}
	if soloActivas {
		where = append(where, "l.activo = 1")
		if isPostgresDialect() {
			where = append(where, "(COALESCE(CAST(l.fecha_inicio AS TEXT), '') = '' OR CAST(l.fecha_inicio AS TIMESTAMP) <= CURRENT_TIMESTAMP)")
			where = append(where, "(COALESCE(CAST(l.fecha_fin AS TEXT), '') = '' OR CAST(l.fecha_fin AS TIMESTAMP) >= CURRENT_TIMESTAMP)")
		} else {
			where = append(where, "(COALESCE(l.fecha_inicio, '') = '' OR datetime(l.fecha_inicio) <= datetime('now','localtime'))")
			where = append(where, "(COALESCE(l.fecha_fin, '') = '' OR datetime(l.fecha_fin) >= datetime('now','localtime'))")
		}
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
		if !isMissingTableError(err) && !isMissingColumnError(err) {
			return nil, err
		}
		if schemaErr := EnsureLicenciasSchema(dbConn); schemaErr != nil {
			return nil, err
		}
		rows, err = querySQLCompat(dbConn, q, args...)
		if err != nil {
			return nil, err
		}
	}
	defer rows.Close()
	var out []Licencia
	for rows.Next() {
		var lic Licencia
		var empresaID sql.NullInt64
		var empresaNombre sql.NullString
		var tipoNombre sql.NullString
		var paisCodigo sql.NullString
		var descripcion sql.NullString
		var modulosHab sql.NullString
		var fechaInicio sql.NullString
		var fechaFin sql.NullString
		var fechaCreacion sql.NullString
		if err := rows.Scan(&lic.ID, &empresaID, &empresaNombre, &lic.TipoID, &tipoNombre, &paisCodigo, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &modulosHab, &lic.EsAdicional, &lic.CodigoFuncion, &lic.SuperRol, &fechaInicio, &fechaFin, &fechaCreacion, &lic.Activo); err != nil {
			return nil, err
		}
		if empresaID.Valid {
			lic.EmpresaID = empresaID.Int64
		}
		if empresaNombre.Valid {
			lic.EmpresaNombre = empresaNombre.String
		}
		if tipoNombre.Valid {
			lic.TipoNombre = tipoNombre.String
		}
		if paisCodigo.Valid && strings.TrimSpace(paisCodigo.String) != "" {
			lic.PaisCodigo = strings.TrimSpace(paisCodigo.String)
		} else {
			lic.PaisCodigo = "CO"
		}
		if descripcion.Valid {
			lic.Descripcion = descripcion.String
		}
		if modulosHab.Valid {
			lic.ModulosHab = modulosHab.String
		}
		if fechaInicio.Valid {
			lic.FechaInicio = fechaInicio.String
		}
		if fechaFin.Valid {
			lic.FechaFin = fechaFin.String
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
	q := `SELECT id, empresa_id, tipo_id, COALESCE(pais_codigo,'CO'), nombre, descripcion, valor, duracion_dias, COALESCE(modulos_habilitados, ''), COALESCE(es_adicional, 0), COALESCE(codigo_funcion, ''), COALESCE(super_rol_habilitado, 0), COALESCE(fecha_inicio, ''), COALESCE(fecha_fin, ''), fecha_creacion, activo FROM licencias WHERE id = ? LIMIT 1`
	scanLicencia := func() (*Licencia, error) {
		row := queryRowSQLCompat(dbConn, q, id)
		var lic Licencia
		var empresaID sql.NullInt64
		var paisCodigo sql.NullString
		var descripcion sql.NullString
		var modulosHab sql.NullString
		var fechaInicio sql.NullString
		var fechaFin sql.NullString
		var fechaCreacion sql.NullString
		if err := row.Scan(&lic.ID, &empresaID, &lic.TipoID, &paisCodigo, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &modulosHab, &lic.EsAdicional, &lic.CodigoFuncion, &lic.SuperRol, &fechaInicio, &fechaFin, &fechaCreacion, &lic.Activo); err != nil {
			return nil, err
		}
		if empresaID.Valid {
			lic.EmpresaID = empresaID.Int64
		}
		if paisCodigo.Valid && strings.TrimSpace(paisCodigo.String) != "" {
			lic.PaisCodigo = strings.TrimSpace(paisCodigo.String)
		} else {
			lic.PaisCodigo = "CO"
		}
		if descripcion.Valid {
			lic.Descripcion = descripcion.String
		}
		if modulosHab.Valid {
			lic.ModulosHab = modulosHab.String
		}
		if fechaInicio.Valid {
			lic.FechaInicio = fechaInicio.String
		}
		if fechaFin.Valid {
			lic.FechaFin = fechaFin.String
		}
		if fechaCreacion.Valid {
			lic.FechaCreacion = fechaCreacion.String
		}
		return &lic, nil
	}

	lic, err := scanLicencia()
	if err == nil {
		return lic, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return nil, err
	}
	if schemaErr := EnsureLicenciasSchema(dbConn); schemaErr != nil {
		return nil, err
	}
	return scanLicencia()
}

// GetActiveLicenciaByEmpresa devuelve la licencia activa vigente mas reciente de una empresa.
func GetActiveLicenciaByEmpresa(dbConn *sql.DB, empresaID int64) (*Licencia, error) {
	if empresaID <= 0 {
		return nil, sql.ErrNoRows
	}
	q := `SELECT id, empresa_id, tipo_id, COALESCE(pais_codigo,'CO'), nombre, descripcion, valor, duracion_dias, COALESCE(modulos_habilitados, ''), COALESCE(es_adicional, 0), COALESCE(codigo_funcion, ''), COALESCE(super_rol_habilitado, 0), COALESCE(fecha_inicio, ''), COALESCE(fecha_fin, ''), fecha_creacion, activo
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
		q = `SELECT id, empresa_id, tipo_id, COALESCE(pais_codigo,'CO'), nombre, descripcion, valor, duracion_dias, COALESCE(modulos_habilitados, ''), COALESCE(es_adicional, 0), COALESCE(codigo_funcion, ''), COALESCE(super_rol_habilitado, 0), COALESCE(fecha_inicio, ''), COALESCE(fecha_fin, ''), fecha_creacion, activo
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
	row := queryRowSQLCompat(dbConn, q, empresaID)
	var lic Licencia
	var empresaIDVal sql.NullInt64
	var paisCodigo sql.NullString
	var descripcion sql.NullString
	var modulosHab sql.NullString
	var fechaInicio sql.NullString
	var fechaFin sql.NullString
	var fechaCreacion sql.NullString
	if err := row.Scan(&lic.ID, &empresaIDVal, &lic.TipoID, &paisCodigo, &lic.Nombre, &descripcion, &lic.Valor, &lic.DuracionDias, &modulosHab, &lic.EsAdicional, &lic.CodigoFuncion, &lic.SuperRol, &fechaInicio, &fechaFin, &fechaCreacion, &lic.Activo); err != nil {
		return nil, err
	}
	if empresaIDVal.Valid {
		lic.EmpresaID = empresaIDVal.Int64
	}
	if paisCodigo.Valid && strings.TrimSpace(paisCodigo.String) != "" {
		lic.PaisCodigo = strings.TrimSpace(paisCodigo.String)
	} else {
		lic.PaisCodigo = "CO"
	}
	if descripcion.Valid {
		lic.Descripcion = descripcion.String
	}
	if modulosHab.Valid {
		lic.ModulosHab = modulosHab.String
	}
	if fechaInicio.Valid {
		lic.FechaInicio = fechaInicio.String
	}
	if fechaFin.Valid {
		lic.FechaFin = fechaFin.String
	}
	if fechaCreacion.Valid {
		lic.FechaCreacion = fechaCreacion.String
	}
	return &lic, nil
}

// UpdateLicencia actualiza campos editables de una licencia
func UpdateLicencia(dbConn *sql.DB, id, tipoID int64, paisCodigo, nombre, descripcion string, valor float64, duracionDias int, modulosHabilitados string, superRolHabilitado int) error {
	return UpdateLicenciaAdvanced(dbConn, id, tipoID, paisCodigo, nombre, descripcion, valor, duracionDias, modulosHabilitados, 0, "", superRolHabilitado)
}

func UpdateLicenciaAdvanced(dbConn *sql.DB, id, tipoID int64, paisCodigo, nombre, descripcion string, valor float64, duracionDias int, modulosHabilitados string, esAdicional int, codigoFuncion string, superRolHabilitado int) error {
	nowExpr := sqlNowExpr()
	query := "UPDATE licencias SET tipo_id = ?, pais_codigo = ?, nombre = ?, descripcion = ?, valor = ?, duracion_dias = ?, modulos_habilitados = ?, es_adicional = ?, codigo_funcion = ?, super_rol_habilitado = ?, fecha_actualizacion = " + nowExpr + " WHERE id = ?"
	fallbackQuery := "UPDATE licencias SET tipo_id = ?, pais_codigo = ?, nombre = ?, descripcion = ?, valor = ?, duracion_dias = ?, modulos_habilitados = ?, es_adicional = ?, codigo_funcion = ?, super_rol_habilitado = ? WHERE id = ?"
	args := []interface{}{tipoID, strings.TrimSpace(paisCodigo), nombre, descripcion, valor, duracionDias, strings.TrimSpace(modulosHabilitados), esAdicional, strings.TrimSpace(codigoFuncion), superRolHabilitado, id}

	_, err := execSQLCompat(dbConn, query, args...)
	if err == nil {
		invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, id)
		return nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return err
	}
	if schemaErr := EnsureLicenciasSchema(dbConn); schemaErr == nil {
		_, retryErr := execSQLCompat(dbConn, query, args...)
		if retryErr == nil {
			invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, id)
			return nil
		}
		err = retryErr
	}
	if isMissingColumnError(err) && strings.Contains(strings.ToLower(err.Error()), "fecha_actualizacion") {
		_, fallbackErr := execSQLCompat(dbConn, fallbackQuery, args...)
		if fallbackErr == nil {
			invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, id)
			return nil
		}
		err = fallbackErr
	}
	return err
}

// LicenciaPermisoPolicy describe las capacidades de acceso habilitadas por licencia activa para una empresa.
type LicenciaPermisoPolicy struct {
	LicenciaID         int64
	Nombre             string
	ModulosHabilitados string
	SuperRolHabilitado bool
}

func mergeLicenciaModules(base string, extra ...string) string {
	if strings.TrimSpace(base) == "" {
		return ""
	}
	seen := map[string]struct{}{}
	out := make([]string, 0)
	for _, chunk := range append([]string{base}, extra...) {
		for _, raw := range strings.Split(chunk, ",") {
			module := strings.ToLower(strings.TrimSpace(raw))
			if module == "" {
				continue
			}
			if _, exists := seen[module]; exists {
				continue
			}
			seen[module] = struct{}{}
			out = append(out, module)
		}
	}
	return strings.Join(out, ",")
}

// GetLicenciaPermisoPolicyByEmpresa resuelve la licencia activa vigente para permisos de una empresa.
func GetLicenciaPermisoPolicyByEmpresa(dbConn *sql.DB, empresaID int64) (*LicenciaPermisoPolicy, error) {
	if dbConn == nil || empresaID <= 0 {
		return nil, nil
	}

	licenciaPermisoPolicyCacheMu.Lock()
	if cached, ok := licenciaPermisoPolicyCache[empresaID]; ok && time.Since(cached.LoadedAt) < licenciaPermisoPolicyCacheTTL {
		licenciaPermisoPolicyCacheMu.Unlock()
		if cached.Policy == nil {
			return nil, nil
		}
		copyPolicy := *cached.Policy
		return &copyPolicy, nil
	}
	licenciaPermisoPolicyCacheMu.Unlock()

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
			licenciaPermisoPolicyCacheMu.Lock()
			licenciaPermisoPolicyCache[empresaID] = cachedLicenciaPermisoPolicy{Policy: nil, LoadedAt: time.Now()}
			licenciaPermisoPolicyCacheMu.Unlock()
			return nil, nil
		}
		if isMissingTableError(err) || isMissingColumnError(err) {
			licenciaPermisoPolicyCacheMu.Lock()
			licenciaPermisoPolicyCache[empresaID] = cachedLicenciaPermisoPolicy{Policy: nil, LoadedAt: time.Now()}
			licenciaPermisoPolicyCacheMu.Unlock()
			return nil, nil
		}
		return nil, err
	}
	item.SuperRolHabilitado = superRolRaw == 1
	addons, addonsErr := ListEmpresaLicenciasAdicionales(dbConn, empresaID, false)
	if addonsErr == nil && strings.TrimSpace(item.ModulosHabilitados) != "" {
		extraModules := make([]string, 0, len(addons))
		for _, addon := range addons {
			if strings.TrimSpace(addon.ModulosHab) != "" {
				extraModules = append(extraModules, addon.ModulosHab)
			}
		}
		item.ModulosHabilitados = mergeLicenciaModules(item.ModulosHabilitados, extraModules...)
	}
	licenciaPermisoPolicyCacheMu.Lock()
	licenciaPermisoPolicyCache[empresaID] = cachedLicenciaPermisoPolicy{Policy: &item, LoadedAt: time.Now()}
	licenciaPermisoPolicyCacheMu.Unlock()
	copyPolicy := item
	return &copyPolicy, nil
}

// DeleteLicencia elimina una licencia por id
func DeleteLicencia(dbConn *sql.DB, id int64) error {
	_, err := execSQLCompat(dbConn, "DELETE FROM licencias WHERE id = ?", id)
	return err
}

// SetLicenciaActivo activa/desactiva una licencia (activo: 1 o 0)
func SetLicenciaActivo(dbConn *sql.DB, id int64, activo int) error {
	nowExpr := sqlNowExpr()
	query := "UPDATE licencias SET activo = ?, fecha_actualizacion = " + nowExpr + " WHERE id = ?"
	_, err := execSQLCompat(dbConn, query, activo, id)
	if err == nil {
		invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, id)
		return nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return err
	}
	if schemaErr := EnsureLicenciasSchema(dbConn); schemaErr == nil {
		_, retryErr := execSQLCompat(dbConn, query, activo, id)
		if retryErr == nil {
			invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, id)
			return nil
		}
		err = retryErr
	}
	if isMissingColumnError(err) && strings.Contains(strings.ToLower(err.Error()), "fecha_actualizacion") {
		_, fallbackErr := execSQLCompat(dbConn, "UPDATE licencias SET activo = ? WHERE id = ?", activo, id)
		if fallbackErr == nil {
			invalidateLicenciaPermisoPolicyCacheForLicencia(dbConn, id)
			return nil
		}
		err = fallbackErr
	}
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
	cacheKey := strings.ToLower(strings.TrimSpace(email))
	if cacheKey != "" {
		adminByEmailCacheMu.Lock()
		if cached, ok := adminByEmailCache[cacheKey]; ok && time.Since(cached.LoadedAt) < adminByEmailCacheTTL {
			adminByEmailCacheMu.Unlock()
			if cached.Admin == nil {
				return nil, sql.ErrNoRows
			}
			copyAdmin := *cached.Admin
			return &copyAdmin, nil
		}
		adminByEmailCacheMu.Unlock()
	}
	// Intentar obtener con la columna 'acepta_contrato' (para bases actualizadas).
	row := queryRowSQLCompat(dbConn, "SELECT id, email, name, role, photo, COALESCE(usuario_creador, ''), fecha_creacion, fecha_actualizacion, estado, COALESCE(acepta_contrato, 0) FROM administradores WHERE email = ? LIMIT 1", email)
	var a Admin
	var photo sql.NullString
	var usuarioCreador sql.NullString
	var acepta sql.NullInt64
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &usuarioCreador, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado, &acepta); err != nil {
		// Fallback: si la columna no existe en este esquema, consultar sin ella.
		if isMissingColumnError(err) {
			row2 := queryRowSQLCompat(dbConn, "SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado FROM administradores WHERE email = ? LIMIT 1", email)
			var photo2 sql.NullString
			if err2 := row2.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo2, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado); err2 != nil {
				if cacheKey != "" && err2 == sql.ErrNoRows {
					adminByEmailCacheMu.Lock()
					adminByEmailCache[cacheKey] = cachedAdminByEmail{Admin: nil, LoadedAt: time.Now()}
					adminByEmailCacheMu.Unlock()
				}
				return nil, err2
			}
			if photo2.Valid {
				a.Photo = photo2.String
			}
			a.AceptaContrato = 0
			if cacheKey != "" {
				adminCopy := a
				adminByEmailCacheMu.Lock()
				adminByEmailCache[cacheKey] = cachedAdminByEmail{Admin: &adminCopy, LoadedAt: time.Now()}
				adminByEmailCacheMu.Unlock()
			}
			out := a
			return &out, nil
		}
		if cacheKey != "" && err == sql.ErrNoRows {
			adminByEmailCacheMu.Lock()
			adminByEmailCache[cacheKey] = cachedAdminByEmail{Admin: nil, LoadedAt: time.Now()}
			adminByEmailCacheMu.Unlock()
		}
		return nil, err
	}
	if photo.Valid {
		a.Photo = photo.String
	}
	if usuarioCreador.Valid {
		a.UsuarioCreador = usuarioCreador.String
	}
	if acepta.Valid {
		a.AceptaContrato = int(acepta.Int64)
	} else {
		a.AceptaContrato = 0
	}
	if cacheKey != "" {
		adminCopy := a
		adminByEmailCacheMu.Lock()
		adminByEmailCache[cacheKey] = cachedAdminByEmail{Admin: &adminCopy, LoadedAt: time.Now()}
		adminByEmailCacheMu.Unlock()
	}
	out := a
	return &out, nil
}

// GetAdminByID devuelve el administrador por id.
func GetAdminByID(dbConn *sql.DB, id int64) (*Admin, error) {
	row := queryRowSQLCompat(dbConn, "SELECT id, email, name, role, photo, COALESCE(usuario_creador, ''), fecha_creacion, fecha_actualizacion, estado, COALESCE(acepta_contrato, 0) FROM administradores WHERE id = ? LIMIT 1", id)
	var a Admin
	var photo sql.NullString
	var usuarioCreador sql.NullString
	var acepta sql.NullInt64
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &usuarioCreador, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado, &acepta); err != nil {
		if isMissingColumnError(err) {
			row2 := queryRowSQLCompat(dbConn, "SELECT id, email, name, role, photo, fecha_creacion, fecha_actualizacion, estado FROM administradores WHERE id = ? LIMIT 1", id)
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
	if usuarioCreador.Valid {
		a.UsuarioCreador = usuarioCreador.String
	}
	if acepta.Valid {
		a.AceptaContrato = int(acepta.Int64)
	}
	return &a, nil
}

// GetAdministradores lista todos los administradores
func GetAdministradores(dbConn *sql.DB) ([]Admin, error) {
	// Intentar consulta que incluya la columna acepta_contrato cuando exista
	rows, err := querySQLCompat(dbConn, "SELECT id, email, name, role, photo, COALESCE(usuario_creador, ''), fecha_creacion, fecha_actualizacion, estado, COALESCE(acepta_contrato, 0) FROM administradores ORDER BY id DESC")
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
		var usuarioCreador sql.NullString
		var acepta sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &usuarioCreador, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado, &acepta); err != nil {
			return nil, err
		}
		if photo.Valid {
			a.Photo = photo.String
		}
		if usuarioCreador.Valid {
			a.UsuarioCreador = usuarioCreador.String
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

// GetAdminByEmailFull devuelve el administrador por email incluyendo campos seguridad (tokens, hash, salt)
func GetAdminByEmailFull(dbConn *sql.DB, email string) (*Admin, error) {
	row := queryRowSQLCompat(dbConn, `SELECT id, email, name, role, photo, COALESCE(usuario_creador, ''), fecha_creacion, fecha_actualizacion, estado, COALESCE(acepta_contrato, 0), COALESCE(telefono, ''), COALESCE(pais, ''), COALESCE(ciudad, ''), COALESCE(email_confirmado, 0), COALESCE(email_confirm_token, ''), COALESCE(email_confirm_expira, ''), COALESCE(email_confirmado_en, ''), COALESCE(password_set, 0), COALESCE(password_hash, ''), COALESCE(password_salt, ''), COALESCE(password_reset_token, ''), COALESCE(password_reset_expira, '') FROM administradores WHERE lower(email) = lower(?) LIMIT 1`, strings.TrimSpace(email))
	var a Admin
	var photo sql.NullString
	var usuarioCreador sql.NullString
	var acepta sql.NullInt64
	var telefono sql.NullString
	var pais sql.NullString
	var ciudad sql.NullString
	var emailConfirmado sql.NullInt64
	var emailConfirmToken sql.NullString
	var emailConfirmExpira sql.NullString
	var emailConfirmadoEn sql.NullString
	var passwordSet sql.NullInt64
	var passwordHash sql.NullString
	var passwordSalt sql.NullString
	var passwordResetToken sql.NullString
	var passwordResetExpira sql.NullString
	if err := row.Scan(&a.ID, &a.Email, &a.Name, &a.Role, &photo, &usuarioCreador, &a.FechaCreacion, &a.FechaActualizacion, &a.Estado, &acepta, &telefono, &pais, &ciudad, &emailConfirmado, &emailConfirmToken, &emailConfirmExpira, &emailConfirmadoEn, &passwordSet, &passwordHash, &passwordSalt, &passwordResetToken, &passwordResetExpira); err != nil {
		if isMissingColumnError(err) {
			// Fallback a la consulta previa
			return GetAdminByEmail(dbConn, email)
		}
		return nil, err
	}
	if photo.Valid {
		a.Photo = photo.String
	}
	a.UsuarioCreador = strings.TrimSpace(usuarioCreador.String)
	a.AceptaContrato = int(acepta.Int64)
	if telefono.Valid {
		a.Telefono = telefono.String
	}
	if pais.Valid {
		a.Pais = pais.String
	}
	if ciudad.Valid {
		a.Ciudad = ciudad.String
	}
	a.EmailConfirmado = int(emailConfirmado.Int64)
	a.EmailConfirmToken = emailConfirmToken.String
	a.EmailConfirmExpira = emailConfirmExpira.String
	a.EmailConfirmadoEn = emailConfirmadoEn.String
	a.PasswordSet = int(passwordSet.Int64)
	a.PasswordHash = passwordHash.String
	a.PasswordSalt = passwordSalt.String
	a.PasswordResetToken = passwordResetToken.String
	a.PasswordResetExpira = passwordResetExpira.String
	return &a, nil
}

// ResolveAdminPrincipalEmail devuelve el email del administrador principal asociado a un administrador.
func ResolveAdminPrincipalEmail(dbConn *sql.DB, email string) (string, error) {
	current := strings.ToLower(strings.TrimSpace(email))
	if current == "" {
		return "", nil
	}
	visited := map[string]bool{}
	for current != "" {
		if visited[current] {
			break
		}
		visited[current] = true
		admin, err := GetAdminByEmailFull(dbConn, current)
		if err != nil {
			if err == sql.ErrNoRows {
				break
			}
			return "", err
		}
		creator := strings.ToLower(strings.TrimSpace(admin.UsuarioCreador))
		if creator == "" || creator == current {
			return current, nil
		}
		current = creator
	}
	return strings.ToLower(strings.TrimSpace(email)), nil
}

// UpdateAdministradorProfile actualiza campos del perfil del administrador identificando por id.
func UpdateAdministradorProfile(dbConn *sql.DB, id int64, name, telefono, email, pais, ciudad string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET name = ?, telefono = ?, email = ?, pais = ?, ciudad = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", strings.TrimSpace(name), strings.TrimSpace(telefono), strings.TrimSpace(email), strings.TrimSpace(pais), strings.TrimSpace(ciudad), id)
	return err
}

// ReassignSessionsAdminEmail actualiza las sesiones activas para reflejar nuevo email de administrador.
func ReassignSessionsAdminEmail(dbConn *sql.DB, oldEmail, newEmail string) error {
	_, err := execSQLCompat(dbConn, "UPDATE sesiones SET admin_email = ? WHERE admin_email = ? AND activo = 1", strings.TrimSpace(newEmail), strings.TrimSpace(oldEmail))
	return err
}

// SetAdministradorConfirmToken actualiza el token de confirmación para un administrador.
func SetAdministradorConfirmToken(dbConn *sql.DB, email, token, expira string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET email_confirm_token = ?, email_confirm_expira = ?, email_confirmado = 0, fecha_actualizacion = "+nowExpr+" WHERE LOWER(COALESCE(email,'')) = LOWER(?)", strings.TrimSpace(token), strings.TrimSpace(expira), strings.TrimSpace(email))
	return err
}

// ConfirmAdministradorByToken confirma el correo de un administrador usando su token.
func ConfirmAdministradorByToken(dbConn *sql.DB, token string) (int64, error) {
	row := dbConn.QueryRow(`SELECT id, COALESCE(email_confirm_expira, '') FROM administradores WHERE email_confirm_token = ? LIMIT 1`, strings.TrimSpace(token))
	var id int64
	var expiraRaw string
	if err := row.Scan(&id, &expiraRaw); err != nil {
		return 0, err
	}
	if expiraRaw != "" {
		if expiraAt, err := time.ParseInLocation("2006-01-02 15:04:05", expiraRaw, time.Local); err == nil {
			if time.Now().After(expiraAt) {
				return 0, fmt.Errorf("token de confirmacion expirado")
			}
		}
	}
	_, err := dbConn.Exec(`UPDATE administradores SET email_confirmado = 1, email_confirmado_en = datetime('now','localtime'), estado = 'activo', email_confirm_token = '', email_confirm_expira = '', fecha_actualizacion = datetime('now','localtime') WHERE id = ?`, id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// SetAdministradorPassword guarda hash y salt y marca password_set.
func SetAdministradorPassword(dbConn *sql.DB, email, hash, salt string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET password_hash = ?, password_salt = ?, password_set = 1, fecha_actualizacion = "+nowExpr+" WHERE LOWER(COALESCE(email,'')) = LOWER(?)", strings.TrimSpace(hash), strings.TrimSpace(salt), strings.TrimSpace(email))
	if err != nil && isMissingColumnError(err) {
		if schemaErr := EnsureAdministradoresAuthSchema(dbConn); schemaErr != nil {
			return schemaErr
		}
		_, err = execSQLCompat(dbConn, "UPDATE administradores SET password_hash = ?, password_salt = ?, password_set = 1, fecha_actualizacion = "+nowExpr+" WHERE LOWER(COALESCE(email,'')) = LOWER(?)", strings.TrimSpace(hash), strings.TrimSpace(salt), strings.TrimSpace(email))
	}
	return err
}

// SetAdministradorPasswordResetToken guarda token de recuperación para el administrador.
func SetAdministradorPasswordResetToken(dbConn *sql.DB, email, token, expira string) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET password_reset_token = ?, password_reset_expira = ?, fecha_actualizacion = "+nowExpr+" WHERE LOWER(COALESCE(email,'')) = LOWER(?)", strings.TrimSpace(token), strings.TrimSpace(expira), strings.TrimSpace(email))
	return err
}

// ClearAdministradorPasswordResetToken por id limpia el token de recuperación.
func ClearAdministradorPasswordResetToken(dbConn *sql.DB, id int64) error {
	nowExpr := sqlNowExpr()
	_, err := execSQLCompat(dbConn, "UPDATE administradores SET password_reset_token = '', password_reset_expira = '', fecha_actualizacion = "+nowExpr+" WHERE id = ?", id)
	return err
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
	AccessSource       string `json:"access_source,omitempty"`
	CompartidaPor      string `json:"compartida_por,omitempty"`
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

// GetEmpresasByUsuarioCreador obtiene las empresas creadas por un administrador.
func GetEmpresasByUsuarioCreador(dbConn *sql.DB, usuarioCreador string) ([]Empresa, error) {
	usuarioCreador = strings.TrimSpace(usuarioCreador)
	if usuarioCreador == "" {
		return []Empresa{}, nil
	}
	rows, err := querySQLCompat(dbConn, "SELECT id, COALESCE(empresa_id, id), nombre, nit, tipo_id, tipo_nombre, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresas WHERE LOWER(COALESCE(usuario_creador, '')) = LOWER(?) ORDER BY id DESC", usuarioCreador)
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

// GetEmpresaByScopeID resuelve una empresa por id fisico o por alcance logico empresa_id.
func GetEmpresaByScopeID(dbConn *sql.DB, empresaID int64) (*Empresa, error) {
	startedAt := time.Now()
	defer func() {
		PerfLogf("[perf][empresa] GetEmpresaByScopeID empresa=%d dur=%s", empresaID, time.Since(startedAt))
	}()
	if empresaID <= 0 {
		return nil, nil
	}
	empresaByScopeCacheMu.Lock()
	if cached, ok := empresaByScopeCache[empresaID]; ok && time.Since(cached.LoadedAt) < empresaByScopeCacheTTL {
		empresaByScopeCacheMu.Unlock()
		if cached.Empresa == nil {
			return nil, nil
		}
		copyEmpresa := *cached.Empresa
		return &copyEmpresa, nil
	}
	empresaByScopeCacheMu.Unlock()
	row := queryRowSQLCompat(dbConn, "SELECT id, COALESCE(empresa_id, id), nombre, nit, tipo_id, tipo_nombre, fecha_creacion, fecha_actualizacion, usuario_creador, estado, observaciones FROM empresas WHERE id = ? OR COALESCE(empresa_id, id) = ? ORDER BY CASE WHEN COALESCE(empresa_id, id) = ? THEN 0 ELSE 1 END, id ASC LIMIT 1", empresaID, empresaID, empresaID)
	var e Empresa
	var resolvedEmpresaID sql.NullInt64
	var nit sql.NullString
	var tipoID sql.NullInt64
	var tipoNombre sql.NullString
	var fechaCre sql.NullString
	var fechaAct sql.NullString
	var usuario sql.NullString
	var estado sql.NullString
	var obs sql.NullString
	if err := row.Scan(&e.ID, &resolvedEmpresaID, &e.Nombre, &nit, &tipoID, &tipoNombre, &fechaCre, &fechaAct, &usuario, &estado, &obs); err != nil {
		if err == sql.ErrNoRows {
			empresaByScopeCacheMu.Lock()
			empresaByScopeCache[empresaID] = cachedEmpresaByScope{Empresa: nil, LoadedAt: time.Now()}
			empresaByScopeCacheMu.Unlock()
			return nil, nil
		}
		return nil, err
	}
	if resolvedEmpresaID.Valid {
		e.EmpresaID = resolvedEmpresaID.Int64
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
	empresaCopy := e
	empresaByScopeCacheMu.Lock()
	empresaByScopeCache[empresaID] = cachedEmpresaByScope{Empresa: &empresaCopy, LoadedAt: time.Now()}
	empresaByScopeCacheMu.Unlock()
	return &empresaCopy, nil
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
	id, err := insertSQLCompat(dbConn, query, licenciaID, empresaID, transactionID, reference, status, rawPayload, discountCode, asesorID)
	if err == nil {
		return id, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return 0, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return 0, err
	}
	return insertSQLCompat(dbConn, query, licenciaID, empresaID, transactionID, reference, status, rawPayload, discountCode, asesorID)
}

// UpdateWompiPaymentRecordByTransaction actualiza una transacción de Wompi usando su transaction_id.
func UpdateWompiPaymentRecordByTransaction(dbConn *sql.DB, transactionID, status, rawPayload string) error {
	nowExpr := sqlNowExpr()
	query := "UPDATE pagos_wompi SET status = ?, raw_payload = ?, fecha_actualizacion = " + nowExpr + " WHERE transaction_id = ?"
	fallbackQuery := "UPDATE pagos_wompi SET status = ?, raw_payload = ? WHERE transaction_id = ?"
	_, err := execSQLCompat(dbConn, query, status, rawPayload, transactionID)
	if err == nil {
		return nil
	}
	if isMissingTableError(err) || isMissingColumnError(err) {
		if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr == nil {
			_, retryErr := execSQLCompat(dbConn, query, status, rawPayload, transactionID)
			if retryErr == nil {
				return nil
			}
			err = retryErr
		}
	}
	if isMissingColumnError(err) && strings.Contains(strings.ToLower(err.Error()), "fecha_actualizacion") {
		_, fallbackErr := execSQLCompat(dbConn, fallbackQuery, status, rawPayload, transactionID)
		if fallbackErr == nil {
			return nil
		}
		err = fallbackErr
	}
	return err
}

// UpdateWompiPaymentRecordByReference actualiza una transaccion de Wompi usando su referencia.
func UpdateWompiPaymentRecordByReference(dbConn *sql.DB, reference, status, rawPayload string) error {
	nowExpr := sqlNowExpr()
	query := "UPDATE pagos_wompi SET status = ?, raw_payload = ?, fecha_actualizacion = " + nowExpr + " WHERE reference = ?"
	fallbackQuery := "UPDATE pagos_wompi SET status = ?, raw_payload = ? WHERE reference = ?"
	_, err := execSQLCompat(dbConn, query, status, rawPayload, reference)
	if err == nil {
		return nil
	}
	if isMissingTableError(err) || isMissingColumnError(err) {
		if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr == nil {
			_, retryErr := execSQLCompat(dbConn, query, status, rawPayload, reference)
			if retryErr == nil {
				return nil
			}
			err = retryErr
		}
	}
	if isMissingColumnError(err) && strings.Contains(strings.ToLower(err.Error()), "fecha_actualizacion") {
		_, fallbackErr := execSQLCompat(dbConn, fallbackQuery, status, rawPayload, reference)
		if fallbackErr == nil {
			return nil
		}
		err = fallbackErr
	}
	return err
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
	read := func() (*WompiPaymentRecord, error) {
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
	rec, err := read()
	if err == nil {
		return rec, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return nil, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return nil, err
	}
	return read()
}

// GetWompiPaymentByReference obtiene una fila de pagos_wompi por reference
func GetWompiPaymentByReference(dbConn *sql.DB, reference string) (*WompiPaymentRecord, error) {
	read := func() (*WompiPaymentRecord, error) {
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
	rec, err := read()
	if err == nil {
		return rec, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return nil, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return nil, err
	}
	return read()
}

// CreateEpaycoPaymentRecord registra una transacción inicial de Epayco en la tabla pagos_epayco.
func CreateEpaycoPaymentRecord(dbConn *sql.DB, licenciaID, empresaID int64, transactionID, reference, status, rawPayload, discountCode, asesorID string) (int64, error) {
	nowExpr := sqlNowExpr()
	query := "INSERT INTO pagos_epayco (licencia_id, empresa_id, transaction_id, reference, status, raw_payload, discount_code, asesor_id, fecha_creacion) VALUES (?, ?, ?, ?, ?, ?, ?, ?, " + nowExpr + ")"
	id, err := insertSQLCompat(dbConn, query, licenciaID, empresaID, transactionID, reference, status, rawPayload, discountCode, asesorID)
	if err == nil {
		return id, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return 0, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return 0, err
	}
	return insertSQLCompat(dbConn, query, licenciaID, empresaID, transactionID, reference, status, rawPayload, discountCode, asesorID)
}

// UpdateEpaycoPaymentRecordByTransaction actualiza una transacción de Epayco usando su transaction_id.
func UpdateEpaycoPaymentRecordByTransaction(dbConn *sql.DB, transactionID, status, rawPayload string) error {
	nowExpr := sqlNowExpr()
	query := "UPDATE pagos_epayco SET status = ?, raw_payload = ?, fecha_actualizacion = " + nowExpr + " WHERE transaction_id = ?"
	fallbackQuery := "UPDATE pagos_epayco SET status = ?, raw_payload = ? WHERE transaction_id = ?"
	_, err := execSQLCompat(dbConn, query, status, rawPayload, transactionID)
	if err == nil {
		return nil
	}
	if isMissingTableError(err) || isMissingColumnError(err) {
		if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr == nil {
			_, retryErr := execSQLCompat(dbConn, query, status, rawPayload, transactionID)
			if retryErr == nil {
				return nil
			}
			err = retryErr
		}
	}
	if isMissingColumnError(err) && strings.Contains(strings.ToLower(err.Error()), "fecha_actualizacion") {
		_, fallbackErr := execSQLCompat(dbConn, fallbackQuery, status, rawPayload, transactionID)
		if fallbackErr == nil {
			return nil
		}
		err = fallbackErr
	}
	return err
}

// UpdateEpaycoPaymentRecordByReference actualiza una transaccion de Epayco usando su reference.
func UpdateEpaycoPaymentRecordByReference(dbConn *sql.DB, reference, status, rawPayload string) error {
	nowExpr := sqlNowExpr()
	query := "UPDATE pagos_epayco SET status = ?, raw_payload = ?, fecha_actualizacion = " + nowExpr + " WHERE reference = ?"
	fallbackQuery := "UPDATE pagos_epayco SET status = ?, raw_payload = ? WHERE reference = ?"
	_, err := execSQLCompat(dbConn, query, status, rawPayload, reference)
	if err == nil {
		return nil
	}
	if isMissingTableError(err) || isMissingColumnError(err) {
		if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr == nil {
			_, retryErr := execSQLCompat(dbConn, query, status, rawPayload, reference)
			if retryErr == nil {
				return nil
			}
			err = retryErr
		}
	}
	if isMissingColumnError(err) && strings.Contains(strings.ToLower(err.Error()), "fecha_actualizacion") {
		_, fallbackErr := execSQLCompat(dbConn, fallbackQuery, status, rawPayload, reference)
		if fallbackErr == nil {
			return nil
		}
		err = fallbackErr
	}
	return err
}

// EpaycoPaymentRecord representa una fila de pagos_epayco
type EpaycoPaymentRecord struct {
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

// GetEpaycoPaymentByTransaction obtiene una fila de pagos_epayco por transaction_id
func GetEpaycoPaymentByTransaction(dbConn *sql.DB, transactionID string) (*EpaycoPaymentRecord, error) {
	read := func() (*EpaycoPaymentRecord, error) {
		row := queryRowSQLCompat(dbConn, `SELECT id, licencia_id, empresa_id, transaction_id, reference, status, raw_payload, discount_code, asesor_id, fecha_creacion FROM pagos_epayco WHERE transaction_id = ? LIMIT 1`, transactionID)
		var r EpaycoPaymentRecord
		if err := row.Scan(&r.ID, &r.LicenciaID, &r.EmpresaID, &r.TransactionID, &r.Reference, &r.Status, &r.RawPayload, &r.DiscountCode, &r.AsesorID, &r.FechaCreacion); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &r, nil
	}
	rec, err := read()
	if err == nil {
		return rec, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return nil, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return nil, err
	}
	return read()
}

// GetEpaycoPaymentByReference obtiene una fila de pagos_epayco por reference
func GetEpaycoPaymentByReference(dbConn *sql.DB, reference string) (*EpaycoPaymentRecord, error) {
	read := func() (*EpaycoPaymentRecord, error) {
		row := queryRowSQLCompat(dbConn, `SELECT id, licencia_id, empresa_id, transaction_id, reference, status, raw_payload, discount_code, asesor_id, fecha_creacion FROM pagos_epayco WHERE reference = ? LIMIT 1`, reference)
		var r EpaycoPaymentRecord
		if err := row.Scan(&r.ID, &r.LicenciaID, &r.EmpresaID, &r.TransactionID, &r.Reference, &r.Status, &r.RawPayload, &r.DiscountCode, &r.AsesorID, &r.FechaCreacion); err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, err
		}
		return &r, nil
	}
	rec, err := read()
	if err == nil {
		return rec, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return nil, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return nil, err
	}
	return read()
}

// GetEpaycoPaymentContext devuelve licencia_id y empresa_id para una transaccion/referencia Epayco.
func GetEpaycoPaymentContext(dbConn *sql.DB, transactionID, reference string) (int64, int64, bool, error) {
	read := func() (int64, int64, bool, error) {
		row := queryRowSQLCompat(dbConn, `
			SELECT licencia_id, empresa_id
			FROM pagos_epayco
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
	licenciaID, empresaID, found, err := read()
	if err == nil {
		return licenciaID, empresaID, found, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return 0, 0, false, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return 0, 0, false, err
	}
	return read()
}

// GetWompiPaymentContext devuelve licencia_id y empresa_id para una transaccion/referencia Wompi.
func GetWompiPaymentContext(dbConn *sql.DB, transactionID, reference string) (int64, int64, bool, error) {
	read := func() (int64, int64, bool, error) {
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
	licenciaID, empresaID, found, err := read()
	if err == nil {
		return licenciaID, empresaID, found, nil
	}
	if !isMissingTableError(err) && !isMissingColumnError(err) {
		return 0, 0, false, err
	}
	if schemaErr := EnsurePaymentGatewaySchema(dbConn); schemaErr != nil {
		return 0, 0, false, err
	}
	return read()
}

func activateLicenciaForEmpresaTx(tx *sql.Tx, licenciaID, empresaID int64, fechaInicio, fechaFin string) (int64, error) {
	if licenciaID <= 0 || empresaID <= 0 {
		return 0, fmt.Errorf("licencia_id y empresa_id son obligatorios")
	}
	var currentEmpresa sql.NullInt64
	var tipoID sql.NullInt64
	var paisCodigo sql.NullString
	var nombre sql.NullString
	var descripcion sql.NullString
	var valor sql.NullFloat64
	var duracionDias sql.NullInt64
	var modulosHabilitados sql.NullString
	var superRol sql.NullInt64
	var usuarioCreador sql.NullString
	var observaciones sql.NullString
	if err := queryRowTxSQLCompat(tx, `SELECT
		empresa_id,
		tipo_id,
		COALESCE(pais_codigo, 'CO'),
		COALESCE(nombre, ''),
		COALESCE(descripcion, ''),
		COALESCE(valor, 0),
		COALESCE(duracion_dias, 0),
		COALESCE(modulos_habilitados, ''),
		COALESCE(super_rol_habilitado, 0),
		COALESCE(usuario_creador, ''),
		COALESCE(observaciones, '')
	FROM licencias
	WHERE id = ?
	LIMIT 1`, licenciaID).Scan(
		&currentEmpresa,
		&tipoID,
		&paisCodigo,
		&nombre,
		&descripcion,
		&valor,
		&duracionDias,
		&modulosHabilitados,
		&superRol,
		&usuarioCreador,
		&observaciones,
	); err != nil {
		return 0, err
	}

	nowExpr := sqlNowExpr()
	if currentEmpresa.Valid && currentEmpresa.Int64 == empresaID {
		if _, err := execTxSQLCompat(tx, "UPDATE licencias SET activo = 1, fecha_inicio = ?, fecha_fin = ?, fecha_actualizacion = "+nowExpr+" WHERE id = ?", fechaInicio, fechaFin, licenciaID); err != nil {
			return 0, err
		}
		return licenciaID, nil
	}

	activePlanQuery := `SELECT id
	FROM licencias
	WHERE empresa_id = ?
		AND COALESCE(activo, 1) = 1
		AND COALESCE(tipo_id, 0) = ?
		AND COALESCE(nombre, '') = ?
		AND ABS(COALESCE(valor, 0) - ?) < 0.0001
		AND COALESCE(duracion_dias, 0) = ?
		AND (COALESCE(fecha_inicio, '') = '' OR datetime(fecha_inicio) <= datetime('now','localtime'))
		AND (COALESCE(fecha_fin, '') = '' OR datetime(fecha_fin) >= datetime('now','localtime'))
	ORDER BY
		CASE WHEN COALESCE(fecha_fin, '') = '' THEN 1 ELSE 0 END DESC,
		datetime(COALESCE(fecha_fin, '9999-12-31 23:59:59')) DESC,
		id DESC
	LIMIT 1`
	if isPostgresDialect() {
		activePlanQuery = `SELECT id
		FROM licencias
		WHERE empresa_id = ?
			AND COALESCE(activo, 1) = 1
			AND COALESCE(tipo_id, 0) = ?
			AND COALESCE(nombre, '') = ?
			AND ABS(COALESCE(valor, 0) - ?) < 0.0001
			AND COALESCE(duracion_dias, 0) = ?
			AND (COALESCE(CAST(fecha_inicio AS TEXT), '') = '' OR CAST(fecha_inicio AS TIMESTAMP) <= CURRENT_TIMESTAMP)
			AND (COALESCE(CAST(fecha_fin AS TEXT), '') = '' OR CAST(fecha_fin AS TIMESTAMP) >= CURRENT_TIMESTAMP)
		ORDER BY
			CASE WHEN COALESCE(CAST(fecha_fin AS TEXT), '') = '' THEN 1 ELSE 0 END DESC,
			COALESCE(CAST(fecha_fin AS TIMESTAMP), TIMESTAMP '9999-12-31 23:59:59') DESC,
			id DESC
		LIMIT 1`
	}
	var existingActivePlanID int64
	err := queryRowTxSQLCompat(tx, activePlanQuery, empresaID, tipoID.Int64, nombre.String, valor.Float64, int(duracionDias.Int64)).Scan(&existingActivePlanID)
	if err == nil && existingActivePlanID > 0 {
		return existingActivePlanID, nil
	}
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	newID, err := insertTxSQLCompat(tx, `INSERT INTO licencias (
		empresa_id,
		tipo_id,
		pais_codigo,
		nombre,
		descripcion,
		valor,
		duracion_dias,
		modulos_habilitados,
		super_rol_habilitado,
		fecha_inicio,
		fecha_fin,
		activo,
		fecha_creacion,
		fecha_actualizacion,
		usuario_creador,
		estado,
		observaciones
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, `+nowExpr+`, `+nowExpr+`, ?, 'activo', ?)`,
		empresaID,
		tipoID.Int64,
		strings.TrimSpace(paisCodigo.String),
		nombre.String,
		descripcion.String,
		valor.Float64,
		int(duracionDias.Int64),
		modulosHabilitados.String,
		int(superRol.Int64),
		fechaInicio,
		fechaFin,
		usuarioCreador.String,
		observaciones.String,
	)
	if err != nil {
		return 0, err
	}
	return newID, nil
}

// ActivateLicenciaForEmpresa asigna y activa una licencia para una empresa, estableciendo fechas de inicio y fin.
// Si la licencia base ya pertenece a otra empresa, crea una copia activa para no mover ni desactivar la licencia anterior.
func ActivateLicenciaForEmpresa(dbConn *sql.DB, licenciaID, empresaID int64, fechaInicio, fechaFin string) error {
	if dbConn == nil {
		dbConn = GetDB()
	}
	if err := EnsureLicenciasSchema(dbConn); err != nil {
		return err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := activateLicenciaForEmpresaTx(tx, licenciaID, empresaID, fechaInicio, fechaFin); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	InvalidateLicenciaPermisoPolicyCacheForEmpresa(empresaID)
	return nil
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
