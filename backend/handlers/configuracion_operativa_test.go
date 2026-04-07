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

func TestEmpresaConfiguracionOperativaHandlerConfigAndRole(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_configuracion_operativa_handler.db")
	if err := dbpkg.EnsureEmpresaConfiguracionOperativaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion operativa schema: %v", err)
	}

	h := EmpresaConfiguracionOperativaHandler(dbEmp)

	baseBody := `{"empresa_id":1,"metodo_pago_efectivo":true,"metodo_pago_tarjeta_credito":true,"metodo_pago_tarjeta_debito":true,"metodo_pago_transferencia_bancaria":false,"metodo_pago_mixto":true,"metodo_pago_codigo_descuento":false,"habilitar_propinas":true,"habilitar_comisiones":true}`
	baseReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?empresa_id=1", strings.NewReader(baseBody))
	baseReq.Header.Set("Content-Type", "application/json")
	baseRR := httptest.NewRecorder()
	h.ServeHTTP(baseRR, baseReq)
	if baseRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for base upsert, got %d body=%s", http.StatusOK, baseRR.Code, baseRR.Body.String())
	}

	roleBody := `{"empresa_id":1,"rol":"cajero","metodo_pago_efectivo":true,"metodo_pago_tarjeta_credito":false,"metodo_pago_tarjeta_debito":false,"metodo_pago_transferencia_bancaria":false,"metodo_pago_mixto":false,"metodo_pago_codigo_descuento":false,"habilitar_propinas":false,"habilitar_comisiones":false}`
	roleReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?action=rol", strings.NewReader(roleBody))
	roleReq.Header.Set("Content-Type", "application/json")
	roleRR := httptest.NewRecorder()
	h.ServeHTTP(roleRR, roleReq)
	if roleRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for role upsert, got %d body=%s", http.StatusOK, roleRR.Code, roleRR.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_operativa?empresa_id=1", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for get, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}

	var cfg dbpkg.EmpresaConfiguracionOperativa
	if err := json.Unmarshal(getRR.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if cfg.MetodoPagoTransferenciaBancaria {
		t.Fatalf("expected transfer disabled at company level, got %+v", cfg)
	}
	if cfg.MetodoPagoCodigoDescuento {
		t.Fatalf("expected codigo_descuento disabled at company level, got %+v", cfg)
	}
	if len(cfg.Roles) != 1 {
		t.Fatalf("expected 1 role row, got %d", len(cfg.Roles))
	}
	if cfg.Roles[0].Rol != "cajero" {
		t.Fatalf("expected role cajero, got %q", cfg.Roles[0].Rol)
	}
	if cfg.Roles[0].MetodoPagoTarjetaCredito || cfg.Roles[0].HabilitarPropinas || cfg.Roles[0].HabilitarComisiones {
		t.Fatalf("expected cajero restrictions persisted, got %+v", cfg.Roles[0])
	}
}

func TestEmpresaConfiguracionOperativaHandlerPoliticaSimulacionHistorialYRollback(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_configuracion_operativa_handler_politica.db")
	if err := dbpkg.EnsureEmpresaConfiguracionOperativaSchema(dbEmp); err != nil {
		t.Fatalf("ensure configuracion operativa schema: %v", err)
	}

	h := EmpresaConfiguracionOperativaHandler(dbEmp)

	baseBody := `{"empresa_id":8,"metodo_pago_efectivo":true,"metodo_pago_tarjeta_credito":true,"metodo_pago_tarjeta_debito":true,"metodo_pago_transferencia_bancaria":true,"metodo_pago_mixto":true,"metodo_pago_codigo_descuento":false,"habilitar_propinas":true,"habilitar_comisiones":true}`
	baseReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?empresa_id=8", strings.NewReader(baseBody))
	baseReq.Header.Set("Content-Type", "application/json")
	baseRR := httptest.NewRecorder()
	h.ServeHTTP(baseRR, baseReq)
	if baseRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for base upsert, got %d body=%s", http.StatusOK, baseRR.Code, baseRR.Body.String())
	}

	var baseResp map[string]interface{}
	if err := json.Unmarshal(baseRR.Body.Bytes(), &baseResp); err != nil {
		t.Fatalf("decode base response: %v", err)
	}
	baseHistorialID := int64(0)
	if rawID, ok := baseResp["historial_id"].(float64); ok {
		baseHistorialID = int64(rawID)
	}
	if baseHistorialID <= 0 {
		t.Fatalf("expected historial_id in base response, got %+v", baseResp)
	}

	roleBody := `{"empresa_id":8,"rol":"cajero","metodo_pago_efectivo":true,"metodo_pago_tarjeta_credito":false,"metodo_pago_tarjeta_debito":false,"metodo_pago_transferencia_bancaria":false,"metodo_pago_mixto":false,"metodo_pago_codigo_descuento":false,"habilitar_propinas":false,"habilitar_comisiones":false}`
	roleReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?action=rol", strings.NewReader(roleBody))
	roleReq.Header.Set("Content-Type", "application/json")
	roleRR := httptest.NewRecorder()
	h.ServeHTTP(roleRR, roleReq)
	if roleRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for role upsert, got %d body=%s", http.StatusOK, roleRR.Code, roleRR.Body.String())
	}

	politicaBody := `{"empresa_id":8,"canal_venta":"app","sucursal_id":2,"turno":"noche","prioridad":5,"metodo_pago_efectivo":false,"metodo_pago_tarjeta_credito":true,"metodo_pago_tarjeta_debito":true,"metodo_pago_transferencia_bancaria":false,"metodo_pago_mixto":false,"metodo_pago_codigo_descuento":false,"habilitar_propinas":true,"habilitar_comisiones":false}`
	politicaReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?action=politica", strings.NewReader(politicaBody))
	politicaReq.Header.Set("Content-Type", "application/json")
	politicaRR := httptest.NewRecorder()
	h.ServeHTTP(politicaRR, politicaReq)
	if politicaRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for politica upsert, got %d body=%s", http.StatusOK, politicaRR.Code, politicaRR.Body.String())
	}

	simReq := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_operativa?action=simular&empresa_id=8&rol=cajero&canal_venta=app&sucursal_id=2&turno=noche", nil)
	simRR := httptest.NewRecorder()
	h.ServeHTTP(simRR, simReq)
	if simRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for simular GET, got %d body=%s", http.StatusOK, simRR.Code, simRR.Body.String())
	}

	var simResp struct {
		Permisos dbpkg.EmpresaConfiguracionOperativaPermisos `json:"permisos_simulados"`
	}
	if err := json.Unmarshal(simRR.Body.Bytes(), &simResp); err != nil {
		t.Fatalf("decode simular response: %v", err)
	}
	if !simResp.Permisos.PoliticaAplicada || simResp.Permisos.Fuente != "politica" {
		t.Fatalf("expected policy source in simulated permissions, got %+v", simResp.Permisos)
	}
	if simResp.Permisos.MetodoPagoEfectivo || !simResp.Permisos.MetodoPagoTarjetaCredito {
		t.Fatalf("expected policy payment mix in simulation, got %+v", simResp.Permisos)
	}

	updateBody := `{"empresa_id":8,"metodo_pago_efectivo":false,"metodo_pago_tarjeta_credito":false,"metodo_pago_tarjeta_debito":false,"metodo_pago_transferencia_bancaria":false,"metodo_pago_mixto":false,"metodo_pago_codigo_descuento":false,"habilitar_propinas":false,"habilitar_comisiones":false}`
	updateReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?empresa_id=8", strings.NewReader(updateBody))
	updateReq.Header.Set("Content-Type", "application/json")
	updateRR := httptest.NewRecorder()
	h.ServeHTTP(updateRR, updateReq)
	if updateRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for base update, got %d body=%s", http.StatusOK, updateRR.Code, updateRR.Body.String())
	}

	rollbackBody := `{"empresa_id":8,"historial_id":` + strconv.FormatInt(baseHistorialID, 10) + `,"observaciones":"rollback de prueba"}`
	rollbackReq := httptest.NewRequest(http.MethodPut, "/api/empresa/configuracion_operativa?action=rollback", strings.NewReader(rollbackBody))
	rollbackReq.Header.Set("Content-Type", "application/json")
	rollbackRR := httptest.NewRecorder()
	h.ServeHTTP(rollbackRR, rollbackReq)
	if rollbackRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for rollback, got %d body=%s", http.StatusOK, rollbackRR.Code, rollbackRR.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_operativa?empresa_id=8", nil)
	getRR := httptest.NewRecorder()
	h.ServeHTTP(getRR, getReq)
	if getRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for get after rollback, got %d body=%s", http.StatusOK, getRR.Code, getRR.Body.String())
	}
	var cfg dbpkg.EmpresaConfiguracionOperativa
	if err := json.Unmarshal(getRR.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode get after rollback: %v", err)
	}
	if !cfg.MetodoPagoEfectivo || !cfg.MetodoPagoTarjetaCredito {
		t.Fatalf("expected base config restored after rollback, got %+v", cfg)
	}

	histReq := httptest.NewRequest(http.MethodGet, "/api/empresa/configuracion_operativa?action=historial&empresa_id=8&limit=10", nil)
	histRR := httptest.NewRecorder()
	h.ServeHTTP(histRR, histReq)
	if histRR.Code != http.StatusOK {
		t.Fatalf("expected status %d for historial list, got %d body=%s", http.StatusOK, histRR.Code, histRR.Body.String())
	}
	var histResp struct {
		Historial []dbpkg.EmpresaConfiguracionOperativaHistorialSnapshot `json:"historial"`
	}
	if err := json.Unmarshal(histRR.Body.Bytes(), &histResp); err != nil {
		t.Fatalf("decode historial response: %v", err)
	}
	if len(histResp.Historial) == 0 {
		t.Fatalf("expected historial rows after operations, got %+v", histResp)
	}
}
