package db

import (
	"database/sql"
	"testing"
)

func ensureRolesPermisosTestBaseSchema(t *testing.T, dbConn *sql.DB) int64 {
	t.Helper()

	_, err := dbConn.Exec(`CREATE TABLE IF NOT EXISTS tipos_de_empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT NOT NULL,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`)
	if err != nil {
		t.Fatalf("create tipos_de_empresas: %v", err)
	}

	_, err = dbConn.Exec(`CREATE TABLE IF NOT EXISTS roles_de_usuario (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		tipo_empresa_id INTEGER NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		fecha_creacion TEXT DEFAULT (datetime('now','localtime')),
		fecha_actualizacion TEXT DEFAULT (datetime('now','localtime')),
		usuario_creador TEXT,
		estado TEXT DEFAULT 'activo',
		observaciones TEXT
	)`)
	if err != nil {
		t.Fatalf("create roles_de_usuario: %v", err)
	}

	res, err := dbConn.Exec(`INSERT INTO tipos_de_empresas (nombre, estado) VALUES ('Tipo QA', 'activo')`)
	if err != nil {
		t.Fatalf("insert tipo empresa: %v", err)
	}
	tipoID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id tipo empresa: %v", err)
	}

	res, err = dbConn.Exec(`INSERT INTO roles_de_usuario (tipo_empresa_id, nombre, descripcion, estado, usuario_creador) VALUES (?, ?, ?, 'activo', 'qa')`, tipoID, "contabilidad", "rol contable")
	if err != nil {
		t.Fatalf("insert rol: %v", err)
	}
	rolID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("last insert id rol: %v", err)
	}
	return rolID
}

func TestRolesPermisosReplaceListAndLookup(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	rolID := ensureRolesPermisosTestBaseSchema(t, dbConn)

	if err := EnsureRolesPermisosSchema(dbConn); err != nil {
		t.Fatalf("EnsureRolesPermisosSchema: %v", err)
	}

	err := ReplaceRolPermisosDeUsuario(
		dbConn,
		rolID,
		[]RolPermisoModulo{
			{RolID: rolID, Modulo: "finanzas", Accion: "A", Permitido: false},
			{RolID: rolID, Modulo: "seguridad", Accion: "R", Permitido: true},
			{RolID: rolID, Modulo: "finanzas", Accion: "A", Permitido: true},
		},
		[]RolPermisoPagina{
			{RolID: rolID, PaginaClave: "linkFinanzas", Permitido: false},
			{RolID: rolID, PaginaClave: "linkReportes", Permitido: true},
		},
		"qa_user",
	)
	if err != nil {
		t.Fatalf("ReplaceRolPermisosDeUsuario: %v", err)
	}

	modulos, err := ListRolPermisosModuloByRolID(dbConn, rolID)
	if err != nil {
		t.Fatalf("ListRolPermisosModuloByRolID: %v", err)
	}
	if len(modulos) != 2 {
		t.Fatalf("expected 2 modulo permisos, got %d", len(modulos))
	}

	paginas, err := ListRolPermisosPaginaByRolID(dbConn, rolID)
	if err != nil {
		t.Fatalf("ListRolPermisosPaginaByRolID: %v", err)
	}
	if len(paginas) != 2 {
		t.Fatalf("expected 2 pagina permisos, got %d", len(paginas))
	}

	found, permitido, err := LookupRolPermisoModuloByRolID(dbConn, rolID, "finanzas", "A")
	if err != nil {
		t.Fatalf("LookupRolPermisoModuloByRolID: %v", err)
	}
	if !found || !permitido {
		t.Fatalf("expected finanzas/A override found=true permitido=true, got found=%v permitido=%v", found, permitido)
	}

	found, permitido, err = LookupRolPermisoModuloByRoleName(dbConn, "contabilidad", "seguridad", "R")
	if err != nil {
		t.Fatalf("LookupRolPermisoModuloByRoleName: %v", err)
	}
	if !found || !permitido {
		t.Fatalf("expected seguridad/R override found=true permitido=true, got found=%v permitido=%v", found, permitido)
	}

	found, permitido, err = LookupRolPermisoPaginaByRoleName(dbConn, "contabilidad", "linkFinanzas")
	if err != nil {
		t.Fatalf("LookupRolPermisoPaginaByRoleName: %v", err)
	}
	if !found || permitido {
		t.Fatalf("expected linkFinanzas override found=true permitido=false, got found=%v permitido=%v", found, permitido)
	}

	found, _, err = LookupRolPermisoPaginaByRolID(dbConn, rolID, "linkInexistente")
	if err != nil {
		t.Fatalf("LookupRolPermisoPaginaByRolID no row: %v", err)
	}
	if found {
		t.Fatalf("expected found=false for missing page override")
	}
}

func TestRolesPermisosFallbackWithoutPermissionTables(t *testing.T) {
	dbConn := openCarritoInventarioTestDB(t)
	rolID := ensureRolesPermisosTestBaseSchema(t, dbConn)

	modulos, err := ListRolPermisosModuloByRolID(dbConn, rolID)
	if err != nil {
		t.Fatalf("ListRolPermisosModuloByRolID fallback: %v", err)
	}
	if len(modulos) != 0 {
		t.Fatalf("expected empty modulo permisos without schema, got %d", len(modulos))
	}

	paginas, err := ListRolPermisosPaginaByRolID(dbConn, rolID)
	if err != nil {
		t.Fatalf("ListRolPermisosPaginaByRolID fallback: %v", err)
	}
	if len(paginas) != 0 {
		t.Fatalf("expected empty pagina permisos without schema, got %d", len(paginas))
	}

	found, permitido, err := LookupRolPermisoModuloByRolID(dbConn, rolID, "finanzas", "A")
	if err != nil {
		t.Fatalf("LookupRolPermisoModuloByRolID fallback: %v", err)
	}
	if found || permitido {
		t.Fatalf("expected found=false permitido=false without schema, got found=%v permitido=%v", found, permitido)
	}

	found, permitido, err = LookupRolPermisoPaginaByRolID(dbConn, rolID, "linkFinanzas")
	if err != nil {
		t.Fatalf("LookupRolPermisoPaginaByRolID fallback: %v", err)
	}
	if found || permitido {
		t.Fatalf("expected page found=false permitido=false without schema, got found=%v permitido=%v", found, permitido)
	}
}
