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
	if math.Abs(monto-208095.24) > 0.02 {
		t.Fatalf("expected monto_total 208095.24, got %.2f", monto)
	}
	diasEquivalentes := calc["dias_equivalentes"].(float64)
	if math.Abs(diasEquivalentes-2.19) > 0.01 {
		t.Fatalf("expected dias_equivalentes about 2.19, got %.2f", diasEquivalentes)
	}
	if int(calc["dias_completos"].(float64)) != 2 {
		t.Fatalf("expected dias_completos 2, got %.0f", calc["dias_completos"].(float64))
	}
	if int(calc["minutos_prorrateo_fuera_ventana"].(float64)) != 240 {
		t.Fatalf("expected minutos_prorrateo_fuera_ventana 240, got %.0f", calc["minutos_prorrateo_fuera_ventana"].(float64))
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

func TestEmpresaTarifasPorDiaHandlerAplicarTodasEstaciones(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_tarifas_por_dia_handler_all.db")
	if err := dbpkg.EnsureEmpresaTarifasPorDiaSchema(dbEmp); err != nil {
		t.Fatalf("ensure tarifas por dia schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	_, err := dbEmp.Exec(`INSERT INTO carritos_compras (empresa_id, codigo, nombre, referencia_externa, estado, estado_carrito, moneda)
	VALUES
		(9, 'EST-9-21', 'Habitacion 21', 'ESTACION_21', 'activo', 'abierto', 'COP'),
		(9, 'EST-9-22', 'Habitacion 22', 'ESTACION_22', 'activo', 'abierto', 'COP')`)
	if err != nil {
		t.Fatalf("seed carritos estaciones: %v", err)
	}

	h := EmpresaTarifasPorDiaHandler(dbEmp)
	bodyCreate := `{"servicio_nombre":"hotel","valor_dia":70000,"hora_check_in":"15:00","hora_check_out":"12:00","moneda":"COP","prioridad":1,"aplicar_automaticamente":true}`
	url := "/api/empresa/tarifas_por_dia?empresa_id=9&action=aplicar_todas_estaciones"
	req := httptest.NewRequest(http.MethodPut, url, strings.NewReader(bodyCreate))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("aplicar_todas_estaciones create expected=%d got=%d body=%s", http.StatusOK, rr.Code, rr.Body.String())
	}

	var respCreate map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &respCreate); err != nil {
		t.Fatalf("decode aplicar_todas_estaciones create: %v", err)
	}
	if int(respCreate["tarifas_creadas"].(float64)) < 2 {
		t.Fatalf("expected tarifas_creadas >= 2, got %.0f", respCreate["tarifas_creadas"].(float64))
	}

	bodyUpdate := `{"servicio_nombre":"hotel","valor_dia":73000,"hora_check_in":"15:00","hora_check_out":"12:00","moneda":"COP","prioridad":2,"aplicar_automaticamente":true}`
	reqUpdate := httptest.NewRequest(http.MethodPut, url, strings.NewReader(bodyUpdate))
	reqUpdate.Header.Set("Content-Type", "application/json")
	rrUpdate := httptest.NewRecorder()
	h.ServeHTTP(rrUpdate, reqUpdate)
	if rrUpdate.Code != http.StatusOK {
		t.Fatalf("aplicar_todas_estaciones update expected=%d got=%d body=%s", http.StatusOK, rrUpdate.Code, rrUpdate.Body.String())
	}

	var respUpdate map[string]interface{}
	if err := json.Unmarshal(rrUpdate.Body.Bytes(), &respUpdate); err != nil {
		t.Fatalf("decode aplicar_todas_estaciones update: %v", err)
	}
	if int(respUpdate["tarifas_actualizadas"].(float64)) < 2 {
		t.Fatalf("expected tarifas_actualizadas >= 2, got %.0f", respUpdate["tarifas_actualizadas"].(float64))
	}
}
