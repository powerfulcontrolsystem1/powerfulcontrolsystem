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

func TestEmpresaVehiculosRegistroHandlerCRUDFlow(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_vehiculos_registro_handler.db")
	if err := dbpkg.EnsureEmpresaVehiculosRegistroSchema(dbEmp); err != nil {
		t.Fatalf("ensure vehiculos schema: %v", err)
	}

	handler := EmpresaVehiculosRegistroHandler(dbEmp)

	createBody := `{"empresa_id":9,"patente":"abc123","tipo_vehiculo":"automovil","conductor_nombre":"Luis Mora","conductor_documento":"123456","motivo_ingreso":"Visita"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/vehiculos_registro", strings.NewReader(createBody))
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

	markSalidaBody := `{"empresa_id":9,"id":` + strconv.FormatInt(id, 10) + `,"observaciones":"Retiro autorizado"}`
	markSalidaReq := httptest.NewRequest(http.MethodPut, "/api/empresa/vehiculos_registro?action=marcar_salida", strings.NewReader(markSalidaBody))
	markSalidaReq.Header.Set("Content-Type", "application/json")
	markSalidaRR := httptest.NewRecorder()
	handler.ServeHTTP(markSalidaRR, markSalidaReq)
	if markSalidaRR.Code != http.StatusOK {
		t.Fatalf("mark salida expected=%d got=%d body=%s", http.StatusOK, markSalidaRR.Code, markSalidaRR.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/vehiculos_registro?empresa_id=9&estado_registro=retirado&limit=50", nil)
	listRR := httptest.NewRecorder()
	handler.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var rows []dbpkg.EmpresaVehiculoRegistro
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].EstadoRegistro != "retirado" {
		t.Fatalf("expected estado_registro retirado, got %q", rows[0].EstadoRegistro)
	}

	disableReq := httptest.NewRequest(http.MethodPut, "/api/empresa/vehiculos_registro?empresa_id=9&id="+strconv.FormatInt(id, 10)+"&action=desactivar", nil)
	disableRR := httptest.NewRecorder()
	handler.ServeHTTP(disableRR, disableReq)
	if disableRR.Code != http.StatusOK {
		t.Fatalf("disable expected=%d got=%d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	listActiveReq := httptest.NewRequest(http.MethodGet, "/api/empresa/vehiculos_registro?empresa_id=9", nil)
	listActiveRR := httptest.NewRecorder()
	handler.ServeHTTP(listActiveRR, listActiveReq)
	if listActiveRR.Code != http.StatusOK {
		t.Fatalf("list active expected=%d got=%d body=%s", http.StatusOK, listActiveRR.Code, listActiveRR.Body.String())
	}
	var rowsActive []dbpkg.EmpresaVehiculoRegistro
	if err := json.Unmarshal(listActiveRR.Body.Bytes(), &rowsActive); err != nil {
		t.Fatalf("decode active response: %v", err)
	}
	if len(rowsActive) != 0 {
		t.Fatalf("expected 0 rows after disable, got %d", len(rowsActive))
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/api/empresa/vehiculos_registro?empresa_id=9&id="+strconv.FormatInt(id, 10), nil)
	deleteRR := httptest.NewRecorder()
	handler.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete expected=%d got=%d body=%s", http.StatusOK, deleteRR.Code, deleteRR.Body.String())
	}
}

func TestEmpresaVehiculosRegistroHandlerConfigYReportePermanencia(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_vehiculos_config_handler.db")
	if err := dbpkg.EnsureEmpresaVehiculosRegistroSchema(dbEmp); err != nil {
		t.Fatalf("ensure vehiculos schema: %v", err)
	}

	handler := EmpresaVehiculosRegistroHandler(dbEmp)

	configBody := `{"empresa_id":19,"pais_codigo":"MX","patente_regex":"^[A-Z0-9]{6,7}$","patente_descripcion":"MX test","evitar_duplicado_activo":true}`
	configReq := httptest.NewRequest(http.MethodPut, "/api/empresa/vehiculos_registro?action=config", strings.NewReader(configBody))
	configReq.Header.Set("Content-Type", "application/json")
	configRR := httptest.NewRecorder()
	handler.ServeHTTP(configRR, configReq)
	if configRR.Code != http.StatusOK {
		t.Fatalf("config expected=%d got=%d body=%s", http.StatusOK, configRR.Code, configRR.Body.String())
	}

	createBody := `{"empresa_id":19,"patente":"ABC1234","tipo_vehiculo":"automovil","conductor_nombre":"Mario","motivo_ingreso":"Visita"}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/vehiculos_registro", strings.NewReader(createBody))
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

	invalidBody := `{"empresa_id":19,"patente":"AB@1234","tipo_vehiculo":"automovil"}`
	invalidReq := httptest.NewRequest(http.MethodPost, "/api/empresa/vehiculos_registro", strings.NewReader(invalidBody))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidRR := httptest.NewRecorder()
	handler.ServeHTTP(invalidRR, invalidReq)
	if invalidRR.Code != http.StatusBadRequest {
		t.Fatalf("invalid patente expected=%d got=%d body=%s", http.StatusBadRequest, invalidRR.Code, invalidRR.Body.String())
	}

	duplicateBody := `{"empresa_id":19,"patente":"ABC-1234","tipo_vehiculo":"automovil"}`
	duplicateReq := httptest.NewRequest(http.MethodPost, "/api/empresa/vehiculos_registro", strings.NewReader(duplicateBody))
	duplicateReq.Header.Set("Content-Type", "application/json")
	duplicateRR := httptest.NewRecorder()
	handler.ServeHTTP(duplicateRR, duplicateReq)
	if duplicateRR.Code != http.StatusConflict {
		t.Fatalf("duplicate expected=%d got=%d body=%s", http.StatusConflict, duplicateRR.Code, duplicateRR.Body.String())
	}

	reporteReq := httptest.NewRequest(http.MethodGet, "/api/empresa/vehiculos_registro?action=permanencia&empresa_id=19&desde=2026-04-01&hasta=2026-04-30", nil)
	reporteRR := httptest.NewRecorder()
	handler.ServeHTTP(reporteRR, reporteReq)
	if reporteRR.Code != http.StatusOK {
		t.Fatalf("reporte expected=%d got=%d body=%s", http.StatusOK, reporteRR.Code, reporteRR.Body.String())
	}
	var reporteRows []dbpkg.EmpresaVehiculoPermanenciaReporteItem
	if err := json.Unmarshal(reporteRR.Body.Bytes(), &reporteRows); err != nil {
		t.Fatalf("decode reporte response: %v", err)
	}
	if len(reporteRows) != 1 {
		t.Fatalf("expected 1 row in reporte, got %d", len(reporteRows))
	}

	markSalidaBody := `{"empresa_id":19,"id":` + strconv.FormatInt(id, 10) + `,"fecha_salida":"2026-04-10 10:00:00"}`
	markSalidaReq := httptest.NewRequest(http.MethodPut, "/api/empresa/vehiculos_registro?action=marcar_salida", strings.NewReader(markSalidaBody))
	markSalidaReq.Header.Set("Content-Type", "application/json")
	markSalidaRR := httptest.NewRecorder()
	handler.ServeHTTP(markSalidaRR, markSalidaReq)
	if markSalidaRR.Code != http.StatusOK {
		t.Fatalf("mark salida expected=%d got=%d body=%s", http.StatusOK, markSalidaRR.Code, markSalidaRR.Body.String())
	}

	recreateBody := `{"empresa_id":19,"patente":"ABC1234","tipo_vehiculo":"camioneta"}`
	recreateReq := httptest.NewRequest(http.MethodPost, "/api/empresa/vehiculos_registro", strings.NewReader(recreateBody))
	recreateReq.Header.Set("Content-Type", "application/json")
	recreateRR := httptest.NewRecorder()
	handler.ServeHTTP(recreateRR, recreateReq)
	if recreateRR.Code != http.StatusCreated {
		t.Fatalf("recreate expected=%d got=%d body=%s", http.StatusCreated, recreateRR.Code, recreateRR.Body.String())
	}
}
