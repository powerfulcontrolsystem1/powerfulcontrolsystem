package handlers

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestEmpresaUsuarioSessionCannotSwitchTenant(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/empresa/productos?empresa_id=13", nil)
	ctx := context.WithValue(req.Context(), "sessionPrincipalType", "empresa_usuario")
	ctx = context.WithValue(ctx, "sessionEmpresaID", int64(12))
	req = req.WithContext(ctx)
	if err := validateEmpresaIDConsistency(req, 13); err == nil {
		t.Fatal("enterprise-user session must not switch to another empresa_id")
	}
	validReq := httptest.NewRequest("GET", "/api/empresa/productos?empresa_id=12", nil)
	validReq = validReq.WithContext(ctx)
	if err := validateEmpresaIDConsistency(validReq, 12); err != nil {
		t.Fatalf("session company should remain valid: %v", err)
	}
}
