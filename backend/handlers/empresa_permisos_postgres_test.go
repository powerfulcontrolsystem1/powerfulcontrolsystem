package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaPermissionRolesPostgresIsolationAndRevocation(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	admin, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal(err)
	}
	schema := fmt.Sprintf("permissions_engine_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		defer admin.Close()
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Errorf("clean fixture: %v", err)
		}
	})
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	conn, err := sql.Open(dbpkg.PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	conn.SetMaxOpenConns(4)
	exec := func(query string, args ...interface{}) {
		t.Helper()
		if _, err := conn.Exec(query, args...); err != nil {
			t.Fatal(err)
		}
	}
	for _, statement := range []string{
		`CREATE TABLE tipos_de_empresas (id BIGINT PRIMARY KEY, nombre TEXT)`,
		`CREATE TABLE roles_de_usuario (id BIGINT PRIMARY KEY, empresa_id BIGINT DEFAULT 0, tipo_empresa_id BIGINT DEFAULT 0, nombre TEXT, descripcion TEXT, origen TEXT DEFAULT 'global', rol_base_id BIGINT DEFAULT 0, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`CREATE TABLE roles_de_usuario_permisos (rol_id BIGINT, modulo TEXT, accion TEXT, permitido INT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE roles_de_usuario_paginas_permisos (rol_id BIGINT, pagina_clave TEXT, permitido INT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE empresa_permisos_modulos (empresa_id BIGINT, modulo TEXT, accion TEXT, permitido INT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE empresa_permisos_paginas (empresa_id BIGINT, pagina_clave TEXT, permitido INT, estado TEXT DEFAULT 'activo')`,
		`CREATE TABLE empresas (id BIGINT PRIMARY KEY, empresa_id BIGINT, nombre TEXT DEFAULT 'QA', nit TEXT, tipo_id BIGINT, tipo_nombre TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`CREATE TABLE licencias (id BIGINT PRIMARY KEY, empresa_id BIGINT, nombre TEXT, modulos_habilitados TEXT, super_rol_habilitado INT, activo INT, fecha_inicio TEXT, fecha_fin TEXT)`,
		`CREATE TABLE empresa_licencias_adicionales (id BIGINT PRIMARY KEY, empresa_id BIGINT, licencia_id BIGINT, activo INT, fecha_inicio TEXT, fecha_fin TEXT)`,
		`CREATE TABLE users (id BIGINT PRIMARY KEY, empresa_id BIGINT, email TEXT, name TEXT, documento_identidad TEXT, rol_usuario_id BIGINT, role TEXT, foto_url TEXT, control_aseo_estaciones INT, email_confirmado INT, email_confirm_token TEXT, email_confirm_expira TEXT, email_confirmado_en TEXT, acepta_contrato INT, contrato_version_aceptada INT, fecha_acepta_contrato TEXT, fecha_creacion TEXT, fecha_actualizacion TEXT, usuario_creador TEXT, estado TEXT DEFAULT 'activo', observaciones TEXT)`,
		`INSERT INTO roles_de_usuario (id,nombre) VALUES (1,'vendedor'),(2,'super_administrador'),(3,'mesero'),(4,'rol_futuro')`,
		`INSERT INTO roles_de_usuario (id,nombre,empresa_id,rol_base_id,origen) VALUES (11,'Ventas ampliadas',101,1,'empresa'),(12,'Ventas ampliadas',202,1,'empresa'),(13,'Base inválida',101,2,'empresa')`,
		`INSERT INTO roles_de_usuario_permisos (rol_id,modulo,accion,permitido) VALUES (11,'inventario','R',1),(11,'inventario','C',1),(11,'ventas','C',0)`,
		`INSERT INTO roles_de_usuario_permisos (rol_id,modulo,accion,permitido) VALUES (3,'ventas','R',1),(3,'ventas','C',1),(3,'ventas','U',1),(3,'clientes','R',1),(3,'clientes','C',1),(3,'inventario','R',1),(3,'facturacion','R',1)`,
		`INSERT INTO roles_de_usuario_paginas_permisos (rol_id,pagina_clave,permitido) VALUES (11,'linkEstaciones',0)`,
		`INSERT INTO empresas (id,empresa_id) VALUES (101,101),(202,202)`,
		`INSERT INTO licencias (id,empresa_id,nombre,modulos_habilitados,super_rol_habilitado,activo) VALUES (101,101,'QA','ventas,inventario,clientes,seguridad',1,1),(202,202,'QA','ventas,inventario,clientes,seguridad',1,1)`,
		`INSERT INTO users (id,empresa_id,email,rol_usuario_id,email_confirmado) VALUES (501,101,'shared@example.invalid',11,1),(502,202,'shared@example.invalid',12,1)`,
	} {
		exec(statement)
	}

	role, rowsA, pagesA, err := loadEmpresaRolePermissionMatrix(conn, 101, 11, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	if role != "vendedor" {
		t.Fatalf("unexpected effective base: %s", role)
	}
	if !findPermissionModuleRowForTest(t, rowsA, permModuleInventario).Create {
		t.Fatal("custom company A grant was ignored")
	}
	if findPermissionModuleRowForTest(t, rowsA, permModuleVentas).Create || pagesA["linkEstaciones"] {
		t.Fatal("custom denial was replaced by the base role")
	}
	_, rowsB, _, err := loadEmpresaRolePermissionMatrix(conn, 202, 12, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	if findPermissionModuleRowForTest(t, rowsB, permModuleInventario).Create {
		t.Fatal("same-name role in company B inherited company A grant")
	}
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 202, 11, "admin_empresa"); err == nil {
		t.Fatal("foreign role accepted")
	}
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 13, "admin_empresa"); err == nil {
		t.Fatal("custom role inherited super administrator")
	}
	_, meseroRows, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 3, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	meseroVentas := findPermissionModuleRowForTest(t, meseroRows, permModuleVentas)
	if !meseroVentas.Read || !meseroVentas.Create || !meseroVentas.Update || meseroVentas.Delete || meseroVentas.Approve {
		t.Fatal("catalogued mesero must receive only its persisted grants")
	}
	_, futureRows, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 4, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range futureRows {
		for _, allowed := range row.Acciones {
			if allowed {
				t.Fatal("new role without grants must default deny")
			}
		}
	}

	// Administrative sessions have no assigned tenant role ID. Their template
	// must still respect the company's type instead of a same-name foreign row.
	exec(`INSERT INTO tipos_de_empresas (id,nombre) VALUES (10,'QA A'),(20,'QA B')`)
	exec(`UPDATE empresas SET tipo_id=CASE WHEN empresa_id=101 THEN 10 ELSE 20 END`)
	exec(`INSERT INTO roles_de_usuario (id,nombre,tipo_empresa_id) VALUES (30,'admin_empresa',0),(31,'admin_empresa',10),(32,'admin_empresa',20)`)
	exec(`INSERT INTO roles_de_usuario_permisos (rol_id,modulo,accion,permitido) VALUES (30,'ventas','C',0),(30,'inventario','C',0),(31,'ventas','C',0),(32,'inventario','C',0)`)
	_, adminA, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 0, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	_, adminB, _, err := loadEmpresaRolePermissionMatrix(conn, 202, 0, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	if findPermissionModuleRowForTest(t, adminA, permModuleVentas).Create || !findPermissionModuleRowForTest(t, adminA, permModuleInventario).Create || !findPermissionModuleRowForTest(t, adminB, permModuleVentas).Create || findPermissionModuleRowForTest(t, adminB, permModuleInventario).Create {
		t.Fatal("administrative template mixed company types or discarded a denial")
	}
	exec(`INSERT INTO roles_de_usuario (id,nombre,tipo_empresa_id) VALUES (33,'admin_empresa',10)`)
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 0, "admin_empresa"); err == nil {
		t.Fatal("ambiguous administrative templates must not select an arbitrary ID")
	}
	exec(`UPDATE roles_de_usuario SET estado='inactivo' WHERE id IN (31,33)`)
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 0, "admin_empresa"); err == nil {
		t.Fatal("inactive typed template must not restore unrestricted defaults or universal grants")
	}
	exec(`UPDATE empresas SET tipo_id=0`)
	_, universalAdmin, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 0, "admin_empresa")
	if err != nil || findPermissionModuleRowForTest(t, universalAdmin, permModuleVentas).Create || findPermissionModuleRowForTest(t, universalAdmin, permModuleInventario).Create {
		t.Fatalf("universal template restrictions were lost: %v", err)
	}
	exec(`INSERT INTO roles_de_usuario (id,nombre,tipo_empresa_id) VALUES (34,'supervisor_sucursal',20)`)
	exec(`INSERT INTO roles_de_usuario_permisos (rol_id,modulo,accion,permitido) VALUES (34,'ventas','R',0)`)
	_, defaultSupervisor, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 0, "supervisor_sucursal")
	if err != nil || !findPermissionModuleRowForTest(t, defaultSupervisor, permModuleVentas).Read {
		t.Fatalf("missing template should preserve known standard policy: %v", err)
	}
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 999, 0, "admin_empresa"); err == nil {
		t.Fatal("missing company must not be treated as a missing optional template")
	}

	exec(`UPDATE roles_de_usuario_permisos SET permitido=0 WHERE rol_id=11 AND modulo='inventario' AND accion='C'`)
	_, revoked, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 11, "admin_empresa")
	if err != nil {
		t.Fatal(err)
	}
	if findPermissionModuleRowForTest(t, revoked, permModuleInventario).Create {
		t.Fatal("next request reused a revoked grant")
	}
	exec(`UPDATE roles_de_usuario SET estado='inactivo' WHERE id=11`)
	if _, _, _, err := loadEmpresaRolePermissionMatrix(conn, 101, 11, "admin_empresa"); err == nil {
		t.Fatal("inactive role fell back to admin")
	}
	exec(`UPDATE roles_de_usuario SET estado='activo' WHERE id=11`)

	newRequest := func(company, principal int64) *http.Request {
		r := httptest.NewRequest(http.MethodGet, "/api/empresa/permisos_contexto", nil)
		ctx := context.WithValue(r.Context(), "sessionPrincipalType", "empresa_usuario")
		ctx = context.WithValue(ctx, "sessionPrincipalID", principal)
		ctx = context.WithValue(ctx, "sessionEmpresaID", company)
		return r.WithContext(ctx)
	}
	if _, err := getEmpresaPermissionSnapshotForRequest(newRequest(101, 502), conn, conn, "shared@example.invalid", 101); err == nil {
		t.Fatal("shared email replaced the tenant principal ID")
	}
	snapshotA, err := getEmpresaPermissionSnapshotForRequest(newRequest(101, 501), conn, conn, "shared@example.invalid", 101)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshotA.CanAccess || snapshotA.RoleID != 11 || snapshotA.RoleModuleActions["ventas|C"] || snapshotA.AllowedPages["linkEstaciones"] {
		t.Fatal("typed tenant user did not receive its exact custom policy")
	}
	snapshotB, err := getEmpresaPermissionSnapshotForRequest(newRequest(202, 502), conn, conn, "shared@example.invalid", 202)
	if err != nil {
		t.Fatal(err)
	}
	if !snapshotB.CanAccess || snapshotB.RoleID != 12 || !snapshotB.RoleModuleActions["ventas|C"] {
		t.Fatal("same corporate email leaked company A restrictions into B")
	}
	exec(`UPDATE roles_de_usuario_permisos SET permitido=1 WHERE rol_id=11 AND modulo='ventas' AND accion='C'`)
	updated, err := getEmpresaPermissionSnapshotForRequest(newRequest(101, 501), conn, conn, "shared@example.invalid", 101)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.RoleModuleActions["ventas|C"] {
		t.Fatal("saved grant not observed on next request")
	}
	exec(`UPDATE roles_de_usuario_permisos SET permitido=0 WHERE rol_id=11 AND modulo='ventas' AND accion='C'`)
	updated, err = getEmpresaPermissionSnapshotForRequest(newRequest(101, 501), conn, conn, "shared@example.invalid", 101)
	if err != nil {
		t.Fatal(err)
	}
	if updated.RoleModuleActions["ventas|C"] {
		t.Fatal("revoked grant survived into next request")
	}
	exec(`UPDATE licencias SET activo=0 WHERE empresa_id=101`)
	withoutLicense, err := getEmpresaPermissionSnapshotForRequest(newRequest(101, 501), conn, conn, "shared@example.invalid", 101)
	if err != nil {
		t.Fatal(err)
	}
	if withoutLicense.RoleModuleActions["ventas|R"] || withoutLicense.RoleModuleActions["inventario|R"] {
		t.Fatal("missing active license granted unrestricted access")
	}
	exec(`UPDATE licencias SET activo=1 WHERE empresa_id=101`)
	exec(`UPDATE users SET estado='inactivo' WHERE empresa_id=101 AND id=501`)
	if _, err := getEmpresaPermissionSnapshotForRequest(newRequest(101, 501), conn, conn, "shared@example.invalid", 101); err == nil {
		t.Fatal("inactive principal retained access")
	}
	exec(`UPDATE users SET estado='activo' WHERE empresa_id=101 AND id=501`)
	exec(`UPDATE empresas SET estado='inactivo' WHERE empresa_id=101`)
	if _, err := getEmpresaPermissionSnapshotForRequest(newRequest(101, 501), conn, conn, "shared@example.invalid", 101); err == nil {
		t.Fatal("inactive company retained access")
	}

	exec(`INSERT INTO empresa_permisos_modulos (empresa_id,modulo,accion,permitido) VALUES (101,'inventario','C',0)`)
	overridesA, _, _, err := loadEmpresaPermissionOverridesStrict(conn, 101)
	if err != nil {
		t.Fatal(err)
	}
	overridesB, _, _, err := loadEmpresaPermissionOverridesStrict(conn, 202)
	if err != nil {
		t.Fatal(err)
	}
	if allowed, exists := overridesA["inventario|C"]; !exists || allowed {
		t.Fatal("company A denial missing")
	}
	if _, exists := overridesB["inventario|C"]; exists {
		t.Fatal("company A ceiling contaminated company B")
	}
	exec(`DROP TABLE empresa_permisos_paginas`)
	if _, _, _, err := loadEmpresaPermissionOverridesStrict(conn, 101); err == nil {
		t.Fatal("missing permissions table failed open")
	}
}
