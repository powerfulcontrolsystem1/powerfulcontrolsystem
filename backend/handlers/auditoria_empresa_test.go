package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
	utilspkg "github.com/you/pos-backend/utils"
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

func TestWithEmpresaVentasPermissionsRegistraAuditoriaAccionCritica(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_ventas.db")
	dbSuper := openTestSQLite(t, "super_auditoria_ventas.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 511, "cajero@ventas.com")
	seedPermsAdmin(t, dbSuper, "cajero@ventas.com", "cajero")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?action=cerrar&empresa_id=511&id=7001", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@ventas.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 511, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 auditoria event, got %d", len(eventos))
	}
	if eventos[0].Modulo != "ventas" {
		t.Fatalf("expected modulo ventas, got %q", eventos[0].Modulo)
	}
	if eventos[0].Accion != "cerrar" {
		t.Fatalf("expected accion cerrar, got %q", eventos[0].Accion)
	}
}

func TestWithEmpresaComprasPermissionsRegistraAuditoriaAccionCritica(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_compras.db")
	dbSuper := openTestSQLite(t, "super_auditoria_compras.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 512, "compras@empresa.com")
	seedPermsAdmin(t, dbSuper, "compras@empresa.com", "compras")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	h := WithEmpresaComprasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/proveedores?action=emitir_orden&empresa_id=512&id=8012", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "compras@empresa.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusAccepted, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 512, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 auditoria event, got %d", len(eventos))
	}
	if eventos[0].Modulo != "compras" {
		t.Fatalf("expected modulo compras, got %q", eventos[0].Modulo)
	}
	if eventos[0].Accion != "emitir_orden" {
		t.Fatalf("expected accion emitir_orden, got %q", eventos[0].Accion)
	}
}

func TestWithEmpresaFacturacionPermissionsRegistraAuditoriaAccionCritica(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_facturacion.db")
	dbSuper := openTestSQLite(t, "super_auditoria_facturacion.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 513, "cajero@facturacion.com")
	seedPermsAdmin(t, dbSuper, "cajero@facturacion.com", "cajero")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	h := WithEmpresaFacturacionPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	})

	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica?action=emitir&empresa_id=513&documento_codigo=FAC-001", strings.NewReader(`{"empresa_id":513}`))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@facturacion.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 513, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 auditoria event, got %d", len(eventos))
	}
	if eventos[0].Modulo != "facturacion" {
		t.Fatalf("expected modulo facturacion, got %q", eventos[0].Modulo)
	}
	if eventos[0].Accion != "emitir" {
		t.Fatalf("expected accion emitir, got %q", eventos[0].Accion)
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

func TestEmpresaAuditoriaEventosHandlerFiltrosAvanzados(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_handler_filtros.db")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	_, err := dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      99,
		Modulo:         "ventas",
		Accion:         "cerrar",
		Recurso:        "carritos_compra",
		RecursoID:      1001,
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/carritos_compra",
		Resultado:      "ok",
		CodigoHTTP:     200,
		UsuarioCreador: "cajero@test.com",
	})
	if err != nil {
		t.Fatalf("create auditoria event 1: %v", err)
	}

	_, err = dbpkg.CreateEmpresaAuditoriaEvento(dbEmp, dbpkg.EmpresaAuditoriaEvento{
		EmpresaID:      99,
		Modulo:         "ventas",
		Accion:         "cerrar",
		Recurso:        "carritos_compra",
		RecursoID:      1002,
		MetodoHTTP:     "PUT",
		Endpoint:       "/api/empresa/carritos_compra",
		Resultado:      "error",
		CodigoHTTP:     409,
		UsuarioCreador: "cajero@test.com",
	})
	if err != nil {
		t.Fatalf("create auditoria event 2: %v", err)
	}

	h := EmpresaAuditoriaEventosHandler(dbEmp)

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/auditoria/eventos?empresa_id=99&modulo=ventas&codigo_http=409&recurso_id=1002&metodo_http=PUT&recurso=carritos&endpoint=carritos_compra&search=cajero&limit=20&offset=0", nil)
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList.Code, rrList.Body.String())
	}
	if got := strings.TrimSpace(rrList.Header().Get("X-Total-Count")); got != "1" {
		t.Fatalf("expected X-Total-Count=1, got %q", got)
	}
	if got := strings.TrimSpace(rrList.Header().Get("X-Page-Limit")); got != "20" {
		t.Fatalf("expected X-Page-Limit=20, got %q", got)
	}
	if got := strings.TrimSpace(rrList.Header().Get("X-Page-Offset")); got != "0" {
		t.Fatalf("expected X-Page-Offset=0, got %q", got)
	}

	var rows []dbpkg.EmpresaAuditoriaEvento
	if err := json.Unmarshal(rrList.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode auditoria list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 auditoria event with advanced filters, got %d", len(rows))
	}
	if rows[0].CodigoHTTP != 409 || rows[0].RecursoID != 1002 {
		t.Fatalf("unexpected row returned, got codigo_http=%d recurso_id=%d", rows[0].CodigoHTTP, rows[0].RecursoID)
	}

	reqInvalid := httptest.NewRequest(http.MethodGet, "/api/empresa/auditoria/eventos?empresa_id=99&codigo_http=abc", nil)
	rrInvalid := httptest.NewRecorder()
	h.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for invalid codigo_http, got %d body=%s", http.StatusBadRequest, rrInvalid.Code, rrInvalid.Body.String())
	}

	reqInvalidDate := httptest.NewRequest(http.MethodGet, "/api/empresa/auditoria/eventos?empresa_id=99&desde=2026-99-99", nil)
	rrInvalidDate := httptest.NewRecorder()
	h.ServeHTTP(rrInvalidDate, reqInvalidDate)
	if rrInvalidDate.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d for invalid desde, got %d body=%s", http.StatusBadRequest, rrInvalidDate.Code, rrInvalidDate.Body.String())
	}
}

func TestWithEmpresaVentasPermissionsRegistraRequestIDDesdeMiddleware(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_reqid.db")
	dbSuper := openTestSQLite(t, "super_auditoria_reqid.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 620, "cajero@reqid.com")
	seedPermsAdmin(t, dbSuper, "cajero@reqid.com", "cajero")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	wrapped := utilspkg.LoggingMiddleware(h)

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?action=cerrar&empresa_id=620&id=77", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "cajero@reqid.com"))
	rr := httptest.NewRecorder()

	wrapped.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNoContent, rr.Code, rr.Body.String())
	}
	reqID := strings.TrimSpace(rr.Header().Get("X-Request-ID"))
	if reqID == "" {
		t.Fatalf("expected response X-Request-ID header")
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 620, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 auditoria event, got %d", len(eventos))
	}
	if eventos[0].RequestID != reqID {
		t.Fatalf("expected request_id=%q, got %q", reqID, eventos[0].RequestID)
	}
}

func TestWithEmpresaVentasPermissionsRegistraIntentoDenegadoCritico(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_auditoria_denegada.db")
	dbSuper := openTestSQLite(t, "super_auditoria_denegada.db")
	ensurePermsEmpresasSchema(t, dbEmp)
	ensurePermsAdminSchema(t, dbSuper)
	seedPermsEmpresa(t, dbEmp, 630, "auditor@denegada.com")
	seedPermsAdmin(t, dbSuper, "auditor@denegada.com", "auditor")
	if err := dbpkg.EnsureEmpresaAuditoriaSchema(dbEmp); err != nil {
		t.Fatalf("ensure auditoria schema: %v", err)
	}

	called := false
	h := WithEmpresaVentasPermissions(dbEmp, dbSuper, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodDelete, "/api/empresa/carritos_compra?empresa_id=630&id=88", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "auditor@denegada.com"))
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusForbidden, rr.Code, rr.Body.String())
	}
	if called {
		t.Fatalf("expected next handler not to be called")
	}

	eventos, err := dbpkg.ListEmpresaAuditoriaEventos(dbEmp, 630, dbpkg.EmpresaAuditoriaEventoFilter{Limit: 20})
	if err != nil {
		t.Fatalf("list auditoria eventos: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 auditoria event, got %d", len(eventos))
	}
	if eventos[0].Accion != "eliminar" {
		t.Fatalf("expected accion eliminar, got %q", eventos[0].Accion)
	}
	if eventos[0].CodigoHTTP != http.StatusForbidden {
		t.Fatalf("expected codigo_http=%d, got %d", http.StatusForbidden, eventos[0].CodigoHTTP)
	}
	if eventos[0].Resultado != "error" {
		t.Fatalf("expected resultado error, got %q", eventos[0].Resultado)
	}
}
