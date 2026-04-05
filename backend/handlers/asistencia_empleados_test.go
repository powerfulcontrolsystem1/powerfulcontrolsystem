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
