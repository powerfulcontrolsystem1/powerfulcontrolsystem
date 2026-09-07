package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnterpriseAIProposalValidationMessageRedactsUnexpectedCause(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/ia/proposal", nil)
	req.Header.Set("X-Request-ID", "req-ai-105")
	message := enterpriseAIProposalValidationMessage(req, errors.New("postgres://user:secret@internal/proposals"))

	if strings.Contains(message, "postgres://") || strings.Contains(message, "secret") {
		t.Fatalf("validation message exposes internal cause: %q", message)
	}
	if message != "Los datos de la propuesta no son validos." {
		t.Fatalf("unexpected safe validation message: %q", message)
	}
}

func TestEnterpriseAIProposalValidationMessagePreservesKnownMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/ia/proposal", nil)
	message := enterpriseAIProposalValidationMessage(req, errors.New("nombre de producto invalido"))
	if message != "nombre de producto invalido" {
		t.Fatalf("known validation message changed: %q", message)
	}
}
