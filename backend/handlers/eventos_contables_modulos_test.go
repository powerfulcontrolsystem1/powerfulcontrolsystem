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
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func seedFacturacionCumplimientoConfig(t *testing.T, dbEmp *sql.DB, empresaID int64) {
	t.Helper()

	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, dbpkg.EmpresaConfiguracionAvanzada{
		EmpresaID:            empresaID,
		TipoDocumentoEmisor:  "NIT",
		NIT:                  "900123456",
		RazonSocial:          "Empresa Facturacion QA SAS",
		PaisCodigo:           "CO",
		AmbienteFE:           "produccion",
		PrefijoFactura:       "FE",
		ResolucionNumero:     "18760000000001",
		ResolucionFechaDesde: "2026-01-01",
		ResolucionFechaHasta: "2030-12-31",
		ConsecutivoDesde:     1,
		ConsecutivoHasta:     999999,
		ProximoConsecutivo:   1,
		UsuarioCreador:       "facturacion@test.com",
		Estado:               "activo",
	}); err != nil {
		t.Fatalf("upsert configuracion avanzada: %v", err)
	}

	if _, err := dbpkg.UpsertFacturacionElectronicaPaisConfig(dbEmp, dbpkg.FacturacionElectronicaPaisConfig{
		EmpresaID:           empresaID,
		PaisCodigo:          "CO",
		Proveedor:           "manual",
		Ambiente:            "produccion",
		TipoDocumentoEmisor: "NIT",
		IdentificadorFiscal: "900123456",
		RazonSocial:         "Empresa Facturacion QA SAS",
		PrefijoFactura:      "FE",
		ResolucionNumero:    "18760000000001",
		Estado:              "activo",
		UsuarioCreador:      "facturacion@test.com",
	}); err != nil {
		t.Fatalf("upsert config facturacion pais: %v", err)
	}
}

func TestEmpresaFacturacionElectronicaEmiteEventoContable(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_facturacion_handler.db")
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaFacturacionElectronicaHandler(dbEmp, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica", strings.NewReader(`{"empresa_id":31,"pais_codigo":"CO","ambiente":"sandbox","proveedor":"manual"}`))
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "facturacion@test.com"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 31, dbpkg.EmpresaEventoContableFilter{Modulo: "facturacion", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos facturacion: %v", err)
	}
	if len(eventos) == 0 {
		t.Fatalf("expected eventos contables de facturacion")
	}
	if eventos[0].Evento != "configuracion_facturacion_actualizada" {
		t.Fatalf("expected evento configuracion_facturacion_actualizada, got %q", eventos[0].Evento)
	}
	if eventos[0].Entidad != "facturacion_electronica_pais" {
		t.Fatalf("expected entidad facturacion_electronica_pais, got %q", eventos[0].Entidad)
	}
	if !strings.Contains(eventos[0].PayloadJSON, `"pais_codigo":"CO"`) {
		t.Fatalf("expected payload with pais_codigo CO, got %s", eventos[0].PayloadJSON)
	}
}

func TestEmpresaProveedoresEmiteEventoContableCompras(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_compras_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaProveedoresHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/proveedores", strings.NewReader(`{"empresa_id":12,"codigo":"PRV-01","nombre":"Proveedor Uno","documento":"900123","catalogo_referencia":"CAT-PRV-01","precio_base_referencial":9800.5,"descuento_porcentaje":6.25,"plazo_pago_dias":21,"condicion_entrega":"Entrega semanal"}`))
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "compras@test.com"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode create proveedor response: %v", err)
	}
	createdID := int64(resp["id"].(float64))
	if createdID <= 0 {
		t.Fatalf("expected proveedor id > 0")
	}

	proveedores, err := dbpkg.GetProveedoresByEmpresa(dbEmp, 12, true)
	if err != nil {
		t.Fatalf("list proveedores after create: %v", err)
	}
	if len(proveedores) != 1 {
		t.Fatalf("expected 1 proveedor, got %d", len(proveedores))
	}
	if proveedores[0].CatalogoReferencia != "CAT-PRV-01" {
		t.Fatalf("expected catalogo_referencia CAT-PRV-01, got %q", proveedores[0].CatalogoReferencia)
	}
	if proveedores[0].PrecioBaseReferencial != 9800.5 {
		t.Fatalf("expected precio_base_referencial 9800.5, got %.2f", proveedores[0].PrecioBaseReferencial)
	}
	if proveedores[0].DescuentoPorcentaje != 6.25 {
		t.Fatalf("expected descuento_porcentaje 6.25, got %.2f", proveedores[0].DescuentoPorcentaje)
	}
	if proveedores[0].PlazoPagoDias != 21 {
		t.Fatalf("expected plazo_pago_dias 21, got %d", proveedores[0].PlazoPagoDias)
	}
	if proveedores[0].CondicionEntrega != "Entrega semanal" {
		t.Fatalf("expected condicion_entrega 'Entrega semanal', got %q", proveedores[0].CondicionEntrega)
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 12, dbpkg.EmpresaEventoContableFilter{Modulo: "compras", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos compras: %v", err)
	}
	if len(eventos) == 0 {
		t.Fatalf("expected eventos contables de compras")
	}
	if eventos[0].Evento != "proveedor_registrado" {
		t.Fatalf("expected evento proveedor_registrado, got %q", eventos[0].Evento)
	}
	if eventos[0].EntidadID != createdID {
		t.Fatalf("expected entidad_id=%d, got %d", createdID, eventos[0].EntidadID)
	}
}

func TestEmpresaProveedoresRechazaCamposComercialesInvalidos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_compras_invalid_payload_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaProveedoresHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/proveedores", strings.NewReader(`{"empresa_id":12,"nombre":"Proveedor Invalido","descuento_porcentaje":120}`))
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "compras@test.com"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
	if !strings.Contains(strings.ToLower(rr.Body.String()), "descuento_porcentaje") {
		t.Fatalf("expected validation error about descuento_porcentaje, got body=%s", rr.Body.String())
	}
}

func TestEmpresaFinanzasEmiteEventosContables(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_finanzas_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	hMov := EmpresaFinanzasMovimientosHandler(dbEmp)
	movReq := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/movimientos", strings.NewReader(`{"empresa_id":44,"tipo_movimiento":"ingreso","concepto":"Ingreso caja","categoria":"ventas","metodo_pago":"efectivo","monto":250000,"total":250000,"moneda":"COP"}`))
	movReq = movReq.WithContext(context.WithValue(movReq.Context(), "adminEmail", "finanzas@test.com"))
	movReq.Header.Set("Content-Type", "application/json")
	movRR := httptest.NewRecorder()
	hMov.ServeHTTP(movRR, movReq)
	if movRR.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, movRR.Code, movRR.Body.String())
	}

	hPer := EmpresaFinanzasPeriodosHandler(dbEmp)
	perCreateReq := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/periodos", strings.NewReader(`{"empresa_id":44,"periodo":"2026-04","fecha_inicio":"2026-04-01","fecha_fin":"2026-04-30"}`))
	perCreateReq = perCreateReq.WithContext(context.WithValue(perCreateReq.Context(), "adminEmail", "finanzas@test.com"))
	perCreateReq.Header.Set("Content-Type", "application/json")
	perCreateRR := httptest.NewRecorder()
	hPer.ServeHTTP(perCreateRR, perCreateReq)
	if perCreateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, perCreateRR.Code, perCreateRR.Body.String())
	}

	perCloseReq := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/periodos?action=cerrar&empresa_id=44&periodo=2026-04", strings.NewReader(`{"autorizado_por":"director.finanzas@test.com","motivo_autorizacion":"cierre contable mensual con bloqueo de cambios","evidencia_autorizacion":"ACTA-CIERRE-2026-04","codigo_autorizacion":"APR-2026-04-001"}`))
	perCloseReq = perCloseReq.WithContext(context.WithValue(perCloseReq.Context(), "adminEmail", "finanzas@test.com"))
	perCloseReq.Header.Set("Content-Type", "application/json")
	perCloseRR := httptest.NewRecorder()
	hPer.ServeHTTP(perCloseRR, perCloseReq)
	if perCloseRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, perCloseRR.Code, perCloseRR.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 44, dbpkg.EmpresaEventoContableFilter{Modulo: "finanzas", Limit: 20})
	if err != nil {
		t.Fatalf("list eventos finanzas: %v", err)
	}
	if len(eventos) < 2 {
		t.Fatalf("expected at least 2 eventos de finanzas, got %d", len(eventos))
	}
	if !hasEventoContable(eventos, "movimiento_ingreso_registrado") {
		t.Fatalf("expected movimiento_ingreso_registrado event")
	}
	if !hasEventoContable(eventos, "periodo_contable_cerrado") {
		t.Fatalf("expected periodo_contable_cerrado event")
	}
	periodoCerradoEvt, ok := findEventoContable(eventos, "periodo_contable_cerrado")
	if !ok {
		t.Fatalf("expected periodo_contable_cerrado event payload")
	}
	if !strings.Contains(periodoCerradoEvt.PayloadJSON, `"evidencia_autorizacion":"ACTA-CIERRE-2026-04"`) {
		t.Fatalf("expected evidencia_autorizacion in payload, got %s", periodoCerradoEvt.PayloadJSON)
	}
	if !strings.Contains(periodoCerradoEvt.PayloadJSON, `"autorizado_por":"director.finanzas@test.com"`) {
		t.Fatalf("expected autorizado_por in payload, got %s", periodoCerradoEvt.PayloadJSON)
	}
}

func TestEmpresaFinanzasMovimientoComprobanteUploadHandler(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_finanzas_comprobante_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	const empresaID int64 = 78
	cleanupComprasComprobantesArtifacts(t, empresaID)

	movID, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "egreso",
		Concepto:        "Pago proveedor soporte",
		Categoria:       "compras",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           180000,
		Total:           180000,
		TipoComprobante: "factura",
		UsuarioCreador:  "finanzas@test.com",
	})
	if err != nil {
		t.Fatalf("create movimiento: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", strconv.FormatInt(empresaID, 10)); err != nil {
		t.Fatalf("write empresa_id field: %v", err)
	}
	if err := writer.WriteField("movimiento_id", strconv.FormatInt(movID, 10)); err != nil {
		t.Fatalf("write movimiento_id field: %v", err)
	}
	part, err := writer.CreateFormFile("archivo", "egreso.png")
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}
	if _, err := part.Write([]byte("png simulado")); err != nil {
		t.Fatalf("write multipart file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}

	hUpload := EmpresaFinanzasMovimientoComprobanteUploadHandler(dbEmp)
	reqUpload := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/movimientos/comprobante", &body)
	reqUpload.Header.Set("Content-Type", writer.FormDataContentType())
	reqUpload = reqUpload.WithContext(context.WithValue(reqUpload.Context(), "adminEmail", "finanzas@test.com"))
	rrUpload := httptest.NewRecorder()
	hUpload.ServeHTTP(rrUpload, reqUpload)
	if rrUpload.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rrUpload.Code, rrUpload.Body.String())
	}

	movs, err := dbpkg.ListEmpresaFinanzasMovimientos(dbEmp, empresaID, dbpkg.EmpresaFinanzasMovimientoFilter{IncludeInactive: true, Limit: 10})
	if err != nil {
		t.Fatalf("list movimientos: %v", err)
	}
	if len(movs) != 1 {
		t.Fatalf("expected 1 movimiento, got %d", len(movs))
	}
	if strings.TrimSpace(movs[0].ComprobanteURL) == "" {
		t.Fatalf("expected comprobante_url persisted")
	}
	absPath := filepath.Join(resolveWebRootDir(), filepath.FromSlash(strings.TrimPrefix(movs[0].ComprobanteURL, "/")))
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("expected uploaded comprobante file: %v", err)
	}
}

func TestEmpresaFinanzasPeriodosRequiereEvidenciaAutorizacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_finanzas_periodo_autorizacion_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	hPer := EmpresaFinanzasPeriodosHandler(dbEmp)
	perCreateReq := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/periodos", strings.NewReader(`{"empresa_id":77,"periodo":"2026-04","fecha_inicio":"2026-04-01","fecha_fin":"2026-04-30"}`))
	perCreateReq = perCreateReq.WithContext(context.WithValue(perCreateReq.Context(), "adminEmail", "finanzas@test.com"))
	perCreateReq.Header.Set("Content-Type", "application/json")
	perCreateRR := httptest.NewRecorder()
	hPer.ServeHTTP(perCreateRR, perCreateReq)
	if perCreateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, perCreateRR.Code, perCreateRR.Body.String())
	}

	reqCloseSinEvidencia := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/periodos?action=cerrar&empresa_id=77&periodo=2026-04", strings.NewReader(`{"autorizado_por":"director.finanzas@test.com","motivo_autorizacion":"cierre contable mensual"}`))
	reqCloseSinEvidencia = reqCloseSinEvidencia.WithContext(context.WithValue(reqCloseSinEvidencia.Context(), "adminEmail", "finanzas@test.com"))
	reqCloseSinEvidencia.Header.Set("Content-Type", "application/json")
	rrCloseSinEvidencia := httptest.NewRecorder()
	hPer.ServeHTTP(rrCloseSinEvidencia, reqCloseSinEvidencia)
	if rrCloseSinEvidencia.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rrCloseSinEvidencia.Code, rrCloseSinEvidencia.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrCloseSinEvidencia.Body.String()), "evidencia_autorizacion") {
		t.Fatalf("expected evidencia_autorizacion validation error, got body=%s", rrCloseSinEvidencia.Body.String())
	}
}

func TestEmpresaFacturacionTransaccionalEmiteEventosContables(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_facturacion_transaccional_handler.db")
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
	seedFacturacionCumplimientoConfig(t, dbEmp, 31)

	h := EmpresaFacturacionElectronicaHandler(dbEmp, nil)

	reqEmitir := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=emitir", strings.NewReader(`{"empresa_id":31,"documento_codigo":"FAC-1001","estado_actual":"borrador","monto_total":120000,"moneda":"COP","periodo_contable":"2026-04"}`))
	reqEmitir = reqEmitir.WithContext(context.WithValue(reqEmitir.Context(), "adminEmail", "facturacion@test.com"))
	reqEmitir.Header.Set("Content-Type", "application/json")
	rrEmitir := httptest.NewRecorder()
	h.ServeHTTP(rrEmitir, reqEmitir)
	if rrEmitir.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrEmitir.Code, rrEmitir.Body.String())
	}
	if !strings.Contains(rrEmitir.Body.String(), `"estado_nuevo":"emitida"`) {
		t.Fatalf("expected estado_nuevo emitida, got body=%s", rrEmitir.Body.String())
	}
	if !strings.Contains(rrEmitir.Body.String(), `"numero_legal":"FE-1"`) {
		t.Fatalf("expected numero_legal FE-1, got body=%s", rrEmitir.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrEmitir.Body.String()), `"cumplimiento_normativo"`) {
		t.Fatalf("expected cumplimiento_normativo block, got body=%s", rrEmitir.Body.String())
	}

	reqAnular := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=anular", strings.NewReader(`{"empresa_id":31,"documento_codigo":"FAC-1001","estado_actual":"emitida","periodo_contable":"2026-04"}`))
	reqAnular = reqAnular.WithContext(context.WithValue(reqAnular.Context(), "adminEmail", "facturacion@test.com"))
	reqAnular.Header.Set("Content-Type", "application/json")
	rrAnular := httptest.NewRecorder()
	h.ServeHTTP(rrAnular, reqAnular)
	if rrAnular.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrAnular.Code, rrAnular.Body.String())
	}

	reqNC := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=nota_credito", strings.NewReader(`{"empresa_id":31,"documento_codigo":"NC-1001","estado_actual":"emitida","monto_total":10000,"moneda":"COP","periodo_contable":"2026-04"}`))
	reqNC = reqNC.WithContext(context.WithValue(reqNC.Context(), "adminEmail", "facturacion@test.com"))
	reqNC.Header.Set("Content-Type", "application/json")
	rrNC := httptest.NewRecorder()
	h.ServeHTTP(rrNC, reqNC)
	if rrNC.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrNC.Code, rrNC.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 31, dbpkg.EmpresaEventoContableFilter{Modulo: "facturacion", Limit: 20})
	if err != nil {
		t.Fatalf("list eventos facturacion: %v", err)
	}
	if !hasEventoContable(eventos, "factura_emitida") {
		t.Fatalf("expected factura_emitida event")
	}
	if !hasEventoContable(eventos, "factura_anulada") {
		t.Fatalf("expected factura_anulada event")
	}
	if !hasEventoContable(eventos, "nota_credito_emitida") {
		t.Fatalf("expected nota_credito_emitida event")
	}

	facturaEmitida, ok := findEventoContable(eventos, "factura_emitida")
	if !ok {
		t.Fatalf("factura_emitida event not found")
	}
	facturaAnulada, ok := findEventoContable(eventos, "factura_anulada")
	if !ok {
		t.Fatalf("factura_anulada event not found")
	}
	if facturaEmitida.EntidadID <= 0 {
		t.Fatalf("expected factura_emitida entidad_id > 0, got %d", facturaEmitida.EntidadID)
	}
	if facturaAnulada.EntidadID != facturaEmitida.EntidadID {
		t.Fatalf("expected factura_anulada entidad_id=%d, got %d", facturaEmitida.EntidadID, facturaAnulada.EntidadID)
	}

	docPersistido, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, 31, "factura_electronica", "FAC-1001")
	if err != nil {
		t.Fatalf("get documento facturacion persistido: %v", err)
	}
	if docPersistido.NumeroLegal != "FE-1" {
		t.Fatalf("expected numero_legal FE-1, got %q", docPersistido.NumeroLegal)
	}
	if docPersistido.CodigoValidacion == "" {
		t.Fatalf("expected codigo_validacion not empty")
	}
}

func TestEmpresaFacturacionTransaccionalReportaEstadoCorreoCliente(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_facturacion_email_cliente_handler.db")
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaClientesSchema(dbEmp); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	seedFacturacionCumplimientoConfig(t, dbEmp, 31)

	clienteID, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{
		EmpresaID:         31,
		TipoDocumento:     "CC",
		NumeroDocumento:   "10000001",
		NombreRazonSocial: "Cliente FE QA",
		Email:             "cliente.fe.qa@test.com",
		UsuarioCreador:    "facturacion@test.com",
	})
	if err != nil {
		t.Fatalf("create cliente: %v", err)
	}

	h := EmpresaFacturacionElectronicaHandler(dbEmp, nil)
	reqEmitir := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=emitir", strings.NewReader(`{"empresa_id":31,"documento_codigo":"FAC-EMAIL-1001","estado_actual":"borrador","monto_total":95000,"moneda":"COP","periodo_contable":"2026-04","cliente_id":`+strconv.FormatInt(clienteID, 10)+`}`))
	reqEmitir = reqEmitir.WithContext(context.WithValue(reqEmitir.Context(), "adminEmail", "facturacion@test.com"))
	reqEmitir.Header.Set("Content-Type", "application/json")
	rrEmitir := httptest.NewRecorder()
	h.ServeHTTP(rrEmitir, reqEmitir)
	if rrEmitir.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrEmitir.Code, rrEmitir.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rrEmitir.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	facturaEmailRaw, ok := resp["factura_email"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected factura_email block, got %+v", resp)
	}
	if intentado, _ := facturaEmailRaw["intentado"].(bool); !intentado {
		t.Fatalf("expected intentado=true, got %+v", facturaEmailRaw)
	}
	if enviado, _ := facturaEmailRaw["enviado"].(bool); enviado {
		t.Fatalf("expected enviado=false without SMTP config, got %+v", facturaEmailRaw)
	}
	if destinatario, _ := facturaEmailRaw["destinatario"].(string); destinatario != "cliente.fe.qa@test.com" {
		t.Fatalf("expected destinatario cliente.fe.qa@test.com, got %+v", facturaEmailRaw)
	}
	if errMsg, _ := facturaEmailRaw["error"].(string); strings.TrimSpace(errMsg) == "" {
		t.Fatalf("expected email error detail when SMTP not configured, got %+v", facturaEmailRaw)
	}
}

func TestEmpresaFacturacionTransaccionalEmitirRechazaSinCumplimientoLegal(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_facturacion_transaccional_sin_cumplimiento_handler.db")
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaFacturacionElectronicaHandler(dbEmp, nil)
	reqEmitir := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=emitir", strings.NewReader(`{"empresa_id":31,"documento_codigo":"FAC-2001","estado_actual":"borrador","monto_total":85000,"moneda":"COP","periodo_contable":"2026-04"}`))
	reqEmitir = reqEmitir.WithContext(context.WithValue(reqEmitir.Context(), "adminEmail", "facturacion@test.com"))
	reqEmitir.Header.Set("Content-Type", "application/json")
	rrEmitir := httptest.NewRecorder()
	h.ServeHTTP(rrEmitir, reqEmitir)
	if rrEmitir.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusUnprocessableEntity, rrEmitir.Code, rrEmitir.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrEmitir.Body.String()), "cumplimiento normativo") {
		t.Fatalf("expected cumplimiento normativo error, got body=%s", rrEmitir.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 31, dbpkg.EmpresaEventoContableFilter{Modulo: "facturacion", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos facturacion: %v", err)
	}
	if hasEventoContable(eventos, "factura_emitida") {
		t.Fatalf("expected no factura_emitida event when cumplimiento fails")
	}
}

func TestEmpresaFacturacionTransaccionalRechazaTransicionInvalida(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_facturacion_transaccional_estado_invalido_handler.db")
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaFacturacionElectronicaHandler(dbEmp, nil)
	req := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=anular", strings.NewReader(`{"empresa_id":31,"documento_codigo":"FAC-9999","estado_actual":"borrador","periodo_contable":"2026-04"}`))
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "facturacion@test.com"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rr.Code, rr.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 31, dbpkg.EmpresaEventoContableFilter{Modulo: "facturacion", Limit: 10})
	if err != nil {
		t.Fatalf("list eventos facturacion: %v", err)
	}
	if hasEventoContable(eventos, "factura_anulada") {
		t.Fatalf("expected no factura_anulada event on invalid transition")
	}
}

func TestEmpresaComprasTransaccionalEmiteEventosContables(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_compras_transaccional_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaProveedoresHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/proveedores", strings.NewReader(`{"empresa_id":12,"codigo":"PRV-02","nombre":"Proveedor Dos","documento":"900222"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create proveedor response: %v", err)
	}
	proveedorID := int64(createResp["id"].(float64))

	reqEmitirOC := httptest.NewRequest(http.MethodPut, "/api/empresa/proveedores?action=emitir_orden&empresa_id=12&id="+strconv.FormatInt(proveedorID, 10), strings.NewReader(`{"documento_codigo":"OC-1001","estado_actual":"borrador","monto_total":500000,"moneda":"COP","periodo_contable":"2026-04"}`))
	reqEmitirOC = reqEmitirOC.WithContext(context.WithValue(reqEmitirOC.Context(), "adminEmail", "compras@test.com"))
	reqEmitirOC.Header.Set("Content-Type", "application/json")
	rrEmitirOC := httptest.NewRecorder()
	h.ServeHTTP(rrEmitirOC, reqEmitirOC)
	if rrEmitirOC.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrEmitirOC.Code, rrEmitirOC.Body.String())
	}

	reqRecepcionar := httptest.NewRequest(http.MethodPut, "/api/empresa/proveedores?action=recepcionar_compra&empresa_id=12&id="+strconv.FormatInt(proveedorID, 10), strings.NewReader(`{"documento_codigo":"OC-1001","estado_actual":"emitida","periodo_contable":"2026-04"}`))
	reqRecepcionar = reqRecepcionar.WithContext(context.WithValue(reqRecepcionar.Context(), "adminEmail", "compras@test.com"))
	reqRecepcionar.Header.Set("Content-Type", "application/json")
	rrRecepcionar := httptest.NewRecorder()
	h.ServeHTTP(rrRecepcionar, reqRecepcionar)
	if rrRecepcionar.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrRecepcionar.Code, rrRecepcionar.Body.String())
	}

	reqContabilizar := httptest.NewRequest(http.MethodPut, "/api/empresa/proveedores?action=contabilizar_compra&empresa_id=12&id="+strconv.FormatInt(proveedorID, 10), strings.NewReader(`{"documento_codigo":"OC-1001","estado_actual":"recepcionada","periodo_contable":"2026-04"}`))
	reqContabilizar = reqContabilizar.WithContext(context.WithValue(reqContabilizar.Context(), "adminEmail", "compras@test.com"))
	reqContabilizar.Header.Set("Content-Type", "application/json")
	rrContabilizar := httptest.NewRecorder()
	h.ServeHTTP(rrContabilizar, reqContabilizar)
	if rrContabilizar.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrContabilizar.Code, rrContabilizar.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 12, dbpkg.EmpresaEventoContableFilter{Modulo: "compras", Limit: 20})
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

	emitida, ok := findEventoContable(eventos, "orden_compra_emitida")
	if !ok {
		t.Fatalf("orden_compra_emitida event not found")
	}
	recepcionada, ok := findEventoContable(eventos, "compra_recepcionada")
	if !ok {
		t.Fatalf("compra_recepcionada event not found")
	}
	contabilizada, ok := findEventoContable(eventos, "compra_contabilizada")
	if !ok {
		t.Fatalf("compra_contabilizada event not found")
	}
	if emitida.EntidadID <= 0 {
		t.Fatalf("expected orden_compra_emitida entidad_id > 0, got %d", emitida.EntidadID)
	}
	if recepcionada.EntidadID != emitida.EntidadID {
		t.Fatalf("expected compra_recepcionada entidad_id=%d, got %d", emitida.EntidadID, recepcionada.EntidadID)
	}
	if contabilizada.EntidadID != emitida.EntidadID {
		t.Fatalf("expected compra_contabilizada entidad_id=%d, got %d", emitida.EntidadID, contabilizada.EntidadID)
	}
}

func TestEmpresaComprasTransaccionalRechazaTransicionInvalida(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_eventos_compras_transaccional_estado_invalido_handler.db")
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	h := EmpresaProveedoresHandler(dbEmp)
	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/proveedores", strings.NewReader(`{"empresa_id":12,"codigo":"PRV-03","nombre":"Proveedor Tres","documento":"900333"}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "compras@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCreate.Code, rrCreate.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create proveedor response: %v", err)
	}
	proveedorID := int64(createResp["id"].(float64))

	reqInvalid := httptest.NewRequest(http.MethodPut, "/api/empresa/proveedores?action=contabilizar_compra&empresa_id=12&id="+strconv.FormatInt(proveedorID, 10), strings.NewReader(`{"documento_codigo":"OC-9999","estado_actual":"emitida","periodo_contable":"2026-04"}`))
	reqInvalid = reqInvalid.WithContext(context.WithValue(reqInvalid.Context(), "adminEmail", "compras@test.com"))
	reqInvalid.Header.Set("Content-Type", "application/json")
	rrInvalid := httptest.NewRecorder()
	h.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rrInvalid.Code, rrInvalid.Body.String())
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 12, dbpkg.EmpresaEventoContableFilter{Modulo: "compras", Limit: 20})
	if err != nil {
		t.Fatalf("list eventos compras: %v", err)
	}
	if hasEventoContable(eventos, "compra_contabilizada") {
		t.Fatalf("expected no compra_contabilizada event on invalid transition")
	}
}

func TestEmpresaFinanzasTableroResumenHandler(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_tablero_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaClientesSchema(dbEmp); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	empresaID := int64(55)
	todayDate := time.Now().Format("2006-01-02")
	todayStamp := time.Now().Format("2006-01-02 15:04:05")

	if _, err := dbEmp.Exec(`INSERT INTO carritos_compras (empresa_id, codigo, nombre, estado_carrito, total, total_pagado, pagado_en, estado) VALUES (?, 'C-551', 'Carrito 551', 'cerrado', 90000, 90000, ?, 'activo')`, empresaID, todayStamp); err != nil {
		t.Fatalf("insert carrito: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO clientes (empresa_id, tipo_documento, numero_documento, nombre_razon_social, estado) VALUES (?, 'CC', '551', 'Cliente 551', 'activo')`, empresaID); err != nil {
		t.Fatalf("insert cliente: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO productos (empresa_id, nombre, stock_minimo, estado) VALUES (?, 'Producto 551', 3, 'activo')`, empresaID); err != nil {
		t.Fatalf("insert producto: %v", err)
	}

	var productoID int64
	if err := dbEmp.QueryRow(`SELECT id FROM productos WHERE empresa_id = ? LIMIT 1`, empresaID).Scan(&productoID); err != nil {
		t.Fatalf("select producto id: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_existencias (empresa_id, producto_id, bodega_id, cantidad, estado) VALUES (?, ?, 1, 2, 'activo')`, empresaID, productoID); err != nil {
		t.Fatalf("insert existencia: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO inventario_movimientos (empresa_id, producto_id, tipo, cantidad, costo_unitario, referencia, fecha_movimiento, estado) VALUES (?, ?, 'entrada', 2, 4000, 'COMP-551', ?, 'activo')`, empresaID, productoID, todayStamp); err != nil {
		t.Fatalf("insert movimiento inventario: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Concepto:        "Ingreso tablero",
		Categoria:       "ventas",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           90000,
		Total:           90000,
		FechaMovimiento: todayStamp,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create movimiento ingreso: %v", err)
	}

	h := EmpresaFinanzasMovimientosHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?action=tablero&empresa_id=55&desde="+todayDate+"&hasta="+todayDate, nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "finanzas@test.com"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode tablero payload: %v", err)
	}
	operativo, _ := payload["operativo"].(map[string]interface{})
	financiero, _ := payload["financiero"].(map[string]interface{})
	if operativo == nil || financiero == nil {
		t.Fatalf("expected operativo and financiero blocks in payload")
	}
	if int64(operativo["ventas_cerradas"].(float64)) != 1 {
		t.Fatalf("expected ventas_cerradas=1, got %v", operativo["ventas_cerradas"])
	}
	if int64(financiero["movimientos_ingresos"].(float64)) != 1 {
		t.Fatalf("expected movimientos_ingresos=1, got %v", financiero["movimientos_ingresos"])
	}
	estadoResultados, _ := payload["estado_resultados"].(map[string]interface{})
	balanceGeneral, _ := payload["balance_general"].(map[string]interface{})
	if estadoResultados == nil || balanceGeneral == nil {
		t.Fatalf("expected estado_resultados and balance_general blocks in payload")
	}
}

func TestEmpresaFinanzasTableroResumenExportHandler(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_tablero_export_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaClientesSchema(dbEmp); err != nil {
		t.Fatalf("ensure clientes schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}

	empresaID := int64(56)
	todayDate := time.Now().Format("2006-01-02")
	todayStamp := time.Now().Format("2006-01-02 15:04:05")

	if _, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:       empresaID,
		TipoMovimiento:  "ingreso",
		Concepto:        "Ingreso export",
		Categoria:       "ventas",
		MetodoPago:      "efectivo",
		Moneda:          "COP",
		Monto:           125000,
		Total:           125000,
		FechaMovimiento: todayStamp,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create movimiento ingreso: %v", err)
	}

	h := EmpresaFinanzasMovimientosHandler(dbEmp)

	reqJSON := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?action=tablero_export&format=json&empresa_id=56&desde="+todayDate+"&hasta="+todayDate, nil)
	reqJSON = reqJSON.WithContext(context.WithValue(reqJSON.Context(), "adminEmail", "finanzas@test.com"))
	rrJSON := httptest.NewRecorder()
	h.ServeHTTP(rrJSON, reqJSON)
	if rrJSON.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrJSON.Code, rrJSON.Body.String())
	}
	if ct := strings.ToLower(rrJSON.Header().Get("Content-Type")); !strings.Contains(ct, "application/json") {
		t.Fatalf("expected content-type application/json, got %q", rrJSON.Header().Get("Content-Type"))
	}
	if disp := strings.ToLower(rrJSON.Header().Get("Content-Disposition")); !strings.Contains(disp, ".json") {
		t.Fatalf("expected content-disposition json filename, got %q", rrJSON.Header().Get("Content-Disposition"))
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rrJSON.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode export json payload: %v", err)
	}
	if _, ok := payload["estado_resultados"].(map[string]interface{}); !ok {
		t.Fatalf("expected estado_resultados block in export json payload")
	}
	if _, ok := payload["balance_general"].(map[string]interface{}); !ok {
		t.Fatalf("expected balance_general block in export json payload")
	}

	reqCSV := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?action=tablero_export&format=csv&empresa_id=56&desde="+todayDate+"&hasta="+todayDate, nil)
	reqCSV = reqCSV.WithContext(context.WithValue(reqCSV.Context(), "adminEmail", "finanzas@test.com"))
	rrCSV := httptest.NewRecorder()
	h.ServeHTTP(rrCSV, reqCSV)
	if rrCSV.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrCSV.Code, rrCSV.Body.String())
	}
	if ct := strings.ToLower(rrCSV.Header().Get("Content-Type")); !strings.Contains(ct, "text/csv") {
		t.Fatalf("expected content-type text/csv, got %q", rrCSV.Header().Get("Content-Type"))
	}
	if disp := strings.ToLower(rrCSV.Header().Get("Content-Disposition")); !strings.Contains(disp, ".csv") {
		t.Fatalf("expected content-disposition csv filename, got %q", rrCSV.Header().Get("Content-Disposition"))
	}
	csvBody := rrCSV.Body.String()
	if !strings.Contains(csvBody, "empresa_id,desde,hasta,generado_en,bloque,metrica,valor") {
		t.Fatalf("expected csv header row in export output")
	}
	if !strings.Contains(csvBody, "estado_resultados,utilidad_operacional") {
		t.Fatalf("expected estado_resultados rows in export csv output")
	}
	if !strings.Contains(csvBody, "balance_general,activos") {
		t.Fatalf("expected balance_general rows in export csv output")
	}

	reqInvalid := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?action=tablero_export&format=xlsx&empresa_id=56", nil)
	reqInvalid = reqInvalid.WithContext(context.WithValue(reqInvalid.Context(), "adminEmail", "finanzas@test.com"))
	rrInvalid := httptest.NewRecorder()
	h.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rrInvalid.Code, rrInvalid.Body.String())
	}
}

func TestEmpresaFinanzasMovimientosHandlerConciliacionBancariaAutomatica(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_conciliacion_bancaria_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	empresaID := int64(120)
	fechaMov := "2026-04-15 11:30:00"
	movID, err := dbpkg.CreateEmpresaFinanzasMovimiento(dbEmp, dbpkg.EmpresaFinanzasMovimiento{
		EmpresaID:         empresaID,
		TipoMovimiento:    "ingreso",
		Concepto:          "Ingreso por transferencia",
		Categoria:         "ventas",
		MetodoPago:        "transferencia_bancaria",
		Moneda:            "COP",
		Monto:             200000,
		Total:             200000,
		ReferenciaExterna: "TRX-H-120",
		FechaMovimiento:   fechaMov,
		UsuarioCreador:    "tester",
	})
	if err != nil {
		t.Fatalf("create movimiento financiero: %v", err)
	}

	h := EmpresaFinanzasMovimientosHandler(dbEmp)
	importBody := `{
		"empresa_id":120,
		"auto_conciliar":true,
		"tolerancia_dias":2,
		"tolerancia_monto":10,
		"movimientos":[
			{
				"tipo_movimiento":"ingreso",
				"periodo_contable":"2026-04",
				"fecha_movimiento":"2026-04-15 11:35:00",
				"referencia_bancaria":"TRX-H-120",
				"monto":200000,
				"total":200000,
				"moneda":"COP"
			},
			{
				"tipo_movimiento":"egreso",
				"periodo_contable":"2026-04",
				"fecha_movimiento":"2026-04-16 09:00:00",
				"referencia_bancaria":"TRX-NO-MATCH-120",
				"monto":75000,
				"total":75000,
				"moneda":"COP"
			}
		]
	}`
	reqImport := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/movimientos?action=importar_extractos_bancarios&empresa_id=120", strings.NewReader(importBody))
	reqImport = reqImport.WithContext(context.WithValue(reqImport.Context(), "adminEmail", "finanzas@test.com"))
	reqImport.Header.Set("Content-Type", "application/json")
	rrImport := httptest.NewRecorder()
	h.ServeHTTP(rrImport, reqImport)
	if rrImport.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rrImport.Code, rrImport.Body.String())
	}

	var importResp map[string]interface{}
	if err := json.Unmarshal(rrImport.Body.Bytes(), &importResp); err != nil {
		t.Fatalf("decode import response: %v", err)
	}
	conciliacionAuto, _ := importResp["conciliacion_automatica"].(map[string]interface{})
	if conciliacionAuto == nil {
		t.Fatalf("expected conciliacion_automatica block in import response")
	}
	if int64(conciliacionAuto["conciliados"].(float64)) != 1 {
		t.Fatalf("expected conciliados=1, got %v", conciliacionAuto["conciliados"])
	}

	reqResumen := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?action=conciliacion_bancaria&empresa_id=120&periodo=2026-04", nil)
	reqResumen = reqResumen.WithContext(context.WithValue(reqResumen.Context(), "adminEmail", "finanzas@test.com"))
	rrResumen := httptest.NewRecorder()
	h.ServeHTTP(rrResumen, reqResumen)
	if rrResumen.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrResumen.Code, rrResumen.Body.String())
	}

	var resumen map[string]interface{}
	if err := json.Unmarshal(rrResumen.Body.Bytes(), &resumen); err != nil {
		t.Fatalf("decode resumen response: %v", err)
	}
	if int64(resumen["total_periodos"].(float64)) != 1 {
		t.Fatalf("expected total_periodos=1, got %v", resumen["total_periodos"])
	}
	filas, _ := resumen["filas"].([]interface{})
	if len(filas) != 1 {
		t.Fatalf("expected 1 fila de conciliacion bancaria, got %d", len(filas))
	}
	fila, _ := filas[0].(map[string]interface{})
	if int64(fila["extractos_total"].(float64)) != 2 {
		t.Fatalf("expected extractos_total=2, got %v", fila["extractos_total"])
	}
	if int64(fila["extractos_conciliados"].(float64)) != 1 {
		t.Fatalf("expected extractos_conciliados=1, got %v", fila["extractos_conciliados"])
	}
	if strings.TrimSpace(fila["estado_conciliacion"].(string)) != "con_pendientes" {
		t.Fatalf("expected estado_conciliacion=con_pendientes, got %v", fila["estado_conciliacion"])
	}

	extractos, err := dbpkg.ListEmpresaFinanzasMovimientosBancarios(dbEmp, empresaID, dbpkg.EmpresaFinanzasMovimientoBancarioFilter{PeriodoContable: "2026-04", Limit: 20})
	if err != nil {
		t.Fatalf("list extractos bancarios: %v", err)
	}
	if len(extractos) != 2 {
		t.Fatalf("expected 2 extractos, got %d", len(extractos))
	}
	matched := false
	for _, it := range extractos {
		if strings.TrimSpace(it.ReferenciaBancaria) == "TRX-H-120" {
			if it.MovimientoFinanzasID != movID {
				t.Fatalf("expected movimiento_finanzas_id=%d for matched extracto, got %d", movID, it.MovimientoFinanzasID)
			}
			matched = true
		}
	}
	if !matched {
		t.Fatalf("expected matched bank extract with referencia TRX-H-120")
	}

	reqExport := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/movimientos?action=conciliacion_bancaria_export&format=csv&empresa_id=120&periodo=2026-04", nil)
	reqExport = reqExport.WithContext(context.WithValue(reqExport.Context(), "adminEmail", "finanzas@test.com"))
	rrExport := httptest.NewRecorder()
	h.ServeHTTP(rrExport, reqExport)
	if rrExport.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrExport.Code, rrExport.Body.String())
	}
	if ct := strings.ToLower(rrExport.Header().Get("Content-Type")); !strings.Contains(ct, "text/csv") {
		t.Fatalf("expected csv content-type, got %q", rrExport.Header().Get("Content-Type"))
	}
}

func TestEmpresaFinanzasAsientosContablesHandlerProcesaPendientes(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_asientos_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaEventoContable(dbEmp, dbpkg.EmpresaEventoContable{
		EmpresaID:       77,
		Modulo:          "finanzas",
		Evento:          "movimiento_ingreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       701,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "ING-701",
		PeriodoContable: time.Now().Format("2006-01"),
		MontoTotal:      50000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"ingreso","categoria":"ventas"}`,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create evento contable: %v", err)
	}

	h := EmpresaFinanzasAsientosContablesHandler(dbEmp)
	reqProcess := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/asientos_contables?action=procesar_asientos&empresa_id=77&limit=20&max_reintentos=5", nil)
	reqProcess = reqProcess.WithContext(context.WithValue(reqProcess.Context(), "adminEmail", "conta@test.com"))
	rrProcess := httptest.NewRecorder()
	h.ServeHTTP(rrProcess, reqProcess)
	if rrProcess.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrProcess.Code, rrProcess.Body.String())
	}

	var processResp map[string]interface{}
	if err := json.Unmarshal(rrProcess.Body.Bytes(), &processResp); err != nil {
		t.Fatalf("decode process response: %v", err)
	}
	if int64(processResp["eventos_procesados"].(float64)) != 1 {
		t.Fatalf("expected eventos_procesados=1, got %v", processResp["eventos_procesados"])
	}
	if int64(processResp["asientos_creados"].(float64)) != 1 {
		t.Fatalf("expected asientos_creados=1, got %v", processResp["asientos_creados"])
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/asientos_contables?empresa_id=77&limit=20", nil)
	reqList = reqList.WithContext(context.WithValue(reqList.Context(), "adminEmail", "conta@test.com"))
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList.Code, rrList.Body.String())
	}

	var asientos []map[string]interface{}
	if err := json.Unmarshal(rrList.Body.Bytes(), &asientos); err != nil {
		t.Fatalf("decode list asientos response: %v", err)
	}
	if len(asientos) != 1 {
		t.Fatalf("expected 1 asiento in listing, got %d", len(asientos))
	}
}

func TestEmpresaFinanzasAsientosContablesHandlerValidaMaxReintentos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_asientos_handler_max_reintentos.db")
	h := EmpresaFinanzasAsientosContablesHandler(dbEmp)

	req := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/asientos_contables?action=procesar_asientos&empresa_id=77&max_reintentos=abc", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@test.com"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestEmpresaFinanzasAsientosContablesHandlerConciliacionPeriodo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_asientos_handler_conciliacion.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaEventosContablesSchema(dbEmp); err != nil {
		t.Fatalf("ensure eventos contables schema: %v", err)
	}

	empresaID := int64(91)
	if _, err := dbpkg.CreateEmpresaEventoContable(dbEmp, dbpkg.EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_ingreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       9101,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "ING-9101",
		PeriodoContable: "2026-04",
		MontoTotal:      77000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"ingreso","categoria":"ventas"}`,
		UsuarioCreador:  "tester",
	}); err != nil {
		t.Fatalf("create evento 1: %v", err)
	}
	pendienteID, err := dbpkg.CreateEmpresaEventoContable(dbEmp, dbpkg.EmpresaEventoContable{
		EmpresaID:       empresaID,
		Modulo:          "finanzas",
		Evento:          "movimiento_egreso_registrado",
		Entidad:         "finanzas_movimiento",
		EntidadID:       9102,
		DocumentoTipo:   "comprobante",
		DocumentoCodigo: "EGR-9102",
		PeriodoContable: "2026-04",
		MontoTotal:      23000,
		Moneda:          "COP",
		PayloadJSON:     `{"tipo_movimiento":"egreso","categoria":"compras"}`,
		UsuarioCreador:  "tester",
	})
	if err != nil {
		t.Fatalf("create evento 2: %v", err)
	}

	if _, err := dbpkg.ProcessEmpresaEventosContablesPendientes(dbEmp, empresaID, "tester", 1); err != nil {
		t.Fatalf("process eventos pendientes: %v", err)
	}
	if _, err := dbEmp.Exec(`UPDATE empresa_eventos_contables SET error_procesamiento = 'fallo temporal', intentos_procesamiento = 2 WHERE id = ?`, pendienteID); err != nil {
		t.Fatalf("mark pending error: %v", err)
	}

	h := EmpresaFinanzasAsientosContablesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/asientos_contables?action=conciliacion_periodo&empresa_id=91&periodo=2026-04&limit=20", nil)
	req = req.WithContext(context.WithValue(req.Context(), "adminEmail", "conta@test.com"))
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode conciliacion payload: %v", err)
	}
	if int64(payload["total_periodos"].(float64)) != 1 {
		t.Fatalf("expected total_periodos=1, got %v", payload["total_periodos"])
	}
	filas, _ := payload["filas"].([]interface{})
	if len(filas) != 1 {
		t.Fatalf("expected 1 fila de conciliacion, got %d", len(filas))
	}
	row, _ := filas[0].(map[string]interface{})
	if strings.TrimSpace(row["periodo_contable"].(string)) != "2026-04" {
		t.Fatalf("expected periodo_contable=2026-04, got %v", row["periodo_contable"])
	}
	if int64(row["eventos_pendientes"].(float64)) != 1 {
		t.Fatalf("expected eventos_pendientes=1, got %v", row["eventos_pendientes"])
	}
	if strings.TrimSpace(row["estado_conciliacion"].(string)) != "con_pendientes" {
		t.Fatalf("expected estado_conciliacion=con_pendientes, got %v", row["estado_conciliacion"])
	}
}

func TestEmpresaFinanzasCierresCajaHandler(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_finanzas_cierres_caja_handler.db")
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	h := EmpresaFinanzasCierresCajaHandler(dbEmp)

	reqCreate := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/cierres_caja", strings.NewReader(`{"empresa_id":66,"sucursal_id":3,"caja_codigo":"Caja-03","turno":"noche","apertura_monto":50000,"ingresos_efectivo":30000,"egresos_efectivo":5000,"retiros_efectivo":2000,"umbral_incidencia":1000}`))
	reqCreate = reqCreate.WithContext(context.WithValue(reqCreate.Context(), "adminEmail", "cajero@test.com"))
	reqCreate.Header.Set("Content-Type", "application/json")
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, reqCreate)
	if rrCreate.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusCreated, rrCreate.Code, rrCreate.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(rrCreate.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create cierre response: %v", err)
	}
	id := int64(createResp["id"].(float64))

	reqInvalidApprove := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cierres_caja?action=aprobar&empresa_id=66&id="+strconv.FormatInt(id, 10), nil)
	reqInvalidApprove = reqInvalidApprove.WithContext(context.WithValue(reqInvalidApprove.Context(), "adminEmail", "supervisor@test.com"))
	rrInvalidApprove := httptest.NewRecorder()
	h.ServeHTTP(rrInvalidApprove, reqInvalidApprove)
	if rrInvalidApprove.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusConflict, rrInvalidApprove.Code, rrInvalidApprove.Body.String())
	}

	reqClose := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cierres_caja?action=cerrar&empresa_id=66&id="+strconv.FormatInt(id, 10)+"&caja_fisica=70000", nil)
	reqClose = reqClose.WithContext(context.WithValue(reqClose.Context(), "adminEmail", "cajero@test.com"))
	rrClose := httptest.NewRecorder()
	h.ServeHTTP(rrClose, reqClose)
	if rrClose.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrClose.Code, rrClose.Body.String())
	}

	reqApprove := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cierres_caja?action=aprobar&empresa_id=66&id="+strconv.FormatInt(id, 10), nil)
	reqApprove = reqApprove.WithContext(context.WithValue(reqApprove.Context(), "adminEmail", "supervisor@test.com"))
	rrApprove := httptest.NewRecorder()
	h.ServeHTTP(rrApprove, reqApprove)
	if rrApprove.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrApprove.Code, rrApprove.Body.String())
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/cierres_caja?empresa_id=66&estado_cierre=aprobado", nil)
	reqList = reqList.WithContext(context.WithValue(reqList.Context(), "adminEmail", "supervisor@test.com"))
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d body=%s", http.StatusOK, rrList.Code, rrList.Body.String())
	}

	var items []map[string]interface{}
	if err := json.Unmarshal(rrList.Body.Bytes(), &items); err != nil {
		t.Fatalf("decode list cierres response: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 cierre aprobado, got %d", len(items))
	}
	if strings.ToLower(strings.TrimSpace(items[0]["estado_cierre"].(string))) != "aprobado" {
		t.Fatalf("expected estado_cierre aprobado, got %v", items[0]["estado_cierre"])
	}
}

func hasEventoContable(items []dbpkg.EmpresaEventoContable, evento string) bool {
	for _, it := range items {
		if it.Evento == evento {
			return true
		}
	}
	return false
}

func findEventoContable(items []dbpkg.EmpresaEventoContable, evento string) (dbpkg.EmpresaEventoContable, bool) {
	for _, it := range items {
		if it.Evento == evento {
			return it, true
		}
	}
	return dbpkg.EmpresaEventoContable{}, false
}
