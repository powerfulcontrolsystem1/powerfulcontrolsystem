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
		action string
	}{
		{http.MethodGet, "consultar_acuse_real"},
		{http.MethodPost, "consultar_acuse_real"},
		{http.MethodPost, "subir_firma"},
		{http.MethodPut, "activar_produccion_local"},
		{http.MethodPost, "enviar_set_pruebas"},
	}
	for _, tc := range tests {
		req := httptest.NewRequest(tc.method, "/api/empresa/facturacion_electronica/dian?action="+tc.action, nil)
		if got := resolveFacturacionPermissionAction(req); got != permActionApprove {
			t.Fatalf("%s %s resolved permission=%q; want approve", tc.method, tc.action, got)
		}
	}
}

func TestDIANConfigWriteRequiresApprovalPermission(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?empresa_id=12", nil)
	if got := resolveFacturacionPermissionAction(req); got != permActionApprove {
		t.Fatalf("DIAN config save resolved permission=%q; want approve", got)
	}
}
