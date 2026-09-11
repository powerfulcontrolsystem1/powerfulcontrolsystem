package db

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"
)

func empresaRolesTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	admin, err := sql.Open(PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal("open isolated PostgreSQL failed")
	}
	schema := fmt.Sprintf("roles_qa_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal("create isolated role schema failed")
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatal("invalid isolated PostgreSQL URL")
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	conn, err := sql.Open(PostgresCompatDriverName(), u.String())
	if err != nil {
		t.Fatal("open isolated schema failed")
	}
	conn.SetMaxOpenConns(12)
	t.Cleanup(func() {
		conn.Close()
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error("cleanup isolated schema failed")
		}
		admin.Close()
	})
	if _, err := conn.Exec(`CREATE TABLE tipos_de_empresas(id BIGINT PRIMARY KEY, nombre TEXT); INSERT INTO tipos_de_empresas VALUES (1, 'QA')`); err != nil {
		t.Fatal(err)
	}
	for _, ensure := range []func(*sql.DB) error{EnsureRolesDeUsuarioSchema, EnsureRolesPermisosSchema, EnsureEmpresaPermisosFinosSchema} {
		if err := ensure(conn); err != nil {
			t.Fatal(err)
		}
	}
	return conn
}

func TestEmpresaRolesPostgresIsolationAndAtomicPermissions(t *testing.T) {
	conn := empresaRolesTestDB(t)
	base, err := CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	roleA, err := CreateEmpresaRolDeUsuario(conn, 71001, "Coordinador QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	roleB, err := CreateEmpresaRolDeUsuario(conn, 71002, "Coordinador QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	modules := []RolPermisoModulo{{Modulo: "ventas", Accion: "U", Permitido: false}, {Modulo: "inventario", Accion: "C", Permitido: true}}
	pages := []RolPermisoPagina{{PaginaClave: "linkEstaciones", Permitido: false}}
	if err := ReplaceEmpresaRolPermisosDeUsuario(conn, 71001, roleA, modules, pages, "qa"); err != nil {
		t.Fatal(err)
	}
	for _, rolID := range []int64{base, roleB} {
		if err := ReplaceEmpresaRolPermisosDeUsuario(conn, 71001, rolID, nil, nil, "qa"); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("global/foreign role mutation err=%v", err)
		}
	}
	if _, err := ListRolPermisosModuloByRolIDEmpresaScope(conn, 71002, roleA); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("foreign role read err=%v", err)
	}
	for _, rolID := range []int64{base, roleB} {
		items, err := ListRolPermisosModuloByRolID(conn, rolID)
		if err != nil || len(items) != 0 {
			t.Fatalf("unrelated role changed: len=%d err=%v", len(items), err)
		}
	}
	if _, err := conn.Exec(`ALTER TABLE roles_de_usuario_paginas_permisos ADD CONSTRAINT roles_qa_reject CHECK (pagina_clave <> 'reject_qa')`); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceEmpresaRolPermisosDeUsuario(conn, 71001, roleA, nil, []RolPermisoPagina{{PaginaClave: "reject_qa"}}, "qa"); err == nil {
		t.Fatal("expected storage constraint failure")
	}
	items, err := ListRolPermisosModuloByRolIDEmpresaScope(conn, 71001, roleA)
	if err != nil || len(items) != 2 {
		t.Fatalf("failed replacement erased old permissions: %v", err)
	}
	if err := SetEmpresaRolDeUsuarioEstado(conn, 71001, roleA, "inactivo"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceEmpresaRolPermisosDeUsuario(conn, 71001, roleA, nil, nil, "qa"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("inactive role mutation err=%v", err)
	}
	if err := SetRolDeUsuarioEstado(conn, base, "inactivo"); err != nil {
		t.Fatal(err)
	}
	if _, err := CreateEmpresaRolDeUsuario(conn, 71001, "Inactive base", "", base, "qa"); err == nil {
		t.Fatal("inactive base accepted")
	}
}

func TestEmpresaRolesPostgresConcurrentCreateAndReplace(t *testing.T) {
	conn := empresaRolesTestDB(t)
	base, err := CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan int64, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, err := CreateEmpresaRolDeUsuario(conn, 71001, "Concurrent role", "", base, "qa")
			if err == nil {
				results <- id
			}
		}()
	}
	wg.Wait()
	close(results)
	var roleID int64
	count := 0
	for id := range results {
		roleID = id
		count++
	}
	if count != 1 {
		t.Fatalf("concurrent creation inserted %d roles", count)
	}
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(allowed bool) {
			defer wg.Done()
			errs <- ReplaceEmpresaRolPermisosDeUsuario(conn, 71001, roleID, []RolPermisoModulo{{Modulo: "ventas", Accion: "R", Permitido: allowed}}, []RolPermisoPagina{{PaginaClave: "linkEstaciones", Permitido: allowed}}, "qa")
		}(i%2 == 0)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	m, err := ListRolPermisosModuloByRolID(conn, roleID)
	if err != nil {
		t.Fatal(err)
	}
	p, err := ListRolPermisosPaginaByRolID(conn, roleID)
	if err != nil {
		t.Fatal(err)
	}
	if len(m) != 1 || len(p) != 1 || m[0].Permitido != p[0].Permitido {
		t.Fatal("concurrent replacement mixed permission snapshots")
	}
}

func TestEmpresaRolesPostgresMissingPermissionTablesFailClosed(t *testing.T) {
	conn := empresaRolesTestDB(t)
	if _, err := conn.Exec(`DROP TABLE roles_de_usuario_permisos, roles_de_usuario_paginas_permisos, empresa_permisos_modulos, empresa_permisos_paginas`); err != nil {
		t.Fatal(err)
	}
	if _, err := ListRolPermisosModuloByRolID(conn, 1); err == nil {
		t.Fatal("missing role modules silently ignored")
	}
	if _, err := ListRolPermisosPaginaByRolID(conn, 1); err == nil {
		t.Fatal("missing role pages silently ignored")
	}
	if _, err := ListEmpresaPermisosModuloByEmpresaID(conn, 1); err == nil {
		t.Fatal("missing enterprise modules silently ignored")
	}
	if _, err := ListEmpresaPermisosPaginaByEmpresaID(conn, 1); err == nil {
		t.Fatal("missing enterprise pages silently ignored")
	}
}
