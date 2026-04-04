package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func seedProveedorComprasTest(t *testing.T, dbEmp *sql.DB, empresaID int64, codigo string) int64 {
	t.Helper()
	id, err := dbpkg.CreateProveedor(dbEmp, dbpkg.Proveedor{
		EmpresaID:      empresaID,
		Codigo:         codigo,
		Nombre:         "Proveedor QA " + codigo,
		Documento:      "900" + codigo,
		UsuarioCreador: "compras@test.com",
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("create proveedor: %v", err)
	}
	return id
}

func decodeBodyMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode body: %v body=%s", err, rr.Body.String())
	}
	return payload
}

func TestEmpresaComprasDocumentosCicloCompleto(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_documentos_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	proveedorID := seedProveedorComprasTest(t, dbEmp, 61, "PRV-COMP-01")
	h := EmpresaComprasDocumentosHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":61,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"OC-6101","accion":"emitir","monto_total":1520000,"moneda":"COP","periodo_contable":"2026-04","observaciones":"emision inicial"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}
	respCreate := decodeBodyMap(t, rrCreate)
	if !respCreate["ok"].(bool) {
		t.Fatalf("expected ok=true")
	}
	resultadoCreate := respCreate["resultado"].(map[string]interface{})
	if resultadoCreate["estado_documento"].(string) != "emitida" {
		t.Fatalf("expected estado_documento emitida, got %v", resultadoCreate["estado_documento"])
	}

	reqRecepcionar := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":61,"documento_codigo":"OC-6101","accion":"recepcionar_compra"}`))
	reqRecepcionar = reqRecepcionar.WithContext(context.WithValue(reqRecepcionar.Context(), "adminEmail", "compras@test.com"))
	reqRecepcionar.Header.Set("Content-Type", "application/json")
	rrRecepcionar := httptest.NewRecorder()
	h.ServeHTTP(rrRecepcionar, reqRecepcionar)
	if rrRecepcionar.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrRecepcionar.Code, rrRecepcionar.Body.String())
	}
	respRecepcionar := decodeBodyMap(t, rrRecepcionar)
	resultadoRecepcionar := respRecepcionar["resultado"].(map[string]interface{})
	if resultadoRecepcionar["estado_documento"].(string) != "recepcionada" {
		t.Fatalf("expected estado_documento recepcionada, got %v", resultadoRecepcionar["estado_documento"])
	}

	reqContabilizar := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":61,"documento_codigo":"OC-6101","accion":"contabilizar_compra"}`))
	reqContabilizar = reqContabilizar.WithContext(context.WithValue(reqContabilizar.Context(), "adminEmail", "compras@test.com"))
	reqContabilizar.Header.Set("Content-Type", "application/json")
	rrContabilizar := httptest.NewRecorder()
	h.ServeHTTP(rrContabilizar, reqContabilizar)
	if rrContabilizar.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrContabilizar.Code, rrContabilizar.Body.String())
	}
	respContabilizar := decodeBodyMap(t, rrContabilizar)
	resultadoContabilizar := respContabilizar["resultado"].(map[string]interface{})
	if resultadoContabilizar["estado_documento"].(string) != "contabilizada" {
		t.Fatalf("expected estado_documento contabilizada, got %v", resultadoContabilizar["estado_documento"])
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/compras/documentos?empresa_id=61&estado_documento=contabilizada", nil)
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList.Code, rrList.Body.String())
	}
	var listado []map[string]interface{}
	if err := json.Unmarshal(rrList.Body.Bytes(), &listado); err != nil {
		t.Fatalf("decode listado: %v body=%s", err, rrList.Body.String())
	}
	if len(listado) != 1 {
		t.Fatalf("expected 1 documento contabilizado, got %d", len(listado))
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 61, dbpkg.EmpresaEventoContableFilter{Modulo: "compras", Limit: 20})
	if err != nil {
		t.Fatalf("list eventos compras: %v", err)
	}
	if !hasEventoContable(eventos, "orden_compra_emitida") {
		t.Fatalf("expected orden_compra_emitida event")
	}
	if !hasEventoContable(eventos, "compra_recepcionada") {
		t.Fatalf("expected compra_recepcionada event")
	}
	if !hasEventoContable(eventos, "compra_contabilizada") {
		t.Fatalf("expected compra_contabilizada event")
	}
}

func TestEmpresaComprasDocumentosActivarYFiltrarInactivos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_documentos_estado_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	proveedorID := seedProveedorComprasTest(t, dbEmp, 62, "PRV-COMP-02")
	h := EmpresaComprasDocumentosHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":62,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"OC-6201","accion":"crear","monto_total":800000,"moneda":"COP"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}

	reqDeactivate := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":62,"documento_codigo":"OC-6201","accion":"activar","activo":false}`))
	reqDeactivate = reqDeactivate.WithContext(context.WithValue(reqDeactivate.Context(), "adminEmail", "compras@test.com"))
	reqDeactivate.Header.Set("Content-Type", "application/json")
	rrDeactivate := httptest.NewRecorder()
	h.ServeHTTP(rrDeactivate, reqDeactivate)
	if rrDeactivate.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusNoContent, rrDeactivate.Code, rrDeactivate.Body.String())
	}

	reqListActive := httptest.NewRequest(http.MethodGet, "/api/empresa/compras/documentos?empresa_id=62", nil)
	rrListActive := httptest.NewRecorder()
	h.ServeHTTP(rrListActive, reqListActive)
	if rrListActive.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrListActive.Code, rrListActive.Body.String())
	}
	var activos []map[string]interface{}
	if err := json.Unmarshal(rrListActive.Body.Bytes(), &activos); err != nil {
		t.Fatalf("decode activos: %v", err)
	}
	if len(activos) != 0 {
		t.Fatalf("expected 0 documentos activos, got %d", len(activos))
	}

	reqListAll := httptest.NewRequest(http.MethodGet, "/api/empresa/compras/documentos?empresa_id=62&include_inactive=1", nil)
	rrListAll := httptest.NewRecorder()
	h.ServeHTTP(rrListAll, reqListAll)
	if rrListAll.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrListAll.Code, rrListAll.Body.String())
	}
	var todos []map[string]interface{}
	if err := json.Unmarshal(rrListAll.Body.Bytes(), &todos); err != nil {
		t.Fatalf("decode todos: %v", err)
	}
	if len(todos) != 1 {
		t.Fatalf("expected 1 documento (incluyendo inactivos), got %d", len(todos))
	}
}

func TestEmpresaComprasDocumentosTransicionInvalida(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_documentos_conflicto_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	proveedorID := seedProveedorComprasTest(t, dbEmp, 63, "PRV-COMP-03")
	h := EmpresaComprasDocumentosHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":63,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"OC-6301","accion":"crear"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}

	reqInvalid := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":63,"documento_codigo":"OC-6301","accion":"contabilizar_compra"}`))
	reqInvalid = reqInvalid.WithContext(context.WithValue(reqInvalid.Context(), "adminEmail", "compras@test.com"))
	reqInvalid.Header.Set("Content-Type", "application/json")
	rrInvalid := httptest.NewRecorder()
	h.ServeHTTP(rrInvalid, reqInvalid)

	if rrInvalid.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rrInvalid.Code, rrInvalid.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrInvalid.Body.String()), "transicion invalida") {
		t.Fatalf("expected transicion invalida message, got body=%s", rrInvalid.Body.String())
	}
}
