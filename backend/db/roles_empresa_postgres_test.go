package db

import (
	"context"
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

func TestEmpresaRolesPostgresBatchOverridesPreserveOrderAndTenant(t *testing.T) {
	conn := empresaRolesTestDB(t)
	base, err := CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	role, err := CreateEmpresaRolDeUsuario(conn, 71001, "Batch QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	other, err := CreateEmpresaRolDeUsuario(conn, 71002, "Batch QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	if err := ReplaceRolPermisosDeUsuario(conn, base, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: true}}, []RolPermisoPagina{{PaginaClave: "linkEstaciones", Permitido: true}}, "qa"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceEmpresaRolPermisosDeUsuario(conn, 71001, role, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: false}}, []RolPermisoPagina{{PaginaClave: "linkEstaciones", Permitido: false}}, "qa"); err != nil {
		t.Fatal(err)
	}
	for _, ids := range [][]int64{{base, role}, {role, base}, {base, base, role}} {
		m, err := ListRolesPermisosModuloByRolIDEmpresaScope(conn, 71001, ids)
		if err != nil {
			t.Fatal(err)
		}
		p, err := ListRolesPermisosPaginaByRolIDEmpresaScope(conn, 71001, ids)
		if err != nil {
			t.Fatal(err)
		}
		first, last := ids[0], ids[len(ids)-1]
		if len(m) != 2 || len(p) != 2 || m[0].RolID != first || p[0].RolID != first || m[1].RolID != last || p[1].RolID != last {
			t.Fatalf("batch did not preserve inheritance order: modules=%v pages=%v", m, p)
		}
		if m[1].Permitido != (last == base) || p[1].Permitido != (last == base) {
			t.Fatal("batch changed explicit deny/grant")
		}
	}
	if m, err := ListRolesPermisosModuloByRolIDEmpresaScope(conn, 71001, []int64{base, other}); !errors.Is(err, sql.ErrNoRows) || m != nil {
		t.Fatal("mixed valid/foreign chain exposed permissions")
	}
	if p, err := ListRolesPermisosPaginaByRolIDEmpresaScope(conn, 71001, []int64{base, other}); !errors.Is(err, sql.ErrNoRows) || p != nil {
		t.Fatal("mixed valid/foreign chain exposed pages")
	}
	m, err := ListRolesPermisosModuloByRolIDEmpresaScope(conn, 71002, []int64{other})
	if err != nil || len(m) != 0 {
		t.Fatal("valid role without overrides must remain an empty matrix")
	}
	if err := SetEmpresaRolDeUsuarioEstado(conn, 71001, role, "inactivo"); err != nil {
		t.Fatal(err)
	}
	if _, err := ListRolesPermisosModuloByRolIDEmpresaScope(conn, 71001, []int64{base, role}); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("inactive role in batch accepted")
	}
}

func TestEmpresaRolesPostgresCatalogUsesCompanyTypeMatrix(t *testing.T) {
	conn := empresaRolesTestDB(t)
	if _, err := conn.Exec(`INSERT INTO tipos_de_empresas VALUES(2,'Taller')`); err != nil {
		t.Fatal(err)
	}
	first, err := CreateRolDeUsuario(conn, 1, "Cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	local, err := CreateRolDeUsuario(conn, 2, "Caja", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	if err := ReplaceRolPermisosDeUsuario(conn, first, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: true}}, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceRolPermisosDeUsuario(conn, local, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: false}}, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	roles, err := GetRolesDeUsuarioCatalogoParaEmpresa(conn, 71001, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != 1 || roles[0].ID != local {
		t.Fatalf("catalog chose different vertical matrix: %+v", roles)
	}
	items, err := ListRolPermisosModuloByRolIDEmpresaScope(conn, 71001, roles[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Permitido {
		t.Fatal("catalog implicitly granted permissions from an older foreign role")
	}
	if got, err := GetRolDeUsuarioByIDEmpresaScope(conn, 71001, first); err != nil || got.ID != first {
		t.Fatal("historical assigned IDs must remain addressable")
	}
}

func TestEmpresaRolesPostgresRevisionPreservesRevocationsAndInheritance(t *testing.T) {
	conn := empresaRolesTestDB(t)
	ctx := context.Background()
	base, err := CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	role, err := CreateEmpresaRolDeUsuario(conn, 71001, "Revision QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	read := func() *EmpresaRolPermisosEstado {
		t.Helper()
		state, err := GetEmpresaRolPermisosEstado(ctx, conn, 71001, role)
		if err != nil {
			t.Fatal(err)
		}
		return state
	}
	initial := read()
	if len(initial.Revision) != 64 || initial.Revision != read().Revision {
		t.Fatal("unchanged policy must have a stable SHA256 revision")
	}
	if err := ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, "", nil, nil, "qa"); !errors.Is(err, ErrRolPermisosRevisionRequired) {
		t.Fatalf("missing revision err=%v", err)
	}
	grant := []RolPermisoModulo{{Modulo: "inventario", Accion: "C", Permitido: true}}
	if err := ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, initial.Revision, grant, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	editorA, editorB := read(), read()
	deny := []RolPermisoModulo{{Modulo: "inventario", Accion: "C", Permitido: false}}
	if err := ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, editorA.Revision, deny, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, editorB.Revision, grant, nil, "qa"); !errors.Is(err, ErrRolPermisosRevisionConflict) {
		t.Fatalf("stale editor restored revoked capability: %v", err)
	}
	revoked := read()
	if len(revoked.Modulos) != 1 || revoked.Modulos[0].Permitido {
		t.Fatal("obsolete replacement changed current denial")
	}
	if err := ReplaceRolPermisosDeUsuario(conn, base, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: false}}, []RolPermisoPagina{{PaginaClave: "linkEstaciones", Permitido: false}}, "qa"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, revoked.Revision, grant, nil, "qa"); !errors.Is(err, ErrRolPermisosRevisionConflict) {
		t.Fatalf("base policy change was not detected: %v", err)
	}
	current := read()
	if len(current.ModulosBase) != 1 || current.ModulosBase[0].Permitido || len(current.PaginasBase) != 1 || current.PaginasBase[0].Permitido {
		t.Fatal("inherited revocation missing from snapshot")
	}
	if err := ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, current.Revision, grant, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceRolPermisosDeUsuario(conn, base, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: true}}, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	updated := read()
	if len(updated.Modulos) != 1 || updated.Modulos[0].Modulo != "inventario" || len(updated.Paginas) != 0 || len(updated.ModulosBase) != 1 || !updated.ModulosBase[0].Permitido {
		t.Fatal("unrelated custom edit froze inherited base policy")
	}
	if _, err := GetEmpresaRolPermisosEstado(ctx, conn, 71002, role); !errors.Is(err, sql.ErrNoRows) {
		t.Fatal("foreign snapshot exposed tenant policy")
	}
}

func TestEmpresaRolesPostgresRevisionSerializesConcurrentEditors(t *testing.T) {
	conn := empresaRolesTestDB(t)
	ctx := context.Background()
	base, err := CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	role, err := CreateEmpresaRolDeUsuario(conn, 71001, "Two editors QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	state, err := GetEmpresaRolPermisosEstado(ctx, conn, 71001, role)
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, allowed := range []bool{true, false} {
		go func(allowed bool) {
			<-start
			results <- ReplaceEmpresaRolPermisosDeUsuarioConRevision(ctx, conn, 71001, role, state.Revision,
				[]RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: allowed}},
				[]RolPermisoPagina{{PaginaClave: "linkEstaciones", Permitido: allowed}}, "qa")
		}(allowed)
	}
	close(start)
	accepted, rejected := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			accepted++
		} else if errors.Is(err, ErrRolPermisosRevisionConflict) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if accepted != 1 || rejected != 1 {
		t.Fatalf("concurrent edits accepted=%d rejected=%d", accepted, rejected)
	}
	after, err := GetEmpresaRolPermisosEstado(ctx, conn, 71001, role)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Modulos) != 1 || len(after.Paginas) != 1 || after.Modulos[0].Permitido != after.Paginas[0].Permitido {
		t.Fatal("concurrent editors mixed module and page policies")
	}
}

func TestGlobalRolesPostgresPermissionReplacementIsAtomic(t *testing.T) {
	conn := empresaRolesTestDB(t)
	base, err := CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	if err := ReplaceRolPermisosDeUsuario(conn, base, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: false}}, nil, "qa"); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(`ALTER TABLE roles_de_usuario_paginas_permisos ADD CONSTRAINT roles_qa_reject CHECK (pagina_clave <> 'reject_qa')`); err != nil {
		t.Fatal(err)
	}
	if err := ReplaceRolPermisosDeUsuario(conn, base, []RolPermisoModulo{{Modulo: "ventas", Accion: "C", Permitido: true}}, []RolPermisoPagina{{PaginaClave: "reject_qa", Permitido: true}}, "qa"); err == nil {
		t.Fatal("expected page constraint failure")
	}
	modules, err := ListRolPermisosModuloByRolID(conn, base)
	if err != nil || len(modules) != 1 || modules[0].Permitido {
		t.Fatal("global module grant persisted despite failing page replacement")
	}
}
