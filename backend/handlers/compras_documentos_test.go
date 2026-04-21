package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func cleanupComprasComprobantesArtifacts(t *testing.T, empresaID int64) {
	t.Helper()
	suffix := "empresa_" + strconv.FormatInt(empresaID, 10)
	bases := []string{
		filepath.Join("web", "uploads", "comprobantes", suffix),
		filepath.Join("..", "web", "uploads", "comprobantes", suffix),
	}
	t.Cleanup(func() {
		for _, base := range bases {
			_ = os.RemoveAll(base)
		}
	})
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

func TestEmpresaComprasDocumentosAprobacionMultinivel(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_documentos_aprobacion_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	proveedorID := seedProveedorComprasTest(t, dbEmp, 64, "6404")
	h := EmpresaComprasDocumentosHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":64,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"OC-6401","accion":"solicitar_aprobacion","requiere_aprobacion":true,"niveles_aprobacion_requeridos":2,"monto_total":1200000,"moneda":"COP"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}
	respCreate := decodeBodyMap(t, rrCreate)
	resultadoCreate := respCreate["resultado"].(map[string]interface{})
	if resultadoCreate["estado_documento"].(string) != "pendiente_aprobacion" {
		t.Fatalf("expected estado_documento pendiente_aprobacion, got %v", resultadoCreate["estado_documento"])
	}
	if !resultadoCreate["requiere_aprobacion"].(bool) {
		t.Fatalf("expected requiere_aprobacion=true")
	}

	reqAprobar1 := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":64,"documento_codigo":"OC-6401","accion":"aprobar_compra","observaciones":"aprobacion nivel 1"}`))
	reqAprobar1 = reqAprobar1.WithContext(context.WithValue(reqAprobar1.Context(), "adminEmail", "aprobador.nivel1@empresa.com"))
	reqAprobar1.Header.Set("Content-Type", "application/json")
	rrAprobar1 := httptest.NewRecorder()
	h.ServeHTTP(rrAprobar1, reqAprobar1)
	if rrAprobar1.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrAprobar1.Code, rrAprobar1.Body.String())
	}
	respAprobar1 := decodeBodyMap(t, rrAprobar1)
	resultadoAprobar1 := respAprobar1["resultado"].(map[string]interface{})
	if resultadoAprobar1["estado_documento"].(string) != "pendiente_aprobacion" {
		t.Fatalf("expected estado_documento pendiente_aprobacion after first approve, got %v", resultadoAprobar1["estado_documento"])
	}
	if int(resultadoAprobar1["nivel_aprobacion_actual"].(float64)) != 1 {
		t.Fatalf("expected nivel_aprobacion_actual 1, got %v", resultadoAprobar1["nivel_aprobacion_actual"])
	}

	reqAprobar2 := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":64,"documento_codigo":"OC-6401","accion":"aprobar_compra","observaciones":"aprobacion nivel 2"}`))
	reqAprobar2 = reqAprobar2.WithContext(context.WithValue(reqAprobar2.Context(), "adminEmail", "aprobador.nivel2@empresa.com"))
	reqAprobar2.Header.Set("Content-Type", "application/json")
	rrAprobar2 := httptest.NewRecorder()
	h.ServeHTTP(rrAprobar2, reqAprobar2)
	if rrAprobar2.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrAprobar2.Code, rrAprobar2.Body.String())
	}
	respAprobar2 := decodeBodyMap(t, rrAprobar2)
	resultadoAprobar2 := respAprobar2["resultado"].(map[string]interface{})
	if resultadoAprobar2["estado_documento"].(string) != "emitida" {
		t.Fatalf("expected estado_documento emitida after second approve, got %v", resultadoAprobar2["estado_documento"])
	}
	if int(resultadoAprobar2["nivel_aprobacion_actual"].(float64)) != 2 {
		t.Fatalf("expected nivel_aprobacion_actual 2, got %v", resultadoAprobar2["nivel_aprobacion_actual"])
	}
}

func TestEmpresaComprasDocumentosRecepcionParcialYValidacionDocumental(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_documentos_recepcion_validacion_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	proveedorID := seedProveedorComprasTest(t, dbEmp, 65, "6505")
	h := EmpresaComprasDocumentosHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":65,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"OC-6501","accion":"emitir_orden","monto_total":980000,"moneda":"COP"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}

	reqRecepcionParcial := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":65,"documento_codigo":"OC-6501","accion":"recepcionar_parcial_compra","recepcion_items":[{"producto_id":1001,"cantidad_ordenada":10,"cantidad_recibida":7,"costo_unitario":20000,"diferencia_motivo":"faltante proveedor"},{"producto_id":1002,"cantidad_ordenada":5,"cantidad_recibida":5,"costo_unitario":15000}]}`))
	reqRecepcionParcial = reqRecepcionParcial.WithContext(context.WithValue(reqRecepcionParcial.Context(), "adminEmail", "compras@test.com"))
	reqRecepcionParcial.Header.Set("Content-Type", "application/json")
	rrRecepcionParcial := httptest.NewRecorder()
	h.ServeHTTP(rrRecepcionParcial, reqRecepcionParcial)
	if rrRecepcionParcial.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrRecepcionParcial.Code, rrRecepcionParcial.Body.String())
	}
	respRecepcionParcial := decodeBodyMap(t, rrRecepcionParcial)
	resultadoRecepcionParcial := respRecepcionParcial["resultado"].(map[string]interface{})
	if resultadoRecepcionParcial["estado_documento"].(string) != "recepcion_parcial" {
		t.Fatalf("expected estado_documento recepcion_parcial, got %v", resultadoRecepcionParcial["estado_documento"])
	}
	recepcionResumen := respRecepcionParcial["recepcion_resumen"].(map[string]interface{})
	if int(recepcionResumen["items_pendientes"].(float64)) != 1 {
		t.Fatalf("expected items_pendientes=1, got %v", recepcionResumen["items_pendientes"])
	}

	reqValidar := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":65,"documento_codigo":"OC-6501","accion":"validar_documentos","proveedor_documento_ref":"9006505","factura_documento_ref":"FAC-6501","entrada_documento_ref":"ENT-6501"}`))
	reqValidar = reqValidar.WithContext(context.WithValue(reqValidar.Context(), "adminEmail", "compras@test.com"))
	reqValidar.Header.Set("Content-Type", "application/json")
	rrValidar := httptest.NewRecorder()
	h.ServeHTTP(rrValidar, reqValidar)
	if rrValidar.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrValidar.Code, rrValidar.Body.String())
	}
	respValidar := decodeBodyMap(t, rrValidar)
	resultadoValidar := respValidar["resultado"].(map[string]interface{})
	if resultadoValidar["validacion_documental_estado"].(string) != "validada" {
		t.Fatalf("expected validacion_documental_estado=validada, got %v", resultadoValidar["validacion_documental_estado"])
	}

	reqRecepcionarCompleta := httptest.NewRequest(http.MethodPut, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":65,"documento_codigo":"OC-6501","accion":"recepcionar_compra","recepcion_items":[{"producto_id":1001,"cantidad_ordenada":10,"cantidad_recibida":10,"costo_unitario":20000},{"producto_id":1002,"cantidad_ordenada":5,"cantidad_recibida":5,"costo_unitario":15000}]}`))
	reqRecepcionarCompleta = reqRecepcionarCompleta.WithContext(context.WithValue(reqRecepcionarCompleta.Context(), "adminEmail", "compras@test.com"))
	reqRecepcionarCompleta.Header.Set("Content-Type", "application/json")
	rrRecepcionarCompleta := httptest.NewRecorder()
	h.ServeHTTP(rrRecepcionarCompleta, reqRecepcionarCompleta)
	if rrRecepcionarCompleta.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrRecepcionarCompleta.Code, rrRecepcionarCompleta.Body.String())
	}
	respRecepcionarCompleta := decodeBodyMap(t, rrRecepcionarCompleta)
	resultadoRecepcionarCompleta := respRecepcionarCompleta["resultado"].(map[string]interface{})
	if resultadoRecepcionarCompleta["estado_documento"].(string) != "recepcionada" {
		t.Fatalf("expected estado_documento recepcionada, got %v", resultadoRecepcionarCompleta["estado_documento"])
	}
}

func TestEmpresaComprasDocumentoComprobanteUploadHandler(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_compras_documentos_comprobante_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	const empresaID int64 = 66
	cleanupComprasComprobantesArtifacts(t, empresaID)
	proveedorID := seedProveedorComprasTest(t, dbEmp, empresaID, "PRV-COMP-04")

	hDocs := EmpresaComprasDocumentosHandler(dbEmp)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos", strings.NewReader(`{"empresa_id":66,"proveedor_id":`+strconv.FormatInt(proveedorID, 10)+`,"documento_codigo":"OC-6601","accion":"crear","monto_total":430000,"moneda":"COP"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	hDocs.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", strconv.FormatInt(empresaID, 10)); err != nil {
		t.Fatalf("write empresa_id field: %v", err)
	}
	if err := writer.WriteField("documento_codigo", "OC-6601"); err != nil {
		t.Fatalf("write documento_codigo field: %v", err)
	}
	if err := writer.WriteField("tipo_documento", "orden_compra"); err != nil {
		t.Fatalf("write tipo_documento field: %v", err)
	}
	part, err := writer.CreateFormFile("archivo", "soporte.pdf")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte("pdf simulado")); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	hUpload := EmpresaComprasDocumentoComprobanteUploadHandler(dbEmp)
	reqUpload := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/documentos/comprobante", &body)
	reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	reqUpload = reqUpload.WithContext(context.WithValue(reqUpload.Context(), "adminEmail", "compras@test.com"))
	rrUpload := httptest.NewRecorder()
	hUpload.ServeHTTP(rrUpload, reqUpload)
	if rrUpload.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rrUpload.Code, rrUpload.Body.String())
	}

	doc, err := dbpkg.GetEmpresaDocumentoCompraByCodigo(dbEmp, empresaID, "orden_compra", "OC-6601")
	if err != nil {
		t.Fatalf("get documento compra: %v", err)
	}
	if strings.TrimSpace(doc.ComprobanteURL) == "" {
		t.Fatalf("expected comprobante_url persisted")
	}
	if !strings.HasSuffix(strings.ToLower(doc.ComprobanteNombre), ".pdf") {
		t.Fatalf("expected comprobante nombre .pdf, got %q", doc.ComprobanteNombre)
	}
	absPath := filepath.Join(resolveWebRootDir(), filepath.FromSlash(strings.TrimPrefix(doc.ComprobanteURL, "/")))
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("expected uploaded comprobante file: %v", err)
	}
}
