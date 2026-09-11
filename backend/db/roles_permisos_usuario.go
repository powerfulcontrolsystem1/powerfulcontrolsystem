package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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
		if _, err := execSQLCompat(dbConn, stmt); err != nil {
			return err
		}
	}
	return nil
}

// GetRolDeUsuarioByID retorna un rol por id.
func GetRolDeUsuarioByID(dbConn *sql.DB, id int64) (*RolDeUsuario, error) {
	return getRolDeUsuarioByIDScoped(dbConn, id, 0, false)
}

func getRolDeUsuarioByIDScoped(dbConn *sql.DB, id, empresaID int64, onlyEmpresa bool) (*RolDeUsuario, error) {
	if dbConn == nil {
		return nil, errors.New("conexion de roles no disponible")
	}
	q := `SELECT
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
	WHERE r.id = ?`
	args := []interface{}{id}
	if empresaID > 0 {
		if onlyEmpresa {
			q += ` AND r.empresa_id = ?`
		} else {
			q += ` AND COALESCE(r.empresa_id, 0) IN (0, ?)`
		}
		args = append(args, empresaID)
	}
	q += ` LIMIT 1`

	item := &RolDeUsuario{}
	err := queryRowSQLCompat(dbConn, q, args...).Scan(
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
	item.Asignable = IsRolDeUsuarioAsignable(item)
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
	return out, rows.Err()
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
	return out, rows.Err()
}

// ListRolPermisosModuloByRolIDEmpresaScope lee únicamente roles globales o propios.
func ListRolPermisosModuloByRolIDEmpresaScope(dbConn *sql.DB, empresaID, rolID int64) ([]RolPermisoModulo, error) {
	return ListRolesPermisosModuloByRolIDEmpresaScope(dbConn, empresaID, []int64{rolID})
}

// ListRolPermisosPaginaByRolIDEmpresaScope aplica el mismo alcance a las páginas.
func ListRolPermisosPaginaByRolIDEmpresaScope(dbConn *sql.DB, empresaID, rolID int64) ([]RolPermisoPagina, error) {
	return ListRolesPermisosPaginaByRolIDEmpresaScope(dbConn, empresaID, []int64{rolID})
}

// ListRolesPermisosModuloByRolIDEmpresaScope carga la cadena de herencia en una
// consulta. LEFT JOIN distingue un rol sin reglas de un ID ajeno o inactivo.
// La salida respeta el orden de rolIDs para que el personalizado prevalezca.
func ListRolesPermisosModuloByRolIDEmpresaScope(dbConn *sql.DB, empresaID int64, rolIDs []int64) ([]RolPermisoModulo, error) {
	if empresaID <= 0 {
		return nil, sql.ErrNoRows
	}
	if dbConn == nil {
		return nil, errors.New("conexion de permisos no disponible")
	}
	return listRolesPermisosModuloScoped(context.Background(), dbConn, empresaID, rolIDs)
}

type rolePermissionQuerier interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

func listRolesPermisosModuloScoped(ctx context.Context, conn rolePermissionQuerier, empresaID int64, rolIDs []int64, incluirInactivos ...bool) ([]RolPermisoModulo, error) {
	clause, args, ids, err := rolesPermissionScope(empresaID, rolIDs, incluirInactivos...)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []RolPermisoModulo{}, nil
	}
	rows, err := conn.QueryContext(ctx, rebindCompatQuery(`SELECT r.id, COALESCE(p.modulo, ''), COALESCE(p.accion, ''), COALESCE(p.permitido, 1)
		FROM roles_de_usuario r
		LEFT JOIN roles_de_usuario_permisos p ON p.rol_id = r.id AND COALESCE(p.estado, 'activo') = 'activo'
		WHERE `+clause+` ORDER BY p.modulo, p.accion`), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byRole := map[int64][]RolPermisoModulo{}
	for rows.Next() {
		var item RolPermisoModulo
		var allowed int
		if err := rows.Scan(&item.RolID, &item.Modulo, &item.Accion, &allowed); err != nil {
			return nil, err
		}
		if _, seen := byRole[item.RolID]; !seen {
			byRole[item.RolID] = nil
		}
		item.Modulo, item.Accion = strings.ToLower(strings.TrimSpace(item.Modulo)), strings.ToUpper(strings.TrimSpace(item.Accion))
		item.Permitido = allowed != 0
		if item.Modulo != "" && isValidPermisoAccion(item.Accion) {
			byRole[item.RolID] = append(byRole[item.RolID], item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(byRole) != len(ids) {
		return nil, sql.ErrNoRows
	}
	out := []RolPermisoModulo{}
	for _, id := range ids {
		out = append(out, byRole[id]...)
	}
	return out, nil
}

// ListRolesPermisosPaginaByRolIDEmpresaScope aplica el mismo alcance y orden a páginas.
func ListRolesPermisosPaginaByRolIDEmpresaScope(dbConn *sql.DB, empresaID int64, rolIDs []int64) ([]RolPermisoPagina, error) {
	if empresaID <= 0 {
		return nil, sql.ErrNoRows
	}
	if dbConn == nil {
		return nil, errors.New("conexion de permisos no disponible")
	}
	return listRolesPermisosPaginaScoped(context.Background(), dbConn, empresaID, rolIDs)
}

func listRolesPermisosPaginaScoped(ctx context.Context, conn rolePermissionQuerier, empresaID int64, rolIDs []int64, incluirInactivos ...bool) ([]RolPermisoPagina, error) {
	clause, args, ids, err := rolesPermissionScope(empresaID, rolIDs, incluirInactivos...)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []RolPermisoPagina{}, nil
	}
	rows, err := conn.QueryContext(ctx, rebindCompatQuery(`SELECT r.id, COALESCE(p.pagina_clave, ''), COALESCE(p.permitido, 1)
		FROM roles_de_usuario r
		LEFT JOIN roles_de_usuario_paginas_permisos p ON p.rol_id = r.id AND COALESCE(p.estado, 'activo') = 'activo'
		WHERE `+clause+` ORDER BY p.pagina_clave`), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	byRole := map[int64][]RolPermisoPagina{}
	for rows.Next() {
		var item RolPermisoPagina
		var allowed int
		if err := rows.Scan(&item.RolID, &item.PaginaClave, &allowed); err != nil {
			return nil, err
		}
		if _, seen := byRole[item.RolID]; !seen {
			byRole[item.RolID] = nil
		}
		item.PaginaClave, item.Permitido = strings.TrimSpace(item.PaginaClave), allowed != 0
		if item.PaginaClave != "" {
			byRole[item.RolID] = append(byRole[item.RolID], item)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(byRole) != len(ids) {
		return nil, sql.ErrNoRows
	}
	out := []RolPermisoPagina{}
	for _, id := range ids {
		out = append(out, byRole[id]...)
	}
	return out, nil
}

func rolesPermissionScope(empresaID int64, input []int64, incluirInactivos ...bool) (string, []interface{}, []int64, error) {
	if empresaID < 0 {
		return "", nil, nil, sql.ErrNoRows
	}
	ids := []int64{}
	seen := map[int64]bool{}
	placeholders := []string{}
	args := []interface{}{}
	scope := `COALESCE(r.empresa_id, 0) = 0`
	if empresaID > 0 {
		scope = `COALESCE(r.empresa_id, 0) IN (0, ?)`
		args = append(args, empresaID)
	}
	for _, id := range input {
		if id <= 0 {
			return "", nil, nil, sql.ErrNoRows
		}
		if seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		args = append(args, id)
		placeholders = append(placeholders, "?")
	}
	if len(incluirInactivos) == 0 || !incluirInactivos[0] {
		scope += ` AND COALESCE(r.estado, 'activo') = 'activo'`
	}
	return scope + ` AND r.id IN (` + strings.Join(placeholders, ",") + `)`, args, ids, nil
}

// ReplaceEmpresaRolPermisosDeUsuario nunca modifica roles globales ni de otro tenant.
func ReplaceEmpresaRolPermisosDeUsuario(dbConn *sql.DB, empresaID, rolID int64, modulos []RolPermisoModulo, paginas []RolPermisoPagina, usuario string) error {
	if empresaID <= 0 {
		return sql.ErrNoRows
	}
	return replaceRolPermisosDeUsuario(dbConn, empresaID, rolID, modulos, paginas, usuario)
}

var ErrRolPermisosRevisionRequired = errors.New("revision de permisos requerida")
var ErrRolPermisosRevisionConflict = errors.New("la politica de permisos fue modificada; recargue antes de guardar")

// EmpresaRolPermisosEstado contiene una lectura coherente de la política propia y
// heredada. Las listas propias son excepciones; una lista vacía conserva herencia.
type EmpresaRolPermisosEstado struct {
	Rol         RolDeUsuario
	Base        RolDeUsuario
	ModulosBase []RolPermisoModulo
	PaginasBase []RolPermisoPagina
	Modulos     []RolPermisoModulo
	Paginas     []RolPermisoPagina
	Revision    string
}

// GetEmpresaRolPermisosEstado usa un único snapshot para que matriz y revisión
// nunca representen políticas leídas a ambos lados de una edición concurrente.
func GetEmpresaRolPermisosEstado(ctx context.Context, dbConn *sql.DB, empresaID, rolID int64) (*EmpresaRolPermisosEstado, error) {
	if dbConn == nil {
		return nil, errors.New("conexion de permisos no disponible")
	}
	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	state, err := loadEmpresaRolPermisosEstado(ctx, tx, empresaID, rolID, false)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return state, nil
}

func loadEmpresaRolPermisosEstado(ctx context.Context, tx *sql.Tx, empresaID, rolID int64, lock bool) (*EmpresaRolPermisosEstado, error) {
	if empresaID <= 0 || rolID <= 0 {
		return nil, sql.ErrNoRows
	}
	state := &EmpresaRolPermisosEstado{Modulos: []RolPermisoModulo{}, Paginas: []RolPermisoPagina{}, ModulosBase: []RolPermisoModulo{}, PaginasBase: []RolPermisoPagina{}}
	query := `SELECT r.id, r.empresa_id, r.tipo_empresa_id, COALESCE(r.nombre, ''),
		COALESCE(r.estado, 'activo'), r.rol_base_id,
		b.id, COALESCE(b.empresa_id, 0), b.tipo_empresa_id, COALESCE(b.nombre, ''), COALESCE(b.estado, 'activo')
		FROM roles_de_usuario r JOIN roles_de_usuario b ON b.id = r.rol_base_id
		WHERE r.id = ? AND r.empresa_id = ? AND COALESCE(b.empresa_id, 0) = 0`
	if lock {
		// Todos los escritores de matrices bloquean el rol. El bloqueo compartido
		// de la base impide que cambie entre comparar la revisión y guardar.
		query += ` FOR UPDATE OF r FOR SHARE OF b`
	}
	err := tx.QueryRowContext(ctx, rebindCompatQuery(query), rolID, empresaID).Scan(
		&state.Rol.ID, &state.Rol.EmpresaID, &state.Rol.TipoEmpresaID, &state.Rol.Nombre, &state.Rol.Estado, &state.Rol.RolBaseID,
		&state.Base.ID, &state.Base.EmpresaID, &state.Base.TipoEmpresaID, &state.Base.Nombre, &state.Base.Estado)
	if err != nil {
		return nil, err
	}
	if !IsRolDeUsuarioAsignable(&state.Rol) || !IsRolDeUsuarioAsignable(&state.Base) {
		return nil, sql.ErrNoRows
	}
	ids := []int64{state.Base.ID, state.Rol.ID}
	modules, err := listRolesPermisosModuloScoped(ctx, tx, empresaID, ids)
	if err != nil {
		return nil, err
	}
	pages, err := listRolesPermisosPaginaScoped(ctx, tx, empresaID, ids)
	if err != nil {
		return nil, err
	}
	for _, item := range modules {
		if item.RolID == state.Base.ID {
			state.ModulosBase = append(state.ModulosBase, item)
		} else {
			state.Modulos = append(state.Modulos, item)
		}
	}
	for _, item := range pages {
		if item.RolID == state.Base.ID {
			state.PaginasBase = append(state.PaginasBase, item)
		} else {
			state.Paginas = append(state.Paginas, item)
		}
	}
	// El orden SQL es determinista. La versión cubre cambios de semántica del
	// contrato y debe incrementarse si cambia la política heredada en código.
	state.Revision, err = rolPermisosRevision(state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

func rolPermisosRevision(state interface{}) (string, error) {
	raw, err := json.Marshal(struct {
		Version string
		State   interface{}
	}{Version: "empresa-roles-v1", State: state})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return hex.EncodeToString(digest[:]), nil
}

// RolPermisosEstado es el snapshot editable de una plantilla global. Los roles
// propios se administran mediante el contrato empresarial que incluye su base.
type RolPermisosEstado struct {
	Rol      RolDeUsuario
	Modulos  []RolPermisoModulo
	Paginas  []RolPermisoPagina
	Revision string
}

func GetRolPermisosEstado(ctx context.Context, dbConn *sql.DB, rolID int64) (*RolPermisosEstado, error) {
	if dbConn == nil {
		return nil, errors.New("conexion de permisos no disponible")
	}
	tx, err := dbConn.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	state, err := loadRolPermisosEstado(ctx, tx, rolID, false)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return state, nil
}

func loadRolPermisosEstado(ctx context.Context, tx *sql.Tx, rolID int64, lock bool) (*RolPermisosEstado, error) {
	query := `SELECT id, tipo_empresa_id, COALESCE(nombre, ''), COALESCE(estado, 'activo')
		FROM roles_de_usuario WHERE id = ? AND COALESCE(empresa_id, 0) = 0`
	if lock {
		query += ` AND COALESCE(estado, 'activo') = 'activo' FOR UPDATE`
	}
	state := &RolPermisosEstado{}
	if err := tx.QueryRowContext(ctx, rebindCompatQuery(query), rolID).Scan(&state.Rol.ID, &state.Rol.TipoEmpresaID, &state.Rol.Nombre, &state.Rol.Estado); err != nil {
		return nil, err
	}
	var err error
	state.Modulos, err = listRolesPermisosModuloScoped(ctx, tx, 0, []int64{rolID}, !lock)
	if err != nil {
		return nil, err
	}
	state.Paginas, err = listRolesPermisosPaginaScoped(ctx, tx, 0, []int64{rolID}, !lock)
	if err != nil {
		return nil, err
	}
	state.Revision, err = rolPermisosRevision(state)
	if err != nil {
		return nil, err
	}
	return state, nil
}

// ReplaceRolPermisosDeUsuarioConRevision serializa ediciones globales y rechaza
// snapshots antiguos antes de que puedan restaurar permisos ya revocados.
func ReplaceRolPermisosDeUsuarioConRevision(ctx context.Context, dbConn *sql.DB, rolID int64, revision string, modulos []RolPermisoModulo, paginas []RolPermisoPagina, usuario string) error {
	if strings.TrimSpace(revision) == "" {
		return ErrRolPermisosRevisionRequired
	}
	if dbConn == nil {
		return errors.New("conexion de permisos no disponible")
	}
	if err := validateRolPermisosInput(rolID, modulos, paginas); err != nil {
		return err
	}
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	state, err := loadRolPermisosEstado(ctx, tx, rolID, true)
	if err != nil {
		return err
	}
	if state.Revision != revision {
		return ErrRolPermisosRevisionConflict
	}
	if err := writeRolPermisosTx(tx, rolID, modulos, paginas, usuario); err != nil {
		return err
	}
	return tx.Commit()
}

// ReplaceEmpresaRolPermisosDeUsuarioConRevision evita sobrescribir decisiones
// nuevas desde un editor obsoleto y conserva sólo excepciones propias explícitas.
func ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx context.Context, dbConn *sql.DB, empresaID, rolID int64, revision string, modulos []RolPermisoModulo, paginas []RolPermisoPagina, usuario string) error {
	if strings.TrimSpace(revision) == "" {
		return ErrRolPermisosRevisionRequired
	}
	if dbConn == nil {
		return errors.New("conexion de permisos no disponible")
	}
	if err := validateRolPermisosInput(rolID, modulos, paginas); err != nil {
		return err
	}
	tx, err := dbConn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	state, err := loadEmpresaRolPermisosEstado(ctx, tx, empresaID, rolID, true)
	if err != nil {
		return err
	}
	if state.Revision != revision {
		return ErrRolPermisosRevisionConflict
	}
	if err := writeRolPermisosTx(tx, rolID, modulos, paginas, usuario); err != nil {
		return err
	}
	return tx.Commit()
}

// ReplaceRolPermisosDeUsuario reemplaza en bloque los permisos por modulo y pagina de un rol.
func ReplaceRolPermisosDeUsuario(dbConn *sql.DB, rolID int64, permisosModulo []RolPermisoModulo, permisosPagina []RolPermisoPagina, usuarioCreador string) error {
	return replaceRolPermisosDeUsuario(dbConn, 0, rolID, permisosModulo, permisosPagina, usuarioCreador)
}

func replaceRolPermisosDeUsuario(dbConn *sql.DB, empresaID, rolID int64, permisosModulo []RolPermisoModulo, permisosPagina []RolPermisoPagina, usuarioCreador string) error {
	if rolID <= 0 {
		return sql.ErrNoRows
	}
	if err := validateRolPermisosInput(rolID, permisosModulo, permisosPagina); err != nil {
		return err
	}
	tx, err := dbConn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var lockedID int64
	if err = tx.QueryRow(`SELECT id FROM roles_de_usuario WHERE id = ? AND (? = 0 OR empresa_id = ?) AND COALESCE(estado, 'activo') = 'activo' FOR UPDATE`, rolID, empresaID, empresaID).Scan(&lockedID); err != nil {
		return err
	}
	if err := writeRolPermisosTx(tx, rolID, permisosModulo, permisosPagina, usuarioCreador); err != nil {
		return err
	}
	return tx.Commit()
}

func writeRolPermisosTx(tx *sql.Tx, rolID int64, permisosModulo []RolPermisoModulo, permisosPagina []RolPermisoPagina, usuarioCreador string) error {
	if _, err := tx.Exec(`DELETE FROM roles_de_usuario_permisos WHERE rol_id = ?`, rolID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM roles_de_usuario_paginas_permisos WHERE rol_id = ?`, rolID); err != nil {
		return err
	}

	modulos := normalizeRolPermisosModulo(permisosModulo, rolID)
	for _, item := range modulos {
		permitido := int64(0)
		if item.Permitido {
			permitido = 1
		}
		if _, err := tx.Exec(`INSERT INTO roles_de_usuario_permisos (
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
		if _, err := tx.Exec(`INSERT INTO roles_de_usuario_paginas_permisos (
			rol_id, pagina_clave, permitido, usuario_creador, estado, fecha_creacion, fecha_actualizacion
		) VALUES (?, ?, ?, ?, 'activo', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
			item.RolID, item.PaginaClave, permitido, usuarioCreador); err != nil {
			return err
		}
	}

	return nil
}

func validateRolPermisosInput(rolID int64, modulos []RolPermisoModulo, paginas []RolPermisoPagina) error {
	seen := make(map[string]bool, len(modulos)+len(paginas))
	for _, item := range modulos {
		key := "m:" + strings.ToLower(strings.TrimSpace(item.Modulo)) + ":" + strings.ToUpper(strings.TrimSpace(item.Accion))
		if (item.RolID != 0 && item.RolID != rolID) || strings.TrimSpace(item.Modulo) == "" || !isValidPermisoAccion(item.Accion) || seen[key] {
			return errors.New("permiso de modulo invalido o repetido")
		}
		seen[key] = true
	}
	for _, item := range paginas {
		key := "p:" + strings.TrimSpace(item.PaginaClave)
		if (item.RolID != 0 && item.RolID != rolID) || strings.TrimSpace(item.PaginaClave) == "" || seen[key] {
			return errors.New("permiso de pagina invalido o repetido")
		}
		seen[key] = true
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
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
