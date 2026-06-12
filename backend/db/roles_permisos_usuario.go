package db

import (
	"database/sql"
	"errors"
	"strings"
)

// RolPermisoModulo representa una regla de permiso por modulo/accion para un rol.
type RolPermisoModulo struct {
	RolID     int64  `json:"rol_id"`
	Modulo    string `json:"modulo"`
	Accion    string `json:"accion"`
	Permitido bool   `json:"permitido"`
}

// RolPermisoPagina representa una regla de permiso por pagina del panel empresa.
type RolPermisoPagina struct {
	RolID       int64  `json:"rol_id"`
	PaginaClave string `json:"pagina_clave"`
	Permitido   bool   `json:"permitido"`
}

// EnsureRolesPermisosSchema crea el esquema de permisos dinamicos de roles.
func EnsureRolesPermisosSchema(dbConn *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS roles_de_usuario_permisos (
			id BIGSERIAL PRIMARY KEY,
			rol_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			accion TEXT NOT NULL,
			permitido INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_de_usuario_permisos_unq
		ON roles_de_usuario_permisos(rol_id, modulo, accion);`,
		`CREATE INDEX IF NOT EXISTS idx_roles_de_usuario_permisos_lookup
		ON roles_de_usuario_permisos(rol_id, estado);`,
		`CREATE TABLE IF NOT EXISTS roles_de_usuario_paginas_permisos (
			id BIGSERIAL PRIMARY KEY,
			rol_id INTEGER NOT NULL,
			pagina_clave TEXT NOT NULL,
			permitido INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_roles_de_usuario_paginas_permisos_unq
		ON roles_de_usuario_paginas_permisos(rol_id, pagina_clave);`,
		`CREATE INDEX IF NOT EXISTS idx_roles_de_usuario_paginas_permisos_lookup
		ON roles_de_usuario_paginas_permisos(rol_id, estado);`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// GetRolDeUsuarioByID retorna un rol por id.
func GetRolDeUsuarioByID(dbConn *sql.DB, id int64) (*RolDeUsuario, error) {
	if err := EnsureRolesDeUsuarioSchema(dbConn); err != nil {
		return nil, err
	}
	const q = `SELECT
		r.id,
		COALESCE(r.empresa_id, 0),
		r.tipo_empresa_id,
		COALESCE(t.nombre, ''),
		COALESCE(r.nombre, ''),
		COALESCE(r.descripcion, ''),
		COALESCE(r.origen, 'global'),
		COALESCE(r.rol_base_id, 0),
		COALESCE(r.fecha_creacion, ''),
		COALESCE(r.fecha_actualizacion, ''),
		COALESCE(r.usuario_creador, ''),
		COALESCE(r.estado, 'activo'),
		COALESCE(r.observaciones, '')
	FROM roles_de_usuario r
	LEFT JOIN tipos_de_empresas t ON t.id = r.tipo_empresa_id
	WHERE r.id = ?
	LIMIT 1`

	item := &RolDeUsuario{}
	err := queryRowSQLCompat(dbConn, q, id).Scan(
		&item.ID,
		&item.EmpresaID,
		&item.TipoEmpresaID,
		&item.TipoEmpresaNombre,
		&item.Nombre,
		&item.Descripcion,
		&item.Origen,
		&item.RolBaseID,
		&item.FechaCreacion,
		&item.FechaActualizacion,
		&item.UsuarioCreador,
		&item.Estado,
		&item.Observaciones,
	)
	if err != nil {
		return nil, err
	}
	item.Personalizado = item.EmpresaID > 0 || strings.EqualFold(strings.TrimSpace(item.Origen), "empresa")
	return item, nil
}

// ResolveRolDeUsuarioIDByNombre resuelve el id mas reciente de un rol activo por nombre.
func ResolveRolDeUsuarioIDByNombre(dbConn *sql.DB, nombreRol string) (int64, error) {
	const q = `SELECT id
	FROM roles_de_usuario
	WHERE lower(trim(nombre)) = lower(trim(?))
		AND COALESCE(empresa_id, 0) = 0
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY id DESC
	LIMIT 1`

	var id int64
	if err := queryRowSQLCompat(dbConn, q, nombreRol).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// ListRolPermisosModuloByRolID lista permisos por modulo/accion para un rol.
func ListRolPermisosModuloByRolID(dbConn *sql.DB, rolID int64) ([]RolPermisoModulo, error) {
	const q = `SELECT
		rol_id,
		COALESCE(modulo, ''),
		COALESCE(accion, ''),
		COALESCE(permitido, 1)
	FROM roles_de_usuario_permisos
	WHERE rol_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY modulo ASC, accion ASC`

	rows, err := dbConn.Query(q, rolID)
	if err != nil {
		if isMissingTableError(err) {
			return []RolPermisoModulo{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]RolPermisoModulo, 0)
	for rows.Next() {
		var item RolPermisoModulo
		var permitidoInt int64
		if err := rows.Scan(&item.RolID, &item.Modulo, &item.Accion, &permitidoInt); err != nil {
			return nil, err
		}
		item.Modulo = strings.ToLower(strings.TrimSpace(item.Modulo))
		item.Accion = strings.ToUpper(strings.TrimSpace(item.Accion))
		item.Permitido = permitidoInt != 0
		if item.Modulo == "" || !isValidPermisoAccion(item.Accion) {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// ListRolPermisosPaginaByRolID lista permisos por pagina para un rol.
func ListRolPermisosPaginaByRolID(dbConn *sql.DB, rolID int64) ([]RolPermisoPagina, error) {
	const q = `SELECT
		rol_id,
		COALESCE(pagina_clave, ''),
		COALESCE(permitido, 1)
	FROM roles_de_usuario_paginas_permisos
	WHERE rol_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY pagina_clave ASC`

	rows, err := dbConn.Query(q, rolID)
	if err != nil {
		if isMissingTableError(err) {
			return []RolPermisoPagina{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]RolPermisoPagina, 0)
	for rows.Next() {
		var item RolPermisoPagina
		var permitidoInt int64
		if err := rows.Scan(&item.RolID, &item.PaginaClave, &permitidoInt); err != nil {
			return nil, err
		}
		item.PaginaClave = strings.TrimSpace(item.PaginaClave)
		item.Permitido = permitidoInt != 0
		if item.PaginaClave == "" {
			continue
		}
		out = append(out, item)
	}
	return out, nil
}

// ReplaceRolPermisosDeUsuario reemplaza en bloque los permisos por modulo y pagina de un rol.
func ReplaceRolPermisosDeUsuario(dbConn *sql.DB, rolID int64, permisosModulo []RolPermisoModulo, permisosPagina []RolPermisoPagina, usuarioCreador string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM roles_de_usuario_permisos WHERE rol_id = ?`, rolID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM roles_de_usuario_paginas_permisos WHERE rol_id = ?`, rolID); err != nil {
		return err
	}

	modulos := normalizeRolPermisosModulo(permisosModulo, rolID)
	for _, item := range modulos {
		permitido := int64(0)
		if item.Permitido {
			permitido = 1
		}
		if _, err = tx.Exec(`INSERT INTO roles_de_usuario_permisos (
			rol_id, modulo, accion, permitido, usuario_creador, estado, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			item.RolID, item.Modulo, item.Accion, permitido, usuarioCreador); err != nil {
			return err
		}
	}

	paginas := normalizeRolPermisosPagina(permisosPagina, rolID)
	for _, item := range paginas {
		permitido := int64(0)
		if item.Permitido {
			permitido = 1
		}
		if _, err = tx.Exec(`INSERT INTO roles_de_usuario_paginas_permisos (
			rol_id, pagina_clave, permitido, usuario_creador, estado, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			item.RolID, item.PaginaClave, permitido, usuarioCreador); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// LookupRolPermisoModuloByRoleName busca override de modulo/accion por nombre de rol.
func LookupRolPermisoModuloByRoleName(dbConn *sql.DB, nombreRol, modulo, accion string) (bool, bool, error) {
	rolID, err := ResolveRolDeUsuarioIDByNombre(dbConn, nombreRol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return LookupRolPermisoModuloByRolID(dbConn, rolID, modulo, accion)
}

// LookupRolPermisoModuloByRolID busca override de modulo/accion por id de rol.
func LookupRolPermisoModuloByRolID(dbConn *sql.DB, rolID int64, modulo, accion string) (bool, bool, error) {
	const q = `SELECT COALESCE(permitido, 1)
	FROM roles_de_usuario_permisos
	WHERE rol_id = ?
		AND lower(trim(modulo)) = lower(trim(?))
		AND upper(trim(accion)) = upper(trim(?))
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY id DESC
	LIMIT 1`

	var permitidoInt int64
	err := dbConn.QueryRow(q, rolID, modulo, accion).Scan(&permitidoInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || isMissingTableError(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, permitidoInt != 0, nil
}

// LookupRolPermisoPaginaByRoleName busca override de pagina por nombre de rol.
func LookupRolPermisoPaginaByRoleName(dbConn *sql.DB, nombreRol, paginaClave string) (bool, bool, error) {
	rolID, err := ResolveRolDeUsuarioIDByNombre(dbConn, nombreRol)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, false, nil
		}
		return false, false, err
	}
	return LookupRolPermisoPaginaByRolID(dbConn, rolID, paginaClave)
}

// LookupRolPermisoPaginaByRolID busca override de pagina por id de rol.
func LookupRolPermisoPaginaByRolID(dbConn *sql.DB, rolID int64, paginaClave string) (bool, bool, error) {
	const q = `SELECT COALESCE(permitido, 1)
	FROM roles_de_usuario_paginas_permisos
	WHERE rol_id = ?
		AND trim(pagina_clave) = trim(?)
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY id DESC
	LIMIT 1`

	var permitidoInt int64
	err := dbConn.QueryRow(q, rolID, paginaClave).Scan(&permitidoInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || isMissingTableError(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, permitidoInt != 0, nil
}

func normalizeRolPermisosModulo(input []RolPermisoModulo, rolID int64) []RolPermisoModulo {
	if len(input) == 0 {
		return []RolPermisoModulo{}
	}
	mapa := make(map[string]RolPermisoModulo, len(input))
	for _, raw := range input {
		modulo := strings.ToLower(strings.TrimSpace(raw.Modulo))
		accion := strings.ToUpper(strings.TrimSpace(raw.Accion))
		if modulo == "" || !isValidPermisoAccion(accion) {
			continue
		}
		key := modulo + "|" + accion
		mapa[key] = RolPermisoModulo{
			RolID:     rolID,
			Modulo:    modulo,
			Accion:    accion,
			Permitido: raw.Permitido,
		}
	}
	out := make([]RolPermisoModulo, 0, len(mapa))
	for _, item := range mapa {
		out = append(out, item)
	}
	return out
}

func normalizeRolPermisosPagina(input []RolPermisoPagina, rolID int64) []RolPermisoPagina {
	if len(input) == 0 {
		return []RolPermisoPagina{}
	}
	mapa := make(map[string]RolPermisoPagina, len(input))
	for _, raw := range input {
		pagina := strings.TrimSpace(raw.PaginaClave)
		if pagina == "" {
			continue
		}
		mapa[pagina] = RolPermisoPagina{
			RolID:       rolID,
			PaginaClave: pagina,
			Permitido:   raw.Permitido,
		}
	}
	out := make([]RolPermisoPagina, 0, len(mapa))
	for _, item := range mapa {
		out = append(out, item)
	}
	return out
}

func isValidPermisoAccion(accion string) bool {
	switch strings.ToUpper(strings.TrimSpace(accion)) {
	case "R", "C", "U", "D", "A":
		return true
	default:
		return false
	}
}

// Postgres-only: isMissingTableError() vive en sql_compat.go
