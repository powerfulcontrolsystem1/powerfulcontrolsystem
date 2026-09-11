package handlers

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	aipkg "github.com/you/pos-backend/ai"
)

// The wrapper has already authorized the company user. A global administrator
// record or legacy role associated with this same email must not replace it.
func enterpriseAIRestrictedRequest(snapshot empresaPermissionSnapshot) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/api/empresa/ai/enterprise?empresa_id=12", nil)
	ctx := context.WithValue(r.Context(), "sessionPrincipalType", "empresa_usuario")
	ctx = context.WithValue(ctx, "sessionPrincipalID", int64(88))
	ctx = context.WithValue(ctx, "sessionEmpresaID", int64(12))
	ctx = context.WithValue(ctx, "adminRole", "super_administrador")
	ctx = context.WithValue(ctx, empresaPermissionSnapshotContextKey{}, snapshot)
	r = r.WithContext(ctx)
	return requestWithTenantContext(r, TenantContext{EmpresaID: 12, AdminEmail: "same-account@example.invalid", AdminRole: "super_administrador", EffectiveRole: snapshot.EffectiveRole})
}

func enterpriseAIRestrictedSnapshot() empresaPermissionSnapshot {
	return empresaPermissionSnapshot{
		EmpresaID: 12, AdminEmail: "same-account@example.invalid", RoleID: 71,
		AdminRole: "cajero", EffectiveRole: "cajero", CanAccess: true,
		AllowedModules: map[string]bool{permModuleInventario: true, permModuleVentas: true, permModuleSeguridad: true},
		RoleModuleActions: map[string]bool{
			permissionModuleActionKey("inventario", "R"): true,
			permissionModuleActionKey("inventario", "C"): false,
			permissionModuleActionKey("ventas", "R"):     true,
			permissionModuleActionKey("ventas", "C"):     false,
			permissionModuleActionKey("ventas", "U"):     false,
			permissionModuleActionKey("finanzas", "R"):   true,
			permissionModuleActionKey("seguridad", "R"):  false,
		},
		AllowedPages: map[string]bool{},
	}
}

func TestEnterpriseAIExecutionContextPreservesCompanyRequestIdentity(t *testing.T) {
	snapshot := enterpriseAIRestrictedSnapshot()
	r := enterpriseAIRestrictedRequest(snapshot)
	ctx, err := enterpriseAIExecutionContext(r, nil, nil, 12, snapshot.AdminEmail)
	if err != nil {
		t.Fatalf("authorized request snapshot should not require an email-based lookup: %v", err)
	}
	if ctx.Role != "cajero" || ctx.EmpresaID != 12 || ctx.UserID != snapshot.AdminEmail {
		t.Fatalf("execution identity escaped the wrapper: role=%q empresa=%d", ctx.Role, ctx.EmpresaID)
	}
	if !reflect.DeepEqual(ctx.Permissions, []string{"inventario:R", "ventas:R"}) {
		t.Fatalf("custom denies or license restrictions were lost: %v", ctx.Permissions)
	}
	if !enterpriseAIRequireTool(ctx, aipkg.ToolCatalogSearchProducts) {
		t.Fatal("authorized inventory read should remain available")
	}
	for _, tool := range []string{aipkg.ToolCatalogCreateProduct, aipkg.ToolSalesAddStationProduct, aipkg.ToolHotelConfigureRoomStation} {
		if enterpriseAIRequireTool(ctx, tool) {
			t.Fatalf("company role unexpectedly gained write tool %q", tool)
		}
	}
	if enterpriseAITariffAdminRole(ctx.Role) {
		t.Fatal("raw super role must not enable administrative tariff tools")
	}
}

func TestEnterpriseAIRequestSnapshotRejectsTenantAndIdentityMismatch(t *testing.T) {
	snapshot := enterpriseAIRestrictedSnapshot()
	r := enterpriseAIRestrictedRequest(snapshot)
	for _, tc := range []struct {
		name      string
		request   *http.Request
		empresaID int64
		user      string
	}{
		{"other_company", r, 99, snapshot.AdminEmail},
		{"other_account", r, 12, "other@example.invalid"},
		{"missing_wrapper", httptest.NewRequest(http.MethodGet, "/", nil), 12, snapshot.AdminEmail},
		{"missing_request", nil, 12, snapshot.AdminEmail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := empresaAIRequestPermissionSnapshot(tc.request, nil, nil, tc.user, tc.empresaID); !errors.Is(err, sql.ErrNoRows) {
				t.Fatalf("unvalidated identity must fail before a database lookup: %v", err)
			}
		})
	}
}

func TestEnterpriseAIChatReadersKeepRestrictedRoleForSharedAdminEmail(t *testing.T) {
	snapshot := enterpriseAIRestrictedSnapshot()
	snapshot.RoleModuleActions[permissionModuleActionKey("inventario", "R")] = false
	r := enterpriseAIRestrictedRequest(snapshot)
	controller := NewEmpresaAIChatController(nil, nil)
	if _, handled := controller.authorizedDirectDocumentResponse(r, 12, snapshot.AdminEmail, "Genera un reporte de productos"); handled {
		t.Fatal("document query must not recover global-admin inventory access by email")
	}
	response, handled, err := controller.authorizedAdminDBDirectResponse(r, 12, snapshot.AdminEmail, "Cuántos usuarios hay registrados")
	if err != nil || !handled || !strings.Contains(response, "No puedo mostrar") {
		t.Fatalf("administrative user counts must be denied without querying users: handled=%v err=%v", handled, err)
	}
	allowed, role, err := controller.authorizedAdministrativeReadRole(r, 12, snapshot.AdminEmail)
	if err != nil || allowed || role != "cajero" {
		t.Fatalf("company history must use the restricted request: allowed=%v role=%q err=%v", allowed, role, err)
	}
	if _, err := normalizeEmpresaAIHistoryScope("empresa", allowed); err == nil {
		t.Fatal("restricted actor must not access company-wide chat history")
	}
	contextText, err := controller.roleScopedChatContext(r, 12, "Resume mi empresa", snapshot.AdminEmail, "", "")
	if err != nil || !strings.Contains(contextText, "herramientas autorizadas") {
		t.Fatalf("restricted context should retain guidance without aggregate database access: %v", err)
	}
}

func TestEnterpriseAIAdministrativeReadRequiresRoleActionAndLicense(t *testing.T) {
	snapshot := enterpriseAIRestrictedSnapshot()
	snapshot.EffectiveRole = "admin_empresa"
	if empresaAISnapshotAllowsAdministrativeRead(snapshot) {
		t.Fatal("administrative role name cannot override an explicit security read denial")
	}
	snapshot.RoleModuleActions[permissionModuleActionKey("seguridad", "R")] = true
	if !empresaAISnapshotAllowsAdministrativeRead(snapshot) {
		t.Fatal("administrative role with company-authorized security read should be allowed")
	}
	snapshot.AllowedModules = map[string]bool{permModuleVentas: true}
	if empresaAISnapshotAllowsAdministrativeRead(snapshot) {
		t.Fatal("administrative reads must respect the license")
	}
	snapshot.AllowedModules = nil
	snapshot.CanAccess = false
	if empresaAISnapshotAllowsAdministrativeRead(snapshot) || len(enterpriseAIPermissionsFromSnapshot(snapshot)) != 0 {
		t.Fatal("a rejected company context cannot expose data or tool permissions")
	}
}

func TestEnterpriseAIConfirmationUsesPermissionsOfNewRequest(t *testing.T) {
	first := enterpriseAIRestrictedSnapshot()
	first.RoleModuleActions[permissionModuleActionKey("inventario", "C")] = true
	before, err := enterpriseAIExecutionContext(enterpriseAIRestrictedRequest(first), nil, nil, 12, first.AdminEmail)
	if err != nil || !enterpriseAIRequireTool(before, aipkg.ToolCatalogCreateProduct) {
		t.Fatalf("initial request should allow the product proposal: %v", err)
	}
	second := enterpriseAIRestrictedSnapshot()
	after, err := enterpriseAIExecutionContext(enterpriseAIRestrictedRequest(second), nil, nil, 12, second.AdminEmail)
	if err != nil {
		t.Fatal(err)
	}
	if enterpriseAIRequireTool(after, aipkg.ToolCatalogCreateProduct) {
		t.Fatal("a new confirmation request must not retain a permission revoked after proposing")
	}
}
