package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestVentaCarritoFacturaYResolucionImpresora(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carrito_factura_impresion.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaImpresorasSchema(dbEmp); err != nil {
		t.Fatalf("ensure impresoras schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	seedFacturacionCumplimientoConfig(t, dbEmp, 71)

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	itemsHandler := EmpresaCarritoItemsHandler(dbEmp)
	facturacionHandler := EmpresaFacturacionElectronicaHandler(dbEmp, nil)
	impresorasHandler := EmpresaImpresorasHandler(dbEmp)
	resolverHandler := EmpresaImpresorasResolverHandler(dbEmp)

	adminCtx := context.WithValue(context.Background(), "adminEmail", "qa-ventas@test.com")

	createCarritoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":71,"nombre":"Caja QA Factura","canal_venta":"mostrador","moneda":"COP"}`))
	createCarritoReq = createCarritoReq.WithContext(adminCtx)
	createCarritoReq.Header.Set("Content-Type", "application/json")
	createCarritoRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(createCarritoRR, createCarritoReq)
	if createCarritoRR.Code != http.StatusCreated {
		t.Fatalf("expected create carrito status %d, got %d body=%s", http.StatusCreated, createCarritoRR.Code, createCarritoRR.Body.String())
	}

	var createCarritoResp map[string]interface{}
	if err := json.Unmarshal(createCarritoRR.Body.Bytes(), &createCarritoResp); err != nil {
		t.Fatalf("decode create carrito response: %v", err)
	}
	carritoID, ok := createCarritoResp["id"].(float64)
	if !ok || carritoID <= 0 {
		t.Fatalf("expected carrito id in response, got %+v", createCarritoResp)
	}

	createItemReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra/items", strings.NewReader(`{"empresa_id":71,"carrito_id":`+strconv.FormatInt(int64(carritoID), 10)+`,"descripcion":"Servicio QA","cantidad":2,"precio_unitario":1500}`))
	createItemReq = createItemReq.WithContext(adminCtx)
	createItemReq.Header.Set("Content-Type", "application/json")
	createItemRR := httptest.NewRecorder()
	itemsHandler.ServeHTTP(createItemRR, createItemReq)
	if createItemRR.Code != http.StatusCreated {
		t.Fatalf("expected create item status %d, got %d body=%s", http.StatusCreated, createItemRR.Code, createItemRR.Body.String())
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=71&id="+strconv.FormatInt(int64(carritoID), 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","total_pagado":3000}`))
	payReq = payReq.WithContext(adminCtx)
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay carrito status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}
	if !strings.Contains(payRR.Body.String(), `"estado_venta":"venta_pagada"`) {
		t.Fatalf("expected venta_pagada in pay response, got body=%s", payRR.Body.String())
	}

	createPrinterReq := httptest.NewRequest(http.MethodPost, "/api/empresa/impresoras?empresa_id=71", strings.NewReader(`{"codigo":"CAJA_FACTURA_71","nombre":"Impresora Factura QA","tipo_conexion":"red","formato_impresion":"carta","es_predeterminada":true,"estado":"activo"}`))
	createPrinterReq = createPrinterReq.WithContext(adminCtx)
	createPrinterReq.Header.Set("Content-Type", "application/json")
	createPrinterRR := httptest.NewRecorder()
	impresorasHandler.ServeHTTP(createPrinterRR, createPrinterReq)
	if createPrinterRR.Code != http.StatusOK {
		t.Fatalf("expected create printer status %d, got %d body=%s", http.StatusOK, createPrinterRR.Code, createPrinterRR.Body.String())
	}

	facturaCodigo := "FAC-CAR-71-0001"
	emitReq := httptest.NewRequest(http.MethodPut, "/api/empresa/facturacion_electronica?action=emitir", strings.NewReader(`{"empresa_id":71,"pais_codigo":"CO","documento_codigo":"`+facturaCodigo+`","estado_actual":"borrador","monto_total":3000,"moneda":"COP","periodo_contable":"2026-04","observaciones":"factura emitida desde la venta del carrito `+strconv.FormatInt(int64(carritoID), 10)+`"}`))
	emitReq = emitReq.WithContext(adminCtx)
	emitReq.Header.Set("Content-Type", "application/json")
	emitRR := httptest.NewRecorder()
	facturacionHandler.ServeHTTP(emitRR, emitReq)
	if emitRR.Code != http.StatusOK {
		t.Fatalf("expected emitir factura status %d, got %d body=%s", http.StatusOK, emitRR.Code, emitRR.Body.String())
	}
	if !strings.Contains(strings.ToLower(emitRR.Body.String()), `"estado_envio":"enviado"`) {
		t.Fatalf("expected estado_envio enviado, got body=%s", emitRR.Body.String())
	}
	if !strings.Contains(emitRR.Body.String(), `"numero_legal":"FE`) {
		t.Fatalf("expected numero_legal with FE prefix, got body=%s", emitRR.Body.String())
	}

	listDocsReq := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica?action=documentos&empresa_id=71&documento="+facturaCodigo, nil)
	listDocsReq = listDocsReq.WithContext(adminCtx)
	listDocsRR := httptest.NewRecorder()
	facturacionHandler.ServeHTTP(listDocsRR, listDocsReq)
	if listDocsRR.Code != http.StatusOK {
		t.Fatalf("expected list documentos status %d, got %d body=%s", http.StatusOK, listDocsRR.Code, listDocsRR.Body.String())
	}

	var listDocsResp struct {
		Items []dbpkg.EmpresaDocumentoFacturacion `json:"items"`
	}
	if err := json.Unmarshal(listDocsRR.Body.Bytes(), &listDocsResp); err != nil {
		t.Fatalf("decode documentos response: %v", err)
	}
	if len(listDocsResp.Items) != 1 {
		t.Fatalf("expected 1 factura document, got %d body=%s", len(listDocsResp.Items), listDocsRR.Body.String())
	}
	if listDocsResp.Items[0].DocumentoCodigo != facturaCodigo {
		t.Fatalf("expected documento_codigo %s, got %s", facturaCodigo, listDocsResp.Items[0].DocumentoCodigo)
	}
	if listDocsResp.Items[0].MontoTotal != 3000 {
		t.Fatalf("expected monto_total 3000, got %v", listDocsResp.Items[0].MontoTotal)
	}

	resolvePrinterReq := httptest.NewRequest(http.MethodGet, "/api/empresa/impresoras/resolver?empresa_id=71&funcionalidad=factura_caja", nil)
	resolvePrinterReq = resolvePrinterReq.WithContext(adminCtx)
	resolvePrinterRR := httptest.NewRecorder()
	resolverHandler.ServeHTTP(resolvePrinterRR, resolvePrinterReq)
	if resolvePrinterRR.Code != http.StatusOK {
		t.Fatalf("expected resolver printer status %d, got %d body=%s", http.StatusOK, resolvePrinterRR.Code, resolvePrinterRR.Body.String())
	}
	if !strings.Contains(resolvePrinterRR.Body.String(), `"ok":true`) {
		t.Fatalf("expected printer resolver ok=true, got body=%s", resolvePrinterRR.Body.String())
	}
	if !strings.Contains(resolvePrinterRR.Body.String(), `"nombre":"Impresora Factura QA"`) {
		t.Fatalf("expected printer name in resolver response, got body=%s", resolvePrinterRR.Body.String())
	}
	if !strings.Contains(resolvePrinterRR.Body.String(), `"fuente":"predeterminada"`) {
		t.Fatalf("expected predeterminada source for factura printer, got body=%s", resolvePrinterRR.Body.String())
	}
}