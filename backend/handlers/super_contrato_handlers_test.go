package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestPublicContratoHandlerReturnsDefaultVersion(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_public_contract_default.db")

	h := PublicContratoHandler(dbSuper)
	req := httptest.NewRequest(http.MethodGet, "/api/public/contrato", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var body struct {
		Contrato dbpkg.SuperContractVersion `json:"contrato"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v body=%s", err, rr.Body.String())
	}
	if body.Contrato.Version != 1 {
		t.Fatalf("expected default contract version 1, got %d", body.Contrato.Version)
	}
	if strings.TrimSpace(body.Contrato.Titulo) == "" {
		t.Fatalf("expected non-empty title")
	}
	if strings.TrimSpace(body.Contrato.Contenido) == "" {
		t.Fatalf("expected non-empty content")
	}
}

func TestSuperContratoHandlerCreatesNewVersionAndHistory(t *testing.T) {
	dbSuper := openTestSQLite(t, "super_contract_editor.db")
	ensureSuperSchema(t, dbSuper)

	if err := dbpkg.UpsertAdministrador(dbSuper, "super@pcs.com", "Super", "super_administrador", ""); err != nil {
		t.Fatalf("upsert super admin: %v", err)
	}
	token := "token-super-contrato"
	if err := dbpkg.CreateSession(dbSuper, "super@pcs.com", "127.0.0.1", "go-test", token); err != nil {
		t.Fatalf("create session: %v", err)
	}

	h := SuperContratoHandler(dbSuper)
	reqGet := httptest.NewRequest(http.MethodGet, "/super/api/contrato", nil)
	reqGet.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, reqGet)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("expected initial GET status %d, got %d body=%s", http.StatusOK, rrGet.Code, rrGet.Body.String())
	}

	body := strings.NewReader(`{"titulo":"Contrato actualizado PCS","resumen":"Resumen actualizado del contrato.","contenido":"1. Objeto\nContrato actualizado para pruebas.\n\n2. Seguridad\nSe exige aceptar la version vigente.","nota_aceptacion":"Acepto la version vigente del contrato.","resumen_cambio":"Se agrega versionado y aceptacion por version."}`)
	reqSave := httptest.NewRequest(http.MethodPut, "/super/api/contrato", body)
	reqSave.Header.Set("Content-Type", "application/json")
	reqSave.AddCookie(&http.Cookie{Name: "session_token", Value: token})
	rrSave := httptest.NewRecorder()
	h.ServeHTTP(rrSave, reqSave)

	if rrSave.Code != http.StatusOK {
		t.Fatalf("expected save status %d, got %d body=%s", http.StatusOK, rrSave.Code, rrSave.Body.String())
	}

	var saveResp struct {
		Saved   bool                      `json:"saved"`
		Current dbpkg.SuperContractVersion `json:"current"`
		History []dbpkg.SuperContractVersion `json:"history"`
	}
	if err := json.Unmarshal(rrSave.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v body=%s", err, rrSave.Body.String())
	}
	if !saveResp.Saved {
		t.Fatalf("expected saved=true, got false body=%s", rrSave.Body.String())
	}
	if saveResp.Current.Version != 2 {
		t.Fatalf("expected current version 2, got %d", saveResp.Current.Version)
	}
	if len(saveResp.History) < 2 {
		t.Fatalf("expected at least 2 history items, got %d", len(saveResp.History))
	}
	if strings.TrimSpace(saveResp.History[0].ResumenCambio) != "Se agrega versionado y aceptacion por version." {
		t.Fatalf("unexpected latest change summary: %q", saveResp.History[0].ResumenCambio)
	}
}