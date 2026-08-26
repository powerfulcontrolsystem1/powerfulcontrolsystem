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
