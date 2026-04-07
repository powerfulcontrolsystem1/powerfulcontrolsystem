package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaCalculadoraHandlerConfigOperacionesFiltrosYExport(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_calculadora_handler.db")
	ensureEmpresaReportesSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaCalculadoraSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaCalculadoraSchema: %v", err)
	}

	handler := EmpresaCalculadoraHandler(dbEmp)
	empresaID := int64(34)

	clienteID, err := dbpkg.CreateCliente(dbEmp, dbpkg.Cliente{
		EmpresaID:         empresaID,
		TipoDocumento:     "CC",
		NumeroDocumento:   "10990034",
		NombreRazonSocial: "Cliente Calculadora QA",
		UsuarioCreador:    "qa_calculadora",
	})
	if err != nil {
		t.Fatalf("CreateCliente: %v", err)
	}

	carritoID, err := dbpkg.CreateCarritoCompra(dbEmp, dbpkg.CarritoCompra{
		EmpresaID:      empresaID,
		Codigo:         "CRT-CALC-001",
		Nombre:         "Carrito Calculadora QA",
		ClienteID:      clienteID,
		EstadoCarrito:  "abierto",
		UsuarioCreador: "qa_calculadora",
	})
	if err != nil {
		t.Fatalf("CreateCarritoCompra: %v", err)
	}

	cotizacionID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":           "COT-CALC-001",
		"cliente_id":       clienteID,
		"cliente_nombre":   "Cliente Calculadora QA",
		"estado_documento": "emitida",
		"fecha_documento":  "2026-04-08 10:00:00",
		"total":            42000,
	}, cfgCotizacionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow cotizacion: %v", err)
	}

	reqCfg := httptest.NewRequest(http.MethodGet, "/api/empresa/calculadora?action=config&empresa_id=34", nil)
	rrCfg := httptest.NewRecorder()
	handler.ServeHTTP(rrCfg, reqCfg)
	if rrCfg.Code != http.StatusOK {
		t.Fatalf("config status=%d body=%s", rrCfg.Code, rrCfg.Body.String())
	}

	var cfgResp struct {
		Configuracion dbpkg.EmpresaCalculadoraConfiguracion `json:"configuracion"`
	}
	if err := json.Unmarshal(rrCfg.Body.Bytes(), &cfgResp); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if !cfgResp.Configuracion.IntegrarCarritos || !cfgResp.Configuracion.IntegrarCotizaciones {
		t.Fatalf("expected default integrations enabled, got %+v", cfgResp.Configuracion)
	}

	reqDisableCot := httptest.NewRequest(http.MethodPut, "/api/empresa/calculadora?action=config&empresa_id=34", strings.NewReader(`{"empresa_id":34,"integrar_cotizaciones":false}`))
	reqDisableCot.Header.Set("Content-Type", "application/json")
	rrDisableCot := httptest.NewRecorder()
	handler.ServeHTTP(rrDisableCot, reqDisableCot)
	if rrDisableCot.Code != http.StatusOK {
		t.Fatalf("disable cotizaciones status=%d body=%s", rrDisableCot.Code, rrDisableCot.Body.String())
	}

	reqCotDisabled := httptest.NewRequest(http.MethodPost, "/api/empresa/calculadora?empresa_id=34", strings.NewReader(`{"empresa_id":34,"expresion":"40+2","resultado":"42","cotizacion_id":`+itoa64(cotizacionID)+`,"fecha_operacion":"2026-04-08 10:00:00"}`))
	reqCotDisabled.Header.Set("Content-Type", "application/json")
	reqCotDisabled.Header.Set("X-Usuario-Email", "luis@empresa.com")
	rrCotDisabled := httptest.NewRecorder()
	handler.ServeHTTP(rrCotDisabled, reqCotDisabled)
	if rrCotDisabled.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 with cotizaciones disabled, got=%d body=%s", rrCotDisabled.Code, rrCotDisabled.Body.String())
	}

	reqOpA := httptest.NewRequest(http.MethodPost, "/api/empresa/calculadora?empresa_id=34", strings.NewReader(`{
		"empresa_id":34,
		"expresion":"12 + 8",
		"resultado":"20",
		"etiquetas":["cierre","mesa_2"],
		"cliente_id":`+itoa64(clienteID)+`,
		"carrito_id":`+itoa64(carritoID)+`,
		"documento_tipo":"carrito",
		"documento_codigo":"CRT-CALC-001",
		"fecha_operacion":"2026-04-07 09:00:00"
	}`))
	reqOpA.Header.Set("Content-Type", "application/json")
	reqOpA.Header.Set("X-Usuario-Email", "ana@empresa.com")
	rrOpA := httptest.NewRecorder()
	handler.ServeHTTP(rrOpA, reqOpA)
	if rrOpA.Code != http.StatusCreated {
		t.Fatalf("create operation A status=%d body=%s", rrOpA.Code, rrOpA.Body.String())
	}

	var opAResp struct {
		Operacion dbpkg.EmpresaCalculadoraOperacion `json:"operacion"`
	}
	if err := json.Unmarshal(rrOpA.Body.Bytes(), &opAResp); err != nil {
		t.Fatalf("decode operation A response: %v", err)
	}
	if opAResp.Operacion.CarritoID != carritoID {
		t.Fatalf("expected carrito_id=%d got=%d", carritoID, opAResp.Operacion.CarritoID)
	}
	if opAResp.Operacion.ClienteID != clienteID {
		t.Fatalf("expected cliente_id=%d got=%d", clienteID, opAResp.Operacion.ClienteID)
	}
	if len(opAResp.Operacion.Etiquetas) == 0 {
		t.Fatalf("expected etiquetas in operation A, got %+v", opAResp.Operacion)
	}
	usuarioOpA := strings.TrimSpace(opAResp.Operacion.UsuarioCreador)
	if usuarioOpA == "" {
		usuarioOpA = "sistema"
	}

	reqEnableCot := httptest.NewRequest(http.MethodPut, "/api/empresa/calculadora?action=config&empresa_id=34", strings.NewReader(`{"empresa_id":34,"integrar_cotizaciones":true}`))
	reqEnableCot.Header.Set("Content-Type", "application/json")
	rrEnableCot := httptest.NewRecorder()
	handler.ServeHTTP(rrEnableCot, reqEnableCot)
	if rrEnableCot.Code != http.StatusOK {
		t.Fatalf("enable cotizaciones status=%d body=%s", rrEnableCot.Code, rrEnableCot.Body.String())
	}

	reqOpB := httptest.NewRequest(http.MethodPost, "/api/empresa/calculadora?empresa_id=34", strings.NewReader(`{
		"empresa_id":34,
		"expresion":"40 + 2",
		"resultado":"42",
		"etiquetas":["cotizacion"],
		"cotizacion_id":`+itoa64(cotizacionID)+`,
		"fecha_operacion":"2026-04-08 10:00:00"
	}`))
	reqOpB.Header.Set("Content-Type", "application/json")
	reqOpB.Header.Set("X-Usuario-Email", "luis@empresa.com")
	rrOpB := httptest.NewRecorder()
	handler.ServeHTTP(rrOpB, reqOpB)
	if rrOpB.Code != http.StatusCreated {
		t.Fatalf("create operation B status=%d body=%s", rrOpB.Code, rrOpB.Body.String())
	}

	var opBResp struct {
		Operacion dbpkg.EmpresaCalculadoraOperacion `json:"operacion"`
	}
	if err := json.Unmarshal(rrOpB.Body.Bytes(), &opBResp); err != nil {
		t.Fatalf("decode operation B response: %v", err)
	}
	if opBResp.Operacion.CotizacionID != cotizacionID {
		t.Fatalf("expected cotizacion_id=%d got=%d", cotizacionID, opBResp.Operacion.CotizacionID)
	}
	if opBResp.Operacion.DocumentoTipo != "cotizacion" {
		t.Fatalf("expected documento_tipo=cotizacion got=%s", opBResp.Operacion.DocumentoTipo)
	}

	reqRefs := httptest.NewRequest(http.MethodGet, "/api/empresa/calculadora?action=referencias&empresa_id=34&limit=20", nil)
	rrRefs := httptest.NewRecorder()
	handler.ServeHTTP(rrRefs, reqRefs)
	if rrRefs.Code != http.StatusOK {
		t.Fatalf("referencias status=%d body=%s", rrRefs.Code, rrRefs.Body.String())
	}
	var refsResp struct {
		Clientes     []map[string]interface{} `json:"clientes"`
		Carritos     []map[string]interface{} `json:"carritos"`
		Cotizaciones []map[string]interface{} `json:"cotizaciones"`
	}
	if err := json.Unmarshal(rrRefs.Body.Bytes(), &refsResp); err != nil {
		t.Fatalf("decode referencias response: %v", err)
	}
	if len(refsResp.Clientes) == 0 || len(refsResp.Carritos) == 0 || len(refsResp.Cotizaciones) == 0 {
		t.Fatalf("expected referencias with clientes/carritos/cotizaciones, got %+v", refsResp)
	}

	reqListFiltered := httptest.NewRequest(http.MethodGet, "/api/empresa/calculadora?empresa_id=34&usuario="+usuarioOpA+"&desde=2026-04-01&hasta=2026-04-07", nil)
	rrListFiltered := httptest.NewRecorder()
	handler.ServeHTTP(rrListFiltered, reqListFiltered)
	if rrListFiltered.Code != http.StatusOK {
		t.Fatalf("list filtered status=%d body=%s", rrListFiltered.Code, rrListFiltered.Body.String())
	}

	var listFilteredResp struct {
		Total int64                               `json:"total"`
		Rows  []dbpkg.EmpresaCalculadoraOperacion `json:"rows"`
	}
	if err := json.Unmarshal(rrListFiltered.Body.Bytes(), &listFilteredResp); err != nil {
		t.Fatalf("decode list filtered response: %v", err)
	}
	if listFilteredResp.Total != 1 || len(listFilteredResp.Rows) != 1 {
		t.Fatalf("expected filtered total/rows=1, got total=%d len=%d", listFilteredResp.Total, len(listFilteredResp.Rows))
	}
	if listFilteredResp.Rows[0].CarritoID != carritoID {
		t.Fatalf("expected filtered row with carrito_id=%d got=%d", carritoID, listFilteredResp.Rows[0].CarritoID)
	}

	reqExport := httptest.NewRequest(http.MethodGet, "/api/empresa/calculadora?action=export&empresa_id=34&format=csv&usuario="+usuarioOpA+"&desde=2026-04-01&hasta=2026-04-30", nil)
	rrExport := httptest.NewRecorder()
	handler.ServeHTTP(rrExport, reqExport)
	if rrExport.Code != http.StatusOK {
		t.Fatalf("export status=%d body=%s", rrExport.Code, rrExport.Body.String())
	}
	if !strings.Contains(strings.ToLower(rrExport.Header().Get("Content-Type")), "text/csv") {
		t.Fatalf("expected csv content type, got=%s", rrExport.Header().Get("Content-Type"))
	}
	exportBody := rrExport.Body.String()
	if !strings.Contains(exportBody, usuarioOpA) || !strings.Contains(exportBody, "12 + 8") {
		t.Fatalf("expected filtered export rows, got=%s", exportBody)
	}

	reqClear := httptest.NewRequest(http.MethodPost, "/api/empresa/calculadora?action=limpiar&empresa_id=34", strings.NewReader(`{"empresa_id":34,"usuario":"`+usuarioOpA+`","etiqueta":"cierre","desde":"2026-04-01","hasta":"2026-04-30"}`))
	reqClear.Header.Set("Content-Type", "application/json")
	rrClear := httptest.NewRecorder()
	handler.ServeHTTP(rrClear, reqClear)
	if rrClear.Code != http.StatusOK {
		t.Fatalf("clear status=%d body=%s", rrClear.Code, rrClear.Body.String())
	}
	var clearResp struct {
		Desactivados int64 `json:"desactivados"`
	}
	if err := json.Unmarshal(rrClear.Body.Bytes(), &clearResp); err != nil {
		t.Fatalf("decode clear response: %v", err)
	}
	if clearResp.Desactivados != 1 {
		t.Fatalf("expected desactivados=1 got=%d", clearResp.Desactivados)
	}

	reqListActive := httptest.NewRequest(http.MethodGet, "/api/empresa/calculadora?empresa_id=34", nil)
	rrListActive := httptest.NewRecorder()
	handler.ServeHTTP(rrListActive, reqListActive)
	if rrListActive.Code != http.StatusOK {
		t.Fatalf("list active status=%d body=%s", rrListActive.Code, rrListActive.Body.String())
	}
	var listActiveResp struct {
		Total int64                               `json:"total"`
		Rows  []dbpkg.EmpresaCalculadoraOperacion `json:"rows"`
	}
	if err := json.Unmarshal(rrListActive.Body.Bytes(), &listActiveResp); err != nil {
		t.Fatalf("decode list active response: %v", err)
	}
	if listActiveResp.Total != 1 || len(listActiveResp.Rows) != 1 {
		t.Fatalf("expected one active row after clear, got total=%d len=%d", listActiveResp.Total, len(listActiveResp.Rows))
	}
	if listActiveResp.Rows[0].CotizacionID != cotizacionID {
		t.Fatalf("expected remaining row with cotizacion_id=%d got=%d", cotizacionID, listActiveResp.Rows[0].CotizacionID)
	}
}

func itoa64(v int64) string {
	return strconv.FormatInt(v, 10)
}
