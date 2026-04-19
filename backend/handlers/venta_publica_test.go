package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaVentaPublicaHandlerConfigCatalogoYToggle(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_venta_publica_handler.db")
	if err := dbpkg.EnsureEmpresaVentaPublicaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaVentaPublicaSchema: %v", err)
	}

	h := EmpresaVentaPublicaHandler(dbEmp)

	configReq := httptest.NewRequest(http.MethodPost, "/api/empresa/venta_publica?empresa_id=137&action=config", strings.NewReader(`{
		"empresa_id":137,
		"empresa_slug":"Tienda 137",
		"nombre_tienda":"Tienda Principal 137",
		"tema_visual":"moderno",
		"moneda":"cop",
		"mostrar_stock":true,
		"wompi_activo":false,
		"epayco_activo":true,
		"epayco_mode":"production",
		"epayco_public_key":"env:EPAYCO_PUBLIC_KEY_EMPRESA_137",
		"epayco_private_key_ref":"env:EPAYCO_PRIVATE_KEY_EMPRESA_137"
	}`))
	configReq.Header.Set("Content-Type", "application/json")
	configRR := httptest.NewRecorder()
	h.ServeHTTP(configRR, configReq)
	if configRR.Code != http.StatusOK {
		t.Fatalf("config status=%d body=%s", configRR.Code, configRR.Body.String())
	}

	var configResp struct {
		Config     dbpkg.EmpresaVentaPublicaConfig `json:"config"`
		PublicPath string                          `json:"public_path"`
	}
	if err := json.Unmarshal(configRR.Body.Bytes(), &configResp); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if configResp.Config.EmpresaSlug != "tienda-137" {
		t.Fatalf("expected normalized slug tienda-137, got=%q", configResp.Config.EmpresaSlug)
	}
	if !configResp.Config.EpaycoActivo {
		t.Fatalf("expected epayco active in config response")
	}
	if configResp.Config.EpaycoMode != "production" {
		t.Fatalf("expected epayco mode production, got=%q", configResp.Config.EpaycoMode)
	}
	if configResp.Config.TemaVisual != "moderno" {
		t.Fatalf("expected tema_visual moderno, got=%q", configResp.Config.TemaVisual)
	}
	if !strings.Contains(configResp.PublicPath, "/tienda-137/venta_publica.html") {
		t.Fatalf("unexpected public path: %q", configResp.PublicPath)
	}

	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/venta_publica?empresa_id=137&action=crear", strings.NewReader(`{
		"empresa_id":137,
		"codigo_publico":"SKU-VP-001",
		"nombre":"Hamburguesa Especial",
		"descripcion":"Con papas y bebida",
		"precio":28000,
		"moneda":"COP",
		"stock_publicado":14,
		"orden_visual":1,
		"destacado":true
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create item status=%d body=%s", createRR.Code, createRR.Body.String())
	}

	var createResp struct {
		Item dbpkg.EmpresaVentaPublicaItem `json:"item"`
	}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Item.ID <= 0 {
		t.Fatalf("expected created item id > 0, got %+v", createResp.Item)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/venta_publica?empresa_id=137&action=catalogo", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list status=%d body=%s", listRR.Code, listRR.Body.String())
	}
	var listResp struct {
		Total int64                           `json:"total"`
		Rows  []dbpkg.EmpresaVentaPublicaItem `json:"rows"`
	}
	if err := json.Unmarshal(listRR.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Rows) != 1 {
		t.Fatalf("expected one active item, total=%d len=%d", listResp.Total, len(listResp.Rows))
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/empresa/venta_publica?empresa_id=137&action=detalle&id="+itoa64(createResp.Item.ID), nil)
	detailRR := httptest.NewRecorder()
	h.ServeHTTP(detailRR, detailReq)
	if detailRR.Code != http.StatusOK {
		t.Fatalf("detail status=%d body=%s", detailRR.Code, detailRR.Body.String())
	}

	deactivateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/venta_publica?empresa_id=137&action=desactivar&id="+itoa64(createResp.Item.ID), nil)
	deactivateRR := httptest.NewRecorder()
	h.ServeHTTP(deactivateRR, deactivateReq)
	if deactivateRR.Code != http.StatusOK {
		t.Fatalf("deactivate status=%d body=%s", deactivateRR.Code, deactivateRR.Body.String())
	}

	listActiveReq := httptest.NewRequest(http.MethodGet, "/api/empresa/venta_publica?empresa_id=137&action=catalogo", nil)
	listActiveRR := httptest.NewRecorder()
	h.ServeHTTP(listActiveRR, listActiveReq)
	if listActiveRR.Code != http.StatusOK {
		t.Fatalf("list active status=%d body=%s", listActiveRR.Code, listActiveRR.Body.String())
	}
	var listActiveResp struct {
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(listActiveRR.Body.Bytes(), &listActiveResp); err != nil {
		t.Fatalf("decode list active response: %v", err)
	}
	if listActiveResp.Total != 0 {
		t.Fatalf("expected active total=0 after deactivate, got=%d", listActiveResp.Total)
	}

	listAllReq := httptest.NewRequest(http.MethodGet, "/api/empresa/venta_publica?empresa_id=137&action=catalogo&include_inactive=1", nil)
	listAllRR := httptest.NewRecorder()
	h.ServeHTTP(listAllRR, listAllReq)
	if listAllRR.Code != http.StatusOK {
		t.Fatalf("list include inactive status=%d body=%s", listAllRR.Code, listAllRR.Body.String())
	}
	var listAllResp struct {
		Total int64                           `json:"total"`
		Rows  []dbpkg.EmpresaVentaPublicaItem `json:"rows"`
	}
	if err := json.Unmarshal(listAllRR.Body.Bytes(), &listAllResp); err != nil {
		t.Fatalf("decode list include inactive response: %v", err)
	}
	if listAllResp.Total != 1 || len(listAllResp.Rows) != 1 {
		t.Fatalf("expected one row when include_inactive=1, total=%d len=%d", listAllResp.Total, len(listAllResp.Rows))
	}
	if listAllResp.Rows[0].Estado != "inactivo" {
		t.Fatalf("expected inactivo state, got=%q", listAllResp.Rows[0].Estado)
	}

	activateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/venta_publica?empresa_id=137&action=activar&id="+itoa64(createResp.Item.ID), nil)
	activateRR := httptest.NewRecorder()
	h.ServeHTTP(activateRR, activateReq)
	if activateRR.Code != http.StatusOK {
		t.Fatalf("activate status=%d body=%s", activateRR.Code, activateRR.Body.String())
	}

	listAfterActivateReq := httptest.NewRequest(http.MethodGet, "/api/empresa/venta_publica?empresa_id=137&action=catalogo", nil)
	listAfterActivateRR := httptest.NewRecorder()
	h.ServeHTTP(listAfterActivateRR, listAfterActivateReq)
	if listAfterActivateRR.Code != http.StatusOK {
		t.Fatalf("list after activate status=%d body=%s", listAfterActivateRR.Code, listAfterActivateRR.Body.String())
	}
	var listAfterActivateResp struct {
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(listAfterActivateRR.Body.Bytes(), &listAfterActivateResp); err != nil {
		t.Fatalf("decode list after activate response: %v", err)
	}
	if listAfterActivateResp.Total != 1 {
		t.Fatalf("expected active total=1 after activate, got=%d", listAfterActivateResp.Total)
	}
}

func TestPublicVentaPublicaHandlerCatalogoYPagoConWompiInactivo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_venta_publica_public_handler.db")
	if err := dbpkg.EnsureEmpresaVentaPublicaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaVentaPublicaSchema: %v", err)
	}

	const empresaID int64 = 202
	if _, err := dbpkg.UpsertEmpresaVentaPublicaConfig(dbEmp, dbpkg.EmpresaVentaPublicaConfig{
		EmpresaID:    empresaID,
		EmpresaSlug:  "restaurante-central",
		NombreTienda: "Restaurante Central",
		TemaVisual:   "minimal",
		Moneda:       "COP",
		MostrarStock: true,
		WompiActivo:  false,
		WompiMode:    "sandbox",
		Estado:       "activo",
	}); err != nil {
		t.Fatalf("UpsertEmpresaVentaPublicaConfig: %v", err)
	}

	itemID, err := dbpkg.CreateEmpresaVentaPublicaItem(dbEmp, dbpkg.EmpresaVentaPublicaItem{
		EmpresaID:      empresaID,
		CodigoPublico:  "VP-202-001",
		Nombre:         "Bandeja Ejecutiva",
		Descripcion:    "Almuerzo del dia",
		Precio:         18000,
		Moneda:         "COP",
		StockPublicado: 25,
		Estado:         "activo",
	})
	if err != nil {
		t.Fatalf("CreateEmpresaVentaPublicaItem: %v", err)
	}
	if itemID <= 0 {
		t.Fatalf("expected item id > 0, got=%d", itemID)
	}

	h := PublicVentaPublicaHandler(dbEmp)

	catalogReq := httptest.NewRequest(http.MethodGet, "/api/public/venta_publica?action=catalogo&empresa_slug=restaurante-central", nil)
	catalogRR := httptest.NewRecorder()
	h.ServeHTTP(catalogRR, catalogReq)
	if catalogRR.Code != http.StatusOK {
		t.Fatalf("catalog status=%d body=%s", catalogRR.Code, catalogRR.Body.String())
	}
	var catalogResp struct {
		EmpresaID   int64                           `json:"empresa_id"`
		EmpresaSlug string                          `json:"empresa_slug"`
		Tienda      map[string]interface{}          `json:"tienda"`
		Items       []dbpkg.EmpresaVentaPublicaItem `json:"items"`
	}
	if err := json.Unmarshal(catalogRR.Body.Bytes(), &catalogResp); err != nil {
		t.Fatalf("decode catalog response: %v", err)
	}
	if catalogResp.EmpresaID != empresaID {
		t.Fatalf("expected empresa_id=%d got=%d", empresaID, catalogResp.EmpresaID)
	}
	if catalogResp.EmpresaSlug != "restaurante-central" {
		t.Fatalf("expected empresa_slug=restaurante-central got=%q", catalogResp.EmpresaSlug)
	}
	if got := strings.TrimSpace(anyString(catalogResp.Tienda["tema_visual"])); got != "minimal" {
		t.Fatalf("expected public tema_visual minimal, got=%q", got)
	}
	if len(catalogResp.Items) != 1 {
		t.Fatalf("expected one public item, got=%d", len(catalogResp.Items))
	}

	payReq := httptest.NewRequest(http.MethodPost, "/api/public/venta_publica?action=crear_pago", strings.NewReader(`{
		"empresa_slug":"restaurante-central",
		"metodo_pago":"wompi_nequi",
		"comprador_nombre":"Cliente Publico",
		"comprador_email":"cliente.publico@test.com",
		"comprador_telefono":"3001234567",
		"accept_terms":true,
		"items":[{"item_id":`+itoa64(itemID)+`,"cantidad":2}]
	}`))
	payReq.Header.Set("Content-Type", "application/json")
	payRR := httptest.NewRecorder()
	h.ServeHTTP(payRR, payReq)
	if payRR.Code != http.StatusPreconditionFailed {
		t.Fatalf("expected status=%d got=%d body=%s", http.StatusPreconditionFailed, payRR.Code, payRR.Body.String())
	}
	var payResp struct {
		OrderID   int64  `json:"order_id"`
		OrderCode string `json:"order_code"`
		Error     string `json:"error"`
	}
	if err := json.Unmarshal(payRR.Body.Bytes(), &payResp); err != nil {
		t.Fatalf("decode pay response: %v", err)
	}
	if payResp.OrderID <= 0 {
		t.Fatalf("expected order_id > 0, got=%d", payResp.OrderID)
	}
	if strings.TrimSpace(payResp.OrderCode) == "" {
		t.Fatalf("expected order_code in response, got=%+v", payResp)
	}

	storedOrder, err := dbpkg.GetEmpresaVentaPublicaOrderByCodigo(dbEmp, empresaID, payResp.OrderCode)
	if err != nil {
		t.Fatalf("GetEmpresaVentaPublicaOrderByCodigo: %v", err)
	}
	if storedOrder.Total != 36000 {
		t.Fatalf("expected total 36000, got=%.2f", storedOrder.Total)
	}
	if storedOrder.EstadoPago != "pendiente" {
		t.Fatalf("expected estado_pago pendiente, got=%q", storedOrder.EstadoPago)
	}

	statusReq := httptest.NewRequest(http.MethodGet, "/api/public/venta_publica?action=estado_pago&empresa_slug=restaurante-central&order_code="+payResp.OrderCode, nil)
	statusRR := httptest.NewRecorder()
	h.ServeHTTP(statusRR, statusReq)
	if statusRR.Code != http.StatusOK {
		t.Fatalf("status query code=%d body=%s", statusRR.Code, statusRR.Body.String())
	}
	var statusResp struct {
		Order dbpkg.EmpresaVentaPublicaOrder `json:"order"`
	}
	if err := json.Unmarshal(statusRR.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("decode status response: %v", err)
	}
	if statusResp.Order.CodigoOrden != payResp.OrderCode {
		t.Fatalf("expected same order code, got=%q want=%q", statusResp.Order.CodigoOrden, payResp.OrderCode)
	}
}

func TestPublicVentaPublicaHandlerEstadoPagoRequiereOrderCode(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_venta_publica_public_handler_bad_request.db")
	if err := dbpkg.EnsureEmpresaVentaPublicaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaVentaPublicaSchema: %v", err)
	}
	if _, err := dbpkg.UpsertEmpresaVentaPublicaConfig(dbEmp, dbpkg.EmpresaVentaPublicaConfig{
		EmpresaID:    333,
		EmpresaSlug:  "hotel-sur",
		NombreTienda: "Hotel Sur",
		Moneda:       "COP",
		WompiActivo:  false,
		WompiMode:    "sandbox",
		Estado:       "activo",
	}); err != nil {
		t.Fatalf("UpsertEmpresaVentaPublicaConfig: %v", err)
	}

	h := PublicVentaPublicaHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/public/venta_publica?action=estado_pago&empresa_slug=hotel-sur", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status=%d got=%d body=%s", http.StatusBadRequest, rr.Code, rr.Body.String())
	}
}

func TestPublicVentaPublicaHandlerCatalogoWithLegacySchemaMissingColumns(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_venta_publica_public_legacy_schema.db")
	if _, err := dbEmp.Exec(`CREATE TABLE empresas (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nombre TEXT,
		estado TEXT
	)`); err != nil {
		t.Fatalf("create empresas table: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO empresas (id, nombre, estado) VALUES (404, 'Hotel Legacy', 'activo')`); err != nil {
		t.Fatalf("insert empresa: %v", err)
	}
	if _, err := dbEmp.Exec(`CREATE TABLE empresa_venta_publica_configuracion (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL UNIQUE,
		empresa_slug TEXT NOT NULL,
		nombre_tienda TEXT,
		moneda TEXT DEFAULT 'COP',
		mostrar_stock INTEGER DEFAULT 1,
		wompi_activo INTEGER DEFAULT 0,
		wompi_mode TEXT DEFAULT 'sandbox',
		estado TEXT DEFAULT 'activo'
	)`); err != nil {
		t.Fatalf("create legacy config table: %v", err)
	}
	if _, err := dbEmp.Exec(`CREATE TABLE empresa_venta_publica_items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		codigo_publico TEXT NOT NULL,
		nombre TEXT NOT NULL,
		descripcion TEXT,
		precio REAL DEFAULT 0,
		moneda TEXT DEFAULT 'COP',
		fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
	)`); err != nil {
		t.Fatalf("create legacy items table: %v", err)
	}
	if _, err := dbEmp.Exec(`CREATE TABLE empresa_venta_publica_ordenes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		empresa_id INTEGER NOT NULL,
		codigo_orden TEXT NOT NULL,
		total REAL DEFAULT 0,
		metodo_pago TEXT DEFAULT 'wompi_nequi',
		estado_pago TEXT DEFAULT 'pendiente',
		fecha_creacion TEXT DEFAULT (datetime('now','localtime'))
	)`); err != nil {
		t.Fatalf("create legacy orders table: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO empresa_venta_publica_configuracion (empresa_id, empresa_slug, nombre_tienda, moneda, mostrar_stock, wompi_activo, wompi_mode, estado)
		VALUES (404, 'hotel-legacy', 'Hotel Legacy', 'COP', 1, 0, 'sandbox', 'activo')`); err != nil {
		t.Fatalf("insert legacy config: %v", err)
	}
	if _, err := dbEmp.Exec(`INSERT INTO empresa_venta_publica_items (empresa_id, codigo_publico, nombre, descripcion, precio, moneda, fecha_creacion)
		VALUES (404, 'LEG-1', 'Habitacion Legacy', 'Tarifa nocturna', 95000, 'COP', datetime('now','localtime'))`); err != nil {
		t.Fatalf("insert legacy item: %v", err)
	}

	if err := dbpkg.EnsureEmpresaVentaPublicaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaVentaPublicaSchema legacy repair: %v", err)
	}

	h := PublicVentaPublicaHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/public/venta_publica?action=catalogo&empresa_slug=hotel-legacy", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("catalog legacy status=%d body=%s", rr.Code, rr.Body.String())
	}

	var resp struct {
		EmpresaID   int64                           `json:"empresa_id"`
		EmpresaSlug string                          `json:"empresa_slug"`
		Items       []dbpkg.EmpresaVentaPublicaItem `json:"items"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode legacy catalog response: %v", err)
	}
	if resp.EmpresaID != 404 {
		t.Fatalf("expected empresa_id=404, got=%d", resp.EmpresaID)
	}
	if resp.EmpresaSlug != "hotel-legacy" {
		t.Fatalf("expected empresa_slug=hotel-legacy, got=%q", resp.EmpresaSlug)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected one legacy item after schema repair, got=%d", len(resp.Items))
	}
	if resp.Items[0].Estado != "activo" {
		t.Fatalf("expected repaired item estado activo, got=%q", resp.Items[0].Estado)
	}
}

func TestResolveVentaPublicaSlugFromHost(t *testing.T) {
	t.Setenv("VENTA_PUBLICA_BASE_DOMAINS", "powerfulcontrolsystem.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "empresa1.powerfulcontrolsystem.com"
	if got := ResolveVentaPublicaSlugFromHost(req); got != "empresa1" {
		t.Fatalf("expected empresa1 got=%q", got)
	}

	reqApex := httptest.NewRequest(http.MethodGet, "/", nil)
	reqApex.Host = "powerfulcontrolsystem.com"
	if got := ResolveVentaPublicaSlugFromHost(reqApex); got != "" {
		t.Fatalf("expected empty slug for apex host, got=%q", got)
	}

	reqForwarded := httptest.NewRequest(http.MethodGet, "/", nil)
	reqForwarded.Host = "127.0.0.1:8080"
	reqForwarded.Header.Set("X-Forwarded-Host", "empresa2.powerfulcontrolsystem.com")
	if got := ResolveVentaPublicaSlugFromHost(reqForwarded); got != "empresa2" {
		t.Fatalf("expected empresa2 from forwarded host, got=%q", got)
	}

	reqNested := httptest.NewRequest(http.MethodGet, "/", nil)
	reqNested.Host = "a.b.powerfulcontrolsystem.com"
	if got := ResolveVentaPublicaSlugFromHost(reqNested); got != "" {
		t.Fatalf("expected empty slug for nested subdomain, got=%q", got)
	}
}

func TestVentaPublicaSlugFromRequestFallsBackToSubdomainHost(t *testing.T) {
	t.Setenv("VENTA_PUBLICA_BASE_DOMAINS", "powerfulcontrolsystem.com")
	req := httptest.NewRequest(http.MethodGet, "/api/public/venta_publica?action=catalogo", nil)
	req.Host = "empresa3.powerfulcontrolsystem.com"

	if got := ventaPublicaSlugFromRequest(req); got != "empresa3" {
		t.Fatalf("expected empresa3, got=%q", got)
	}
}

func anyString(value interface{}) string {
	if value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return ""
}
