package handlers

import (
	"bytes"
	"mime/multipart"
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

func TestValidateEmpresaIDConsistencyRejectsRepeatedQueryManipulation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/recurso?empresa_id=11&empresa_id=22", nil)
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("repeated query empresa_id was accepted")
	}
}

func TestValidateEmpresaIDConsistencyRejectsRepeatedHeaderManipulation(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/recurso", nil)
	req.Header.Add("X-Empresa-ID", "11")
	req.Header.Add("X-Empresa-ID", "22")
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("repeated header empresa_id was accepted")
	}
}

func TestValidateEmpresaIDConsistencyRejectsRepeatedFormManipulation(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/recurso", strings.NewReader("empresa_id=11&empresa_id=22"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("repeated form empresa_id was accepted")
	}
}

func TestValidateEmpresaIDConsistencyRejectsRepeatedMultipartManipulation(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("empresa_id", "11")
	_ = writer.WriteField("empresa_id", "22")
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/recurso", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if err := validateEmpresaIDConsistency(req, 11); err == nil {
		t.Fatal("repeated multipart empresa_id was accepted")
	}
}
