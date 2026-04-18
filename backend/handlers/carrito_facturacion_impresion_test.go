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
	if _, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, dbpkg.EmpresaConfiguracionAvanzada{
		EmpresaID:          71,
		ModoDocumentoVenta: "factura_electronica",
		UsuarioCreador:     "facturacion@test.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("upsert modo_documento_venta factura_electronica: %v", err)
	}

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	itemsHandler := EmpresaCarritoItemsHandler(dbEmp)
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
	if !strings.Contains(payRR.Body.String(), `"tipo_documento":"factura_electronica"`) {
		t.Fatalf("expected auto factura_electronica document in pay response, got body=%s", payRR.Body.String())
	}
	if !strings.Contains(payRR.Body.String(), `"numero_legal":"FE-`) {
		t.Fatalf("expected auto numero_legal with FE prefix in pay response, got body=%s", payRR.Body.String())
	}

	docAuto, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{EmpresaID: 71, TipoDocumento: "factura_electronica"})
	if err != nil {
		t.Fatalf("list auto factura documents: %v", err)
	}
	if len(docAuto) != 1 {
		t.Fatalf("expected 1 auto factura document, got %d", len(docAuto))
	}
	if docAuto[0].EstadoDocumento != "emitida" {
		t.Fatalf("expected auto factura emitida, got %s", docAuto[0].EstadoDocumento)
	}

	createPrinterReq := httptest.NewRequest(http.MethodPost, "/api/empresa/impresoras?empresa_id=71", strings.NewReader(`{"codigo":"CAJA_FACTURA_71","nombre":"Impresora Factura QA","tipo_conexion":"red","formato_impresion":"carta","es_predeterminada":true,"estado":"activo"}`))
	createPrinterReq = createPrinterReq.WithContext(adminCtx)
	createPrinterReq.Header.Set("Content-Type", "application/json")
	createPrinterRR := httptest.NewRecorder()
	impresorasHandler.ServeHTTP(createPrinterRR, createPrinterReq)
	if createPrinterRR.Code != http.StatusOK {
		t.Fatalf("expected create printer status %d, got %d body=%s", http.StatusOK, createPrinterRR.Code, createPrinterRR.Body.String())
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

func TestVentaCarritoGeneraComprobantePagoSegunConfiguracion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_carrito_comprobante_pago.db")
	ensureClientesSchema(t, dbEmp)
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaFacturacionElectronicaSchema(dbEmp); err != nil {
		t.Fatalf("ensure facturacion schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaConfiguracionAvanzadaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion avanzada schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("ensure documentos transaccionales schema: %v", err)
	}
	if _, err := dbpkg.UpsertEmpresaConfiguracionAvanzada(dbEmp, dbpkg.EmpresaConfiguracionAvanzada{
		EmpresaID:          72,
		ModoDocumentoVenta: "comprobante_pago",
		UsuarioCreador:     "qa-ventas@test.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("upsert modo_documento_venta comprobante_pago: %v", err)
	}

	carritosHandler := EmpresaCarritosCompraHandler(dbEmp)
	itemsHandler := EmpresaCarritoItemsHandler(dbEmp)
	adminCtx := context.WithValue(context.Background(), "adminEmail", "qa-ventas@test.com")

	createCarritoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra", strings.NewReader(`{"empresa_id":72,"nombre":"Caja QA Comprobante","canal_venta":"mostrador","moneda":"COP"}`))
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

	createItemReq := httptest.NewRequest(http.MethodPost, "/api/empresa/carritos_compra/items", strings.NewReader(`{"empresa_id":72,"carrito_id":`+strconv.FormatInt(int64(carritoID), 10)+`,"descripcion":"Servicio QA","cantidad":1,"precio_unitario":4200}`))
	createItemReq = createItemReq.WithContext(adminCtx)
	createItemReq.Header.Set("Content-Type", "application/json")
	createItemRR := httptest.NewRecorder()
	itemsHandler.ServeHTTP(createItemRR, createItemReq)
	if createItemRR.Code != http.StatusCreated {
		t.Fatalf("expected create item status %d, got %d body=%s", http.StatusCreated, createItemRR.Code, createItemRR.Body.String())
	}

	payReq := httptest.NewRequest(http.MethodPut, "/api/empresa/carritos_compra?empresa_id=72&id="+strconv.FormatInt(int64(carritoID), 10)+"&action=pagar_estacion", strings.NewReader(`{"metodo_pago":"efectivo","total_pagado":4200}`))
	payReq = payReq.WithContext(adminCtx)
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	carritosHandler.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusOK {
		t.Fatalf("expected pay carrito status %d, got %d body=%s", http.StatusOK, payRR.Code, payRR.Body.String())
	}
	if !strings.Contains(payRR.Body.String(), `"tipo_documento":"comprobante_pago"`) {
		t.Fatalf("expected auto comprobante_pago document in pay response, got body=%s", payRR.Body.String())
	}

	docs, err := dbpkg.ListEmpresaDocumentosFacturacionByEmpresa(dbEmp, dbpkg.EmpresaDocumentoFacturacionListFilter{EmpresaID: 72, TipoDocumento: "comprobante_pago"})
	if err != nil {
		t.Fatalf("list comprobante documents: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 comprobante document, got %d", len(docs))
	}
	if docs[0].EstadoDocumento != "emitida" {
		t.Fatalf("expected comprobante emitido, got %s", docs[0].EstadoDocumento)
	}
	if docs[0].NumeroLegal == "" {
		t.Fatalf("expected comprobante numero_legal, got empty record %+v", docs[0])
	}
	if docs[0].CodigoValidacion != "" {
		t.Fatalf("expected comprobante without codigo_validacion, got %s", docs[0].CodigoValidacion)
	}
}
