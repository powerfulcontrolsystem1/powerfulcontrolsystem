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

func TestEmpresaAsistenciaEmpleadosHandlerCRUDFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_asistencia_handler.db")
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	handler := EmpresaAsistenciaEmpleadosHandler(dbEmp)

	createBody := `{"empresa_id":12,"empleado_codigo":"EMP-001","empleado_nombre":"Ana Perez","empleado_documento":"10101010","cargo":"Recepcion","turno":"manana","fecha_asistencia":"2026-04-04"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create expected=%d got=%d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	id := int64(createResp["id"].(float64))
	if id <= 0 {
		t.Fatalf("expected id > 0, got %d", id)
	}

	markEntradaBody := `{"empresa_id":12,"id":` + strconv.FormatInt(id, 10) + `,"hora_entrada":"08:05:00","minutos_tarde":5}`
	markEntradaReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados?action=marcar_entrada", strings.NewReader(markEntradaBody))
	markEntradaReq.Header.Set("Content-Type", "application/json")
	markEntradaRR := httptest.NewRecorder()
	handler.ServeHTTP(markEntradaRR, markEntradaReq)
	if markEntradaRR.Code != http.StatusOK {
		t.Fatalf("mark entrada expected=%d got=%d body=%s", http.StatusOK, markEntradaRR.Code, markEntradaRR.Body.String())
	}

	markSalidaBody := `{"empresa_id":12,"id":` + strconv.FormatInt(id, 10) + `,"hora_salida":"17:10:00"}`
	markSalidaReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados?action=marcar_salida", strings.NewReader(markSalidaBody))
	markSalidaReq.Header.Set("Content-Type", "application/json")
	markSalidaRR := httptest.NewRecorder()
	handler.ServeHTTP(markSalidaRR, markSalidaReq)
	if markSalidaRR.Code != http.StatusOK {
		t.Fatalf("mark salida expected=%d got=%d body=%s", http.StatusOK, markSalidaRR.Code, markSalidaRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/asistencia_empleados?empresa_id=12&desde=2026-04-01&hasta=2026-04-30&limit=50", nil)
	listRR := httptest.NewRecorder()
	handler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var rows []dbpkg.EmpresaAsistenciaEmpleado
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].HoraEntrada == "" || rows[0].HoraSalida == "" {
		t.Fatalf("expected entrada/salida, got entrada=%q salida=%q", rows[0].HoraEntrada, rows[0].HoraSalida)
	}
	if rows[0].HorasTrabajadas <= 0 {
		t.Fatalf("expected horas_trabajadas > 0, got %.2f", rows[0].HorasTrabajadas)
	}

	updateBody := `{"empresa_id":12,"id":` + strconv.FormatInt(id, 10) + `,"empleado_codigo":"EMP-001","empleado_nombre":"Ana Perez","empleado_documento":"10101010","cargo":"Recepcion Principal","turno":"manana","fecha_asistencia":"2026-04-04","hora_entrada":"08:00:00","hora_salida":"17:00:00","minutos_tarde":0,"horas_trabajadas":9,"estado_asistencia":"presente","novedad":"turno normal","observaciones":"ok"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	handler.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update expected=%d got=%d body=%s", http.StatusOK, updateRR.Code, updateRR.Body.String())
	}

	disableReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados?empresa_id=12&id="+strconv.FormatInt(id, 10)+"&action=desactivar", nil)
	disableRR := httptest.NewRecorder()
	handler.ServeHTTP(disableRR, disableReq)
	if disableRR.Code != http.StatusOK {
		t.Fatalf("disable expected=%d got=%d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	listActiveReq := httptest.NewRequest(http.MethodGet, "/api/empresa/asistencia_empleados?empresa_id=12", nil)
	listActiveRR := httptest.NewRecorder()
	handler.ServeHTTP(listActiveRR, listActiveReq)
	if listActiveRR.Code != http.StatusOK {
		t.Fatalf("list active expected=%d got=%d body=%s", http.StatusOK, listActiveRR.Code, listActiveRR.Body.String())
	}
	var rowsActive []dbpkg.EmpresaAsistenciaEmpleado
	if err := json.Unmarshal(listActiveRR.Body.Bytes(), &rowsActive); err != nil {
		t.Fatalf("decode list active response: %v", err)
	}
	if len(rowsActive) != 0 {
		t.Fatalf("expected 0 active rows after disable, got %d", len(rowsActive))
	}

	listAllReq := httptest.NewRequest(http.MethodGet, "/api/empresa/asistencia_empleados?empresa_id=12&include_inactive=1", nil)
	listAllRR := httptest.NewRecorder()
	handler.ServeHTTP(listAllRR, listAllReq)
	if listAllRR.Code != http.StatusOK {
		t.Fatalf("list all expected=%d got=%d body=%s", http.StatusOK, listAllRR.Code, listAllRR.Body.String())
	}
	var rowsAll []dbpkg.EmpresaAsistenciaEmpleado
	if err := json.Unmarshal(listAllRR.Body.Bytes(), &rowsAll); err != nil {
		t.Fatalf("decode list all response: %v", err)
	}
	if len(rowsAll) != 1 {
		t.Fatalf("expected 1 row in include_inactive list, got %d", len(rowsAll))
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/asistencia_empleados?empresa_id=12&id="+strconv.FormatInt(id, 10), nil)
	deleteRR := httptest.NewRecorder()
	handler.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete expected=%d got=%d body=%s", http.StatusOK, deleteRR.Code, deleteRR.Body.String())
	}
}

func TestEmpresaAsistenciaEmpleadosHandlerConfigTurnosYTolerancia(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_asistencia_config_handler.db")
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	handler := EmpresaAsistenciaEmpleadosHandler(dbEmp)

	configBody := `{"empresa_id":22,"tolerancia_entrada_minutos":15,"hora_inicio_turno_manana":"08:00:00","hora_inicio_turno_tarde":"14:00:00","hora_inicio_turno_noche":"22:00:00","permitir_turno_nocturno":false,"permitir_turno_cruzado":false}`
	configReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados?action=config", strings.NewReader(configBody))
	configReq.Header.Set("Content-Type", "application/json")
	configRR := httptest.NewRecorder()
	handler.ServeHTTP(configRR, configReq)
	if configRR.Code != http.StatusOK {
		t.Fatalf("config expected=%d got=%d body=%s", http.StatusOK, configRR.Code, configRR.Body.String())
	}

	createBodyA := `{"empresa_id":22,"empleado_codigo":"EMP-A","empleado_nombre":"Empleado A","empleado_documento":"DOC-A","cargo":"Caja","turno":"manana","fecha_asistencia":"2026-04-07","hora_entrada":"08:10:00","hora_salida":"16:10:00"}`
	createReqA := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados", strings.NewReader(createBodyA))
	createReqA.Header.Set("Content-Type", "application/json")
	createRRA := httptest.NewRecorder()
	handler.ServeHTTP(createRRA, createReqA)
	if createRRA.Code != http.StatusCreated {
		t.Fatalf("create A expected=%d got=%d body=%s", http.StatusCreated, createRRA.Code, createRRA.Body.String())
	}

	createBodyB := `{"empresa_id":22,"empleado_codigo":"EMP-B","empleado_nombre":"Empleado B","empleado_documento":"DOC-B","cargo":"Caja","turno":"manana","fecha_asistencia":"2026-04-08","hora_entrada":"08:30:00","hora_salida":"16:30:00"}`
	createReqB := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados", strings.NewReader(createBodyB))
	createReqB.Header.Set("Content-Type", "application/json")
	createRRB := httptest.NewRecorder()
	handler.ServeHTTP(createRRB, createReqB)
	if createRRB.Code != http.StatusCreated {
		t.Fatalf("create B expected=%d got=%d body=%s", http.StatusCreated, createRRB.Code, createRRB.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/asistencia_empleados?empresa_id=22&desde=2026-04-01&hasta=2026-04-30&limit=100", nil)
	listRR := httptest.NewRecorder()
	handler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var rows []dbpkg.EmpresaAsistenciaEmpleado
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode rows: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("expected at least 2 rows, got %d", len(rows))
	}

	var minutosA int
	var minutosB int
	for _, row := range rows {
		switch strings.TrimSpace(row.EmpleadoCodigo) {
		case "EMP-A":
			minutosA = row.MinutosTarde
		case "EMP-B":
			minutosB = row.MinutosTarde
		}
	}
	if minutosA != 0 {
		t.Fatalf("expected EMP-A minutos_tarde=0 by tolerance, got %d", minutosA)
	}
	if minutosB != 15 {
		t.Fatalf("expected EMP-B minutos_tarde=15, got %d", minutosB)
	}

	createNocturno := `{"empresa_id":22,"empleado_codigo":"EMP-N","empleado_nombre":"Empleado N","empleado_documento":"DOC-N","cargo":"Noche","turno":"noche","fecha_asistencia":"2026-04-09","hora_entrada":"22:05:00"}`
	createNocturnoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados", strings.NewReader(createNocturno))
	createNocturnoReq.Header.Set("Content-Type", "application/json")
	createNocturnoRR := httptest.NewRecorder()
	handler.ServeHTTP(createNocturnoRR, createNocturnoReq)
	if createNocturnoRR.Code != http.StatusBadRequest {
		t.Fatalf("expected nocturno disabled status=%d got=%d body=%s", http.StatusBadRequest, createNocturnoRR.Code, createNocturnoRR.Body.String())
	}

	createCruzado := `{"empresa_id":22,"empleado_codigo":"EMP-C","empleado_nombre":"Empleado C","empleado_documento":"DOC-C","cargo":"Tarde","turno":"tarde","fecha_asistencia":"2026-04-10","hora_entrada":"14:00:00","hora_salida":"06:00:00"}`
	createCruzadoReq := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados", strings.NewReader(createCruzado))
	createCruzadoReq.Header.Set("Content-Type", "application/json")
	createCruzadoRR := httptest.NewRecorder()
	handler.ServeHTTP(createCruzadoRR, createCruzadoReq)
	if createCruzadoRR.Code != http.StatusBadRequest {
		t.Fatalf("expected turno cruzado disabled status=%d got=%d body=%s", http.StatusBadRequest, createCruzadoRR.Code, createCruzadoRR.Body.String())
	}
}

func TestEmpresaAsistenciaEmpleadosHandlerCierrePeriodoBloqueaEdicion(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_asistencia_cierre_periodo.db")
	if err := dbpkg.EnsureEmpresaAsistenciaSchema(dbEmp); err != nil {
		t.Fatalf("ensure asistencia schema: %v", err)
	}

	handler := EmpresaAsistenciaEmpleadosHandler(dbEmp)

	createBody := `{"empresa_id":31,"empleado_codigo":"EMP-LOCK","empleado_nombre":"Empleado Lock","empleado_documento":"DOC-LOCK","cargo":"Recepcion","turno":"manana","fecha_asistencia":"2026-04-12"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	handler.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create expected=%d got=%d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}
	var createResp map[string]interface{}
	if err := json.Unmarshal(createRR.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	id := int64(createResp["id"].(float64))

	cierreBody := `{"empresa_id":31,"periodo_desde":"2026-04-01","periodo_hasta":"2026-04-30","motivo":"cierre nomina abril"}`
	cierreReq := httptest.NewRequest(http.MethodPost, "/api/empresa/asistencia_empleados?action=cerrar_periodo", strings.NewReader(cierreBody))
	cierreReq.Header.Set("Content-Type", "application/json")
	cierreRR := httptest.NewRecorder()
	handler.ServeHTTP(cierreRR, cierreReq)
	if cierreRR.Code != http.StatusCreated {
		t.Fatalf("cierre expected=%d got=%d body=%s", http.StatusCreated, cierreRR.Code, cierreRR.Body.String())
	}

	updateBody := `{"empresa_id":31,"id":` + strconv.FormatInt(id, 10) + `,"empleado_codigo":"EMP-LOCK","empleado_nombre":"Empleado Lock","empleado_documento":"DOC-LOCK","cargo":"Recepcion","turno":"manana","fecha_asistencia":"2026-04-12","hora_entrada":"08:00:00"}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	handler.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusConflict {
		t.Fatalf("update expected conflict=%d got=%d body=%s", http.StatusConflict, updateRR.Code, updateRR.Body.String())
	}

	entradaBody := `{"empresa_id":31,"id":` + strconv.FormatInt(id, 10) + `,"hora_entrada":"08:00:00"}`
	entradaReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados?action=marcar_entrada", strings.NewReader(entradaBody))
	entradaReq.Header.Set("Content-Type", "application/json")
	entradaRR := httptest.NewRecorder()
	handler.ServeHTTP(entradaRR, entradaReq)
	if entradaRR.Code != http.StatusConflict {
		t.Fatalf("mark entrada expected conflict=%d got=%d body=%s", http.StatusConflict, entradaRR.Code, entradaRR.Body.String())
	}

	toggleReq := httptest.NewRequest(http.MethodPut, "/api/empresa/asistencia_empleados?empresa_id=31&id="+strconv.FormatInt(id, 10)+"&action=desactivar", nil)
	toggleRR := httptest.NewRecorder()
	handler.ServeHTTP(toggleRR, toggleReq)
	if toggleRR.Code != http.StatusConflict {
		t.Fatalf("toggle expected conflict=%d got=%d body=%s", http.StatusConflict, toggleRR.Code, toggleRR.Body.String())
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/asistencia_empleados?empresa_id=31&id="+strconv.FormatInt(id, 10), nil)
	deleteRR := httptest.NewRecorder()
	handler.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusConflict {
		t.Fatalf("delete expected conflict=%d got=%d body=%s", http.StatusConflict, deleteRR.Code, deleteRR.Body.String())
	}

	periodosReq := httptest.NewRequest(http.MethodGet, "/api/empresa/asistencia_empleados?action=periodos_cerrados&empresa_id=31", nil)
	periodosRR := httptest.NewRecorder()
	handler.ServeHTTP(periodosRR, periodosReq)
	if periodosRR.Code != http.StatusOK {
		t.Fatalf("periodos expected=%d got=%d body=%s", http.StatusOK, periodosRR.Code, periodosRR.Body.String())
	}
	var periodos []dbpkg.EmpresaAsistenciaPeriodoCierre
	if err := json.Unmarshal(periodosRR.Body.Bytes(), &periodos); err != nil {
		t.Fatalf("decode periodos: %v", err)
	}
	if len(periodos) != 1 {
		t.Fatalf("expected 1 periodo cerrado, got %d", len(periodos))
	}
}
