package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestEmpresaInventoryIdempotentMutationRequiresKeyBeforeDatabase(t *testing.T) {
	called := false
	handler := EmpresaInventoryIdempotentMutation(nil, "adjustment", func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/ajustar?empresa_id=12", strings.NewReader(`{"empresa_id":12}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusBadRequest)
	}
	if called {
		t.Fatal("inventory mutation executed without Idempotency-Key")
	}
	if !strings.Contains(rec.Body.String(), "Idempotency-Key") {
		t.Fatalf("missing public idempotency error: %q", rec.Body.String())
	}
}

func TestProductosCatalogDuplicateReturnsConflictWithoutDatabaseDetails(t *testing.T) {
	tests := []struct {
		entity  string
		detail  string
		message string
	}{
		{"categoria", `pq: duplicate key value violates unique constraint "ux_categorias_productos_empresa_codigo"`, "Ya existe una categoria"},
		{"producto", `UNIQUE constraint failed: productos.empresa_id, productos.sku`, "Ya existe un producto"},
		{"proveedor", `pq: duplicate key value violates unique constraint "ux_proveedores_empresa_nombre"`, "Ya existe un proveedor"},
		{"servicio", `UNIQUE constraint failed: servicios.empresa_id, servicios.codigo`, "Ya existe un servicio"},
	}
	for _, tc := range tests {
		t.Run(tc.entity, func(t *testing.T) {
			rec := httptest.NewRecorder()
			if !writeProductosCatalogDuplicate(rec, tc.entity, errors.New(tc.detail)) {
				t.Fatal("duplicate was not classified")
			}
			if rec.Code != http.StatusConflict {
				t.Fatalf("status=%d, want %d", rec.Code, http.StatusConflict)
			}
			if !strings.Contains(rec.Body.String(), tc.message) {
				t.Fatalf("public message=%q", rec.Body.String())
			}
			if strings.Contains(strings.ToLower(rec.Body.String()), "constraint") || strings.Contains(strings.ToLower(rec.Body.String()), "pq:") {
				t.Fatalf("database detail leaked: %q", rec.Body.String())
			}
		})
	}

	rec := httptest.NewRecorder()
	if writeProductosCatalogDuplicate(rec, "categoria", errors.New("database unavailable")) {
		t.Fatal("non-duplicate error was classified as duplicate")
	}
}

func TestBodegaNoEncontradaUsa404Seguro(t *testing.T) {
	rec := httptest.NewRecorder()
	if !writeBodegaNoEncontrada(rec, sql.ErrNoRows) {
		t.Fatal("sql.ErrNoRows was not classified")
	}
	if rec.Code != http.StatusNotFound || !strings.Contains(rec.Body.String(), "bodega no encontrada") {
		t.Fatalf("response=%d %q", rec.Code, rec.Body.String())
	}
	if writeBodegaNoEncontrada(httptest.NewRecorder(), errors.New("database unavailable")) {
		t.Fatal("internal error was classified as not found")
	}
}

func TestEmpresaInventoryIdempotentMutationAllowsReadsWithoutKey(t *testing.T) {
	called := false
	handler := EmpresaInventoryIdempotentMutation(nil, "advanced", func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario_avanzado?empresa_id=12", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d, want %d", rec.Code, http.StatusNoContent)
	}
	if !called {
		t.Fatal("inventory read was not delegated")
	}
}

func TestEmpresaPurchasesIdempotentMutationRequiresKeyBeforeDatabase(t *testing.T) {
	called := false
	handler := EmpresaPurchasesIdempotentMutation(nil, "advanced", func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/compras_avanzadas?empresa_id=12", strings.NewReader(`{"action":"recepcion","empresa_id":12}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest || called {
		t.Fatalf("status=%d called=%v; mutation must stop before database", rec.Code, called)
	}
	if !strings.Contains(rec.Body.String(), "Idempotency-Key") {
		t.Fatalf("missing public idempotency error: %q", rec.Body.String())
	}
}

func TestEmpresaPrinterQueueIdempotentMutationRequiresKeyBeforeDatabase(t *testing.T) {
	called := false
	handler := EmpresaPrinterQueueIdempotentMutation(nil, func(w http.ResponseWriter, r *http.Request) {
		called = true
	})
	req := httptest.NewRequest(http.MethodPut, "/api/empresa/impresoras?empresa_id=12&action=cola_trabajo", strings.NewReader(`{"empresa_id":12,"contenido":"prueba"}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest || called {
		t.Fatalf("status=%d called=%v; queue creation must stop before database", rec.Code, called)
	}
	if !strings.Contains(rec.Body.String(), "Idempotency-Key") {
		t.Fatalf("missing public idempotency error: %q", rec.Body.String())
	}
}

func TestEmpresaPrinterQueueIdempotentMutationLeavesConfigurationUpsertsCompatible(t *testing.T) {
	called := false
	handler := EmpresaPrinterQueueIdempotentMutation(nil, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusNoContent)
	})
	req := httptest.NewRequest(http.MethodPut, "/api/empresa/impresoras?empresa_id=12&action=funcionalidad", strings.NewReader(`{"empresa_id":12}`))
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent || !called {
		t.Fatalf("status=%d called=%v; configuration upsert must remain compatible", rec.Code, called)
	}
}

func TestInventoryFrontendProductionContracts(t *testing.T) {
	pagePath := filepath.Join("..", "..", "web", "administrar_empresa", "administrar_productos.html")
	rawPage, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatalf("read %s: %v", pagePath, err)
	}
	page := string(rawPage)
	for _, required := range []string{
		"inventarioMutationHeaders('cycle-count'",
		"inventarioMutationHeaders('transfer'",
		"inventarioMutationHeaders('adjustment'",
		"inventarioMutationHeaders('product-change'",
		"/api/empresa/inventario/conteo_ciclico?empresa_id=",
		"/api/empresa/inventario/transferir?empresa_id=",
		"/api/empresa/inventario/ajustar?empresa_id=",
		"/api/empresa/inventario/cambiar_producto?empresa_id=",
		"ensureInventarioMutationReference('conteoReferencia', 'CNT-UI')",
		"ensureInventarioMutationReference('transferReferencia', 'TR-UI')",
		"ensureInventarioMutationReference('ajusteReferencia', 'AJ-UI')",
		"ensureInventarioMutationReference('cambioReferencia', 'CAM-UI')",
		"created.resultado && created.resultado.variacion",
		"await loadInventarioResumen();",
		"function publicAPIError(",
		`accept=".png,.jpg,.jpeg,.gif,.webp"`,
	} {
		if !strings.Contains(page, required) {
			t.Errorf("administrar_productos.html missing production contract %q", required)
		}
	}
	if strings.Count(page, "<section") != strings.Count(page, "</section>") {
		t.Fatalf("unbalanced section markup: open=%d close=%d", strings.Count(page, "<section"), strings.Count(page, "</section>"))
	}

	advancedPath := filepath.Join("..", "..", "web", "js", "inventario_avanzado.js")
	rawAdvanced, err := os.ReadFile(advancedPath)
	if err != nil {
		t.Fatalf("read %s: %v", advancedPath, err)
	}
	advanced := string(rawAdvanced)
	for _, required := range []string{
		"Idempotency-Key",
		"function readJSONResponse(",
		`typeof data.error === "string"`,
		`function loadProductos(){`,
		`function loadBodegas(){`,
		`function loadLotes(){`,
		`function loadSeriales(){`,
		`Promise.all([dashboard, loadProductos(), loadBodegas(), loadLotes(), loadSeriales(), loadReservas()])`,
	} {
		if !strings.Contains(advanced, required) {
			t.Errorf("inventario_avanzado.js missing production contract %q", required)
		}
	}
	advancedPagePath := filepath.Join("..", "..", "web", "administrar_empresa", "inventario_avanzado.html")
	rawAdvancedPage, err := os.ReadFile(advancedPagePath)
	if err != nil {
		t.Fatalf("read %s: %v", advancedPagePath, err)
	}
	advancedPage := string(rawAdvancedPage)
	for _, required := range []string{
		`class="form-input iav-product-select"`,
		`class="form-input iav-bodega-select"`,
		`class="form-input iav-lote-select"`,
	} {
		if !strings.Contains(advancedPage, required) {
			t.Errorf("inventario_avanzado.html missing production selector %q", required)
		}
	}
	if strings.Contains(advancedPage, `id="btnSeed"`) || strings.Contains(advanced, `seed_demo`) {
		t.Error("production inventory UI must not expose demo-data creation")
	}
	handlerRaw, err := os.ReadFile("inventario_avanzado.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(handlerRaw), `case "seed_demo"`) {
		t.Error("production inventory API must not expose demo-data creation")
	}
}

func TestInventoryCriticalRoutesRequireIdempotency(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	mainSource := string(raw)
	for _, operation := range []string{`"cycle_count"`, `"transfer"`, `"adjustment"`, `"product_change"`, `"advanced"`} {
		if !strings.Contains(mainSource, "EmpresaInventoryIdempotentMutation(dbEmpresas, "+operation) {
			t.Errorf("main.go missing inventory idempotency wrapper for %s", operation)
		}
	}
	if !strings.Contains(mainSource, `EmpresaPurchasesIdempotentMutation(dbEmpresas, "advanced"`) {
		t.Error("main.go missing durable idempotency for advanced purchases")
	}
	if !strings.Contains(mainSource, `EmpresaPrinterQueueIdempotentMutation(dbEmpresas`) {
		t.Error("main.go missing durable idempotency for printer queue creation")
	}
}

func TestPrinterQueueFrontendProductionContracts(t *testing.T) {
	pagePath := filepath.Join("..", "..", "web", "administrar_empresa", "configuracion_impresora.html")
	rawPage, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	page := string(rawPage)
	for _, required := range []string{
		`function readPrinterResponse(res)`,
		`function ensurePrinterQueueMutationKey()`,
		`'Idempotency-Key': ensurePrinterQueueMutationKey()`,
		`queueMutationInFlight`,
		`submitButton.setAttribute('aria-busy', 'true')`,
	} {
		if !strings.Contains(page, required) {
			t.Errorf("configuracion_impresora.html missing production contract %q", required)
		}
	}

	handlerPath := filepath.Join("empresa_impresoras.go")
	rawHandler, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rawHandler), `http.Error(w, err.Error(), http.StatusBadRequest)`) {
		t.Fatal("printer handler must not expose raw internal errors")
	}

	cartPath := filepath.Join("..", "..", "web", "administrar_empresa", "carrito_de_compras.html")
	rawCart, err := os.ReadFile(cartPath)
	if err != nil {
		t.Fatal(err)
	}
	cart := string(rawCart)
	if strings.Count(cart, `/api/empresa/impresoras/resolver?`) != 1 {
		t.Fatal("cart must centralize printer resolution in one request helper")
	}
	for _, required := range []string{
		`async function resolvePrinterOperation(funcionalidad, itemPayload)`,
		`resolvePrinterOperation('orden_servicio', itemPayload)`,
		`resolvePrinterOperation('ticket_cobro')`,
		`resolvePrinterOperation(drawer.printerFunction || 'cajon_monedero')`,
	} {
		if !strings.Contains(cart, required) {
			t.Errorf("carrito_de_compras.html missing centralized printer contract %q", required)
		}
	}
}

func TestAdvancedPurchasesInventoryProductionContracts(t *testing.T) {
	pagePath := filepath.Join("..", "..", "web", "administrar_empresa", "compras_avanzadas.html")
	rawPage, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	page := string(rawPage)
	for _, required := range []string{
		`id="itemNombre1" class="form-input producto-select"`,
		`id="recItemID" class="form-input"`,
		`id="recBodega" class="form-input bodega-select"`,
		`id="recLote"`,
		`id="btnAddReqItems"`,
		`id="reqItemsDraftBody"`,
		`Puedes agregar tantos productos como necesites.`,
		`id="btnAddRecItem"`,
		`id="recItemsDraftBody"`,
		`Todos se registran de forma atomica bajo el mismo documento de recepcion.`,
		`Recibir y actualizar inventario`,
		`La recepción se marca total únicamente cuando no quedan cantidades pendientes.`,
	} {
		if !strings.Contains(page, required) {
			t.Errorf("compras_avanzadas.html missing %q", required)
		}
	}
	if strings.Contains(page, `id="btnSeed"`) {
		t.Error("production purchases UI must not expose demo data seeding")
	}
	if strings.Contains(page, `id="recEstado"`) {
		t.Error("receipt completion must be calculated by the backend, not selected by the user")
	}

	jsPath := filepath.Join("..", "..", "web", "js", "compras_avanzadas.js")
	rawJS, err := os.ReadFile(jsPath)
	if err != nil {
		t.Fatal(err)
	}
	js := string(rawJS)
	if strings.Contains(js, `seed_demo`) {
		t.Error("production purchases JavaScript must not retain demo-data helpers")
	}
	for _, required := range []string{
		`"Idempotency-Key":idempotencyKey(action, reference)`,
		`producto_id: productID`,
		`bodega_id:num("recBodega")`,
		`lote: val("recLote")`,
		`var recepcionItemsDraft = [];`,
		`var requisicionItemsDraft = [];`,
		`function addRequisitionItems(){`,
		`function renderRequisitionDraft(){`,
		`function addReceptionItem(){`,
		`items:items`,
		`function publicError(status, raw)`,
		`Promise.all([loadProveedores(), loadProductos(), loadBodegas()])`,
	} {
		if !strings.Contains(js, required) {
			t.Errorf("compras_avanzadas.js missing %q", required)
		}
	}

	dbPath := filepath.Join("..", "db", "compras_avanzadas.go")
	rawDB, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	dbSource := string(rawDB)
	if strings.Contains(dbSource, `SeedEmpresaComprasAvanzadasDemo`) {
		t.Error("production purchases database layer must not retain demo-data creation")
	}
	for _, required := range []string{
		`WHERE empresa_id=? AND id=?`,
		`FOR UPDATE OF i`,
		`len(req.Items) == 0 || len(req.Items) > 500`,
		`len(rec.Items) == 0 || len(rec.Items) > 500`,
		`seenProducts := make(map[int64]struct{}, len(req.Items))`,
		`ErrInventarioEntidadNoDisponible`,
		`COALESCE(cantidad_recibida,0)+? <= COALESCE(cantidad_solicitada,0)`,
		`validateBodegaEmpresaTx(tx, rec.EmpresaID, rec.BodegaID)`,
		`validateProductoEmpresaTx(tx, rec.EmpresaID, productoID)`,
		`upsertExistenciaTx(tx, rec.EmpresaID, productoID, rec.BodegaID`,
		`registerCostoLoteTx(tx, rec.EmpresaID, productoID, rec.BodegaID`,
		`insertMovimientoTx(tx, InventarioMovimiento{`,
		`upsertLoteAvanzadoCompraTx`,
		`la requisicion no se puede editar en estado`,
		`la requisicion no se puede cotizar en estado`,
		`la requisicion no se puede aprobar en estado`,
		`la cotizacion no esta seleccionada para recepcion`,
	} {
		if !strings.Contains(dbSource, required) {
			t.Errorf("compras_avanzadas.go missing atomic inventory contract %q", required)
		}
	}
	handlerPath := filepath.Join("compras_avanzadas.go")
	rawHandler, err := os.ReadFile(handlerPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(rawHandler), `case "seed_demo"`) {
		t.Error("production purchases API must not expose demo-data creation")
	}
}

func TestPurchaseAIFilePickerMatchesBackendAllowlist(t *testing.T) {
	pagePath := filepath.Join("..", "..", "web", "administrar_empresa", "compras.html")
	rawPage, err := os.ReadFile(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	page := string(rawPage)
	inputStart := strings.Index(page, `id="comprobanteArchivo"`)
	if inputStart < 0 {
		t.Fatal("purchase AI file picker is missing")
	}
	inputTag := page[inputStart:]
	if inputEnd := strings.Index(inputTag, ">"); inputEnd >= 0 {
		inputTag = inputTag[:inputEnd]
	}
	if !strings.Contains(inputTag, `accept=".png,.jpg,.jpeg,.webp,.pdf,.xml"`) {
		t.Fatal("purchase AI file picker must match the verified backend types")
	}
	if !strings.Contains(page, `throw new Error(publicError(resp.status, errorText))`) {
		t.Fatal("purchase UI must not expose raw HTML or internal API errors")
	}
	for _, required := range []string{
		`function resolverSoporteCompraDuplicado(soporte)`,
		`&action=soportes&registro=activo`,
		`actual.duplicado_soporte_id`,
		`soporteCompraTieneExtraccion(soporte)`,
		`se reutilizo de forma segura una extraccion IA existente`,
	} {
		if !strings.Contains(page, required) {
			t.Errorf("purchase AI duplicate flow missing contract %q", required)
		}
	}
	for _, forbidden := range []string{".gif", ".txt", ".csv", ".doc", ".xlsx"} {
		if strings.Contains(inputTag, forbidden) {
			t.Errorf("purchase AI file picker still advertises unsupported type %s", forbidden)
		}
	}
	fixturePath := filepath.Join("testdata", "qa_factura_compra.xml")
	fixture, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(strings.TrimSpace(string(fixture)), `<?xml`) || !strings.Contains(string(fixture), `<DocumentoSintetico>true</DocumentoSintetico>`) {
		t.Fatal("synthetic purchase fixture must be explicit XML without commercial validity")
	}
}

func TestInventoryProductionRoutesAreRegisteredOnce(t *testing.T) {
	mainPath := filepath.Join("..", "main.go")
	raw, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, route := range []string{
		"/api/empresa/bodegas",
		"/api/empresa/categorias_productos",
		"/api/empresa/productos",
		"/api/empresa/recetas_productos",
		"/api/empresa/inventario/existencias",
		"/api/empresa/inventario/movimientos",
		"/api/empresa/inventario/transferir",
		"/api/empresa/inventario/ajustar",
		"/api/empresa/inventario_avanzado",
		"/api/empresa/compras_avanzadas",
		"/api/empresa/soportes_compras_ia",
		"/api/empresa/impresoras",
		"/api/empresa/impresoras/resolver",
	} {
		marker := `http.HandleFunc("` + route + `"`
		if count := strings.Count(source, marker); count != 1 {
			t.Errorf("inventory route %s registered %d times, want exactly once", route, count)
		}
	}
}

func TestInventoryProductionPagesHaveUniqueMarkupIDs(t *testing.T) {
	scriptBlock := regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	idPattern := regexp.MustCompile(`\bid\s*=\s*["']([^"']+)["']`)
	for _, name := range []string{
		"administrar_productos.html",
		"compras.html",
		"compras_avanzadas.html",
		"inventario_avanzado.html",
		"soportes_compras_ia.html",
		"configuracion_impresora.html",
		"carrito_de_compras.html",
		"recetas_productos.html",
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", name))
			if err != nil {
				t.Fatal(err)
			}
			markup := scriptBlock.ReplaceAll(raw, nil)
			seen := map[string]bool{}
			for _, match := range idPattern.FindAllSubmatch(markup, -1) {
				id := string(match[1])
				if seen[id] {
					t.Errorf("duplicate markup id %q", id)
				}
				seen[id] = true
			}
		})
	}
}

func TestInventoryProductionPagesHaveUniqueNamedFunctions(t *testing.T) {
	functionPattern := regexp.MustCompile(`\bfunction\s+([A-Za-z_$][A-Za-z0-9_$]*)\s*\(`)
	for _, name := range []string{
		"administrar_productos.html",
		"compras.html",
		"compras_avanzadas.html",
		"inventario_avanzado.html",
		"soportes_compras_ia.html",
		"configuracion_impresora.html",
		"carrito_de_compras.html",
		"recetas_productos.html",
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "web", "administrar_empresa", name))
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]bool{}
			for _, match := range functionPattern.FindAllSubmatch(raw, -1) {
				functionName := string(match[1])
				if seen[functionName] {
					t.Errorf("duplicate named function %q", functionName)
				}
				seen[functionName] = true
			}
		})
	}
}

func TestProductosHandlersDoNotExposeRawInternalErrors(t *testing.T) {
	raw, err := os.ReadFile("productos.go")
	if err != nil {
		t.Fatal(err)
	}
	source := string(raw)
	for _, forbidden := range []string{
		`http.Error(w, err.Error(),`,
		`"failed to list productos: "+err.Error()`,
		`"failed to transfer stock: "+err.Error()`,
	} {
		if strings.Contains(source, forbidden) {
			t.Errorf("productos handler exposes raw internal error through %q", forbidden)
		}
	}
	for _, required := range []string{
		`func writeProductosInternalError`,
		`[productos_inventario] operation=%s request_id=%s error_type=%T`,
		`func productosPublicError`,
		`No se pudo completar la operación de productos e inventario.`,
	} {
		if !strings.Contains(source, required) {
			t.Errorf("productos public error contract missing %q", required)
		}
	}
}
