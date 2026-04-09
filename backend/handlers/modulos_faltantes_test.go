package handlers

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
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

func ensureModulosFaltantesHandlerSchema(t *testing.T, dbEmp *sql.DB) {
	t.Helper()
	if err := dbpkg.EnsureEmpresaModulosFaltantesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaModulosFaltantesSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaFinanzasSchema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaDocumentosTransaccionalesSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaDocumentosTransaccionalesSchema: %v", err)
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

func TestEmpresaIntegracionesAPIsHandlerRotarCredencialYMonitoreo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_integraciones_apis_rotacion_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(520)
	probeServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer probeServer.Close()

	idOK, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgIntegracionesAPIs.Table, empresaID, map[string]interface{}{
		"codigo":             "API-ROT-001",
		"nombre_integracion": "Conector ERP",
		"tipo_integracion":   "rest",
		"base_url":           probeServer.URL,
		"api_key_ref":        "env:INTEGRACION_OLD",
		"estado_integracion": "activa",
	}, cfgIntegracionesAPIs.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow api rotacion: %v", err)
	}

	_, err = dbpkg.CreateEmpresaGenericRow(dbEmp, cfgIntegracionesAPIs.Table, empresaID, map[string]interface{}{
		"codigo":             "API-ROT-002",
		"nombre_integracion": "Conector sin endpoint",
		"tipo_integracion":   "rest",
		"base_url":           "",
		"api_key_ref":        "env:INTEGRACION_EMPTY",
		"estado_integracion": "inactiva",
	}, cfgIntegracionesAPIs.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow api sin endpoint: %v", err)
	}

	handler := EmpresaIntegracionesAPIsHandler(dbEmp)

	reqRotate := httptest.NewRequest(http.MethodPut, "/api/empresa/integraciones/apis?action=rotar_credencial", strings.NewReader(`{"empresa_id":520,"id":`+strconv.FormatInt(idOK, 10)+`,"nueva_credencial_ref":"env:INTEGRACION_NEW","validar":true}`))
	rrRotate := httptest.NewRecorder()
	handler.ServeHTTP(rrRotate, reqRotate)
	if rrRotate.Code != http.StatusOK {
		t.Fatalf("rotar_credencial status=%d body=%s", rrRotate.Code, rrRotate.Body.String())
	}

	rowAfterRotate, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgIntegracionesAPIs.Table, empresaID, idOK)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID after rotate: %v", err)
	}
	if got := strings.TrimSpace(genericStringValue(rowAfterRotate["api_key_ref"])); got != "env:INTEGRACION_NEW" {
		t.Fatalf("api_key_ref esperado=env:INTEGRACION_NEW obtenido=%s", got)
	}

	reqRotateInvalid := httptest.NewRequest(http.MethodPut, "/api/empresa/integraciones/apis?action=rotar_credencial", strings.NewReader(`{"empresa_id":520,"id":`+strconv.FormatInt(idOK, 10)+`,"nueva_credencial_ref":"token-plano-123"}`))
	rrRotateInvalid := httptest.NewRecorder()
	handler.ServeHTTP(rrRotateInvalid, reqRotateInvalid)
	if rrRotateInvalid.Code != http.StatusBadRequest {
		t.Fatalf("rotar_credencial invalida debe retornar 400, obtenido=%d body=%s", rrRotateInvalid.Code, rrRotateInvalid.Body.String())
	}

	reqMonitoreo := httptest.NewRequest(http.MethodGet, "/api/empresa/integraciones/apis?empresa_id=520&action=monitoreo&latencia_alerta_ms=1&stale_hours=1", nil)
	rrMonitoreo := httptest.NewRecorder()
	handler.ServeHTTP(rrMonitoreo, reqMonitoreo)
	if rrMonitoreo.Code != http.StatusOK {
		t.Fatalf("monitoreo status=%d body=%s", rrMonitoreo.Code, rrMonitoreo.Body.String())
	}

	respMonitoreo := decodeBodyAsMap(t, rrMonitoreo)
	alertas, _ := respMonitoreo["alertas"].([]interface{})
	if len(alertas) == 0 {
		t.Fatalf("monitoreo debe generar alertas para endpoint vacio o sin sync")
	}

	tipoEncontrado := false
	for _, raw := range alertas {
		row, _ := raw.(map[string]interface{})
		tipo := strings.ToLower(strings.TrimSpace(genericStringValue(row["tipo"])))
		if tipo == "endpoint_invalido" {
			tipoEncontrado = true
			break
		}
	}
	if !tipoEncontrado {
		t.Fatalf("monitoreo debe incluir alerta endpoint_invalido, alertas=%v", alertas)
	}
}

func TestEmpresaIntegracionesBancosHandlerRotarCredencial(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_integraciones_bancos_rotacion_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(521)
	id, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgIntegracionesBancos.Table, empresaID, map[string]interface{}{
		"codigo":             "BANK-ROT-001",
		"banco_nombre":       "Banco Rotacion",
		"numero_cuenta":      "789456123",
		"credencial_ref":     "env:BANK_CRED_OLD",
		"estado_integracion": "activa",
	}, cfgIntegracionesBancos.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow bancos rotacion: %v", err)
	}

	handler := EmpresaIntegracionesBancosHandler(dbEmp)
	reqRotate := httptest.NewRequest(http.MethodPut, "/api/empresa/integraciones/bancos?action=rotar_credencial", strings.NewReader(`{"empresa_id":521,"id":`+strconv.FormatInt(id, 10)+`,"nueva_credencial_ref":"vault:finanzas/banco_rotacion"}`))
	rrRotate := httptest.NewRecorder()
	handler.ServeHTTP(rrRotate, reqRotate)
	if rrRotate.Code != http.StatusOK {
		t.Fatalf("rotar_credencial bancos status=%d body=%s", rrRotate.Code, rrRotate.Body.String())
	}

	rowAfterRotate, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgIntegracionesBancos.Table, empresaID, id)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID bancos after rotate: %v", err)
	}
	if got := strings.TrimSpace(genericStringValue(rowAfterRotate["credencial_ref"])); got != "vault:finanzas/banco_rotacion" {
		t.Fatalf("credencial_ref esperado=vault:finanzas/banco_rotacion obtenido=%s", got)
	}
}

func TestEmpresaDocumentosGestionHandlerVersionadoYControlAcceso(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_documentos_gestion_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(522)
	idBase, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDocumentosGestion.Table, empresaID, map[string]interface{}{
		"codigo":           "DOC-522-001",
		"modulo":           "finanzas",
		"entidad":          "cuentas_por_cobrar",
		"entidad_id":       912,
		"documento_codigo": "DOCREF-522-01",
		"nombre_documento": "Soporte de cartera",
		"tipo_documento":   "pdf",
		"mime_type":        "application/pdf",
		"url_archivo":      "https://files.local/doc-v1.pdf",
		"hash_archivo":     "hash-v1",
		"tamano_bytes":     1200,
		"version":          "1",
		"estado_documento": "vigente",
		"estado":           "activo",
	}, cfgDocumentosGestion.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow documento base: %v", err)
	}

	handler := EmpresaDocumentosGestionHandler(dbEmp)
	reqVersionar := httptest.NewRequest(http.MethodPut, "/api/empresa/documentos/gestion?action=versionar", strings.NewReader(`{"empresa_id":522,"id":`+strconv.FormatInt(idBase, 10)+`,"url_archivo":"https://files.local/doc-v2.pdf","hash_archivo":"hash-v2","observaciones":"actualizacion de cierre mensual"}`))
	reqVersionar.Header.Set("X-Admin-Role", "contabilidad")
	reqVersionar.Header.Set("X-Admin-Email", "qa-contabilidad@empresa.test")
	rrVersionar := httptest.NewRecorder()
	handler.ServeHTTP(rrVersionar, reqVersionar)
	if rrVersionar.Code != http.StatusCreated {
		t.Fatalf("versionar status=%d body=%s", rrVersionar.Code, rrVersionar.Body.String())
	}

	respVersionar := decodeBodyAsMap(t, rrVersionar)
	idNuevo := int64(respVersionar["id_nuevo"].(float64))
	if idNuevo <= 0 {
		t.Fatalf("id_nuevo invalido en versionado: %v", respVersionar["id_nuevo"])
	}
	if got := int64(respVersionar["version_nueva"].(float64)); got != 2 {
		t.Fatalf("version_nueva esperada=2 obtenida=%d", got)
	}

	rowBase, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, idBase)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID base after versionar: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(rowBase["estado_documento"]))); got != "historico" {
		t.Fatalf("estado_documento base esperado=historico obtenido=%s", got)
	}

	rowNuevo, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgDocumentosGestion.Table, empresaID, idNuevo)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID nuevo after versionar: %v", err)
	}
	if got := strings.TrimSpace(genericStringValue(rowNuevo["version"])); got != "2" {
		t.Fatalf("version nuevo esperado=2 obtenido=%s", got)
	}

	reqVersiones := httptest.NewRequest(http.MethodGet, "/api/empresa/documentos/gestion?empresa_id=522&action=versiones&documento_codigo=DOCREF-522-01", nil)
	reqVersiones.Header.Set("X-Admin-Role", "contabilidad")
	rrVersiones := httptest.NewRecorder()
	handler.ServeHTTP(rrVersiones, reqVersiones)
	if rrVersiones.Code != http.StatusOK {
		t.Fatalf("versiones status=%d body=%s", rrVersiones.Code, rrVersiones.Body.String())
	}
	respVersiones := decodeBodyAsMap(t, rrVersiones)
	itemsVersiones, _ := respVersiones["items"].([]interface{})
	if len(itemsVersiones) < 2 {
		t.Fatalf("versiones esperadas>=2 obtenidas=%d", len(itemsVersiones))
	}

	reqAccesoPermitido := httptest.NewRequest(http.MethodGet, "/api/empresa/documentos/gestion?empresa_id=522&action=acceso&id="+strconv.FormatInt(idNuevo, 10)+"&permiso=U", nil)
	reqAccesoPermitido.Header.Set("X-Admin-Role", "contabilidad")
	rrAccesoPermitido := httptest.NewRecorder()
	handler.ServeHTTP(rrAccesoPermitido, reqAccesoPermitido)
	if rrAccesoPermitido.Code != http.StatusOK {
		t.Fatalf("acceso permitido status=%d body=%s", rrAccesoPermitido.Code, rrAccesoPermitido.Body.String())
	}
	respAccesoPermitido := decodeBodyAsMap(t, rrAccesoPermitido)
	if allowed, _ := respAccesoPermitido["acceso_permitido"].(bool); !allowed {
		t.Fatalf("acceso_permitido esperado=true para rol contabilidad")
	}

	reqAccesoDenegado := httptest.NewRequest(http.MethodGet, "/api/empresa/documentos/gestion?empresa_id=522&action=acceso&id="+strconv.FormatInt(idNuevo, 10)+"&permiso=U", nil)
	reqAccesoDenegado.Header.Set("X-Admin-Role", "inventario")
	rrAccesoDenegado := httptest.NewRecorder()
	handler.ServeHTTP(rrAccesoDenegado, reqAccesoDenegado)
	if rrAccesoDenegado.Code != http.StatusOK {
		t.Fatalf("acceso denegado status=%d body=%s", rrAccesoDenegado.Code, rrAccesoDenegado.Body.String())
	}
	respAccesoDenegado := decodeBodyAsMap(t, rrAccesoDenegado)
	if allowed, _ := respAccesoDenegado["acceso_permitido"].(bool); allowed {
		t.Fatalf("acceso_permitido esperado=false para rol inventario con permiso U en modulo finanzas")
	}

	reqRepositorioDenegado := httptest.NewRequest(http.MethodGet, "/api/empresa/documentos/gestion?empresa_id=522&action=repositorio&permiso=U&include_denegados=1", nil)
	reqRepositorioDenegado.Header.Set("X-Admin-Role", "inventario")
	rrRepositorioDenegado := httptest.NewRecorder()
	handler.ServeHTTP(rrRepositorioDenegado, reqRepositorioDenegado)
	if rrRepositorioDenegado.Code != http.StatusOK {
		t.Fatalf("repositorio denegado status=%d body=%s", rrRepositorioDenegado.Code, rrRepositorioDenegado.Body.String())
	}
	respRepositorio := decodeBodyAsMap(t, rrRepositorioDenegado)
	itemsRepo, _ := respRepositorio["items"].([]interface{})
	if len(itemsRepo) == 0 {
		t.Fatalf("repositorio con include_denegados=1 debe devolver items")
	}
	primero, _ := itemsRepo[0].(map[string]interface{})
	if allowed, _ := primero["acceso_permitido"].(bool); allowed {
		t.Fatalf("repositorio esperado con acceso_permitido=false para rol inventario y permiso U")
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

func TestEmpresaVentasCotizacionesConversionPedidoYDocumentoFinal(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_conversion_ventas_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(62)
	cotizacionID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":           "COT-62-001",
		"cliente_id":       1001,
		"cliente_nombre":   "Cliente Conversion",
		"estado_documento": "emitida",
		"fecha_documento":  "2026-04-02 08:30:00",
		"vigencia_hasta":   "2026-04-10",
		"subtotal":         100000,
		"impuesto_total":   19000,
		"total":            119000,
		"moneda":           "COP",
	}, cfgCotizacionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow cotizacion conversion: %v", err)
	}

	hCotizaciones := EmpresaVentasCotizacionesHandler(dbEmp)
	bodyPedido, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         cotizacionID,
	})
	reqPedido := httptest.NewRequest(http.MethodPost, "/api/empresa/ventas/cotizaciones?action=convertir_pedido", strings.NewReader(string(bodyPedido)))
	reqPedido.Header.Set("Content-Type", "application/json")
	rrPedido := httptest.NewRecorder()
	hCotizaciones.ServeHTTP(rrPedido, reqPedido)
	if rrPedido.Code != http.StatusOK {
		t.Fatalf("convertir_pedido status=%d body=%s", rrPedido.Code, rrPedido.Body.String())
	}
	respPedido := decodeBodyAsMap(t, rrPedido)
	pedidoID := anyToInt64(respPedido["pedido_id"])
	if pedidoID <= 0 {
		t.Fatalf("pedido_id invalido en respuesta: %s", rrPedido.Body.String())
	}
	if autoAprobada, _ := respPedido["cotizacion_auto_aprobada"].(bool); !autoAprobada {
		t.Fatalf("se esperaba cotizacion_auto_aprobada=true para cotizacion emitida")
	}

	cotizacionActualizada, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionID)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID cotizacion actualizada: %v", err)
	}
	if got := strings.ToLower(genericStringValue(cotizacionActualizada["estado_documento"])); got != "convertida" {
		t.Fatalf("estado cotizacion esperado=convertida obtenido=%s", got)
	}
	if got := anyToInt64(cotizacionActualizada["convertido_pedido_id"]); got != pedidoID {
		t.Fatalf("convertido_pedido_id esperado=%d obtenido=%d", pedidoID, got)
	}

	pedido, err := dbpkg.GetEmpresaGenericRowByID(dbEmp, cfgPedidosVenta.Table, empresaID, pedidoID)
	if err != nil {
		t.Fatalf("GetEmpresaGenericRowByID pedido: %v", err)
	}
	if got := anyToInt64(pedido["cotizacion_id"]); got != cotizacionID {
		t.Fatalf("cotizacion_id en pedido esperado=%d obtenido=%d", cotizacionID, got)
	}

	bodyDocumento, _ := json.Marshal(map[string]interface{}{
		"empresa_id":     empresaID,
		"id":             cotizacionID,
		"tipo_documento": "factura_electronica",
	})
	reqDocumento := httptest.NewRequest(http.MethodPost, "/api/empresa/ventas/cotizaciones?action=convertir_documento_final", strings.NewReader(string(bodyDocumento)))
	reqDocumento.Header.Set("Content-Type", "application/json")
	rrDocumento := httptest.NewRecorder()
	hCotizaciones.ServeHTTP(rrDocumento, reqDocumento)
	if rrDocumento.Code != http.StatusOK {
		t.Fatalf("convertir_documento_final (cotizacion) status=%d body=%s", rrDocumento.Code, rrDocumento.Body.String())
	}
	respDocumento := decodeBodyAsMap(t, rrDocumento)
	docRaw, ok := respDocumento["documento_final"].(map[string]interface{})
	if !ok {
		t.Fatalf("documento_final no presente en respuesta: %s", rrDocumento.Body.String())
	}
	documentoCodigo := strings.TrimSpace(genericStringValue(docRaw["documento_codigo"]))
	if documentoCodigo == "" {
		t.Fatalf("documento_codigo vacio en documento_final: %s", rrDocumento.Body.String())
	}

	docPersistido, err := dbpkg.GetEmpresaDocumentoFacturacionByCodigo(dbEmp, empresaID, "factura_electronica", documentoCodigo)
	if err != nil {
		t.Fatalf("GetEmpresaDocumentoFacturacionByCodigo: %v", err)
	}
	if docPersistido.EntidadRelacionadaID != pedidoID {
		t.Fatalf("entidad_relacionada_id esperado=%d obtenido=%d", pedidoID, docPersistido.EntidadRelacionadaID)
	}

	hPedidos := EmpresaVentasPedidosHandler(dbEmp)
	bodyPedidoDocumento, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         pedidoID,
	})
	reqPedidoDocumento := httptest.NewRequest(http.MethodPost, "/api/empresa/ventas/pedidos?action=convertir_documento_final", strings.NewReader(string(bodyPedidoDocumento)))
	reqPedidoDocumento.Header.Set("Content-Type", "application/json")
	rrPedidoDocumento := httptest.NewRecorder()
	hPedidos.ServeHTTP(rrPedidoDocumento, reqPedidoDocumento)
	if rrPedidoDocumento.Code != http.StatusOK {
		t.Fatalf("convertir_documento_final (pedido) status=%d body=%s", rrPedidoDocumento.Code, rrPedidoDocumento.Body.String())
	}

	var docsCount int64
	if err := dbEmp.QueryRow(`SELECT COUNT(1)
	FROM empresa_facturacion_documentos
	WHERE empresa_id = ?
	  AND COALESCE(entidad_relacionada_id, 0) = ?
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID, pedidoID).Scan(&docsCount); err != nil {
		t.Fatalf("count documentos por pedido: %v", err)
	}
	if docsCount != 1 {
		t.Fatalf("se esperaba 1 documento final por pedido, obtenido=%d", docsCount)
	}
}

func TestEmpresaVentasCotizacionesEmbudoYAlertasSLA(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_embudo_ventas_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(63)

	cotizacionSLAID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":           "COT-63-001",
		"cliente_nombre":   "Cliente SLA Cotizacion",
		"estado_documento": "emitida",
		"fecha_documento":  "2026-04-01 08:00:00",
		"vigencia_hasta":   "2026-04-02",
		"total":            50000,
	}, cfgCotizacionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("Create cotizacion SLA: %v", err)
	}

	cotizacionPedidoSLAID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":           "COT-63-002",
		"cliente_nombre":   "Cliente SLA Pedido",
		"estado_documento": "convertida",
		"fecha_documento":  "2026-04-01 09:00:00",
		"total":            80000,
	}, cfgCotizacionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("Create cotizacion pedido SLA: %v", err)
	}

	pedidoSLAID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, map[string]interface{}{
		"codigo":         "PED-63-002",
		"cliente_nombre": "Cliente SLA Pedido",
		"cotizacion_id":  cotizacionPedidoSLAID,
		"fecha_pedido":   "2026-04-01 10:00:00",
		"estado_pedido":  "confirmado",
		"total":          80000,
	}, cfgPedidosVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("Create pedido SLA: %v", err)
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionPedidoSLAID, map[string]interface{}{
		"convertido_pedido_id": pedidoSLAID,
	}, cfgCotizacionesVenta.AllowedColumns); err != nil {
		t.Fatalf("Update cotizacion con pedido SLA: %v", err)
	}

	cotizacionCompletaID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, map[string]interface{}{
		"codigo":           "COT-63-003",
		"cliente_nombre":   "Cliente Completo",
		"estado_documento": "convertida",
		"fecha_documento":  "2026-04-03 11:00:00",
		"total":            120000,
	}, cfgCotizacionesVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("Create cotizacion completa: %v", err)
	}

	pedidoCompletoID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgPedidosVenta.Table, empresaID, map[string]interface{}{
		"codigo":         "PED-63-003",
		"cliente_nombre": "Cliente Completo",
		"cotizacion_id":  cotizacionCompletaID,
		"fecha_pedido":   "2026-04-03 12:00:00",
		"estado_pedido":  "entregado",
		"total":          120000,
	}, cfgPedidosVenta.AllowedColumns)
	if err != nil {
		t.Fatalf("Create pedido completo: %v", err)
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgCotizacionesVenta.Table, empresaID, cotizacionCompletaID, map[string]interface{}{
		"convertido_pedido_id": pedidoCompletoID,
	}, cfgCotizacionesVenta.AllowedColumns); err != nil {
		t.Fatalf("Update cotizacion completa con pedido: %v", err)
	}

	if _, err := dbpkg.UpsertEmpresaDocumentoFacturacion(dbEmp, dbpkg.EmpresaDocumentoFacturacion{
		EmpresaID:            empresaID,
		TipoDocumento:        "factura_electronica",
		DocumentoCodigo:      "FV-63-003",
		EstadoDocumento:      "emitida",
		EventoUltimo:         "emitir",
		MontoTotal:           120000,
		Moneda:               "COP",
		FechaDocumento:       "2026-04-03",
		EntidadRelacionadaID: pedidoCompletoID,
		UsuarioCreador:       "qa_embudo",
		Estado:               "activo",
		Observaciones:        "Documento final de prueba",
	}); err != nil {
		t.Fatalf("UpsertEmpresaDocumentoFacturacion: %v", err)
	}

	handler := EmpresaVentasCotizacionesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/ventas/cotizaciones?action=embudo&empresa_id=63&limit=50&sla_cotizacion_horas=1&sla_pedido_horas=1", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("embudo ventas status=%d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeBodyAsMap(t, rr)
	summary, ok := resp["summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("summary no disponible en embudo: %s", rr.Body.String())
	}
	if got := anyToInt64(summary["cotizaciones_total"]); got < 3 {
		t.Fatalf("cotizaciones_total esperado>=3 obtenido=%d", got)
	}
	if got := anyToInt64(summary["cotizaciones_convertidas_pedido"]); got < 2 {
		t.Fatalf("cotizaciones_convertidas_pedido esperado>=2 obtenido=%d", got)
	}
	if got := anyToInt64(summary["cotizaciones_documento_final"]); got < 1 {
		t.Fatalf("cotizaciones_documento_final esperado>=1 obtenido=%d", got)
	}
	if got := anyToInt64(summary["alertas_total"]); got < 2 {
		t.Fatalf("alertas_total esperado>=2 obtenido=%d", got)
	}

	items, ok := resp["items"].([]interface{})
	if !ok {
		t.Fatalf("items no disponible en embudo: %s", rr.Body.String())
	}

	encontroCotizacionSLA := false
	encontroPedidoSLA := false
	encontroDocumentoFinal := false
	for _, raw := range items {
		row := raw.(map[string]interface{})
		id := anyToInt64(row["cotizacion_id"])
		if id == cotizacionSLAID {
			alertaTipo := strings.ToLower(genericStringValue(row["alerta_tipo"]))
			if strings.Contains(alertaTipo, "cotizacion") {
				encontroCotizacionSLA = true
			}
		}
		if id == cotizacionPedidoSLAID {
			alertaTipo := strings.ToLower(genericStringValue(row["alerta_tipo"]))
			if strings.Contains(alertaTipo, "pedido_sla_vencido") {
				encontroPedidoSLA = true
			}
		}
		if id == cotizacionCompletaID {
			if got := strings.ToLower(genericStringValue(row["conversion_etapa"])); got == "documento_final" {
				encontroDocumentoFinal = true
			}
		}
	}

	if !encontroCotizacionSLA {
		t.Fatalf("no se detecto alerta de cotizacion SLA/vigencia para cotizacion %d", cotizacionSLAID)
	}
	if !encontroPedidoSLA {
		t.Fatalf("no se detecto alerta pedido_sla_vencido para cotizacion %d", cotizacionPedidoSLAID)
	}
	if !encontroDocumentoFinal {
		t.Fatalf("no se detecto conversion_etapa=documento_final para cotizacion %d", cotizacionCompletaID)
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

func marshalRSAPrivateKeyPEM(t *testing.T, key *rsa.PrivateKey) string {
	t.Helper()
	if key == nil {
		t.Fatalf("rsa key nil")
	}
	block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}
	return string(pem.EncodeToMemory(block))
}

func TestEmpresaDIANColombiaHandlerFirmaEnvioYAcuseReal(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_dian_real_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(59)
	t.Setenv("DIAN_TOKEN_TEST", "token-dian-qa")

	dianServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.TrimSpace(r.Header.Get("Authorization")) != "Bearer token-dian-qa" {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"token invalido"}`))
			return
		}

		switch r.Method {
		case http.MethodPost:
			payload := map[string]interface{}{}
			_ = json.NewDecoder(r.Body).Decode(&payload)
			if strings.TrimSpace(genericStringValue(payload["xml_firmado"])) == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"xml_firmado requerido"}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"acuse":"aceptado","track_id":"TRK-5901"}`))
			return

		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"ok":true,"estado":"aceptado","message":"acuse disponible"}`))
			return
		}

		w.WriteHeader(http.StatusMethodNotAllowed)
	}))
	defer dianServer.Close()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	privateKeyPEM := marshalRSAPrivateKeyPEM(t, privateKey)

	_, err = dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, empresaID, map[string]interface{}{
		"codigo":                "DIAN-REAL-59",
		"nit":                   "900590001",
		"razon_social":          "Empresa DIAN QA 59",
		"tipo_ambiente":         "habilitacion",
		"software_id":           "SW-59",
		"software_pin":          "PIN-59",
		"prefijo":               "FE",
		"resolucion_numero":     "187600000590",
		"rango_desde":           1,
		"rango_hasta":           999999,
		"consecutivo_actual":    1,
		"url_dian":              dianServer.URL,
		"token_emisor_ref":      "env:DIAN_TOKEN_TEST",
		"certificado_clave_ref": privateKeyPEM,
		"estado_dian":           "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow dian: %v", err)
	}

	handler := EmpresaDIANColombiaHandler(dbEmp)

	reqFirma := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=firmar_xml_real", strings.NewReader(`{"empresa_id":59,"documento_codigo":"FV-5901","xml":"<Invoice><ID>FV-5901</ID></Invoice>"}`))
	reqFirma.Header.Set("Content-Type", "application/json")
	rrFirma := httptest.NewRecorder()
	handler.ServeHTTP(rrFirma, reqFirma)
	if rrFirma.Code != http.StatusOK {
		t.Fatalf("firmar_xml_real status=%d body=%s", rrFirma.Code, rrFirma.Body.String())
	}
	respFirma := decodeBodyAsMap(t, rrFirma)
	if ok, _ := respFirma["ok"].(bool); !ok {
		t.Fatalf("firma DIAN no ok: %s", rrFirma.Body.String())
	}
	if strings.TrimSpace(genericStringValue(respFirma["firma_base64"])) == "" {
		t.Fatalf("firma_base64 vacia en respuesta: %s", rrFirma.Body.String())
	}

	xmlFirmado := genericStringValue(respFirma["xml_firmado"])
	reqEnviar := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=enviar_documento_real", strings.NewReader(`{"empresa_id":59,"documento_codigo":"FV-5901","total":"150000","xml_firmado":`+strconv.Quote(xmlFirmado)+`}`))
	reqEnviar.Header.Set("Content-Type", "application/json")
	rrEnviar := httptest.NewRecorder()
	handler.ServeHTTP(rrEnviar, reqEnviar)
	if rrEnviar.Code != http.StatusOK {
		t.Fatalf("enviar_documento_real status=%d body=%s", rrEnviar.Code, rrEnviar.Body.String())
	}
	respEnviar := decodeBodyAsMap(t, rrEnviar)
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(respEnviar["estado_dian"]))); got != "aceptado" {
		t.Fatalf("estado_dian esperado=aceptado obtenido=%s body=%s", got, rrEnviar.Body.String())
	}
	if contingencia, _ := respEnviar["contingencia_activa"].(bool); contingencia {
		t.Fatalf("contingencia_activa no esperada en envio exitoso")
	}

	reqAcuse := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica/dian?action=consultar_acuse_real&empresa_id=59&documento_codigo=FV-5901", nil)
	rrAcuse := httptest.NewRecorder()
	handler.ServeHTTP(rrAcuse, reqAcuse)
	if rrAcuse.Code != http.StatusOK {
		t.Fatalf("consultar_acuse_real status=%d body=%s", rrAcuse.Code, rrAcuse.Body.String())
	}
	respAcuse := decodeBodyAsMap(t, rrAcuse)
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(respAcuse["acuse_estado"]))); got != "aceptado" {
		t.Fatalf("acuse_estado esperado=aceptado obtenido=%s body=%s", got, rrAcuse.Body.String())
	}

	cfg, err := getEmpresaDIANConfig(dbEmp, empresaID)
	if err != nil {
		t.Fatalf("getEmpresaDIANConfig: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(cfg["estado_dian"]))); got != "aceptado" {
		t.Fatalf("estado_dian en DB esperado=aceptado obtenido=%s", got)
	}
}

func TestEmpresaDIANColombiaHandlerContingenciaYReconexion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_dian_contingencia_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(60)
	_, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, empresaID, map[string]interface{}{
		"codigo":             "DIAN-REAL-60",
		"nit":                "900600001",
		"razon_social":       "Empresa DIAN QA 60",
		"tipo_ambiente":      "habilitacion",
		"software_id":        "SW-60",
		"software_pin":       "PIN-60",
		"prefijo":            "FE",
		"resolucion_numero":  "187600000600",
		"rango_desde":        1,
		"rango_hasta":        999999,
		"consecutivo_actual": 1,
		"url_dian":           "http://127.0.0.1:1",
		"estado_dian":        "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow dian contingencia: %v", err)
	}

	handler := EmpresaDIANColombiaHandler(dbEmp)

	reqEnviar := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=enviar_documento_real", strings.NewReader(`{"empresa_id":60,"documento_codigo":"FV-6001","xml_firmado":"<Invoice><ID>FV-6001</ID></Invoice>","total":"1000"}`))
	reqEnviar.Header.Set("Content-Type", "application/json")
	rrEnviar := httptest.NewRecorder()
	handler.ServeHTTP(rrEnviar, reqEnviar)
	if rrEnviar.Code != http.StatusOK {
		t.Fatalf("enviar_documento_real contingencia status=%d body=%s", rrEnviar.Code, rrEnviar.Body.String())
	}
	respEnviar := decodeBodyAsMap(t, rrEnviar)
	if contingencia, _ := respEnviar["contingencia_activa"].(bool); !contingencia {
		t.Fatalf("se esperaba contingencia_activa=true body=%s", rrEnviar.Body.String())
	}

	cfgContingencia, err := getEmpresaDIANConfig(dbEmp, empresaID)
	if err != nil {
		t.Fatalf("getEmpresaDIANConfig contingencia: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(cfgContingencia["estado_dian"]))); got != "contingencia" {
		t.Fatalf("estado_dian esperado=contingencia obtenido=%s", got)
	}

	reconnectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer reconnectServer.Close()

	id := anyToInt64(cfgContingencia["id"])
	if id <= 0 {
		t.Fatalf("id DIAN invalido para actualizar url")
	}
	if err := dbpkg.UpdateEmpresaGenericRow(dbEmp, cfgDIAN.Table, empresaID, id, map[string]interface{}{"url_dian": reconnectServer.URL}, cfgDIAN.AllowedColumns); err != nil {
		t.Fatalf("UpdateEmpresaGenericRow url_dian: %v", err)
	}

	reqRecon := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=reconexion_dian", strings.NewReader(`{"empresa_id":60}`))
	reqRecon.Header.Set("Content-Type", "application/json")
	rrRecon := httptest.NewRecorder()
	handler.ServeHTTP(rrRecon, reqRecon)
	if rrRecon.Code != http.StatusOK {
		t.Fatalf("reconexion_dian status=%d body=%s", rrRecon.Code, rrRecon.Body.String())
	}
	respRecon := decodeBodyAsMap(t, rrRecon)
	if ok, _ := respRecon["ok"].(bool); !ok {
		t.Fatalf("reconexion_dian esperaba ok=true body=%s", rrRecon.Body.String())
	}
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(respRecon["estado_dian"]))); got != "reconectado" {
		t.Fatalf("estado_dian esperado=reconectado obtenido=%s body=%s", got, rrRecon.Body.String())
	}
}

func TestEmpresaDIANColombiaHandlerEnviarSetPruebas(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_dian_set_pruebas_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(61)
	enviados := make([]string, 0)

	dianServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"metodo no permitido"}`))
			return
		}

		payload := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		documento := strings.TrimSpace(genericStringValue(payload["documento_codigo"]))
		tipoDocumento := strings.TrimSpace(genericStringValue(payload["documento_tipo"]))
		if documento == "" || tipoDocumento == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"documento o tipo faltante"}`))
			return
		}
		enviados = append(enviados, documento+"|"+tipoDocumento)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"acuse":"aceptado","message":"ok"}`))
	}))
	defer dianServer.Close()

	_, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, empresaID, map[string]interface{}{
		"codigo":             "DIAN-SET-61",
		"nit":                "900610001",
		"razon_social":       "Empresa Set QA 61",
		"tipo_ambiente":      "habilitacion",
		"software_id":        "SW-61",
		"software_pin":       "PIN-61",
		"test_set_id":        "TESTSET-61",
		"prefijo":            "SETP",
		"resolucion_numero":  "187600000610",
		"rango_desde":        100,
		"rango_hasta":        999,
		"consecutivo_actual": 100,
		"url_dian":           dianServer.URL,
		"estado_dian":        "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow dian set: %v", err)
	}

	handler := EmpresaDIANColombiaHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=enviar_set_pruebas", strings.NewReader(`{"empresa_id":61,"facturas_electronicas":3,"notas_debito":1,"notas_credito":1,"total_documentos":5,"total_por_documento":"2500"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("enviar_set_pruebas status=%d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeBodyAsMap(t, rr)
	if ok, _ := resp["ok"].(bool); !ok {
		t.Fatalf("respuesta no ok: %s", rr.Body.String())
	}
	if got := int(anyToInt64(resp["procesados"])); got != 5 {
		t.Fatalf("procesados esperado=5 obtenido=%d body=%s", got, rr.Body.String())
	}

	resumen, _ := resp["resumen"].(map[string]interface{})
	if got := int(anyToInt64(resumen["aceptado"])); got != 5 {
		t.Fatalf("aceptados esperados=5 obtenido=%d body=%s", got, rr.Body.String())
	}

	detalles, _ := resp["detalles"].([]interface{})
	if len(detalles) != 5 {
		t.Fatalf("detalles esperados=5 obtenido=%d body=%s", len(detalles), rr.Body.String())
	}

	if len(enviados) != 5 {
		t.Fatalf("envios recibidos esperados=5 obtenido=%d", len(enviados))
	}

	cfgActualizada, err := getEmpresaDIANConfig(dbEmp, empresaID)
	if err != nil {
		t.Fatalf("getEmpresaDIANConfig set: %v", err)
	}
	if got := anyToInt64(cfgActualizada["consecutivo_actual"]); got != 105 {
		t.Fatalf("consecutivo_actual esperado=105 obtenido=%d", got)
	}
}

func TestEmpresaDIANColombiaHandlerSoftwareCompartidoMultiempresa(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_dian_software_compartido_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	t.Setenv("DIAN_SHARED_SOFTWARE_ID", "SW-SHARED-01")
	t.Setenv("DIAN_SHARED_SOFTWARE_PIN", "PIN-SHARED-01")
	t.Setenv("DIAN_TOKEN_EMP_62", "token-emp-62")
	t.Setenv("DIAN_TOKEN_EMP_63", "token-emp-63")

	receivedSoftware := make([]string, 0)
	receivedNIT := make([]string, 0)
	receivedTokens := make([]string, 0)

	dianServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"metodo no permitido"}`))
			return
		}

		payload := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&payload)

		nit := strings.TrimSpace(genericStringValue(payload["nit"]))
		softwareID := strings.TrimSpace(genericStringValue(payload["software_id"]))
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))

		receivedNIT = append(receivedNIT, nit)
		receivedSoftware = append(receivedSoftware, softwareID)
		receivedTokens = append(receivedTokens, token)

		expectedToken := ""
		switch nit {
		case "900620001":
			expectedToken = "token-emp-62"
		case "900630001":
			expectedToken = "token-emp-63"
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"nit inesperado"}`))
			return
		}

		if softwareID != "SW-SHARED-01" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"software_id invalido"}`))
			return
		}
		if token != expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"ok":false,"estado":"rechazado","message":"token invalido"}`))
			return
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"acuse":"aceptado","message":"ok"}`))
	}))
	defer dianServer.Close()

	_, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, 62, map[string]interface{}{
		"codigo":                      "DIAN-SH-62",
		"nit":                         "900620001",
		"razon_social":                "Empresa SaaS 62",
		"tipo_ambiente":               "habilitacion",
		"usar_software_compartido":    1,
		"software_id":                 "SW-LOCAL-62",
		"software_pin":                "PIN-LOCAL-62",
		"software_id_compartido_ref":  "env:DIAN_SHARED_SOFTWARE_ID",
		"software_pin_compartido_ref": "env:DIAN_SHARED_SOFTWARE_PIN",
		"prefijo":                     "SETP",
		"resolucion_numero":           "187600000620",
		"rango_desde":                 1,
		"rango_hasta":                 99999,
		"consecutivo_actual":          1,
		"url_dian":                    dianServer.URL,
		"token_emisor_ref":            "env:DIAN_TOKEN_EMP_62",
		"estado_dian":                 "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow empresa 62: %v", err)
	}

	_, err = dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, 63, map[string]interface{}{
		"codigo":                      "DIAN-SH-63",
		"nit":                         "900630001",
		"razon_social":                "Empresa SaaS 63",
		"tipo_ambiente":               "habilitacion",
		"usar_software_compartido":    1,
		"software_id":                 "SW-LOCAL-63",
		"software_pin":                "PIN-LOCAL-63",
		"software_id_compartido_ref":  "env:DIAN_SHARED_SOFTWARE_ID",
		"software_pin_compartido_ref": "env:DIAN_SHARED_SOFTWARE_PIN",
		"prefijo":                     "SETP",
		"resolucion_numero":           "187600000630",
		"rango_desde":                 1,
		"rango_hasta":                 99999,
		"consecutivo_actual":          1,
		"url_dian":                    dianServer.URL,
		"token_emisor_ref":            "env:DIAN_TOKEN_EMP_63",
		"estado_dian":                 "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow empresa 63: %v", err)
	}

	handler := EmpresaDIANColombiaHandler(dbEmp)

	req62 := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=enviar_documento_real", strings.NewReader(`{"empresa_id":62,"documento_codigo":"FV-6201","xml_firmado":"<Invoice><ID>FV-6201</ID></Invoice>","total":"20000"}`))
	req62.Header.Set("Content-Type", "application/json")
	rr62 := httptest.NewRecorder()
	handler.ServeHTTP(rr62, req62)
	if rr62.Code != http.StatusOK {
		t.Fatalf("envio empresa 62 status=%d body=%s", rr62.Code, rr62.Body.String())
	}
	resp62 := decodeBodyAsMap(t, rr62)
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(resp62["software_modo"]))); got != "compartido" {
		t.Fatalf("software_modo esperado=compartido obtenido=%s body=%s", got, rr62.Body.String())
	}
	if got := strings.TrimSpace(genericStringValue(resp62["software_id"])); got != "SW-SHARED-01" {
		t.Fatalf("software_id esperado=SW-SHARED-01 obtenido=%s", got)
	}

	req63 := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=enviar_documento_real", strings.NewReader(`{"empresa_id":63,"documento_codigo":"FV-6301","xml_firmado":"<Invoice><ID>FV-6301</ID></Invoice>","total":"30000"}`))
	req63.Header.Set("Content-Type", "application/json")
	rr63 := httptest.NewRecorder()
	handler.ServeHTTP(rr63, req63)
	if rr63.Code != http.StatusOK {
		t.Fatalf("envio empresa 63 status=%d body=%s", rr63.Code, rr63.Body.String())
	}
	resp63 := decodeBodyAsMap(t, rr63)
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(resp63["software_modo"]))); got != "compartido" {
		t.Fatalf("software_modo esperado=compartido obtenido=%s body=%s", got, rr63.Body.String())
	}
	if got := strings.TrimSpace(genericStringValue(resp63["software_id"])); got != "SW-SHARED-01" {
		t.Fatalf("software_id esperado=SW-SHARED-01 obtenido=%s", got)
	}

	if len(receivedSoftware) != 2 || len(receivedNIT) != 2 || len(receivedTokens) != 2 {
		t.Fatalf("captura de requests incompleta software=%d nit=%d token=%d", len(receivedSoftware), len(receivedNIT), len(receivedTokens))
	}
	if receivedSoftware[0] != "SW-SHARED-01" || receivedSoftware[1] != "SW-SHARED-01" {
		t.Fatalf("software compartido no aplicado en ambos envios: %v", receivedSoftware)
	}
	if !(receivedNIT[0] != receivedNIT[1]) {
		t.Fatalf("se esperaban NIT distintos por empresa, obtenido=%v", receivedNIT)
	}
}

func TestEmpresaDIANColombiaHandlerGuiaOnboardingYValidarCredenciales(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_dian_onboarding_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	t.Setenv("DIAN_SHARED_SOFTWARE_ID", "SW-SHARED-ONBOARD")
	t.Setenv("DIAN_SHARED_SOFTWARE_PIN", "PIN-SHARED-ONBOARD")
	t.Setenv("DIAN_TOKEN_EMP_64", "token-emp-64")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	privateKeyPEM := marshalRSAPrivateKeyPEM(t, privateKey)

	_, err = dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, 64, map[string]interface{}{
		"codigo":                      "DIAN-ONBOARD-64",
		"nit":                         "900640001",
		"razon_social":                "Empresa Onboarding 64",
		"tipo_ambiente":               "habilitacion",
		"usar_software_compartido":    1,
		"software_id_compartido_ref":  "env:DIAN_SHARED_SOFTWARE_ID",
		"software_pin_compartido_ref": "env:DIAN_SHARED_SOFTWARE_PIN",
		"prefijo":                     "SETP",
		"resolucion_numero":           "187600000640",
		"rango_desde":                 1,
		"rango_hasta":                 99999,
		"consecutivo_actual":          1,
		"url_dian":                    "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
		"token_emisor_ref":            "env:DIAN_TOKEN_EMP_64",
		"certificado_clave_ref":       privateKeyPEM,
		"estado_dian":                 "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow empresa 64: %v", err)
	}

	handler := EmpresaDIANColombiaHandler(dbEmp)

	reqGuia := httptest.NewRequest(http.MethodGet, "/api/empresa/facturacion_electronica/dian?action=guia_onboarding&empresa_id=64", nil)
	rrGuia := httptest.NewRecorder()
	handler.ServeHTTP(rrGuia, reqGuia)
	if rrGuia.Code != http.StatusOK {
		t.Fatalf("guia_onboarding status=%d body=%s", rrGuia.Code, rrGuia.Body.String())
	}
	respGuia := decodeBodyAsMap(t, rrGuia)
	if got := strings.ToLower(strings.TrimSpace(genericStringValue(respGuia["software_modo"]))); got != "compartido" {
		t.Fatalf("software_modo esperado=compartido obtenido=%s", got)
	}
	pasos, _ := respGuia["pasos"].([]interface{})
	if len(pasos) < 5 {
		t.Fatalf("guia_onboarding debe devolver pasos operativos, obtenido=%d", len(pasos))
	}

	reqValidar := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=validar_credenciales", strings.NewReader(`{"empresa_id":64}`))
	reqValidar.Header.Set("Content-Type", "application/json")
	rrValidar := httptest.NewRecorder()
	handler.ServeHTTP(rrValidar, reqValidar)
	if rrValidar.Code != http.StatusOK {
		t.Fatalf("validar_credenciales status=%d body=%s", rrValidar.Code, rrValidar.Body.String())
	}
	respValidar := decodeBodyAsMap(t, rrValidar)
	if ok, _ := respValidar["ok"].(bool); !ok {
		t.Fatalf("validar_credenciales debe retornar ok=true body=%s", rrValidar.Body.String())
	}
	checks, _ := respValidar["checks"].(map[string]interface{})
	firma, _ := checks["firma_digital"].(map[string]interface{})
	if okFirma, _ := firma["ok"].(bool); !okFirma {
		t.Fatalf("firma_digital debe ser valida body=%s", rrValidar.Body.String())
	}
}

func TestEmpresaDIANColombiaHandlerSubirFirma(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_dian_subir_firma_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	_, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDIAN.Table, 65, map[string]interface{}{
		"codigo":             "DIAN-UP-65",
		"nit":                "900650001",
		"razon_social":       "Empresa Upload 65",
		"tipo_ambiente":      "habilitacion",
		"software_id":        "SW-65",
		"software_pin":       "PIN-65",
		"prefijo":            "SETP",
		"resolucion_numero":  "187600000650",
		"rango_desde":        1,
		"rango_hasta":        99999,
		"consecutivo_actual": 1,
		"url_dian":           "https://vpfe-hab.dian.gov.co/WcfDianCustomerServices.svc?wsdl",
		"estado_dian":        "pendiente",
	}, cfgDIAN.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow empresa 65: %v", err)
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	privateKeyPEM := marshalRSAPrivateKeyPEM(t, privateKey)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("empresa_id", "65"); err != nil {
		t.Fatalf("writer.WriteField empresa_id: %v", err)
	}
	part, err := writer.CreateFormFile("archivo_firma", "empresa65.pem")
	if err != nil {
		t.Fatalf("writer.CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte(privateKeyPEM)); err != nil {
		t.Fatalf("part.Write: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("writer.Close: %v", err)
	}

	handler := EmpresaDIANColombiaHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/facturacion_electronica/dian?action=subir_firma", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("subir_firma status=%d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeBodyAsMap(t, rr)
	if ok, _ := resp["ok"].(bool); !ok {
		t.Fatalf("subir_firma debe retornar ok=true body=%s", rr.Body.String())
	}
	ref := strings.TrimSpace(genericStringValue(resp["certificado_clave_ref"]))
	if !strings.HasPrefix(strings.ToLower(ref), "file:") {
		t.Fatalf("certificado_clave_ref debe guardarse como file:, obtenido=%s", ref)
	}
	absPath := strings.TrimSpace(ref[5:])
	if absPath == "" {
		t.Fatalf("ruta file: vacia en certificado_clave_ref")
	}
	if _, err := os.Stat(absPath); err != nil {
		t.Fatalf("archivo de firma no existe en ruta guardada: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(absPath)
		_ = os.Remove(filepath.Dir(absPath))
	})

	cfg, err := getEmpresaDIANConfig(dbEmp, 65)
	if err != nil {
		t.Fatalf("getEmpresaDIANConfig empresa 65: %v", err)
	}
	if got := strings.TrimSpace(genericStringValue(cfg["certificado_clave_ref"])); got != ref {
		t.Fatalf("certificado_clave_ref en DB no coincide, esperado=%s obtenido=%s", ref, got)
	}
}

func TestEmpresaFinanzasPlanCuentasPlantillasYAplicacion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_finanzas_plan_cuentas_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(120)
	handler := EmpresaFinanzasPlanCuentasHandler(dbEmp)

	reqPlantillas := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/plan_cuentas?action=plantillas&tipo_empresa=hotel", nil)
	rrPlantillas := httptest.NewRecorder()
	handler.ServeHTTP(rrPlantillas, reqPlantillas)
	if rrPlantillas.Code != http.StatusOK {
		t.Fatalf("plantillas plan_cuentas status=%d body=%s", rrPlantillas.Code, rrPlantillas.Body.String())
	}
	respPlantillas := decodeBodyAsMap(t, rrPlantillas)
	if got := strings.ToLower(genericStringValue(respPlantillas["tipo_empresa"])); got != "hotel" {
		t.Fatalf("tipo_empresa esperado=hotel obtenido=%s", got)
	}
	itemsRaw, ok := respPlantillas["items"].([]interface{})
	if !ok || len(itemsRaw) == 0 {
		t.Fatalf("items de plantilla hotel no disponibles: %s", rrPlantillas.Body.String())
	}
	encontroHospedaje := false
	for _, raw := range itemsRaw {
		row, _ := raw.(map[string]interface{})
		if strings.TrimSpace(genericStringValue(row["Codigo"])) == "413540" || strings.TrimSpace(genericStringValue(row["codigo"])) == "413540" {
			encontroHospedaje = true
			break
		}
	}
	if !encontroHospedaje {
		t.Fatalf("la plantilla hotel debe incluir la cuenta 413540")
	}

	bodyAplicar, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"tipo_empresa": "hotel",
		"sobrescribir": true,
	})
	reqAplicar := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/plan_cuentas?action=aplicar_plantilla&empresa_id=120", strings.NewReader(string(bodyAplicar)))
	reqAplicar.Header.Set("Content-Type", "application/json")
	rrAplicar := httptest.NewRecorder()
	handler.ServeHTTP(rrAplicar, reqAplicar)
	if rrAplicar.Code != http.StatusOK {
		t.Fatalf("aplicar_plantilla status=%d body=%s", rrAplicar.Code, rrAplicar.Body.String())
	}
	respAplicar := decodeBodyAsMap(t, rrAplicar)
	if created := anyToInt64(respAplicar["creadas"]); created <= 0 {
		t.Fatalf("se esperaban cuentas creadas en aplicar_plantilla, obtenido=%d body=%s", created, rrAplicar.Body.String())
	}

	var cuentaHotel int64
	if err := dbEmp.QueryRow(`SELECT COUNT(1)
	FROM empresa_plan_cuentas
	WHERE empresa_id = ?
	  AND codigo = '413540'
	  AND LOWER(COALESCE(estado, 'activo')) = 'activo'`, empresaID).Scan(&cuentaHotel); err != nil {
		t.Fatalf("count cuenta hotel 413540: %v", err)
	}
	if cuentaHotel <= 0 {
		t.Fatalf("se esperaba la cuenta 413540 creada para empresa_id=%d", empresaID)
	}

	bodyAplicar2, _ := json.Marshal(map[string]interface{}{
		"empresa_id":   empresaID,
		"tipo_empresa": "hotel",
	})
	reqAplicar2 := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/plan_cuentas?action=aplicar_plantilla&empresa_id=120", strings.NewReader(string(bodyAplicar2)))
	reqAplicar2.Header.Set("Content-Type", "application/json")
	rrAplicar2 := httptest.NewRecorder()
	handler.ServeHTTP(rrAplicar2, reqAplicar2)
	if rrAplicar2.Code != http.StatusOK {
		t.Fatalf("reaplicar plantilla status=%d body=%s", rrAplicar2.Code, rrAplicar2.Body.String())
	}
	respAplicar2 := decodeBodyAsMap(t, rrAplicar2)
	if omitidas := anyToInt64(respAplicar2["omitidas"]); omitidas <= 0 {
		t.Fatalf("se esperaban cuentas omitidas al reaplicar sin sobrescribir, obtenido=%d body=%s", omitidas, rrAplicar2.Body.String())
	}
}

func TestEmpresaFinanzasCuentasCobrarConciliacionPagosReales(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_finanzas_cxc_conciliacion_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(121)
	cxcID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCxC.Table, empresaID, map[string]interface{}{
		"codigo":            "CXC-121-001",
		"cliente_nombre":    "Cliente Conciliacion",
		"documento_codigo":  "FAC-121-001",
		"fecha_emision":     "2026-04-05",
		"fecha_vencimiento": "2026-04-30",
		"periodo_contable":  "2026-04",
		"valor_original":    1000,
		"saldo":             1000,
		"estado_cartera":    "pendiente",
	}, cfgCxC.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow CxC: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO empresa_finanzas_movimientos (
		empresa_id,
		tipo_movimiento,
		codigo,
		fecha_movimiento,
		periodo_contable,
		total,
		total_neto,
		tercero_nombre,
		referencia_externa,
		numero_comprobante,
		estado,
		usuario_creador
	) VALUES (?, 'ingreso', 'ING-121-001', '2026-04-20 11:00:00', '2026-04', 400, 400, 'Cliente Conciliacion', 'FAC-121-001', 'FAC-121-001', 'activo', 'qa_mod20')`, empresaID); err != nil {
		t.Fatalf("insert movimiento ingreso conciliacion: %v", err)
	}

	handler := EmpresaFinanzasCuentasCobrarHandler(dbEmp)
	req := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/cuentas_cobrar?action=conciliar_pagos&empresa_id=121&periodo=2026-04", strings.NewReader(`{"empresa_id":121}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("conciliar_pagos CxC status=%d body=%s", rr.Code, rr.Body.String())
	}
	resp := decodeBodyAsMap(t, rr)
	if conciliados := anyToInt64(resp["conciliados"]); conciliados <= 0 {
		t.Fatalf("se esperaba al menos 1 registro conciliado, obtenido=%d body=%s", conciliados, rr.Body.String())
	}

	var valorPagado float64
	var saldo float64
	var estadoCartera string
	var conciliadoEn string
	var referenciaPagos string
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(valor_pagado, 0),
		COALESCE(saldo, 0),
		COALESCE(estado_cartera, ''),
		COALESCE(conciliado_en, ''),
		COALESCE(referencia_pagos_json, '')
	FROM empresa_cuentas_por_cobrar
	WHERE empresa_id = ? AND id = ?`, empresaID, cxcID).Scan(&valorPagado, &saldo, &estadoCartera, &conciliadoEn, &referenciaPagos); err != nil {
		t.Fatalf("query CxC conciliada: %v", err)
	}
	if valorPagado < 399.99 || valorPagado > 400.01 {
		t.Fatalf("valor_pagado esperado=400 obtenido=%.2f", valorPagado)
	}
	if saldo < 599.99 || saldo > 600.01 {
		t.Fatalf("saldo esperado=600 obtenido=%.2f", saldo)
	}
	if got := strings.ToLower(strings.TrimSpace(estadoCartera)); got != "parcial" {
		t.Fatalf("estado_cartera esperado=parcial obtenido=%s", got)
	}
	if strings.TrimSpace(conciliadoEn) == "" {
		t.Fatalf("conciliado_en debe registrarse tras conciliacion automatica")
	}
	if !strings.Contains(strings.ToUpper(referenciaPagos), "ING-121-001") {
		t.Fatalf("referencia_pagos_json debe incluir movimiento ING-121-001, obtenido=%s", referenciaPagos)
	}
}

func TestEmpresaFinanzasCarteraBloqueoPeriodoCerrado(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_finanzas_cartera_bloqueo_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(122)
	handler := EmpresaFinanzasCuentasPagarHandler(dbEmp)

	if err := dbpkg.SetEmpresaFinanzasPeriodoEstado(dbEmp, empresaID, "2026-03", "cerrado", "qa_mod20", "cierre mensual QA"); err != nil {
		t.Fatalf("SetEmpresaFinanzasPeriodoEstado 2026-03: %v", err)
	}

	bodyCreateClosed, _ := json.Marshal(map[string]interface{}{
		"empresa_id":        empresaID,
		"codigo":            "CXP-122-001",
		"proveedor_nombre":  "Proveedor Periodo Cerrado",
		"documento_codigo":  "FCXP-122-001",
		"fecha_emision":     "2026-03-10",
		"fecha_vencimiento": "2026-03-25",
		"periodo_contable":  "2026-03",
		"valor_original":    900,
		"saldo":             900,
	})
	reqCreateClosed := httptest.NewRequest(http.MethodPost, "/api/empresa/finanzas/cuentas_pagar?empresa_id=122", strings.NewReader(string(bodyCreateClosed)))
	reqCreateClosed.Header.Set("Content-Type", "application/json")
	rrCreateClosed := httptest.NewRecorder()
	handler.ServeHTTP(rrCreateClosed, reqCreateClosed)
	if rrCreateClosed.Code != http.StatusConflict {
		t.Fatalf("create CxP en periodo cerrado debe retornar 409, obtenido=%d body=%s", rrCreateClosed.Code, rrCreateClosed.Body.String())
	}

	cxpID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgCxP.Table, empresaID, map[string]interface{}{
		"codigo":            "CXP-122-002",
		"proveedor_nombre":  "Proveedor Bloqueo",
		"documento_codigo":  "FCXP-122-002",
		"fecha_emision":     "2026-04-10",
		"fecha_vencimiento": "2026-04-20",
		"periodo_contable":  "2026-04",
		"valor_original":    500,
		"saldo":             500,
	}, cfgCxP.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow CxP abierto: %v", err)
	}

	if err := dbpkg.SetEmpresaFinanzasPeriodoEstado(dbEmp, empresaID, "2026-04", "cerrado", "qa_mod20", "cierre mensual QA"); err != nil {
		t.Fatalf("SetEmpresaFinanzasPeriodoEstado 2026-04: %v", err)
	}

	reqValidar := httptest.NewRequest(http.MethodGet, "/api/empresa/finanzas/cuentas_pagar?action=validar_cierre_periodo&empresa_id=122&periodo=2026-04", nil)
	rrValidar := httptest.NewRecorder()
	handler.ServeHTTP(rrValidar, reqValidar)
	if rrValidar.Code != http.StatusOK {
		t.Fatalf("validar_cierre_periodo status=%d body=%s", rrValidar.Code, rrValidar.Body.String())
	}
	respValidar := decodeBodyAsMap(t, rrValidar)
	if cerrado, _ := respValidar["cerrado"].(bool); !cerrado {
		t.Fatalf("validar_cierre_periodo esperaba cerrado=true body=%s", rrValidar.Body.String())
	}

	bodyUpdate, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         cxpID,
		"saldo":      100,
	})
	reqUpdate := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cuentas_pagar?empresa_id=122&id="+strconv.FormatInt(cxpID, 10), strings.NewReader(string(bodyUpdate)))
	reqUpdate.Header.Set("Content-Type", "application/json")
	rrUpdate := httptest.NewRecorder()
	handler.ServeHTTP(rrUpdate, reqUpdate)
	if rrUpdate.Code != http.StatusConflict {
		t.Fatalf("update CxP en periodo cerrado debe retornar 409, obtenido=%d body=%s", rrUpdate.Code, rrUpdate.Body.String())
	}

	reqEstado := httptest.NewRequest(http.MethodPut, "/api/empresa/finanzas/cuentas_pagar?empresa_id=122&id="+strconv.FormatInt(cxpID, 10)+"&action=desactivar", nil)
	rrEstado := httptest.NewRecorder()
	handler.ServeHTTP(rrEstado, reqEstado)
	if rrEstado.Code != http.StatusConflict {
		t.Fatalf("cambio de estado CxP en periodo cerrado debe retornar 409, obtenido=%d body=%s", rrEstado.Code, rrEstado.Body.String())
	}

	reqDelete := httptest.NewRequest(http.MethodDelete, "/api/empresa/finanzas/cuentas_pagar?empresa_id=122&id="+strconv.FormatInt(cxpID, 10), nil)
	rrDelete := httptest.NewRecorder()
	handler.ServeHTTP(rrDelete, reqDelete)
	if rrDelete.Code != http.StatusConflict {
		t.Fatalf("delete CxP en periodo cerrado debe retornar 409, obtenido=%d body=%s", rrDelete.Code, rrDelete.Body.String())
	}
}

func TestEmpresaInventarioLotesSeriesBloqueoAutomaticoVencido(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_lotes_vencidos_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(123)
	loteID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgLotesSeries.Table, empresaID, map[string]interface{}{
		"producto_id":         901,
		"bodega_id":           1,
		"codigo_lote_serie":   "LOT-EXP-001",
		"fecha_vencimiento":   "2000-01-01",
		"cantidad_inicial":    10,
		"cantidad_disponible": 10,
		"estado_lote":         "activo",
	}, cfgLotesSeries.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow lotes vencidos: %v", err)
	}

	handler := EmpresaInventarioLotesSeriesHandler(dbEmp)
	bodyReserva, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         loteID,
		"cantidad":   1,
	})
	reqReserva := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/lotes_series?action=reservar&empresa_id=123", strings.NewReader(string(bodyReserva)))
	reqReserva.Header.Set("Content-Type", "application/json")
	rrReserva := httptest.NewRecorder()
	handler.ServeHTTP(rrReserva, reqReserva)
	if rrReserva.Code != http.StatusConflict {
		t.Fatalf("reserva en lote vencido debe retornar 409, obtenido=%d body=%s", rrReserva.Code, rrReserva.Body.String())
	}

	var estadoLote string
	var bloqueadoVenta int64
	var bloqueoMotivo string
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(estado_lote, ''),
		COALESCE(bloqueado_venta, 0),
		COALESCE(bloqueo_motivo, '')
	FROM inventario_lotes_series
	WHERE empresa_id = ? AND id = ?`, empresaID, loteID).Scan(&estadoLote, &bloqueadoVenta, &bloqueoMotivo); err != nil {
		t.Fatalf("query lote vencido actualizado: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(estadoLote)); got != "vencido" {
		t.Fatalf("estado_lote esperado=vencido obtenido=%s", got)
	}
	if bloqueadoVenta != 1 {
		t.Fatalf("bloqueado_venta esperado=1 obtenido=%d", bloqueadoVenta)
	}
	if strings.TrimSpace(bloqueoMotivo) == "" {
		t.Fatalf("bloqueo_motivo debe registrarse para lote vencido")
	}

	reqValidar := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/lotes_series?action=validar_disponibilidad&empresa_id=123&id="+strconv.FormatInt(loteID, 10)+"&cantidad=1", nil)
	rrValidar := httptest.NewRecorder()
	handler.ServeHTTP(rrValidar, reqValidar)
	if rrValidar.Code != http.StatusOK {
		t.Fatalf("validar_disponibilidad status=%d body=%s", rrValidar.Code, rrValidar.Body.String())
	}
	respValidar := decodeBodyAsMap(t, rrValidar)
	if disponible, _ := respValidar["disponible_para_reserva"].(bool); disponible {
		t.Fatalf("disponible_para_reserva debe ser false para lote vencido")
	}
}

func TestEmpresaInventarioLotesSeriesTrazabilidadCicloVenta(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_lotes_trazabilidad_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(124)
	loteID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgLotesSeries.Table, empresaID, map[string]interface{}{
		"producto_id":         902,
		"bodega_id":           2,
		"codigo_lote_serie":   "LOT-TRZ-001",
		"fecha_vencimiento":   "2099-12-31",
		"cantidad_inicial":    10,
		"cantidad_disponible": 10,
		"estado_lote":         "activo",
	}, cfgLotesSeries.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow lotes trazabilidad: %v", err)
	}

	handler := EmpresaInventarioLotesSeriesHandler(dbEmp)

	bodyReserva, _ := json.Marshal(map[string]interface{}{
		"empresa_id":        empresaID,
		"id":                loteID,
		"cantidad":          3,
		"referencia_tipo":   "carrito",
		"referencia_codigo": "CAR-124-001",
		"cliente_id":        501,
		"cliente_nombre":    "Cliente Lote",
	})
	reqReserva := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/lotes_series?action=reservar&empresa_id=124", strings.NewReader(string(bodyReserva)))
	reqReserva.Header.Set("Content-Type", "application/json")
	rrReserva := httptest.NewRecorder()
	handler.ServeHTTP(rrReserva, reqReserva)
	if rrReserva.Code != http.StatusOK {
		t.Fatalf("reserva lote status=%d body=%s", rrReserva.Code, rrReserva.Body.String())
	}

	bodyVenta, _ := json.Marshal(map[string]interface{}{
		"empresa_id":        empresaID,
		"id":                loteID,
		"cantidad":          2,
		"referencia_tipo":   "factura",
		"referencia_codigo": "FAC-124-001",
		"cliente_id":        501,
		"cliente_nombre":    "Cliente Lote",
	})
	reqVenta := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/lotes_series?action=vender&empresa_id=124", strings.NewReader(string(bodyVenta)))
	reqVenta.Header.Set("Content-Type", "application/json")
	rrVenta := httptest.NewRecorder()
	handler.ServeHTTP(rrVenta, reqVenta)
	if rrVenta.Code != http.StatusOK {
		t.Fatalf("venta lote status=%d body=%s", rrVenta.Code, rrVenta.Body.String())
	}

	bodyLiberar, _ := json.Marshal(map[string]interface{}{
		"empresa_id":        empresaID,
		"id":                loteID,
		"cantidad":          1,
		"referencia_tipo":   "carrito",
		"referencia_codigo": "CAR-124-001",
	})
	reqLiberar := httptest.NewRequest(http.MethodPost, "/api/empresa/inventario/lotes_series?action=liberar_reserva&empresa_id=124", strings.NewReader(string(bodyLiberar)))
	reqLiberar.Header.Set("Content-Type", "application/json")
	rrLiberar := httptest.NewRecorder()
	handler.ServeHTTP(rrLiberar, reqLiberar)
	if rrLiberar.Code != http.StatusOK {
		t.Fatalf("liberar_reserva lote status=%d body=%s", rrLiberar.Code, rrLiberar.Body.String())
	}

	var disponible float64
	var reservado float64
	var vendido float64
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(cantidad_disponible, 0),
		COALESCE(reservado_cantidad, 0),
		COALESCE(vendido_cantidad, 0)
	FROM inventario_lotes_series
	WHERE empresa_id = ? AND id = ?`, empresaID, loteID).Scan(&disponible, &reservado, &vendido); err != nil {
		t.Fatalf("query lote post-operaciones: %v", err)
	}
	if disponible < 7.99 || disponible > 8.01 {
		t.Fatalf("cantidad_disponible esperada=8 obtenida=%.2f", disponible)
	}
	if reservado < -0.01 || reservado > 0.01 {
		t.Fatalf("reservado_cantidad esperada=0 obtenida=%.2f", reservado)
	}
	if vendido < 1.99 || vendido > 2.01 {
		t.Fatalf("vendido_cantidad esperado=2 obtenido=%.2f", vendido)
	}

	reqTrazabilidad := httptest.NewRequest(http.MethodGet, "/api/empresa/inventario/lotes_series?action=trazabilidad&empresa_id=124&id="+strconv.FormatInt(loteID, 10), nil)
	rrTrazabilidad := httptest.NewRecorder()
	handler.ServeHTTP(rrTrazabilidad, reqTrazabilidad)
	if rrTrazabilidad.Code != http.StatusOK {
		t.Fatalf("trazabilidad lote status=%d body=%s", rrTrazabilidad.Code, rrTrazabilidad.Body.String())
	}
	respTrazabilidad := decodeBodyAsMap(t, rrTrazabilidad)
	if totalMov := anyToInt64(respTrazabilidad["total_movimientos"]); totalMov < 3 {
		t.Fatalf("trazabilidad esperaba al menos 3 movimientos, obtenido=%d body=%s", totalMov, rrTrazabilidad.Body.String())
	}
}

func TestEmpresaComprasDevolucionesProveedorContabilizarImpactoCompleto(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_devolucion_proveedor_contable_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(125)
	devolucionID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgDevProveedor.Table, empresaID, map[string]interface{}{
		"codigo":                  "DPROV-125-001",
		"proveedor_nombre":        "Proveedor Contable",
		"documento_compra_codigo": "OC-125-001",
		"fecha_devolucion":        "2026-04-07",
		"motivo":                  "Devolucion por calidad",
		"estado_devolucion":       "aprobada",
		"subtotal":                800,
		"impuesto_total":          200,
		"total":                   1000,
		"moneda":                  "COP",
	}, cfgDevProveedor.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow devolucion proveedor: %v", err)
	}

	handler := EmpresaComprasDevolucionesProveedorHandler(dbEmp)
	bodyContabilizar, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         devolucionID,
	})
	reqContabilizar := httptest.NewRequest(http.MethodPost, "/api/empresa/compras/devoluciones_proveedor?action=contabilizar&empresa_id=125", strings.NewReader(string(bodyContabilizar)))
	reqContabilizar.Header.Set("Content-Type", "application/json")
	rrContabilizar := httptest.NewRecorder()
	handler.ServeHTTP(rrContabilizar, reqContabilizar)
	if rrContabilizar.Code != http.StatusOK {
		t.Fatalf("contabilizar devolucion proveedor status=%d body=%s", rrContabilizar.Code, rrContabilizar.Body.String())
	}
	respContabilizar := decodeBodyAsMap(t, rrContabilizar)
	movimientoID := anyToInt64(respContabilizar["impacto_contable_movimiento_id"])
	eventoID := anyToInt64(respContabilizar["impacto_contable_evento_id"])
	if movimientoID <= 0 {
		t.Fatalf("impacto_contable_movimiento_id debe ser > 0, obtenido=%d body=%s", movimientoID, rrContabilizar.Body.String())
	}
	if eventoID <= 0 {
		t.Fatalf("impacto_contable_evento_id debe ser > 0, obtenido=%d body=%s", eventoID, rrContabilizar.Body.String())
	}

	var estadoDevolucion string
	var periodoContable string
	var totalReintegrado float64
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(estado_devolucion, ''),
		COALESCE(periodo_contable, ''),
		COALESCE(total_reintegrado, 0)
	FROM empresa_devoluciones_proveedor
	WHERE empresa_id = ? AND id = ?`, empresaID, devolucionID).Scan(&estadoDevolucion, &periodoContable, &totalReintegrado); err != nil {
		t.Fatalf("query devolucion proveedor contabilizada: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(estadoDevolucion)); got != "contabilizada" {
		t.Fatalf("estado_devolucion esperado=contabilizada obtenido=%s", got)
	}
	if strings.TrimSpace(periodoContable) == "" {
		t.Fatalf("periodo_contable debe registrarse en devolucion contabilizada")
	}
	if totalReintegrado < 999.99 || totalReintegrado > 1000.01 {
		t.Fatalf("total_reintegrado esperado=1000 obtenido=%.2f", totalReintegrado)
	}

	var tipoMovimiento string
	var referenciaExterna string
	var totalMovimiento float64
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(tipo_movimiento, ''),
		COALESCE(referencia_externa, ''),
		COALESCE(total, 0)
	FROM empresa_finanzas_movimientos
	WHERE empresa_id = ? AND id = ?`, empresaID, movimientoID).Scan(&tipoMovimiento, &referenciaExterna, &totalMovimiento); err != nil {
		t.Fatalf("query movimiento financiero devolucion: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(tipoMovimiento)); got != "ingreso" {
		t.Fatalf("tipo_movimiento esperado=ingreso obtenido=%s", got)
	}
	if strings.TrimSpace(referenciaExterna) != "DPROV-125-001" {
		t.Fatalf("referencia_externa esperada=DPROV-125-001 obtenida=%s", referenciaExterna)
	}
	if totalMovimiento < 999.99 || totalMovimiento > 1000.01 {
		t.Fatalf("total movimiento esperado=1000 obtenido=%.2f", totalMovimiento)
	}

	var eventosCount int64
	if err := dbEmp.QueryRow(`SELECT COALESCE(COUNT(1), 0)
	FROM empresa_eventos_contables
	WHERE empresa_id = ?
	  AND entidad_id = ?
	  AND LOWER(COALESCE(entidad, '')) = 'devolucion_proveedor'
	  AND LOWER(COALESCE(evento, '')) = 'devolucion_proveedor_contabilizada'`, empresaID, devolucionID).Scan(&eventosCount); err != nil {
		t.Fatalf("count eventos contables devolucion proveedor: %v", err)
	}
	if eventosCount <= 0 {
		t.Fatalf("se esperaba al menos un evento contable para devolucion_proveedor, obtenido=%d", eventosCount)
	}
}

func TestEmpresaRRHHVacacionesSaldoYAprobacionJerarquica(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_rrhh_aprobacion_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaNominaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaNominaSchema: %v", err)
	}

	empresaID := int64(126)
	resNomina, err := dbEmp.Exec(`INSERT INTO empresa_nomina_empleados (
		empresa_id,
		empleado_id,
		empleado_nombre,
		fecha_ingreso,
		salario_basico_mensual,
		estado,
		usuario_creador
	) VALUES (?, ?, ?, ?, ?, 'activo', 'qa_rrhh')`, empresaID, 4001, "Empleado RRHH QA", "2024-01-01", 2200000)
	if err != nil {
		t.Fatalf("insert empleado nomina RRHH: %v", err)
	}
	empleadoNominaID, _ := resNomina.LastInsertId()

	rrhhID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgRRHHVacLic.Table, empresaID, map[string]interface{}{
		"codigo":                     "RRHH-126-001",
		"empleado_id":                4001,
		"empleado_nomina_id":         empleadoNominaID,
		"empleado_nombre":            "Empleado RRHH QA",
		"tipo_novedad":               "vacacion",
		"fecha_inicio":               "2026-04-10",
		"fecha_fin":                  "2026-04-14",
		"dias":                       5,
		"estado_novedad":             "solicitada",
		"nivel_aprobacion_actual":    0,
		"nivel_aprobacion_requerido": 2,
	}, cfgRRHHVacLic.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow RRHH aprobacion: %v", err)
	}

	handler := EmpresaRRHHVacacionesLicenciasHandler(dbEmp)
	reqSaldo := httptest.NewRequest(http.MethodGet, "/api/empresa/rrhh/vacaciones_licencias?action=resumen_saldo&empresa_id=126&id="+strconv.FormatInt(rrhhID, 10)+"&fecha_corte=2026-04-14", nil)
	rrSaldo := httptest.NewRecorder()
	handler.ServeHTTP(rrSaldo, reqSaldo)
	if rrSaldo.Code != http.StatusOK {
		t.Fatalf("resumen_saldo RRHH status=%d body=%s", rrSaldo.Code, rrSaldo.Body.String())
	}
	respSaldo := decodeBodyAsMap(t, rrSaldo)
	saldo, ok := respSaldo["saldo"].(map[string]interface{})
	if !ok {
		t.Fatalf("saldo no disponible en respuesta RRHH: %s", rrSaldo.Body.String())
	}
	if diasAcumulados := ventasAnyToFloat64(saldo["dias_acumulados"]); diasAcumulados <= 0 {
		t.Fatalf("dias_acumulados debe ser > 0, obtenido=%.2f", diasAcumulados)
	}

	bodyAprobarNivel1, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         rrhhID,
		"comentario": "aprobacion nivel 1",
	})
	reqAprobarNivel1 := httptest.NewRequest(http.MethodPost, "/api/empresa/rrhh/vacaciones_licencias?action=aprobar&empresa_id=126", strings.NewReader(string(bodyAprobarNivel1)))
	reqAprobarNivel1.Header.Set("Content-Type", "application/json")
	rrAprobarNivel1 := httptest.NewRecorder()
	handler.ServeHTTP(rrAprobarNivel1, reqAprobarNivel1)
	if rrAprobarNivel1.Code != http.StatusOK {
		t.Fatalf("aprobar nivel 1 RRHH status=%d body=%s", rrAprobarNivel1.Code, rrAprobarNivel1.Body.String())
	}
	respAprobarNivel1 := decodeBodyAsMap(t, rrAprobarNivel1)
	if nivelActual := anyToInt64(respAprobarNivel1["nivel_actual"]); nivelActual != 1 {
		t.Fatalf("nivel_actual esperado=1 tras primera aprobacion, obtenido=%d", nivelActual)
	}
	if estado := strings.ToLower(strings.TrimSpace(genericStringValue(respAprobarNivel1["estado_novedad"]))); estado != "en_aprobacion" {
		t.Fatalf("estado_novedad esperado=en_aprobacion tras nivel 1, obtenido=%s", estado)
	}

	bodyAprobarNivel2, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         rrhhID,
		"comentario": "aprobacion nivel 2",
	})
	reqAprobarNivel2 := httptest.NewRequest(http.MethodPost, "/api/empresa/rrhh/vacaciones_licencias?action=aprobar&empresa_id=126", strings.NewReader(string(bodyAprobarNivel2)))
	reqAprobarNivel2.Header.Set("Content-Type", "application/json")
	rrAprobarNivel2 := httptest.NewRecorder()
	handler.ServeHTTP(rrAprobarNivel2, reqAprobarNivel2)
	if rrAprobarNivel2.Code != http.StatusOK {
		t.Fatalf("aprobar nivel 2 RRHH status=%d body=%s", rrAprobarNivel2.Code, rrAprobarNivel2.Body.String())
	}
	respAprobarNivel2 := decodeBodyAsMap(t, rrAprobarNivel2)
	if nivelActual := anyToInt64(respAprobarNivel2["nivel_actual"]); nivelActual != 2 {
		t.Fatalf("nivel_actual esperado=2 tras segunda aprobacion, obtenido=%d", nivelActual)
	}
	if estado := strings.ToLower(strings.TrimSpace(genericStringValue(respAprobarNivel2["estado_novedad"]))); estado != "aprobada" {
		t.Fatalf("estado_novedad esperado=aprobada tras nivel 2, obtenido=%s", estado)
	}

	var estadoNovedad string
	var nivelActualDB int64
	var saldoAntes float64
	var saldoDespues float64
	var saldoSnapshotJSON string
	var historialJSON string
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(estado_novedad, ''),
		COALESCE(nivel_aprobacion_actual, 0),
		COALESCE(saldo_dias_antes, 0),
		COALESCE(saldo_dias_despues, 0),
		COALESCE(saldo_snapshot_json, ''),
		COALESCE(historial_aprobaciones_json, '')
	FROM empresa_rrhh_vacaciones_licencias
	WHERE empresa_id = ? AND id = ?`, empresaID, rrhhID).Scan(&estadoNovedad, &nivelActualDB, &saldoAntes, &saldoDespues, &saldoSnapshotJSON, &historialJSON); err != nil {
		t.Fatalf("query RRHH aprobacion final: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(estadoNovedad)); got != "aprobada" {
		t.Fatalf("estado_novedad en DB esperado=aprobada, obtenido=%s", got)
	}
	if nivelActualDB != 2 {
		t.Fatalf("nivel_aprobacion_actual en DB esperado=2, obtenido=%d", nivelActualDB)
	}
	if saldoAntes < saldoDespues {
		t.Fatalf("saldo_dias_antes debe ser >= saldo_dias_despues, antes=%.2f despues=%.2f", saldoAntes, saldoDespues)
	}
	if strings.TrimSpace(saldoSnapshotJSON) == "" {
		t.Fatalf("saldo_snapshot_json debe registrarse tras aprobacion final")
	}
	if strings.TrimSpace(historialJSON) == "" {
		t.Fatalf("historial_aprobaciones_json debe registrarse en aprobaciones RRHH")
	}
}

func TestEmpresaRRHHVacacionesVincularNominaPeriodo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_rrhh_vincular_nomina_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaNominaSchema(dbEmp); err != nil {
		t.Fatalf("EnsureEmpresaNominaSchema: %v", err)
	}

	empresaID := int64(127)
	resNomina, err := dbEmp.Exec(`INSERT INTO empresa_nomina_empleados (
		empresa_id,
		empleado_id,
		empleado_nombre,
		fecha_ingreso,
		salario_basico_mensual,
		estado,
		usuario_creador
	) VALUES (?, ?, ?, ?, ?, 'activo', 'qa_rrhh')`, empresaID, 4101, "Empleado Nomina RRHH", "2024-03-01", 2400000)
	if err != nil {
		t.Fatalf("insert empleado nomina vincular: %v", err)
	}
	empleadoNominaID, _ := resNomina.LastInsertId()

	resLiquidacion, err := dbEmp.Exec(`INSERT INTO empresa_nomina_liquidaciones (
		empresa_id,
		empleado_nomina_id,
		empleado_id,
		empleado_nombre,
		periodo_desde,
		periodo_hasta,
		dias_liquidados,
		neto_pagar,
		estado,
		usuario_creador
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'activo', 'qa_rrhh')`, empresaID, empleadoNominaID, 4101, "Empleado Nomina RRHH", "2026-04-01", "2026-04-15", 15, 1500000)
	if err != nil {
		t.Fatalf("insert liquidacion nomina RRHH: %v", err)
	}
	liquidacionID, _ := resLiquidacion.LastInsertId()

	rrhhID, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgRRHHVacLic.Table, empresaID, map[string]interface{}{
		"codigo":                     "RRHH-127-001",
		"empleado_id":                4101,
		"empleado_nomina_id":         empleadoNominaID,
		"empleado_nombre":            "Empleado Nomina RRHH",
		"tipo_novedad":               "vacacion",
		"fecha_inicio":               "2026-04-10",
		"fecha_fin":                  "2026-04-12",
		"dias":                       3,
		"estado_novedad":             "aprobada",
		"nivel_aprobacion_actual":    1,
		"nivel_aprobacion_requerido": 1,
	}, cfgRRHHVacLic.AllowedColumns)
	if err != nil {
		t.Fatalf("CreateEmpresaGenericRow RRHH vincular nomina: %v", err)
	}

	handler := EmpresaRRHHVacacionesLicenciasHandler(dbEmp)
	bodyVincular, _ := json.Marshal(map[string]interface{}{
		"empresa_id": empresaID,
		"id":         rrhhID,
	})
	reqVincular := httptest.NewRequest(http.MethodPost, "/api/empresa/rrhh/vacaciones_licencias?action=vincular_nomina&empresa_id=127", strings.NewReader(string(bodyVincular)))
	reqVincular.Header.Set("Content-Type", "application/json")
	rrVincular := httptest.NewRecorder()
	handler.ServeHTTP(rrVincular, reqVincular)
	if rrVincular.Code != http.StatusOK {
		t.Fatalf("vincular_nomina RRHH status=%d body=%s", rrVincular.Code, rrVincular.Body.String())
	}
	respVincular := decodeBodyAsMap(t, rrVincular)
	if got := anyToInt64(respVincular["nomina_liquidacion_id"]); got != liquidacionID {
		t.Fatalf("nomina_liquidacion_id esperado=%d obtenido=%d body=%s", liquidacionID, got, rrVincular.Body.String())
	}

	var estadoNovedad string
	var nominaLiquidacionID int64
	var nominaPeriodoDesde string
	var nominaPeriodoHasta string
	var nominaVinculadaEn string
	var nominaVinculadaPor string
	if err := dbEmp.QueryRow(`SELECT
		COALESCE(estado_novedad, ''),
		COALESCE(nomina_liquidacion_id, 0),
		COALESCE(nomina_periodo_desde, ''),
		COALESCE(nomina_periodo_hasta, ''),
		COALESCE(nomina_vinculada_en, ''),
		COALESCE(nomina_vinculada_por, '')
	FROM empresa_rrhh_vacaciones_licencias
	WHERE empresa_id = ? AND id = ?`, empresaID, rrhhID).Scan(&estadoNovedad, &nominaLiquidacionID, &nominaPeriodoDesde, &nominaPeriodoHasta, &nominaVinculadaEn, &nominaVinculadaPor); err != nil {
		t.Fatalf("query RRHH vinculada nomina: %v", err)
	}
	if got := strings.ToLower(strings.TrimSpace(estadoNovedad)); got != "contabilizada" {
		t.Fatalf("estado_novedad esperado=contabilizada obtenido=%s", got)
	}
	if nominaLiquidacionID != liquidacionID {
		t.Fatalf("nomina_liquidacion_id en DB esperado=%d obtenido=%d", liquidacionID, nominaLiquidacionID)
	}
	if strings.TrimSpace(nominaPeriodoDesde) != "2026-04-01" {
		t.Fatalf("nomina_periodo_desde esperado=2026-04-01 obtenido=%s", nominaPeriodoDesde)
	}
	if strings.TrimSpace(nominaPeriodoHasta) != "2026-04-15" {
		t.Fatalf("nomina_periodo_hasta esperado=2026-04-15 obtenido=%s", nominaPeriodoHasta)
	}
	if strings.TrimSpace(nominaVinculadaEn) == "" {
		t.Fatalf("nomina_vinculada_en debe registrarse tras vincular_nomina")
	}
	if strings.TrimSpace(nominaVinculadaPor) == "" {
		t.Fatalf("nomina_vinculada_por debe registrarse tras vincular_nomina")
	}
}

func TestEmpresaProduccionOrdenesPlanCapacidad(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_produccion_capacidad_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(128)
	if _, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgProduccionOrdenes.Table, empresaID, map[string]interface{}{
		"codigo":              "OP-128-001",
		"producto_nombre":     "Producto Capacidad A",
		"cantidad_programada": 150,
		"cantidad_producida":  60,
		"fecha_programada":    "2000-01-01",
		"estado_orden":        "en_proceso",
	}, cfgProduccionOrdenes.AllowedColumns); err != nil {
		t.Fatalf("CreateEmpresaGenericRow produccion 1: %v", err)
	}
	if _, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgProduccionOrdenes.Table, empresaID, map[string]interface{}{
		"codigo":              "OP-128-002",
		"producto_nombre":     "Producto Capacidad B",
		"cantidad_programada": 80,
		"cantidad_producida":  80,
		"fecha_programada":    "2000-01-02",
		"estado_orden":        "cerrado",
	}, cfgProduccionOrdenes.AllowedColumns); err != nil {
		t.Fatalf("CreateEmpresaGenericRow produccion 2: %v", err)
	}

	handler := EmpresaProduccionOrdenesHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/produccion/ordenes?action=plan_capacidad&empresa_id=128&desde=2000-01-01&hasta=2000-01-05&meta_diaria=100", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("plan_capacidad status=%d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeBodyAsMap(t, rr)
	resumen, ok := resp["resumen"].(map[string]interface{})
	if !ok {
		t.Fatalf("resumen no disponible en plan_capacidad: %s", rr.Body.String())
	}
	if got := anyToInt64(resumen["ordenes_total"]); got != 2 {
		t.Fatalf("ordenes_total esperado=2 obtenido=%d", got)
	}
	if got := anyToInt64(resumen["ordenes_atrasadas"]); got < 1 {
		t.Fatalf("ordenes_atrasadas esperado>=1 obtenido=%d", got)
	}
	if got := ventasAnyToFloat64(resumen["capacidad_objetivo_total"]); got <= 0 {
		t.Fatalf("capacidad_objetivo_total debe ser > 0 obtenido=%.2f", got)
	}

	itemsRaw, ok := resp["items"].([]interface{})
	if !ok || len(itemsRaw) != 2 {
		t.Fatalf("items esperado=2 obtenido=%v", resp["items"])
	}

	encontroAlertaAtraso := false
	for _, raw := range itemsRaw {
		item := raw.(map[string]interface{})
		if strings.EqualFold(genericStringValue(item["codigo"]), "OP-128-001") {
			if strings.Contains(strings.ToLower(genericStringValue(item["alerta_tipo"])), "atrasada") {
				encontroAlertaAtraso = true
			}
		}
	}
	if !encontroAlertaAtraso {
		t.Fatalf("se esperaba alerta de atraso para OP-128-001")
	}
}

func TestEmpresaLogisticaEnviosSeguimientoHitos(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_modulos_faltantes_logistica_hitos_handler.db")
	ensureModulosFaltantesHandlerSchema(t, dbEmp)

	empresaID := int64(129)
	if _, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgLogisticaEnvios.Table, empresaID, map[string]interface{}{
		"codigo":            "ENV-129-001",
		"cliente_nombre":    "Cliente Sin Salida",
		"direccion_entrega": "Direccion 1",
		"fecha_programada":  "2000-01-01 08:00:00",
		"estado_envio":      "programado",
	}, cfgLogisticaEnvios.AllowedColumns); err != nil {
		t.Fatalf("CreateEmpresaGenericRow logistica 1: %v", err)
	}
	if _, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgLogisticaEnvios.Table, empresaID, map[string]interface{}{
		"codigo":            "ENV-129-002",
		"cliente_nombre":    "Cliente Tardio",
		"direccion_entrega": "Direccion 2",
		"fecha_programada":  "2000-01-01 08:00:00",
		"fecha_salida":      "2000-01-01 09:00:00",
		"fecha_entrega":     "2000-01-03 12:00:00",
		"estado_envio":      "entregado",
	}, cfgLogisticaEnvios.AllowedColumns); err != nil {
		t.Fatalf("CreateEmpresaGenericRow logistica 2: %v", err)
	}
	if _, err := dbpkg.CreateEmpresaGenericRow(dbEmp, cfgLogisticaEnvios.Table, empresaID, map[string]interface{}{
		"codigo":            "ENV-129-003",
		"cliente_nombre":    "Cliente Vigente",
		"direccion_entrega": "Direccion 3",
		"fecha_programada":  "2099-01-01 08:00:00",
		"estado_envio":      "programado",
	}, cfgLogisticaEnvios.AllowedColumns); err != nil {
		t.Fatalf("CreateEmpresaGenericRow logistica 3: %v", err)
	}

	handler := EmpresaLogisticaEnviosHandler(dbEmp)
	req := httptest.NewRequest(http.MethodGet, "/api/empresa/logistica/envios?action=seguimiento_hitos&empresa_id=129&desde=1999-12-31&hasta=2100-01-01&sla_horas=12", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("seguimiento_hitos status=%d body=%s", rr.Code, rr.Body.String())
	}

	resp := decodeBodyAsMap(t, rr)
	resumen, ok := resp["resumen"].(map[string]interface{})
	if !ok {
		t.Fatalf("resumen no disponible en seguimiento_hitos: %s", rr.Body.String())
	}
	if got := anyToInt64(resumen["envios_total"]); got != 3 {
		t.Fatalf("envios_total esperado=3 obtenido=%d", got)
	}
	if got := anyToInt64(resumen["incumplidos"]); got < 2 {
		t.Fatalf("incumplidos esperado>=2 obtenido=%d", got)
	}

	itemsRaw, ok := resp["items"].([]interface{})
	if !ok || len(itemsRaw) != 3 {
		t.Fatalf("items esperado=3 obtenido=%v", resp["items"])
	}

	encontroSinSalida := false
	encontroEntregaTardia := false
	for _, raw := range itemsRaw {
		item := raw.(map[string]interface{})
		codigo := strings.ToUpper(strings.TrimSpace(genericStringValue(item["codigo"])))
		alertaTipo := strings.ToLower(strings.TrimSpace(genericStringValue(item["alerta_tipo"])))
		switch codigo {
		case "ENV-129-001":
			if strings.Contains(alertaTipo, "sin_salida") {
				encontroSinSalida = true
			}
		case "ENV-129-002":
			if strings.Contains(alertaTipo, "entrega_tardia") {
				encontroEntregaTardia = true
			}
		}
	}

	if !encontroSinSalida {
		t.Fatalf("se esperaba alerta sin_salida para ENV-129-001")
	}
	if !encontroEntregaTardia {
		t.Fatalf("se esperaba alerta entrega_tardia para ENV-129-002")
	}
}
