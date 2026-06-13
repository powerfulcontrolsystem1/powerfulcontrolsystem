package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPermissionChangeRequiresApprovalForSecurityEndpoints(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
		action string
		want   bool
	}{
		{
			name:   "crear usuario operativo no requiere codigo de aprobacion",
			method: http.MethodPost,
			path:   "/api/empresa/usuarios",
			action: permActionCreate,
			want:   false,
		},
		{
			name:   "actualizar usuario operativo no requiere codigo de aprobacion",
			method: http.MethodPut,
			path:   "/api/empresa/usuarios?id=10",
			action: permActionUpdate,
			want:   false,
		},
		{
			name:   "reenviar confirmacion de usuario no requiere codigo de aprobacion",
			method: http.MethodPut,
			path:   "/api/empresa/usuarios?action=reenviar_confirmacion&id=10",
			action: permActionApprove,
			want:   false,
		},
		{
			name:   "cambiar roles de usuario requiere codigo de aprobacion",
			method: http.MethodPost,
			path:   "/api/empresa/roles_de_usuario",
			action: permActionCreate,
			want:   true,
		},
		{
			name:   "cambiar permisos finos requiere codigo de aprobacion",
			method: http.MethodPut,
			path:   "/api/empresa/permisos_empresa",
			action: permActionUpdate,
			want:   true,
		},
		{
			name:   "leer permisos finos no requiere codigo de aprobacion",
			method: http.MethodGet,
			path:   "/api/empresa/permisos_empresa?empresa_id=44",
			action: permActionRead,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			got := permissionChangeRequiresApproval(permModuleSeguridad, req, tt.action)
			if got != tt.want {
				t.Fatalf("permissionChangeRequiresApproval() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultHiddenEnterpriseIAPagesRequireExplicitCompanyEnable(t *testing.T) {
	pages := map[string]bool{
		"linkReportes":              true,
		"linkCentroIAEmpresarial":   true,
		"linkRentaIA":               true,
		"linkSoportesComprasIA":     true,
		"linkSoportesComprasIAMenu": true,
	}

	hidden := applyDefaultHiddenEnterpriseIAPages(pages, map[string]bool{})
	for _, page := range []string{"linkCentroIAEmpresarial", "linkRentaIA", "linkSoportesComprasIA", "linkSoportesComprasIAMenu"} {
		if hidden[page] {
			t.Fatalf("%s debe quedar oculto por defecto", page)
		}
	}
	if !hidden["linkReportes"] {
		t.Fatal("una pagina no IA no debe cambiar por el default IA")
	}

	enabled := applyDefaultHiddenEnterpriseIAPages(pages, map[string]bool{"linkRentaIA": true})
	if !enabled["linkRentaIA"] {
		t.Fatal("linkRentaIA debe poder mostrarse cuando la empresa lo habilita explicitamente")
	}
	if enabled["linkCentroIAEmpresarial"] {
		t.Fatal("linkCentroIAEmpresarial debe seguir oculto si no tiene habilitacion explicita")
	}
}

func TestCajeroFinanzasManualAPIRequestSoloMovimientosPermitidos(t *testing.T) {
	if !isCajeroFinanzasManualAPIRequest("cajero", "/api/empresa/finanzas/movimientos", http.MethodPost, "") {
		t.Fatal("cajero debe poder llegar al handler de movimientos manuales para validar configuracion operativa")
	}
	if !isCajeroFinanzasManualAPIRequest("cajero", "/api/empresa/finanzas/movimientos", http.MethodPut, "anular") {
		t.Fatal("cajero debe poder llegar al handler de anulacion manual para validar configuracion operativa")
	}
	if isCajeroFinanzasManualAPIRequest("cajero", "/api/empresa/finanzas/movimientos", http.MethodPost, "importar_bancario") {
		t.Fatal("cajero no debe saltarse permisos para importar extractos bancarios")
	}
	if isCajeroFinanzasManualAPIRequest("cajero", "/api/empresa/finanzas/breb_qr", http.MethodPost, "") {
		t.Fatal("cajero no debe saltarse permisos de otras rutas de finanzas")
	}
	if isCajeroFinanzasManualAPIRequest("contador", "/api/empresa/finanzas/movimientos", http.MethodPost, "") {
		t.Fatal("la excepcion operativa aplica solo al rol cajero")
	}
}

func TestResolveAdminPermissionRoleForContextPreservaSuperAdminReservado(t *testing.T) {
	got := resolveAdminPermissionRoleForContext(nil, "powerfulcontrolsystem@gmail.com", "cajero")
	if got != "super_administrador" {
		t.Fatalf("rol contexto correo reservado = %q, want super_administrador", got)
	}
}

func TestResolveAdminPermissionRoleForSnapshotPreservaSuperAdminReservado(t *testing.T) {
	got := resolveAdminPermissionRoleForSnapshot("powerfulcontrolsystem@gmail.com", "cajero")
	if got != "super_administrador" {
		t.Fatalf("rol snapshot correo reservado = %q, want super_administrador", got)
	}
}
