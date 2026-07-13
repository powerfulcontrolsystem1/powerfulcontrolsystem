package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnterpriseAIIsClosedByDefault(t *testing.T) {
	t.Setenv("AI_ENTERPRISE_ORCHESTRATOR_ENABLED", "")
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/ai/enterprise?empresa_id=1", nil)
	rr := httptest.NewRecorder()
	EmpresaAIEnterpriseHandler(nil, nil)(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected closed feature flag, got %d", rr.Code)
	}
}

func TestDecodeEnterpriseJSONRejectsUnknownOrTrailingFields(t *testing.T) {
	for _, body := range []string{`{"known":"ok","unexpected":true}`, `{"known":"ok"}{}`} {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
		rr := httptest.NewRecorder()
		var dst struct {
			Known string `json:"known"`
		}
		if err := decodeEnterpriseJSON(rr, req, &dst, 1024); err == nil {
			t.Fatalf("expected strict JSON rejection for %q", body)
		}
	}
}

func TestEnterpriseAgentModeFailsClosed(t *testing.T) {
	t.Setenv("AI_ENTERPRISE_ORCHESTRATOR_ENABLED", "true")
	t.Setenv("AI_AGENT_MODE_ENABLED", "")
	if enterpriseAIAgentModeEnabled() {
		t.Fatal("agent mode unexpectedly enabled")
	}
}
