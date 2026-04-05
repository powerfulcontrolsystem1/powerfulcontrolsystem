package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
