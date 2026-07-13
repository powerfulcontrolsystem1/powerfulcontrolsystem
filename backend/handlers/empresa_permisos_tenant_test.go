package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateEmpresaIDConsistencyRejectsQueryManipulation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/recurso?empresa_id=22", nil)
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("cross-tenant empresa_id in query was accepted")
	}
}

func TestValidateEmpresaIDConsistencyRejectsHeaderManipulation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/recurso?empresa_id=11", nil)
	req.Header.Set("X-Empresa-ID", "22")
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("cross-tenant empresa_id in header was accepted")
	}
}

func TestValidateEmpresaIDConsistencyRejectsJSONManipulation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/recurso?empresa_id=11", strings.NewReader(`{"empresa_id":22,"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("cross-tenant empresa_id in JSON was accepted")
	}
}

func TestValidateEmpresaIDConsistencyAcceptsMatchingSources(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/recurso?empresa_id=11", strings.NewReader(`{"empresa_id":11,"id":1}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Empresa-ID", "11")
	if err := validateEmpresaIDConsistency(req, 11); err != nil {
		t.Fatalf("matching validated tenant sources rejected: %v", err)
	}
}
