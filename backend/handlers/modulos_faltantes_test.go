package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func ensureModulosFaltantesHandlerSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaModulosFaltantesSchema: %v", err)
	}
}

func decodeBodyAsMap(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	out := map[string]interface{}{}
	if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal response: %v body=%s", err, rr.Body.String())
	}
	return out
}

func containsJSONStringValue(values []interface{}, expected string) bool {
	expected = strings.ToLower(strings.TrimSpace(expected))
	for _, value := range values {
		candidate := strings.ToLower(strings.TrimSpace(value.(string)))
		if candidate == expected {
			return true
		}
	}
	return false
}

func TestEmpresaIntegracionesAPIsHandlerHealthAndSync(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_integraciones_apis_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(51)
	probeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer probeServer.Close()

	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgIntegracionesAPIs.Table, empresaID, map[string]interface{}{
		"codigo":             "API-TEST-001",
		"nombre_integracion": "ERP API",
		"tipo_integracion":   "rest",
		"base_url":           probeServer.URL,
		"estado_integracion": "inactiva",
	}, cfgIntegracionesAPIs.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow: %v", err)
	}

	handler := EmpresaIntegracionesAPIsHandler(dbEmp)

	reqHealth := httptest.NewRequest(http.MethodGet, "/api/empresa/integraciones/apis?empresa_id=51&action=health_check&id="+strconv.FormatInt(id, 10), nil)
	rrHealth := httptest.NewRecorder()
	handler.ServeHTTP(rrHealth, reqHealth)
	if rrHealth.Code != http.StatusOK {
		t.Fatalf("health_check status=%d body=%s", rrHealth.Code, rrHealth.Body.String())
	}

	healthResp := decodeBodyAsMap(t, rrHealth)
	if ok, _ := healthResp["ok"].(bool); !ok {
		t.Fatalf("health_check response not ok: %s", rrHealth.Body.String())
	}
	results, _ := healthResp["resultados"].([]interface{})
	if len(results) != 1 {
		t.Fatalf("health_check resultados esperados=1 obtenidos=%d", len(results))
	}
	first := results[0].(map[string]interface{})
	if strings.ToLower(strings.TrimSpace(first["estado_integracion"].(string))) != "activa" {
		t.Fatalf("estado_integracion esperado=activa obtenido=%v", first["estado_integracion"])
	}
	if reachable, _ := first["reachable"].(bool); !reachable {
		t.Fatalf("reachable esperado=true obtenido=%v", first["reachable"])
	}

	rowAfterHealth, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgIntegracionesAPIs.Table, empresaID, id)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID after health: %v", err)
	}
	if got := strings.ToLower(genericStringValue(rowAfterHealth["estado_integracion"])); got != "activa" {
		t.Fatalf("estado_integracion en DB esperado=activa obtenido=%s", got)
	}
	if got := strings.TrimSpace(genericStringValue(rowAfterHealth["respuesta_ultimo_sync"])); got == "" {
		t.Fatalf("respuesta_ultimo_sync debe registrarse tras health_check")
	}

	reqSync := httptest.NewRequest(http.MethodPost, "/api/empresa/integraciones/apis?empresa_id=51&action=sync_manual&id="+strconv.FormatInt(id, 10), nil)
	rrSync := httptest.NewRecorder()
	handler.ServeHTTP(rrSync, reqSync)
	if rrSync.Code != http.StatusOK {
		t.Fatalf("sync_manual status=%d body=%s", rrSync.Code, rrSync.Body.String())
	}

	rowAfterSync, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgIntegracionesAPIs.Table, empresaID, id)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID after sync: %v", err)
	}
	if got := strings.TrimSpace(genericStringValue(rowAfterSync["ultima_sincronizacion"])); got == "" {
		t.Fatalf("ultima_sincronizacion debe actualizarse en sync_manual")
	}
}

func TestEmpresaIntegracionesBancosHandlerSyncAndEstado(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_integraciones_bancos_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(52)
	probeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("auth required"))
	}))
	defer probeServer.Close()

	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgIntegracionesBancos.Table, empresaID, map[string]interface{}{
		"codigo":             "BANK-TEST-001",
		"banco_nombre":       "Banco Test",
		"numero_cuenta":      "123456",
		"api_endpoint":       probeServer.URL,
		"estado_integracion": "inactiva",
	}, cfgIntegracionesBancos.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow bancos: %v", err)
	}

	handler := EmpresaIntegracionesBancosHandler(dbEmp)
	reqSync := httptest.NewRequest(http.MethodPost, "/api/empresa/integraciones/bancos?empresa_id=52&action=sync_manual&id="+strconv.FormatInt(id, 10), nil)
	rrSync := httptest.NewRecorder()
	handler.ServeHTTP(rrSync, reqSync)
	if rrSync.Code != http.StatusOK {
		t.Fatalf("sync_manual bancos status=%d body=%s", rrSync.Code, rrSync.Body.String())
	}

	rowAfterSync, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgIntegracionesBancos.Table, empresaID, id)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID bancos: %v", err)
	}
	if got := strings.ToLower(genericStringValue(rowAfterSync["estado_integracion"])); got != "activa" {
		t.Fatalf("estado_integracion banco esperado=activa obtenido=%s", got)
	}
	if got := strings.TrimSpace(genericStringValue(rowAfterSync["ultima_conciliacion"])); got == "" {
		t.Fatalf("ultima_conciliacion debe actualizarse en sync_manual")
	}

	reqEstado := httptest.NewRequest(http.MethodGet, "/api/empresa/integraciones/bancos?empresa_id=52&action=estado&id="+strconv.FormatInt(id, 10), nil)
	rrEstado := httptest.NewRecorder()
	handler.ServeHTTP(rrEstado, reqEstado)
	if rrEstado.Code != http.StatusOK {
		t.Fatalf("estado bancos status=%d body=%s", rrEstado.Code, rrEstado.Body.String())
	}

	estadoResp := decodeBodyAsMap(t, rrEstado)
	items, _ := estadoResp["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("estado bancos items esperados=1 obtenidos=%d", len(items))
	}
}

func TestEmpresaVentasCotizacionesStateMachine(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_cotizaciones_sm_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(53)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":           "COT-TEST-001",
		"cliente_nombre":   "Cliente SM",
		"estado_documento": "borrador",
	}, cfgCotizacionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow cotizacion: %v", err)
	}

	handler := EmpresaVentasCotizacionesHandler(dbEmp)

	bodyValid, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "emitida",
		"motivo":       "aprobacion comercial",
	})
	reqValid := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/cotizaciones?action=transicionar", strings.NewReader(string(bodyValid)))
	rrValid := httptest.NewRecorder()
	handler.ServeHTTP(rrValid, reqValid)
	if rrValid.Code != http.StatusOK {
		t.Fatalf("transicion valida status=%d body=%s", rrValid.Code, rrValid.Body.String())
	}

	rowAfterValid, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgCotizacionesVenta.Table, empresaID, id)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID cotizacion: %v", err)
	}
	if got := strings.ToLower(genericStringValue(rowAfterValid["estado_documento"])); got != "emitida" {
		t.Fatalf("estado_documento esperado=emitida obtenido=%s", got)
	}

	bodyInvalid, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "convertida",
	})
	reqInvalid := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/cotizaciones?action=transicionar", strings.NewReader(string(bodyInvalid)))
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusConflict {
		t.Fatalf("transicion invalida debe retornar 409, obtenido=%d body=%s", rrInvalid.Code, rrInvalid.Body.String())
	}

	reqTransitions := httptest.NewRequest(http.MethodGet, "/api/empresa/ventas/cotizaciones?empresa_id=53&action=transiciones&id="+strconv.FormatInt(id, 10), nil)
	rrTransitions := httptest.NewRecorder()
	handler.ServeHTTP(rrTransitions, reqTransitions)
	if rrTransitions.Code != http.StatusOK {
		t.Fatalf("transiciones status=%d body=%s", rrTransitions.Code, rrTransitions.Body.String())
	}
	respTransitions := decodeBodyAsMap(t, rrTransitions)
	allowed, _ := respTransitions["transiciones_disponibles"].([]interface{})
	if !containsJSONStringValue(allowed, "aprobada") {
		t.Fatalf("se esperaba transicion disponible a 'aprobada': %v", allowed)
	}
}

func TestEmpresaCRMLeadsStateMachine(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_crm_leads_sm_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(54)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCRMLeads.Table, empresaID, map[string]interface{}{
		"codigo":      "LEAD-TEST-001",
		"nombre":      "Lead Pipeline",
		"estado_lead": "nuevo",
	}, cfgCRMLeads.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow lead: %v", err)
	}

	handler := EmpresaCRMLeadsHandler(dbEmp)

	bodyInvalid, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "ganado",
	})
	reqInvalid := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/leads?action=transicionar", strings.NewReader(string(bodyInvalid)))
	rrInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalid, reqInvalid)
	if rrInvalid.Code != http.StatusConflict {
		t.Fatalf("transicion CRM invalida debe retornar 409, obtenido=%d body=%s", rrInvalid.Code, rrInvalid.Body.String())
	}

	bodyContactado, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "contactado",
	})
	reqContactado := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/leads?action=transicionar", strings.NewReader(string(bodyContactado)))
	rrContactado := httptest.NewRecorder()
	handler.ServeHTTP(rrContactado, reqContactado)
	if rrContactado.Code != http.StatusOK {
		t.Fatalf("transicion CRM a contactado status=%d body=%s", rrContactado.Code, rrContactado.Body.String())
	}

	bodyCalificado, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "calificado",
	})
	reqCalificado := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/leads?action=transicionar", strings.NewReader(string(bodyCalificado)))
	rrCalificado := httptest.NewRecorder()
	handler.ServeHTTP(rrCalificado, reqCalificado)
	if rrCalificado.Code != http.StatusOK {
		t.Fatalf("transicion CRM a calificado status=%d body=%s", rrCalificado.Code, rrCalificado.Body.String())
	}

	reqEstado := httptest.NewRequest(http.MethodGet, "/api/empresa/crm/leads?empresa_id=54&action=estado&id="+strconv.FormatInt(id, 10), nil)
	rrEstado := httptest.NewRecorder()
	handler.ServeHTTP(rrEstado, reqEstado)
	if rrEstado.Code != http.StatusOK {
		t.Fatalf("estado CRM status=%d body=%s", rrEstado.Code, rrEstado.Body.String())
	}
	respEstado := decodeBodyAsMap(t, rrEstado)
	items, _ := respEstado["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("estado CRM items esperados=1 obtenidos=%d", len(items))
	}
	row := items[0].(map[string]interface{})
	if got := strings.ToLower(strings.TrimSpace(row["estado_actual"].(string))); got != "calificado" {
		t.Fatalf("estado_actual esperado=calificado obtenido=%s", got)
	}
}

func TestEmpresaVentasPedidosStateMachine(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_pedidos_sm_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(55)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, map[string]interface{}{
		"codigo":         "PED-TEST-001",
		"cliente_nombre": "Cliente Pedido",
		"estado_pedido":  "borrador",
	}, cfgPedidosVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow pedido: %v", err)
	}

	handler := EmpresaVentasPedidosHandler(dbEmp)

	bodyConfirmado, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "confirmado",
	})
	reqConfirmado := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/pedidos?action=transicionar", strings.NewReader(string(bodyConfirmado)))
	rrConfirmado := httptest.NewRecorder()
	handler.ServeHTTP(rrConfirmado, reqConfirmado)
	if rrConfirmado.Code != http.StatusOK {
		t.Fatalf("transicion pedido a confirmado status=%d body=%s", rrConfirmado.Code, rrConfirmado.Body.String())
	}

	bodyPreparacion, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "en_preparacion",
	})
	reqPreparacion := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/pedidos?action=transicionar", strings.NewReader(string(bodyPreparacion)))
	rrPreparacion := httptest.NewRecorder()
	handler.ServeHTTP(rrPreparacion, reqPreparacion)
	if rrPreparacion.Code != http.StatusOK {
		t.Fatalf("transicion pedido a en_preparacion status=%d body=%s", rrPreparacion.Code, rrPreparacion.Body.String())
	}

	bodyInvalido, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "cerrado",
	})
	reqInvalido := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/pedidos?action=transicionar", strings.NewReader(string(bodyInvalido)))
	rrInvalido := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalido, reqInvalido)
	if rrInvalido.Code != http.StatusConflict {
		t.Fatalf("transicion invalida pedido debe retornar 409, obtenido=%d body=%s", rrInvalido.Code, rrInvalido.Body.String())
	}

	reqTransitions := httptest.NewRequest(http.MethodGet, "/api/empresa/ventas/pedidos?empresa_id=55&action=transiciones&id="+strconv.FormatInt(id, 10), nil)
	rrTransitions := httptest.NewRecorder()
	handler.ServeHTTP(rrTransitions, reqTransitions)
	if rrTransitions.Code != http.StatusOK {
		t.Fatalf("transiciones pedido status=%d body=%s", rrTransitions.Code, rrTransitions.Body.String())
	}
	respTransitions := decodeBodyAsMap(t, rrTransitions)
	allowed, _ := respTransitions["transiciones_disponibles"].([]interface{})
	if !containsJSONStringValue(allowed, "despachado") {
		t.Fatalf("se esperaba transicion disponible a 'despachado': %v", allowed)
	}
}

func TestEmpresaVentasDevolucionesStateMachine(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_devoluciones_sm_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(56)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDevolucionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":            "DEVV-TEST-001",
		"motivo":            "Ajuste por calidad",
		"estado_devolucion": "borrador",
	}, cfgDevolucionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow devolucion: %v", err)
	}

	handler := EmpresaVentasDevolucionesHandler(dbEmp)

	bodySolicitada, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "solicitada",
	})
	reqSolicitada := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/devoluciones?action=transicionar", strings.NewReader(string(bodySolicitada)))
	rrSolicitada := httptest.NewRecorder()
	handler.ServeHTTP(rrSolicitada, reqSolicitada)
	if rrSolicitada.Code != http.StatusOK {
		t.Fatalf("transicion devolucion a solicitada status=%d body=%s", rrSolicitada.Code, rrSolicitada.Body.String())
	}

	bodyAprobada, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "aprobada",
	})
	reqAprobada := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/devoluciones?action=transicionar", strings.NewReader(string(bodyAprobada)))
	rrAprobada := httptest.NewRecorder()
	handler.ServeHTTP(rrAprobada, reqAprobada)
	if rrAprobada.Code != http.StatusOK {
		t.Fatalf("transicion devolucion a aprobada status=%d body=%s", rrAprobada.Code, rrAprobada.Body.String())
	}

	bodyInvalido, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "cerrada",
	})
	reqInvalido := httptest.NewRequest(http.MethodPut, "/api/empresa/ventas/devoluciones?action=transicionar", strings.NewReader(string(bodyInvalido)))
	rrInvalido := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalido, reqInvalido)
	if rrInvalido.Code != http.StatusConflict {
		t.Fatalf("transicion invalida devolucion debe retornar 409, obtenido=%d body=%s", rrInvalido.Code, rrInvalido.Body.String())
	}

	reqEstado := httptest.NewRequest(http.MethodGet, "/api/empresa/ventas/devoluciones?empresa_id=56&action=estado&id="+strconv.FormatInt(id, 10), nil)
	rrEstado := httptest.NewRecorder()
	handler.ServeHTTP(rrEstado, reqEstado)
	if rrEstado.Code != http.StatusOK {
		t.Fatalf("estado devolucion status=%d body=%s", rrEstado.Code, rrEstado.Body.String())
	}
	respEstado := decodeBodyAsMap(t, rrEstado)
	items, _ := respEstado["items"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("estado devolucion items esperados=1 obtenidos=%d", len(items))
	}
	row := items[0].(map[string]interface{})
	if got := strings.ToLower(strings.TrimSpace(row["estado_actual"].(string))); got != "aprobada" {
		t.Fatalf("estado_actual esperado=aprobada obtenido=%s", got)
	}
}

func TestEmpresaCRMInteraccionesStateMachine(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_crm_interacciones_sm_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(57)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCRMInteracciones.Table, empresaID, map[string]interface{}{
		"codigo":             "INT-TEST-001",
		"tipo_interaccion":   "llamada",
		"resumen":            "Primer contacto",
		"estado_interaccion": "abierta",
	}, cfgCRMInteracciones.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow interaccion: %v", err)
	}

	handler := EmpresaCRMInteraccionesHandler(dbEmp)

	bodyProgreso, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "en_progreso",
	})
	reqProgreso := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/interacciones?action=transicionar", strings.NewReader(string(bodyProgreso)))
	rrProgreso := httptest.NewRecorder()
	handler.ServeHTTP(rrProgreso, reqProgreso)
	if rrProgreso.Code != http.StatusOK {
		t.Fatalf("transicion interaccion a en_progreso status=%d body=%s", rrProgreso.Code, rrProgreso.Body.String())
	}

	bodyCerrada, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "cerrada",
	})
	reqCerrada := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/interacciones?action=transicionar", strings.NewReader(string(bodyCerrada)))
	rrCerrada := httptest.NewRecorder()
	handler.ServeHTTP(rrCerrada, reqCerrada)
	if rrCerrada.Code != http.StatusOK {
		t.Fatalf("transicion interaccion a cerrada status=%d body=%s", rrCerrada.Code, rrCerrada.Body.String())
	}

	bodyInvalido, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "cancelada",
	})
	reqInvalido := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/interacciones?action=transicionar", strings.NewReader(string(bodyInvalido)))
	rrInvalido := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalido, reqInvalido)
	if rrInvalido.Code != http.StatusConflict {
		t.Fatalf("transicion invalida interaccion debe retornar 409, obtenido=%d body=%s", rrInvalido.Code, rrInvalido.Body.String())
	}

	reqTransitions := httptest.NewRequest(http.MethodGet, "/api/empresa/crm/interacciones?empresa_id=57&action=transiciones&id="+strconv.FormatInt(id, 10), nil)
	rrTransitions := httptest.NewRecorder()
	handler.ServeHTTP(rrTransitions, reqTransitions)
	if rrTransitions.Code != http.StatusOK {
		t.Fatalf("transiciones interaccion status=%d body=%s", rrTransitions.Code, rrTransitions.Body.String())
	}
	respTransitions := decodeBodyAsMap(t, rrTransitions)
	allowed, _ := respTransitions["transiciones_disponibles"].([]interface{})
	if !containsJSONStringValue(allowed, "reabierta") {
		t.Fatalf("se esperaba transicion disponible a 'reabierta': %v", allowed)
	}
}

func TestEmpresaCRMCampanasStateMachine(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_crm_campanas_sm_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(58)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCRMCampanas.Table, empresaID, map[string]interface{}{
		"codigo":         "CAMP-TEST-001",
		"nombre":         "Campana Primavera",
		"canal":          "email",
		"estado_campana": "planificada",
	}, cfgCRMCampanas.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow campana: %v", err)
	}

	handler := EmpresaCRMCampanasHandler(dbEmp)

	bodyActiva, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "activa",
	})
	reqActiva := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/campanas?action=transicionar", strings.NewReader(string(bodyActiva)))
	rrActiva := httptest.NewRecorder()
	handler.ServeHTTP(rrActiva, reqActiva)
	if rrActiva.Code != http.StatusOK {
		t.Fatalf("transicion campana a activa status=%d body=%s", rrActiva.Code, rrActiva.Body.String())
	}

	bodyPausada, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "pausada",
	})
	reqPausada := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/campanas?action=transicionar", strings.NewReader(string(bodyPausada)))
	rrPausada := httptest.NewRecorder()
	handler.ServeHTTP(rrPausada, reqPausada)
	if rrPausada.Code != http.StatusOK {
		t.Fatalf("transicion campana a pausada status=%d body=%s", rrPausada.Code, rrPausada.Body.String())
	}

	bodyFinalizada, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "finalizada",
	})
	reqFinalizada := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/campanas?action=transicionar", strings.NewReader(string(bodyFinalizada)))
	rrFinalizada := httptest.NewRecorder()
	handler.ServeHTTP(rrFinalizada, reqFinalizada)
	if rrFinalizada.Code != http.StatusOK {
		t.Fatalf("transicion campana a finalizada status=%d body=%s", rrFinalizada.Code, rrFinalizada.Body.String())
	}

	bodyInvalido, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"id":           id,
		"nuevo_estado": "activa",
	})
	reqInvalido := httptest.NewRequest(http.MethodPut, "/api/empresa/crm/campanas?action=transicionar", strings.NewReader(string(bodyInvalido)))
	rrInvalido := httptest.NewRecorder()
	handler.ServeHTTP(rrInvalido, reqInvalido)
	if rrInvalido.Code != http.StatusConflict {
		t.Fatalf("transicion invalida campana debe retornar 409, obtenido=%d body=%s", rrInvalido.Code, rrInvalido.Body.String())
	}

	reqTransitions := httptest.NewRequest(http.MethodGet, "/api/empresa/crm/campanas?empresa_id=58&action=transiciones&id="+strconv.FormatInt(id, 10), nil)
	rrTransitions := httptest.NewRecorder()
	handler.ServeHTTP(rrTransitions, reqTransitions)
	if rrTransitions.Code != http.StatusOK {
		t.Fatalf("transiciones campana status=%d body=%s", rrTransitions.Code, rrTransitions.Body.String())
	}
	respTransitions := decodeBodyAsMap(t, rrTransitions)
	allowed, _ := respTransitions["transiciones_disponibles"].([]interface{})
	if !containsJSONStringValue(allowed, "archivada") {
		t.Fatalf("se esperaba transicion disponible a 'archivada': %v", allowed)
	}
}
