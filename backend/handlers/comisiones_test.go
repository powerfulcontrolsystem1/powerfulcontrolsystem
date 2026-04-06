package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaComisionesServicioHandlerConfigAndReporte(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_comisiones_handler.db")
	ensureCarritosVentasSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaProductosSchema(dbEmp); err != nil {
		t.Fatalf("ensure productos schema: %v", err)
	}

	h := EmpresaComisionesServicioHandler(dbEmp)

	cfgBody := `{"empresa_id":1,"habilitar_comisiones":true,"porcentaje_comision":15,"filtro_servicio":"lavado","aplicar_automaticamente":true}`
	putReq := httptest.NewRequest(http.MethodPut, "/api/empresa/comisiones?empresa_id=1", strings.NewReader(cfgBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRR := httptest.NewRecorder()
	h.ServeHTTP(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected config upsert status %d, got %d body=%s", http.StatusOK, putRR.Code, putRR.Body.String())
	}

	if _, err := dbpkg.CreateEmpresaComisionServicioMovimiento(dbEmp, dbpkg.EmpresaComisionServicioMovimiento{
		EmpresaID:          1,
		CarritoID:          100,
		CarritoItemID:      200,
		ServicioID:         300,
		ServicioCodigo:     "LAV-001",
		ServicioNombre:     "Lavado premium",
		ServicioCategoria:  "lavado",
		UsuarioOrigen:      "cajero@empresa.com",
		UsuarioLavador:     "lavador@empresa.com",
		VentaReferencia:    "EST-1-001",
		Moneda:             "COP",
		BaseServicio:       10000,
		PorcentajeComision: 15,
		MontoComision:      1500,
		UsuarioCreador:     "cajero@empresa.com",
		Estado:             "activo",
	}); err != nil {
		t.Fatalf("seed comision movement: %v", err)
	}

	reportReq := httptest.NewRequest(http.MethodGet, "/api/empresa/comisiones?empresa_id=1&action=reporte&usuario_lavador=lavador&limit=50", nil)
	reportRR := httptest.NewRecorder()
	h.ServeHTTP(reportRR, reportReq)
	if reportRR.Code != http.StatusOK {
		t.Fatalf("expected report status %d, got %d body=%s", http.StatusOK, reportRR.Code, reportRR.Body.String())
	}

	var report dbpkg.EmpresaComisionesServicioReporte
	if err := json.Unmarshal(reportRR.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report response: %v", err)
	}
	if report.Resumen.CantidadMovimientos == 0 {
		t.Fatalf("expected movimientos > 0, got %+v", report.Resumen)
	}
	if len(report.Lavadores) == 0 {
		t.Fatal("expected lavadores rows in report")
	}
	if report.Lavadores[0].UsuarioLavador != "lavador@empresa.com" {
		t.Fatalf("expected lavador lavador@empresa.com, got %q", report.Lavadores[0].UsuarioLavador)
	}
}

func TestEmpresaComisionesServicioHandlerEscalasYAjustes(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_comisiones_handler_escalas_ajustes.db")
	if err := dbpkg.EnsureEmpresaComisionesServicioSchema(dbEmp); err != nil {
		t.Fatalf("ensure comisiones schema: %v", err)
	}

	h := EmpresaComisionesServicioHandler(dbEmp)

	createEscala := `{"empresa_id":3,"rol_operacion":"cajero","servicio_filtro":"premium","porcentaje_comision":22,"tope_comision":1800,"prioridad":1}`
	reqEscala := httptest.NewRequest(http.MethodPost, "/api/empresa/comisiones?action=escala&empresa_id=3", strings.NewReader(createEscala))
	reqEscala.Header.Set("Content-Type", "application/json")
	rrEscala := httptest.NewRecorder()
	h.ServeHTTP(rrEscala, reqEscala)
	if rrEscala.Code != http.StatusCreated {
		t.Fatalf("create escala expected=%d got=%d body=%s", http.StatusCreated, rrEscala.Code, rrEscala.Body.String())
	}

	listEscalasReq := httptest.NewRequest(http.MethodGet, "/api/empresa/comisiones?action=escalas&empresa_id=3&limit=20", nil)
	listEscalasRR := httptest.NewRecorder()
	h.ServeHTTP(listEscalasRR, listEscalasReq)
	if listEscalasRR.Code != http.StatusOK {
		t.Fatalf("list escalas expected=%d got=%d body=%s", http.StatusOK, listEscalasRR.Code, listEscalasRR.Body.String())
	}
	var escalas []dbpkg.EmpresaComisionServicioEscala
	if err := json.Unmarshal(listEscalasRR.Body.Bytes(), &escalas); err != nil {
		t.Fatalf("decode escalas: %v", err)
	}
	if len(escalas) != 1 {
		t.Fatalf("expected 1 escala, got %d", len(escalas))
	}

	createAjuste := `{"empresa_id":3,"usuario_lavador":"emp-qa-1","monto_ajuste":900,"motivo":"ajuste operativo validado"}`
	reqAjuste := httptest.NewRequest(http.MethodPost, "/api/empresa/comisiones?action=ajuste_manual&empresa_id=3", strings.NewReader(createAjuste))
	reqAjuste.Header.Set("Content-Type", "application/json")
	rrAjuste := httptest.NewRecorder()
	h.ServeHTTP(rrAjuste, reqAjuste)
	if rrAjuste.Code != http.StatusCreated {
		t.Fatalf("create ajuste expected=%d got=%d body=%s", http.StatusCreated, rrAjuste.Code, rrAjuste.Body.String())
	}

	var ajusteResp map[string]interface{}
	if err := json.Unmarshal(rrAjuste.Body.Bytes(), &ajusteResp); err != nil {
		t.Fatalf("decode ajuste response: %v", err)
	}
	movID, _ := ajusteResp["id"].(float64)
	if movID <= 0 {
		t.Fatalf("expected ajuste id > 0, got %v", ajusteResp["id"])
	}
	movimientoID := int64(movID)

	approveBody := fmt.Sprintf(`{"empresa_id":3,"movimiento_id":%d,"observaciones":"aprobado por prueba"}`, movimientoID)
	approveReq := httptest.NewRequest(http.MethodPut, "/api/empresa/comisiones?action=aprobar_ajuste&empresa_id=3", strings.NewReader(approveBody))
	approveReq.Header.Set("Content-Type", "application/json")
	approveRR := httptest.NewRecorder()
	h.ServeHTTP(approveRR, approveReq)
	if approveRR.Code != http.StatusOK {
		t.Fatalf("approve ajuste expected=%d got=%d body=%s", http.StatusOK, approveRR.Code, approveRR.Body.String())
	}

	listMovReq := httptest.NewRequest(http.MethodGet, "/api/empresa/comisiones?action=movimientos&empresa_id=3&solo_ajustes=true&ajuste_estado=aprobado&include_inactive=true&limit=20", nil)
	listMovRR := httptest.NewRecorder()
	h.ServeHTTP(listMovRR, listMovReq)
	if listMovRR.Code != http.StatusOK {
		t.Fatalf("list movimientos expected=%d got=%d body=%s", http.StatusOK, listMovRR.Code, listMovRR.Body.String())
	}
	var movs []dbpkg.EmpresaComisionServicioMovimiento
	if err := json.Unmarshal(listMovRR.Body.Bytes(), &movs); err != nil {
		t.Fatalf("decode movimientos: %v", err)
	}
	if len(movs) != 1 {
		t.Fatalf("expected 1 ajuste aprobado, got %d", len(movs))
	}
	if movs[0].AjusteEstado != dbpkg.EmpresaComisionServicioAjusteAprobado {
		t.Fatalf("expected ajuste_estado aprobado, got %s", movs[0].AjusteEstado)
	}
}
