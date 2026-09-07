package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTenantContextIsCanonicalEmpresaSource(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/recurso?empresa_id=999", nil)
	req.Header.Set("X-Empresa-ID", "888")
	req = requestWithTenantContext(req, TenantContext{
		EmpresaID:     12,
		AdminEmail:    "ADMIN@EXAMPLE.INVALID",
		AdminRole:     "administrador",
		EffectiveRole: "cajero",
		Module:        "ventas",
		Action:        "CREATE",
	})

	tenant, ok := TenantContextFromRequest(req)
	if !ok || tenant.EmpresaID != 12 {
		t.Fatalf("unexpected tenant context: ok=%v tenant=%+v", ok, tenant)
	}
	if tenant.AdminEmail != "admin@example.invalid" {
		t.Fatalf("tenant email was not normalized: %q", tenant.AdminEmail)
	}
	if got := parseEmpresaIDFromContext(req); got != 12 {
		t.Fatalf("legacy parser returned client supplied tenant %d, want 12", got)
	}
}

func TestTenantContextRejectsMissingValidatedEmpresa(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/recurso?empresa_id=99", nil)
	if _, ok := TenantContextFromRequest(req); ok {
		t.Fatal("unvalidated client empresa_id became a tenant context")
	}
}
