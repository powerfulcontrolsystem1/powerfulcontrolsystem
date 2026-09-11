package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaRolPermissionHandlerRequiresValidatedTenant(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/empresa/roles_de_usuario?empresa_id=1&action=permisos&rol_id=1", nil)
	r.Header.Set("X-Admin-Role", "admin_empresa")
	w := httptest.NewRecorder()
	EmpresaRolDeUsuarioPermisosHandler(nil)(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("unvalidated request status=%d", w.Code)
	}
}

func TestEmpresaRoleEditorPostgresRoundTripPreservesInheritedDenials(t *testing.T) {
	dsn := os.Getenv("PCS_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("PCS_TEST_POSTGRES_DSN is not configured")
	}
	admin, err := sql.Open(dbpkg.PostgresCompatDriverName(), dsn)
	if err != nil {
		t.Fatal("open isolated PostgreSQL failed")
	}
	schema := fmt.Sprintf("role_editor_%d", time.Now().UnixNano())
	if _, err := admin.Exec("CREATE SCHEMA " + schema); err != nil {
		admin.Close()
		t.Fatal(err)
	}
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
	t.Cleanup(func() {
		conn.Close()
		if _, err := admin.Exec("DROP SCHEMA " + schema + " CASCADE"); err != nil {
			t.Error(err)
		}
		admin.Close()
	})
	if _, err := conn.Exec(`CREATE TABLE tipos_de_empresas(id BIGINT PRIMARY KEY, nombre TEXT); INSERT INTO tipos_de_empresas VALUES(1,'QA')`); err != nil {
		t.Fatal(err)
	}
	for _, ensure := range []func(*sql.DB) error{dbpkg.EnsureRolesDeUsuarioSchema, dbpkg.EnsureRolesPermisosSchema} {
		if err := ensure(conn); err != nil {
			t.Fatal(err)
		}
	}
	base, err := dbpkg.CreateRolDeUsuario(conn, 1, "cajero", "", "qa")
	if err != nil {
		t.Fatal(err)
	}
	role, err := dbpkg.CreateEmpresaRolDeUsuario(conn, 71001, "Caja personalizada QA", "", base, "qa")
	if err != nil {
		t.Fatal(err)
	}
	_, before, beforePages, err := loadEmpresaRolePermissionMatrix(conn, 71001, role, "sin_rol")
	if err != nil {
		t.Fatal(err)
	}
	target := fmt.Sprintf("/api/empresa/roles_de_usuario?empresa_id=71001&action=permisos&rol_id=%d", role)
	req := requestWithTenantContext(httptest.NewRequest(http.MethodGet, target, nil), TenantContext{EmpresaID: 71001, AdminEmail: "qa@example.invalid"})
	w := httptest.NewRecorder()
	EmpresaRolDeUsuarioPermisosHandler(conn)(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("editor GET status=%d", w.Code)
	}
	var data struct {
		Modulos []permissionModuleMatrixRow `json:"modulos"`
		Paginas []permissionPageAccessRow   `json:"paginas"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatal(err)
	}
	beforeByKey := map[string]bool{}
	for _, row := range before {
		for _, action := range permissionActionsCatalogOrdered {
			beforeByKey[permissionModuleActionKey(row.Modulo, action)] = row.Acciones[action]
		}
	}
	payload := rolPermisosUpsertPayload{RolID: role, PermisosModulo: []rolPermisoModuloPayload{}, PermisosPagina: []rolPermisoPaginaPayload{}}
	for _, row := range data.Modulos {
		for _, action := range permissionActionsCatalogOrdered {
			key := permissionModuleActionKey(row.Modulo, action)
			if row.Acciones[action] != beforeByKey[key] {
				t.Fatalf("editor differs from authorization for %s", key)
			}
			allowed := row.Acciones[action]
			if key == "inventario|C" {
				allowed = true
			}
			payload.PermisosModulo = append(payload.PermisosModulo, rolPermisoModuloPayload{Modulo: row.Modulo, Accion: action, Permitido: allowed})
		}
	}
	for _, page := range data.Paginas {
		if page.Permitido != beforePages[page.PaginaClave] {
			t.Fatalf("editor differs from role page %s", page.PaginaClave)
		}
		payload.PermisosPagina = append(payload.PermisosPagina, rolPermisoPaginaPayload{PaginaClave: page.PaginaClave, Permitido: page.Permitido})
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	req = requestWithTenantContext(httptest.NewRequest(http.MethodPut, target, bytes.NewReader(raw)), TenantContext{EmpresaID: 71001, AdminEmail: "qa@example.invalid"})
	w = httptest.NewRecorder()
	EmpresaRolDeUsuarioPermisosHandler(conn)(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("editor PUT status=%d", w.Code)
	}
	_, after, _, err := loadEmpresaRolePermissionMatrix(conn, 71001, role, "sin_rol")
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range after {
		for _, action := range permissionActionsCatalogOrdered {
			key := permissionModuleActionKey(row.Modulo, action)
			want := beforeByKey[key]
			if key == "inventario|C" {
				want = true
			}
			if row.Acciones[action] != want {
				t.Fatalf("saving one checkbox changed unrelated capability %s", key)
			}
		}
	}
}

func TestEmpresaRolPermissionPayloadAcceptsWholeCatalogAndRejectsUnknown(t *testing.T) {
	payload := rolPermisosUpsertPayload{RolID: 7}
	for _, modulo := range permissionModulesCatalogOrdered {
		for _, action := range permissionActionsCatalogOrdered {
			payload.PermisosModulo = append(payload.PermisosModulo, rolPermisoModuloPayload{Modulo: modulo, Accion: action, Permitido: true})
		}
	}
	for _, rule := range permissionPagesCatalogOrdered {
		payload.PermisosPagina = append(payload.PermisosPagina, rolPermisoPaginaPayload{PaginaClave: rule.PaginaClave, Permitido: true})
	}
	modulos, paginas, err := validateEmpresaRolPermissionPayload(7, payload)
	if err != nil || len(modulos) != len(payload.PermisosModulo) || len(paginas) != len(payload.PermisosPagina) {
		t.Fatalf("catalog rejected: %v", err)
	}
	for _, cases := range []rolPermisosUpsertPayload{
		{PermisosModulo: []rolPermisoModuloPayload{{Modulo: "platform_root", Accion: "R"}}},
		{PermisosModulo: []rolPermisoModuloPayload{{Modulo: "ventas", Accion: "super"}}},
		{PermisosModulo: []rolPermisoModuloPayload{{Modulo: "ventas", Accion: "R"}, {Modulo: " VENTAS ", Accion: "r"}}},
		{PermisosPagina: []rolPermisoPaginaPayload{{PaginaClave: "superDashboard"}}},
		{PermisosPagina: []rolPermisoPaginaPayload{{PaginaClave: "linkEstaciones"}, {PaginaClave: " linkEstaciones "}}},
	} {
		if cases.PermisosModulo == nil {
			cases.PermisosModulo = []rolPermisoModuloPayload{}
		}
		if cases.PermisosPagina == nil {
			cases.PermisosPagina = []rolPermisoPaginaPayload{}
		}
		if _, _, err := validateEmpresaRolPermissionPayload(7, cases); err == nil {
			t.Fatal("invalid or duplicate capability accepted")
		}
	}
	if _, _, err := validateEmpresaRolPermissionPayload(7, rolPermisosUpsertPayload{}); err == nil {
		t.Fatal("missing lists would silently reset existing permissions")
	}
	if _, _, err := validateEmpresaRolPermissionPayload(7, rolPermisosUpsertPayload{PermisosModulo: []rolPermisoModuloPayload{}, PermisosPagina: []rolPermisoPaginaPayload{}}); err != nil {
		t.Fatalf("explicit inheritance reset rejected: %v", err)
	}
}

func TestEmpresaRolPermissionHandlerRejectsAmbiguousRoleID(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/api/empresa/roles_de_usuario?empresa_id=1&rol_id=1&rol_id=2", nil)
	r = requestWithTenantContext(r, TenantContext{EmpresaID: 1, AdminEmail: "synthetic@example.invalid"})
	w := httptest.NewRecorder()
	EmpresaRolDeUsuarioPermisosHandler(nil)(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ambiguous role status=%d", w.Code)
	}
}
