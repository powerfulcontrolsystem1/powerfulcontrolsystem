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

func TestEmpresaTarifasPorMinutosHandlerCRUDAndCalculo(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_tarifas_por_minutos_handler.db")
	if err := dbpkg.EnsureEmpresaTarifasPorMinutosSchema(dbEmp); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}

	h := EmpresaTarifasPorMinutosHandler(dbEmp)

	createBody := `{"empresa_id":7,"estacion_id":12,"estacion_codigo":"EST-7-12","estacion_nombre":"Habitacion 12","dia_semana_desde":1,"dia_semana_hasta":4,"minutos_base":120,"valor_base":30000,"minutos_extra":60,"valor_extra":15000,"moneda":"COP","prioridad":1}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/empresa/tarifas_por_minutos", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createRR := httptest.NewRecorder()
	h.ServeHTTP(createRR, createReq)
	if createRR.Code != http.StatusCreated {
		t.Fatalf("create expected=%d got=%d body=%s", http.StatusCreated, createRR.Code, createRR.Body.String())
	}

	var created dbpkg.EmpresaTarifaPorMinutos
	if err := json.Unmarshal(createRR.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if created.ID <= 0 {
		t.Fatalf("expected id > 0, got %d", created.ID)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_minutos?empresa_id=7&estacion_id=12&dia_semana=2&limit=50", nil)
	listRR := httptest.NewRecorder()
	h.ServeHTTP(listRR, listReq)
	if listRR.Code != http.StatusOK {
		t.Fatalf("list expected=%d got=%d body=%s", http.StatusOK, listRR.Code, listRR.Body.String())
	}
	var rows []dbpkg.EmpresaTarifaPorMinutos
	if err := json.Unmarshal(listRR.Body.Bytes(), &rows); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	updateBody := `{"id":` + strconv.FormatInt(created.ID, 10) + `,"empresa_id":7,"estacion_id":12,"estacion_codigo":"EST-7-12","estacion_nombre":"Habitacion 12","dia_semana_desde":1,"dia_semana_hasta":4,"minutos_base":120,"valor_base":32000,"minutos_extra":60,"valor_extra":16000,"moneda":"COP","prioridad":1}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/tarifas_por_minutos", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	h.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("update expected=%d got=%d body=%s", http.StatusOK, updateRR.Code, updateRR.Body.String())
	}

	calcReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_minutos?empresa_id=7&action=calcular&estacion_id=12&dia_semana=2&minutos_consumidos=190", nil)
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
	monto := calc["monto_total"].(float64)
	if math.Abs(monto-64000) > 0.001 {
		t.Fatalf("expected monto_total 64000, got %.2f", monto)
	}

	disableURL := "/api/empresa/tarifas_por_minutos?empresa_id=7&id=" + strconv.FormatInt(created.ID, 10) + "&action=desactivar"
	disableReq := httptest.NewRequest(http.MethodPut, disableURL, nil)
	disableRR := httptest.NewRecorder()
	h.ServeHTTP(disableRR, disableReq)
	if disableRR.Code != http.StatusOK {
		t.Fatalf("disable expected=%d got=%d body=%s", http.StatusOK, disableRR.Code, disableRR.Body.String())
	}

	aplicableReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_minutos?empresa_id=7&action=aplicable&estacion_id=12&dia_semana=2", nil)
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

	deleteURL := "/api/empresa/tarifas_por_minutos?empresa_id=7&id=" + strconv.FormatInt(created.ID, 10)
	deleteReq := httptest.NewRequest(http.MethodDelete, deleteURL, nil)
	deleteRR := httptest.NewRecorder()
	h.ServeHTTP(deleteRR, deleteReq)
	if deleteRR.Code != http.StatusOK {
		t.Fatalf("delete expected=%d got=%d body=%s", http.StatusOK, deleteRR.Code, deleteRR.Body.String())
	}

	detailReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_minutos?empresa_id=7&action=detalle&id="+strconv.FormatInt(created.ID, 10), nil)
	detailRR := httptest.NewRecorder()
	h.ServeHTTP(detailRR, detailReq)
	if detailRR.Code != http.StatusNotFound {
		t.Fatalf("detail after delete expected=%d got=%d body=%s", http.StatusNotFound, detailRR.Code, detailRR.Body.String())
	}
}
