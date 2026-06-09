package db

import (
	"database/sql"
	"strings"
)

// EmpresaPermisoModulo representa una regla fina de techo por empresa para modulo/accion.
type EmpresaPermisoModulo struct {
	EmpresaID int64  `json:"empresa_id"`
	Modulo    string `json:"modulo"`
	Accion    string `json:"accion"`
	Permitido bool   `json:"permitido"`
}

// EmpresaPermisoPagina representa una regla fina de visibilidad por empresa para una pagina o funcion.
type EmpresaPermisoPagina struct {
	EmpresaID   int64  `json:"empresa_id"`
	PaginaClave string `json:"pagina_clave"`
	Permitido   bool   `json:"permitido"`
}

// EnsureEmpresaPermisosFinosSchema crea el esquema de restricciones finas por empresa.
func EnsureEmpresaPermisosFinosSchema(dbConn *sql.DB) error {
	if dbConn == nil {
		return nil
	}
	statements := []string{
		`CREATE TABLE IF NOT EXISTS empresa_permisos_modulos (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			modulo TEXT NOT NULL,
			accion TEXT NOT NULL,
			permitido INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empresa_permisos_modulos_unq
		ON empresa_permisos_modulos(empresa_id, modulo, accion);`,
		`CREATE INDEX IF NOT EXISTS idx_empresa_permisos_modulos_lookup
		ON empresa_permisos_modulos(empresa_id, estado);`,
		`CREATE TABLE IF NOT EXISTS empresa_permisos_paginas (
			id BIGSERIAL PRIMARY KEY,
			empresa_id INTEGER NOT NULL,
			pagina_clave TEXT NOT NULL,
			permitido INTEGER NOT NULL DEFAULT 1,
			fecha_creacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			fecha_actualizacion TEXT DEFAULT (CURRENT_TIMESTAMP),
			usuario_creador TEXT,
			estado TEXT DEFAULT 'activo',
			observaciones TEXT
		);`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_empresa_permisos_paginas_unq
		ON empresa_permisos_paginas(empresa_id, pagina_clave);`,
		`CREATE INDEX IF NOT EXISTS idx_empresa_permisos_paginas_lookup
		ON empresa_permisos_paginas(empresa_id, estado);`,
	}
	for _, stmt := range statements {
		if _, err := dbConn.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

// ListEmpresaPermisosModuloByEmpresaID lista reglas finas por modulo/accion para una empresa.
func ListEmpresaPermisosModuloByEmpresaID(dbConn *sql.DB, empresaID int64) ([]EmpresaPermisoModulo, error) {
	const q = `SELECT
		empresa_id,
		COALESCE(modulo, ''),
		COALESCE(accion, ''),
		COALESCE(permitido, 1)
	FROM empresa_permisos_modulos
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY modulo ASC, accion ASC`

	rows, err := dbConn.Query(q, empresaID)
	if err != nil {
		if isMissingTableError(err) {
			return []EmpresaPermisoModulo{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaPermisoModulo, 0)
	for rows.Next() {
		var item EmpresaPermisoModulo
		var permitidoInt int64
		if err := rows.Scan(&item.EmpresaID, &item.Modulo, &item.Accion, &permitidoInt); err != nil {
			return nil, err
		}
		item.Modulo = strings.ToLower(strings.TrimSpace(item.Modulo))
		item.Accion = strings.ToUpper(strings.TrimSpace(item.Accion))
		item.Permitido = permitidoInt != 0
		if item.EmpresaID <= 0 || item.Modulo == "" || !isValidPermisoAccion(item.Accion) {
			continue
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ListEmpresaPermisosPaginaByEmpresaID lista reglas finas por pagina para una empresa.
func ListEmpresaPermisosPaginaByEmpresaID(dbConn *sql.DB, empresaID int64) ([]EmpresaPermisoPagina, error) {
	const q = `SELECT
		empresa_id,
		COALESCE(pagina_clave, ''),
		COALESCE(permitido, 1)
	FROM empresa_permisos_paginas
	WHERE empresa_id = ?
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY pagina_clave ASC`

	rows, err := dbConn.Query(q, empresaID)
	if err != nil {
		if isMissingTableError(err) {
			return []EmpresaPermisoPagina{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]EmpresaPermisoPagina, 0)
	for rows.Next() {
		var item EmpresaPermisoPagina
		var permitidoInt int64
		if err := rows.Scan(&item.EmpresaID, &item.PaginaClave, &permitidoInt); err != nil {
			return nil, err
		}
		item.PaginaClave = strings.TrimSpace(item.PaginaClave)
		item.Permitido = permitidoInt != 0
		if item.EmpresaID <= 0 || item.PaginaClave == "" {
			continue
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ReplaceEmpresaPermisosFinos reemplaza las reglas finas de una empresa.
func ReplaceEmpresaPermisosFinos(dbConn *sql.DB, empresaID int64, permisosModulo []EmpresaPermisoModulo, permisosPagina []EmpresaPermisoPagina, usuarioCreador string) error {
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.Exec(`DELETE FROM empresa_permisos_modulos WHERE empresa_id = ?`, empresaID); err != nil {
		return err
	}
	if _, err = tx.Exec(`DELETE FROM empresa_permisos_paginas WHERE empresa_id = ?`, empresaID); err != nil {
		return err
	}

	modulos := normalizeEmpresaPermisosModulo(permisosModulo, empresaID)
	for _, item := range modulos {
		permitido := int64(0)
		if item.Permitido {
			permitido = 1
		}
		if _, err = tx.Exec(`INSERT INTO empresa_permisos_modulos (
			empresa_id, modulo, accion, permitido, usuario_creador, estado, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			item.EmpresaID, item.Modulo, item.Accion, permitido, usuarioCreador); err != nil {
			return err
		}
	}

	paginas := normalizeEmpresaPermisosPagina(permisosPagina, empresaID)
	for _, item := range paginas {
		permitido := int64(0)
		if item.Permitido {
			permitido = 1
		}
		if _, err = tx.Exec(`INSERT INTO empresa_permisos_paginas (
			empresa_id, pagina_clave, permitido, usuario_creador, estado, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			item.EmpresaID, item.PaginaClave, permitido, usuarioCreador); err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

// LookupEmpresaPermisoModulo busca una regla fina de modulo/accion por empresa.
func LookupEmpresaPermisoModulo(dbConn *sql.DB, empresaID int64, modulo, accion string) (bool, bool, error) {
	const q = `SELECT COALESCE(permitido, 1)
	FROM empresa_permisos_modulos
	WHERE empresa_id = ?
		AND lower(trim(modulo)) = lower(trim(?))
		AND upper(trim(accion)) = upper(trim(?))
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY id DESC
	LIMIT 1`

	var permitidoInt int64
	err := dbConn.QueryRow(q, empresaID, modulo, accion).Scan(&permitidoInt)
	if err != nil {
		if err == sql.ErrNoRows || isMissingTableError(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, permitidoInt != 0, nil
}

// LookupEmpresaPermisoPagina busca una regla fina de pagina por empresa.
func LookupEmpresaPermisoPagina(dbConn *sql.DB, empresaID int64, paginaClave string) (bool, bool, error) {
	const q = `SELECT COALESCE(permitido, 1)
	FROM empresa_permisos_paginas
	WHERE empresa_id = ?
		AND trim(pagina_clave) = trim(?)
		AND COALESCE(estado, 'activo') = 'activo'
	ORDER BY id DESC
	LIMIT 1`

	var permitidoInt int64
	err := dbConn.QueryRow(q, empresaID, paginaClave).Scan(&permitidoInt)
	if err != nil {
		if err == sql.ErrNoRows || isMissingTableError(err) {
			return false, false, nil
		}
		return false, false, err
	}
	return true, permitidoInt != 0, nil
}

func normalizeEmpresaPermisosModulo(input []EmpresaPermisoModulo, empresaID int64) []EmpresaPermisoModulo {
	if len(input) == 0 {
		return []EmpresaPermisoModulo{}
	}
	mapa := make(map[string]EmpresaPermisoModulo, len(input))
	for _, raw := range input {
		modulo := strings.ToLower(strings.TrimSpace(raw.Modulo))
		accion := strings.ToUpper(strings.TrimSpace(raw.Accion))
		if modulo == "" || !isValidPermisoAccion(accion) {
			continue
		}
		key := modulo + "|" + accion
		mapa[key] = EmpresaPermisoModulo{
			EmpresaID: empresaID,
			Modulo:    modulo,
			Accion:    accion,
			Permitido: raw.Permitido,
		}
	}
	out := make([]EmpresaPermisoModulo, 0, len(mapa))
	for _, item := range mapa {
		out = append(out, item)
	}
	return out
}

func normalizeEmpresaPermisosPagina(input []EmpresaPermisoPagina, empresaID int64) []EmpresaPermisoPagina {
	if len(input) == 0 {
		return []EmpresaPermisoPagina{}
	}
	mapa := make(map[string]EmpresaPermisoPagina, len(input))
	for _, raw := range input {
		pagina := strings.TrimSpace(raw.PaginaClave)
		if pagina == "" {
			continue
		}
		mapa[pagina] = EmpresaPermisoPagina{
			EmpresaID:   empresaID,
			PaginaClave: pagina,
			Permitido:   raw.Permitido,
		}
	}
	out := make([]EmpresaPermisoPagina, 0, len(mapa))
	for _, item := range mapa {
		out = append(out, item)
	}
	return out
}
