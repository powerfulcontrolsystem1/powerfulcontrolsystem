package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnterpriseAIIsClosedByDefault(t *testing.T) {
	t.Setenv("AI_ENTERPRISE_ORCHESTRATOR_ENABLED", "")
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/ai/enterprise?empresa_id=1", nil)
	rr := httptest.NewRecorder()
	EmpresaAIEnterpriseHandler(nil)(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected closed feature flag, got %d", rr.Code)
	}
}
