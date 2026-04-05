package handlers

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaTarifasPorDiaHandlerCRUDAndCalculo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_tarifas_por_dia_handler.db")
	if err := dbpkg.EnsureEmpresaTarifasPorDiaSchema(dbEmp); err != nil {
		t.Fatalf("ensure tarifas por dia schema: %v", err)
	}

	h := EmpresaTarifasPorDiaHandler(dbEmp)

	createBody := `{"empresa_id":9,"estacion_id":12,"estacion_codigo":"EST-9-12","estacion_nombre":"Habitacion 12","servicio_nombre":"hotel","valor_dia":90000,"hora_check_in":"15:00","hora_check_out":"12:00","moneda":"COP","prioridad":1,"aplicar_automaticamente":true}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/tarifas_por_dia", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create expected=%d got=%d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var created dbpkg.EmpresaTarifaPorDia
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected id > 0, got %d", created.ID)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_dia?empresa_id=9&estacion_id=12&limit=50", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var rows []dbpkg.EmpresaTarifaPorDia
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	updateBody := `{"id":` + strconv.FormatInt(created.ID, 10) + `,"empresa_id":9,"estacion_id":12,"estacion_codigo":"EST-9-12","estacion_nombre":"Habitacion 12","servicio_nombre":"hotel","valor_dia":95000,"hora_check_in":"15:00","hora_check_out":"12:00","moneda":"COP","prioridad":1,"aplicar_automaticamente":true}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/tarifas_por_dia", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	h.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update expected=%d got=%d body=%s", http.StatusOK, updateRR.Code, updateRR.Body.String())
	}

	calcReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_dia?empresa_id=9&action=calcular&estacion_id=12&activado_en=2026-04-01%2016:00:00&fecha_corte=2026-04-03%2013:00:00", nil)
	calcRR := httptest.NewRecorder()
	h.ServeHTTP(calcRR, calcReq)
	if calcRR.Code != http.StatusOK {
		t.Fatalf("calcular expected=%d got=%d body=%s", http.StatusOK, calcRR.Code, calcRR.Body.String())
	}
	var calc map[string]interface{}
	if err := json.Unmarshal(calcRR.Body.Bytes(), &calc); err != nil {
		t.Fatalf("decode calcular response: %v", err)
	}
	if !calc["ok"].(bool) {
		t.Fatal("expected ok=true in calcular response")
	}
	dias := int(calc["dias_cobrados"].(float64))
	if dias != 3 {
		t.Fatalf("expected dias_cobrados 3, got %d", dias)
	}
	monto := calc["monto_total"].(float64)
	if math.Abs(monto-285000) > 0.001 {
		t.Fatalf("expected monto_total 285000, got %.2f", monto)
	}

	disableURL := "/api/empresa/tarifas_por_dia?empresa_id=9&id=" + strconv.FormatInt(created.ID, 10) + "&action=desactivar"
	disableReq := httptest.NewRequest(http.MethodPut, disableURL, nil)
	disableRR := httptest.NewRecorder()
	h.ServeHTTP(disableRR, disableReq)
	if disableRR.Code != http.StatusOK {
		t.Fatalf("disable expected=%d got=%d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	aplicableReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_dia?empresa_id=9&action=aplicable&estacion_id=12", nil)
	aplicableRR := httptest.NewRecorder()
	h.ServeHTTP(aplicableRR, aplicableReq)
	if aplicableRR.Code != http.StatusOK {
		t.Fatalf("aplicable expected=%d got=%d body=%s", http.StatusOK, aplicableRR.Code, aplicableRR.Body.String())
	}
	var aplicable map[string]interface{}
	if err := json.Unmarshal(aplicableRR.Body.Bytes(), &aplicable); err != nil {
		t.Fatalf("decode aplicable response: %v", err)
	}
	if aplicable["ok"].(bool) {
		t.Fatal("expected ok=false in aplicable response after disable")
	}

	deleteURL := "/api/empresa/tarifas_por_dia?empresa_id=9&id=" + strconv.FormatInt(created.ID, 10)
	deleteReq := httptest.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete expected=%d got=%d body=%s", http.StatusOK, deleteRR.Code, deleteRR.Body.String())
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_dia?empresa_id=9&action=detalle&id="+strconv.FormatInt(created.ID, 10), nil)
	detailRR := httptest.NewRecorder()
	h.ServeHTTP(detailRR, detailReq)
	if detailRR.Code != http.StatusNotFound {
		t.Fatalf("detail after delete expected=%d got=%d body=%s", http.StatusNotFound, detailRR.Code, detailRR.Body.String())
	}
}
