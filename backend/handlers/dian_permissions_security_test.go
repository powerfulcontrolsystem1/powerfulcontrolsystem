package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDIANEffectfulActionsAlwaysRequireApprovalPermission(t *testing.T) {
	t.Parallel()
	tests := []struct {
		method string
		path   string
		action string
	}{
		{http.MethodGet, "/api/empresa/facturacion_electronica/dian", "consultar_acuse_real"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "consultar_acuse_real"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "subir_firma"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "upload_firma"},
		{http.MethodPut, "/api/empresa/facturacion_electronica/dian", "activar_produccion_local"},
		{http.MethodPut, "/api/empresa/facturacion_electronica/dian", "pasar_a_produccion_local"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "enviar_set_pruebas"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "test_habilitacion"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "validar_credenciales"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "preflight_documento"},
		{http.MethodPost, "/api/empresa/facturacion_electronica/dian", "consultar_rango_numeracion"},
		{http.MethodPost, "/api/empresa/facturacion_electronica", "facturar_desde_venta"},
		{http.MethodPost, "/api/empresa/facturacion_electronica", "emitir_nomina_electronica"},
		{http.MethodPut, "/api/empresa/facturacion_electronica", "reenviar_dian"},
		{http.MethodPost, "/api/empresa/facturacion_electronica", "reconciliar_aceptados_local"},
		{http.MethodPost, "/api/empresa/facturacion_electronica", "configuracion_documentos_dian"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, tc.path+"?action="+tc.action, nil)
		if got := resolveFacturacionPermissionAction(req); got != permActionApprove {
			t.Fatalf("%s %s resolved permission=%q; want approve", tc.method, tc.action, got)
		}
	}
}

func TestNominaElectronicaRequiresPayrollApprovalInAdditionToFiscalApproval(t *testing.T) {
	t.Parallel()
	allowed := empresaPermissionSnapshot{
		CanAccess:      true,
		AllowedModules: map[string]bool{permModuleNominaSueldos: true},
		RoleModuleActions: map[string]bool{
			permissionModuleActionKey(permModuleNominaSueldos, permActionApprove): true,
		},
		AllowedPages: map[string]bool{"linkNominaSueldos": true},
	}
	if ok, reason := empresaPermissionSnapshotAllowsAdditionalModule(allowed, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos"); !ok {
		t.Fatalf("complete payroll authorization rejected: %s", reason)
	}
	withoutApproval := allowed
	withoutApproval.RoleModuleActions = map[string]bool{}
	if ok, _ := empresaPermissionSnapshotAllowsAdditionalModule(withoutApproval, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos"); ok {
		t.Fatal("electronic payroll emission accepted without payroll approval")
	}
	withoutLicense := allowed
	withoutLicense.AllowedModules = map[string]bool{permModuleFacturacion: true}
	if ok, _ := empresaPermissionSnapshotAllowsAdditionalModule(withoutLicense, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos"); ok {
		t.Fatal("electronic payroll emission accepted without payroll license")
	}
	withoutPage := allowed
	withoutPage.AllowedPages = map[string]bool{}
	if ok, _ := empresaPermissionSnapshotAllowsAdditionalModule(withoutPage, permModuleNominaSueldos, permActionApprove, "linkNominaSueldos"); ok {
		t.Fatal("electronic payroll emission accepted with payroll page disabled")
	}
}

func TestNominaEndpointResolvesPayrollPagePermission(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/nomina?empresa_id=12", nil)
	if got := resolvePermissionPageKeyForRequest(req); got != "linkNominaSueldos" {
		t.Fatalf("payroll endpoint page permission=%q; want linkNominaSueldos", got)
	}
}

func TestFacturacionNominaPrivacyClassifiesEntirePayrollFamily(t *testing.T) {
	t.Parallel()
	for _, documentType := range []string{
		"nomina_electronica",
		"nomina",
		"documento_soporte_pago_nomina_electronica",
		"nota_ajuste_nomina_electronica",
	} {
		if !facturacionDocumentoEsFamiliaNomina(documentType) {
			t.Fatalf("payroll document type %q was not classified as sensitive payroll data", documentType)
		}
	}
	for _, documentType := range []string{"factura_electronica", "nota_credito", "documento_soporte"} {
		if facturacionDocumentoEsFamiliaNomina(documentType) {
			t.Fatalf("non-payroll document type %q was classified as payroll", documentType)
		}
	}
}

func TestDIANExpiryReadDoesNotRequireApprovalWhenNotificationIsDisabled(t *testing.T) {
	t.Parallel()
	readOnly := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica/dian?action=vencimiento_certificado&notificar=0", nil)
	if got := resolveFacturacionPermissionAction(readOnly); got != permActionRead {
		t.Fatalf("expiry read permission=%q; want read", got)
	}
	notify := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica/dian?action=vencimiento_certificado&notificar=1", nil)
	if got := resolveFacturacionPermissionAction(notify); got != permActionApprove {
		t.Fatalf("expiry notification permission=%q; want approve", got)
	}
}

func TestDIANConfigWriteRequiresApprovalPermission(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?empresa_id=12", nil)
	if got := resolveFacturacionPermissionAction(req); got != permActionApprove {
		t.Fatalf("DIAN config save resolved permission=%q; want approve", got)
	}
}
