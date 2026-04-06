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
	if calc["trazabilidad_contable_id"].(float64) <= 0 {
		t.Fatalf("expected trazabilidad_contable_id > 0, got %v", calc["trazabilidad_contable_id"])
	}
	if strings.TrimSpace(calc["documento_codigo"].(string)) == "" {
		t.Fatalf("expected documento_codigo in calcular response")
	}

	eventos, err := dbpkg.ListEmpresaEventosContables(dbEmp, 7, dbpkg.EmpresaEventoContableFilter{Evento: "tarifa_por_minutos_calculada", Limit: 20})
	if err != nil {
		t.Fatalf("list eventos contables trazabilidad: %v", err)
	}
	if len(eventos) != 1 {
		t.Fatalf("expected 1 evento contable de trazabilidad, got %d", len(eventos))
	}
	if !eventos[0].Procesado {
		t.Fatalf("expected evento procesado=true")
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

func TestEmpresaTarifasPorMinutosHandlerConfigYAplicacionMasiva(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_tarifas_por_minutos_handler_config.db")
	if err := dbpkg.EnsureEmpresaTarifasPorMinutosSchema(dbEmp); err != nil {
		t.Fatalf("ensure tarifas por minutos schema: %v", err)
	}
	if err := dbpkg.EnsureEmpresaCarritosSchema(dbEmp); err != nil {
		t.Fatalf("ensure carritos schema: %v", err)
	}

	if _, err := dbEmp.Exec(`INSERT INTO carritos_compras (empresa_id, codigo, nombre, referencia_externa, estado, estado_carrito, moneda)
	VALUES
		(9, 'EST-9-101', 'Estacion 101', 'ESTACION_101', 'activo', 'abierto', 'COP'),
		(9, 'EST-9-102', 'Estacion 102', 'ESTACION_102', 'activo', 'abierto', 'COP')`); err != nil {
		t.Fatalf("seed carritos estaciones: %v", err)
	}

	h := EmpresaTarifasPorMinutosHandler(dbEmp)

	configBody := `{"empresa_id":9,"redondeo_modo":"arriba","redondeo_unidad":100,"monto_minimo_diario":50000,"monto_maximo_diario":60000}`
	configReq := httptest.NewRequest(http.MethodPut, "/api/empresa/tarifas_por_minutos?action=config", strings.NewReader(configBody))
	configReq.Header.Set("Content-Type", "application/json")
	configRR := httptest.NewRecorder()
	h.ServeHTTP(configRR, configReq)
	if configRR.Code != http.StatusOK {
		t.Fatalf("config expected=%d got=%d body=%s", http.StatusOK, configRR.Code, configRR.Body.String())
	}

	var cfg dbpkg.EmpresaTarifaPorMinutosConfiguracion
	if err := json.Unmarshal(configRR.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if cfg.RedondeoModo != "arriba" {
		t.Fatalf("expected redondeo_modo arriba, got %q", cfg.RedondeoModo)
	}

	applyBody := `{"empresa_id":9,"dia_semana_desde":1,"dia_semana_hasta":7,"minutos_base":120,"valor_base":30000,"minutos_extra":60,"valor_extra":15000,"moneda":"COP","prioridad":1,"estado":"activo"}`
	applyReq := httptest.NewRequest(http.MethodPut, "/api/empresa/tarifas_por_minutos?action=aplicar_todas_estaciones", strings.NewReader(applyBody))
	applyReq.Header.Set("Content-Type", "application/json")
	applyRR := httptest.NewRecorder()
	h.ServeHTTP(applyRR, applyReq)
	if applyRR.Code != http.StatusOK {
		t.Fatalf("aplicar_todas_estaciones expected=%d got=%d body=%s", http.StatusOK, applyRR.Code, applyRR.Body.String())
	}

	var applyResp map[string]interface{}
	if err := json.Unmarshal(applyRR.Body.Bytes(), &applyResp); err != nil {
		t.Fatalf("decode aplicar_todas_estaciones response: %v", err)
	}
	if int(applyResp["tarifas_creadas"].(float64)) < 2 {
		t.Fatalf("expected at least 2 tarifas_creadas, got %v", applyResp["tarifas_creadas"])
	}

	calcReq := httptest.NewRequest(http.MethodGet, "/api/empresa/tarifas_por_minutos?empresa_id=9&action=calcular&estacion_id=101&dia_semana=2&minutos_consumidos=120.5", nil)
	calcRR := httptest.NewRecorder()
	h.ServeHTTP(calcRR, calcReq)
	if calcRR.Code != http.StatusOK {
		t.Fatalf("calcular with decimal expected=%d got=%d body=%s", http.StatusOK, calcRR.Code, calcRR.Body.String())
	}

	var calc map[string]interface{}
	if err := json.Unmarshal(calcRR.Body.Bytes(), &calc); err != nil {
		t.Fatalf("decode calcular decimal response: %v", err)
	}
	if !calc["monto_minimo_aplicado"].(bool) {
		t.Fatalf("expected monto_minimo_aplicado=true")
	}
	if math.Abs(calc["monto_total"].(float64)-50000) > 0.001 {
		t.Fatalf("expected monto_total 50000 by min limit, got %.2f", calc["monto_total"].(float64))
	}

	rows, err := dbpkg.ListEmpresaTarifasPorMinutos(dbEmp, 9, dbpkg.EmpresaTarifaPorMinutosFilter{EstacionID: 102, DiaSemana: 2, Limit: 20})
	if err != nil {
		t.Fatalf("list tarifas station 102: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 tarifa for station 102, got %d", len(rows))
	}
}
