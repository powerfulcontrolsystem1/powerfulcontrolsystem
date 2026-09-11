package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
