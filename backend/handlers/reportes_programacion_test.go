package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestEmpresaReportesHandlerPlantillasVersionado(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_plantillas_m31.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	handler := EmpresaReportesHandler(dbEmp)
	empresaID := int64(801)

	payloadV1 := `{"empresa_id":801,"codigo":"TPL-VENTAS","nombre":"Plantilla Ventas V1","dataset_key":"operativo_ventas_detalle","formato":"csv","columnas":["fecha_pago","cliente_nombre","total"],"config":{"separador":";"},"marcar_vigente":true}`
	reqV1 := httptest.NewRequest(http.MethodPost, "/api/empresa/reportes?action=plantillas&empresa_id=801", strings.NewReader(payloadV1))
	reqV1.Header.Set("Content-Type", "application/json")
	rrV1 := httptest.NewRecorder()
	handler.ServeHTTP(rrV1, reqV1)
	if rrV1.Code != http.StatusCreated {
		t.Fatalf("crear plantilla v1 status=%d body=%s", rrV1.Code, rrV1.Body.String())
	}

	var respV1 map[string]interface{}
	if err := json.Unmarshal(rrV1.Body.Bytes(), &respV1); err != nil {
		t.Fatalf("unmarshal plantilla v1: %v", err)
	}
	plantillaV1, _ := respV1["plantilla"].(map[string]interface{})
	if reporteDatasetToInt64(plantillaV1["version"]) != 1 {
		t.Fatalf("version plantilla v1 esperada=1 obtenida=%v", plantillaV1["version"])
	}

	payloadV2 := `{"empresa_id":801,"codigo":"TPL-VENTAS","nombre":"Plantilla Ventas V2","dataset_key":"operativo_ventas_detalle","formato":"csv","columnas":["fecha_pago","cliente_nombre","total","metodo_pago"],"config":{"separador":";"},"marcar_vigente":true}`
	reqV2 := httptest.NewRequest(http.MethodPost, "/api/empresa/reportes?action=plantillas&empresa_id=801", strings.NewReader(payloadV2))
	reqV2.Header.Set("Content-Type", "application/json")
	rrV2 := httptest.NewRecorder()
	handler.ServeHTTP(rrV2, reqV2)
	if rrV2.Code != http.StatusCreated {
		t.Fatalf("crear plantilla v2 status=%d body=%s", rrV2.Code, rrV2.Body.String())
	}

	var respV2 map[string]interface{}
	if err := json.Unmarshal(rrV2.Body.Bytes(), &respV2); err != nil {
		t.Fatalf("unmarshal plantilla v2: %v", err)
	}
	plantillaV2, _ := respV2["plantilla"].(map[string]interface{})
	if reporteDatasetToInt64(plantillaV2["version"]) != 2 {
		t.Fatalf("version plantilla v2 esperada=2 obtenida=%v", plantillaV2["version"])
	}

	reqList := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=plantillas&empresa_id=801&codigo=TPL-VENTAS&solo_vigente=1", nil)
	rrList := httptest.NewRecorder()
	handler.ServeHTTP(rrList, reqList)
	if rrList.Code != http.StatusOK {
		t.Fatalf("listar plantillas vigentes status=%d body=%s", rrList.Code, rrList.Body.String())
	}

	var listResp map[string]interface{}
	if err := json.Unmarshal(rrList.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal listar plantillas: %v", err)
	}
	items, _ := listResp["plantillas"].([]interface{})
	if len(items) != 1 {
		t.Fatalf("plantillas vigentes esperadas=1 obtenidas=%d", len(items))
	}
	item, _ := items[0].(map[string]interface{})
	if reporteDatasetToInt64(item["empresa_id"]) != empresaID {
		t.Fatalf("empresa_id esperado=%d obtenido=%v", empresaID, item["empresa_id"])
	}
	if reporteDatasetToInt64(item["version"]) != 2 {
		t.Fatalf("version vigente esperada=2 obtenida=%v", item["version"])
	}
}

func TestEmpresaReportesHandlerProgramacionEjecucionYConsistencia(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_reportes_programacion_m31.db")
	ensureEmpresaReportesSchema(t, dbEmp)

	handler := EmpresaReportesHandler(dbEmp)
	empresaID := int64(811)

	payloadTpl := `{"empresa_id":811,"codigo":"TPL-TABLERO","nombre":"Plantilla Tablero","dataset_key":"empresarial_tablero","formato":"json","columnas":["empresa_id","ingresos_ventas","ticket_promedio"],"marcar_vigente":true}`
	reqTpl := httptest.NewRequest(http.MethodPost, "/api/empresa/reportes?action=plantillas&empresa_id=811", strings.NewReader(payloadTpl))
	reqTpl.Header.Set("Content-Type", "application/json")
	rrTpl := httptest.NewRecorder()
	handler.ServeHTTP(rrTpl, reqTpl)
	if rrTpl.Code != http.StatusCreated {
		t.Fatalf("crear plantilla status=%d body=%s", rrTpl.Code, rrTpl.Body.String())
	}

	payloadProg := `{"empresa_id":811,"nombre":"Programacion diaria tablero","dataset":"empresarial_tablero","frecuencia":"diario","hora_envio":"07:30","timezone":"America/Bogota","formatos":["json","csv","txt","xls","pdf"],"destinatarios":["contabilidad@empresa.test","gerencia@empresa.test"],"template_codigo":"TPL-TABLERO","validar_consistencia":true,"activa":true,"parametros":{"max_rows":100}}`
	reqProg := httptest.NewRequest(http.MethodPost, "/api/empresa/reportes?action=programacion&empresa_id=811", strings.NewReader(payloadProg))
	reqProg.Header.Set("Content-Type", "application/json")
	rrProg := httptest.NewRecorder()
	handler.ServeHTTP(rrProg, reqProg)
	if rrProg.Code != http.StatusCreated {
		t.Fatalf("crear programacion status=%d body=%s", rrProg.Code, rrProg.Body.String())
	}

	var progResp map[string]interface{}
	if err := json.Unmarshal(rrProg.Body.Bytes(), &progResp); err != nil {
		t.Fatalf("unmarshal programacion: %v", err)
	}
	progItem, _ := progResp["programacion"].(map[string]interface{})
	programacionID := reporteDatasetToInt64(progItem["id"])
	if programacionID <= 0 {
		t.Fatalf("programacion id invalido: %v", progItem["id"])
	}

	reqExec := httptest.NewRequest(http.MethodPost, "/api/empresa/reportes?action=ejecutar_programacion&empresa_id=811", strings.NewReader(`{"programacion_id":`+strconv.FormatInt(programacionID, 10)+`}`))
	reqExec.Header.Set("Content-Type", "application/json")
	rrExec := httptest.NewRecorder()
	handler.ServeHTTP(rrExec, reqExec)
	if rrExec.Code != http.StatusOK {
		t.Fatalf("ejecutar programacion status=%d body=%s", rrExec.Code, rrExec.Body.String())
	}

	var execResp map[string]interface{}
	if err := json.Unmarshal(rrExec.Body.Bytes(), &execResp); err != nil {
		t.Fatalf("unmarshal ejecutar programacion: %v", err)
	}
	if !reportesToBool(execResp["ok"], false) {
		t.Fatalf("respuesta ejecutar programacion sin ok=true: %+v", execResp)
	}
	if reporteDatasetToInt64(execResp["programacion_id"]) != programacionID {
		t.Fatalf("programacion_id esperado=%d obtenido=%v", programacionID, execResp["programacion_id"])
	}
	consistencia, _ := execResp["consistencia"].(map[string]interface{})
	if !reportesToBool(consistencia["consistente"], false) {
		t.Fatalf("consistencia esperada=true obtenido=%v payload=%+v", consistencia["consistente"], consistencia)
	}

	reqEjecuciones := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=ejecuciones&empresa_id=811&programacion_id="+strconv.FormatInt(programacionID, 10), nil)
	rrEjecuciones := httptest.NewRecorder()
	handler.ServeHTTP(rrEjecuciones, reqEjecuciones)
	if rrEjecuciones.Code != http.StatusOK {
		t.Fatalf("listar ejecuciones status=%d body=%s", rrEjecuciones.Code, rrEjecuciones.Body.String())
	}
	var ejecucionesResp map[string]interface{}
	if err := json.Unmarshal(rrEjecuciones.Body.Bytes(), &ejecucionesResp); err != nil {
		t.Fatalf("unmarshal ejecuciones: %v", err)
	}
	if reporteDatasetToInt64(ejecucionesResp["total"]) < 1 {
		t.Fatalf("total ejecuciones esperado>=1 obtenido=%v", ejecucionesResp["total"])
	}

	reqConsistencia := httptest.NewRequest(http.MethodGet, "/api/empresa/reportes?action=validar_consistencia&empresa_id=811&dataset=empresarial_tablero&formatos=json,csv,txt,xls,pdf&template_codigo=TPL-TABLERO", nil)
	rrConsistencia := httptest.NewRecorder()
	handler.ServeHTTP(rrConsistencia, reqConsistencia)
	if rrConsistencia.Code != http.StatusOK {
		t.Fatalf("validar consistencia status=%d body=%s", rrConsistencia.Code, rrConsistencia.Body.String())
	}
	var consistenciaResp map[string]interface{}
	if err := json.Unmarshal(rrConsistencia.Body.Bytes(), &consistenciaResp); err != nil {
		t.Fatalf("unmarshal consistencia: %v", err)
	}
	if !reportesToBool(consistenciaResp["consistente"], false) {
		t.Fatalf("consistencia final esperada=true obtenido=%v payload=%+v", consistenciaResp["consistente"], consistenciaResp)
	}
	if reporteDatasetToInt64(consistenciaResp["empresa_id"]) != empresaID {
		t.Fatalf("empresa_id consistencia esperado=%d obtenido=%v", empresaID, consistenciaResp["empresa_id"])
	}
}
