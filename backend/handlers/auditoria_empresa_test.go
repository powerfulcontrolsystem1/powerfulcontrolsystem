package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestWithEmpresaFinanzasPermissionsRegistraAuditoriaAccionCritica(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_permisos.db")
	dbSuper := openTestSQLite(t, "super_auditoria_permisos.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 501, "conta@audit.com")
	seedPermsAdmin(t, dbSuper, "conta@audit.com", "contabilidad")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/movimientos", strings.NewReader(`{"empresa_id":501,"tipo_movimiento":"ingreso","total":1200}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@audit.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 501, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 auditoria event, got %d", len(eventos))
	}
	if eventos[0].Modulo != "finanzas" {
		t.Fatalf("expected modulo finanzas, got %q", eventos[0].Modulo)
	}
	if eventos[0].Accion != "crear" {
		t.Fatalf("expected accion crear, got %q", eventos[0].Accion)
	}
	if eventos[0].Resultado != "ok" {
		t.Fatalf("expected resultado ok, got %q", eventos[0].Resultado)
	}
	if eventos[0].CodigoHTTP != http.StatusCreated {
		t.Fatalf("expected codigo_http=%d, got %d", http.StatusCreated, eventos[0].CodigoHTTP)
	}
}

func TestWithEmpresaFinanzasPermissionsNoRegistraLectura(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_read.db")
	dbSuper := openTestSQLite(t, "super_auditoria_read.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 502, "auditor@audit.com")
	seedPermsAdmin(t, dbSuper, "auditor@audit.com", "auditor")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	h := WithEmpresaFinanzasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	})

	req := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?empresa_id=502", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "auditor@audit.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 502, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 0 {
		t.Fatalf("expected 0 auditoria events for read action, got %d", len(eventos))
	}
}

func TestEmpresaAuditoriaEventosHandlerConsultaYPurga(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_handler.db")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	_, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      88,
		Modulo:         "seguridad",
		Accion:         "actualizar",
		Recurso:        "usuarios",
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/usuarios",
		Resultado:      "ok",
		CodigoHTTP:     200,
		FechaEvento:    "2024-01-01 00:00:00",
		UsuarioCreador: "admin@test.com",
	})
	if err != nil {
		t.Fatalf("create old auditoria event: %v", err)
	}
	_, err = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      88,
		Modulo:         "seguridad",
		Accion:         "crear",
		Recurso:        "usuarios",
		MetodoHTTP:     "POST",
		Endpoint:       "/api/empresa/usuarios",
		Resultado:      "ok",
		CodigoHTTP:     201,
		UsuarioCreador: "admin@test.com",
	})
	if err != nil {
		t.Fatalf("create fresh auditoria event: %v", err)
	}

	h := EmpresaAuditoriaEventosHandler(dbEmp)

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/auditoria/eventos?empresa_id=88&modulo=seguridad&limit=20", nil)
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList.Code, rrList.Body.String())
	}

	var rows []dbpkg.EmpresaAuditoriaEvento
	if err := json.Unmarshal(rrList.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode auditoria list response: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 auditoria events before purge, got %d", len(rows))
	}

	reqPurge := httptest.NewRequest(http.MethodPut, "/api/empresa/auditoria/eventos?action=retener&empresa_id=88&retencion_dias=30", nil)
	rrPurge := httptest.NewRecorder()
	h.ServeHTTP(rrPurge, reqPurge)
	if rrPurge.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrPurge.Code, rrPurge.Body.String())
	}

	var purgeResp map[string]interface{}
	if err := json.Unmarshal(rrPurge.Body.Bytes(), &purgeResp); err != nil {
		t.Fatalf("decode purge response: %v", err)
	}
	if int64(purgeResp["eliminados"].(float64)) < 1 {
		t.Fatalf("expected at least 1 deleted row in purge, got %v", purgeResp["eliminados"])
	}

	reqList2 := httptest.NewRequest(http.MethodGet, "/api/empresa/auditoria/eventos?empresa_id=88&modulo=seguridad&limit=20", nil)
	rrList2 := httptest.NewRecorder()
	h.ServeHTTP(rrList2, reqList2)
	if rrList2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList2.Code, rrList2.Body.String())
	}
	if err := json.Unmarshal(rrList2.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode auditoria list response after purge: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 auditoria event after purge, got %d", len(rows))
	}
}
