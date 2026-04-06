package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestEmpresaPropinasHandlerConfigAndReporte(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_propinas_handler.db")
	ensureCarritosVentasSchema(t, dbEmp)
	ensureEmpresaUsersSchema(t, dbEmp)

	if _, err := dbEmp.Exec(`INSERT INTO users (email, name, role, empresa_id, estado) VALUES
		('meseroa@empresa.com', 'Mesero A', 'cajero', 1, 'activo'),
		('meserob@empresa.com', 'Mesero B', 'cajero', 1, 'activo')`); err != nil {
		t.Fatalf("seed users: %v", err)
	}

	h := EmpresaPropinasHandler(dbEmp)

	cfgBody := `{"empresa_id":1,"habilitar_propina":true,"porcentaje_propina":12,"modo_distribucion":"universal","aplicar_automaticamente":true}`
	putReq := httptest.NewRequest(http.MethodPut, "/api/empresa/propinas?empresa_id=1", strings.NewReader(cfgBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRR := httptest.NewRecorder()
	h.ServeHTTP(putRR, putReq)
	if putRR.Code != http.StatusOK {
		t.Fatalf("expected config upsert status %d, got %d body=%s", http.StatusOK, putRR.Code, putRR.Body.String())
	}

	cfgReq := httptest.NewRequest(http.MethodGet, "/api/empresa/propinas?empresa_id=1&action=config", nil)
	cfgRR := httptest.NewRecorder()
	h.ServeHTTP(cfgRR, cfgReq)
	if cfgRR.Code != http.StatusOK {
		t.Fatalf("expected config get status %d, got %d body=%s", http.StatusOK, cfgRR.Code, cfgRR.Body.String())
	}

	var cfg dbpkg.EmpresaPropinasConfiguracion
	if err := json.Unmarshal(cfgRR.Body.Bytes(), &cfg); err != nil {
		t.Fatalf("decode config response: %v", err)
	}
	if !cfg.HabilitarPropina {
		t.Fatalf("expected cfg habilitar_propina=true, got %+v", cfg)
	}
	if cfg.ModoDistribucion != dbpkg.EmpresaPropinaModoUniversal {
		t.Fatalf("expected modo universal, got %q", cfg.ModoDistribucion)
	}

	if _, err := dbpkg.CreateEmpresaPropinaMovimiento(dbEmp, dbpkg.EmpresaPropinaMovimiento{
		EmpresaID:         1,
		CarritoID:         101,
		VentaReferencia:   "CAJ-101",
		UsuarioOrigen:     "meseroa@empresa.com",
		ModoDistribucion:  dbpkg.EmpresaPropinaModoUniversal,
		Moneda:            "COP",
		BaseCobro:         25000,
		PorcentajePropina: 12,
		MontoPropina:      3000,
		UsuarioCreador:    "meseroa@empresa.com",
	}); err != nil {
		t.Fatalf("seed universal tip movement: %v", err)
	}

	reportReq := httptest.NewRequest(http.MethodGet, "/api/empresa/propinas?empresa_id=1&action=reporte&limit=50", nil)
	reportRR := httptest.NewRecorder()
	h.ServeHTTP(reportRR, reportReq)
	if reportRR.Code != http.StatusOK {
		t.Fatalf("expected report status %d, got %d body=%s", http.StatusOK, reportRR.Code, reportRR.Body.String())
	}

	var report dbpkg.EmpresaPropinasReporte
	if err := json.Unmarshal(reportRR.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode report response: %v", err)
	}
	if report.Resumen.CantidadMovimientos == 0 {
		t.Fatalf("expected movimientos > 0, got %+v", report.Resumen)
	}
	if report.Resumen.UsuariosActivos != 2 {
		t.Fatalf("expected usuarios activos 2, got %d", report.Resumen.UsuariosActivos)
	}
	if len(report.Usuarios) == 0 {
		t.Fatal("expected usuarios rows in report")
	}
	if len(report.Movimientos) == 0 {
		t.Fatal("expected movimientos rows in report")
	}
}

func TestEmpresaPropinasHandlerAjusteManualYConciliacionCierre(t *testing.T) {
	dbEmp := openTestSQLite(t, "empresas_propinas_handler_ajuste_cierre.db")
	ensureCarritosVentasSchema(t, dbEmp)
	ensureEmpresaUsersSchema(t, dbEmp)
	if err := dbpkg.EnsureEmpresaFinanzasSchema(dbEmp); err != nil {
		t.Fatalf("ensure finanzas schema: %v", err)
	}

	h := EmpresaPropinasHandler(dbEmp)

	cfgBody := `{"empresa_id":1,"habilitar_propina":true,"porcentaje_propina":10,"modo_distribucion":"por_usuario","aplicar_automaticamente":true,"tratamiento_fiscal":"gravada","porcentaje_impuesto_propina":19}`
	putCfgReq := httptest.NewRequest(http.MethodPut, "/api/empresa/propinas?empresa_id=1", strings.NewReader(cfgBody))
	putCfgReq.Header.Set("Content-Type", "application/json")
	putCfgRR := httptest.NewRecorder()
	h.ServeHTTP(putCfgRR, putCfgReq)
	if putCfgRR.Code != http.StatusOK {
		t.Fatalf("expected config status %d, got %d body=%s", http.StatusOK, putCfgRR.Code, putCfgRR.Body.String())
	}

	fechaOperacion := time.Now().Format("2006-01-02")
	cierreID, err := dbpkg.CreateEmpresaCierreCaja(dbEmp, dbpkg.EmpresaCierreCaja{
		EmpresaID:      1,
		SucursalID:     1,
		CajaCodigo:     "CAJA-1",
		Turno:          "manana",
		FechaOperacion: fechaOperacion,
		EstadoCierre:   "cerrado",
		UsuarioCreador: "admin@empresa.com",
	})
	if err != nil {
		t.Fatalf("create cierre caja: %v", err)
	}

	if _, err := dbpkg.CreateEmpresaPropinaMovimiento(dbEmp, dbpkg.EmpresaPropinaMovimiento{
		EmpresaID:         1,
		CarritoID:         401,
		CierreCajaID:      cierreID,
		VentaReferencia:   "VEN-401",
		UsuarioOrigen:     "cajero@empresa.com",
		UsuarioAsignado:   "cajero@empresa.com",
		ModoDistribucion:  dbpkg.EmpresaPropinaModoPorUsuario,
		Moneda:            "COP",
		BaseCobro:         10000,
		PorcentajePropina: 10,
		MontoPropina:      1000,
		FechaMovimiento:   fechaOperacion + " 10:30:00",
		UsuarioCreador:    "cajero@empresa.com",
	}); err != nil {
		t.Fatalf("seed movement: %v", err)
	}

	ajusteBody := fmt.Sprintf(`{"empresa_id":1,"cierre_caja_id":%d,"usuario_asignado":"cajero@empresa.com","modo_distribucion":"por_usuario","monto_ajuste":500,"motivo":"regularizacion de caja"}`,
		cierreID,
	)
	putAjusteReq := httptest.NewRequest(http.MethodPut, "/api/empresa/propinas?action=ajuste_manual", strings.NewReader(ajusteBody))
	putAjusteReq.Header.Set("Content-Type", "application/json")
	putAjusteReq.Header.Set("X-Admin-Email", "auditor@empresa.com")
	putAjusteRR := httptest.NewRecorder()
	h.ServeHTTP(putAjusteRR, putAjusteReq)
	if putAjusteRR.Code != http.StatusOK {
		t.Fatalf("expected ajuste status %d, got %d body=%s", http.StatusOK, putAjusteRR.Code, putAjusteRR.Body.String())
	}

	var ajusteResp map[string]interface{}
	if err := json.Unmarshal(putAjusteRR.Body.Bytes(), &ajusteResp); err != nil {
		t.Fatalf("decode ajuste response: %v", err)
	}
	if gotOK, _ := ajusteResp["ok"].(bool); !gotOK {
		t.Fatalf("expected ok=true in ajuste response, got %+v", ajusteResp)
	}

	concReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/empresa/propinas?empresa_id=1&action=conciliacion_cierre&cierre_caja_id=%d", cierreID), nil)
	concReq.Header.Set("X-Admin-Email", "auditor@empresa.com")
	concRR := httptest.NewRecorder()
	h.ServeHTTP(concRR, concReq)
	if concRR.Code != http.StatusOK {
		t.Fatalf("expected conciliacion status %d, got %d body=%s", http.StatusOK, concRR.Code, concRR.Body.String())
	}

	var conciliacion dbpkg.EmpresaPropinaConciliacionCierre
	if err := json.Unmarshal(concRR.Body.Bytes(), &conciliacion); err != nil {
		t.Fatalf("decode conciliacion response: %v", err)
	}
	if conciliacion.CierreCajaID != cierreID {
		t.Fatalf("expected cierre_id=%d, got %+v", cierreID, conciliacion)
	}
	if conciliacion.CantidadMovimientos < 2 {
		t.Fatalf("expected at least 2 movimientos conciliados, got %+v", conciliacion)
	}
	if conciliacion.TotalAjustes <= 0 {
		t.Fatalf("expected ajustes > 0, got %+v", conciliacion)
	}

	var auditoriaCount int
	if err := dbEmp.QueryRow(`SELECT COUNT(1)
	FROM empresa_auditoria_eventos
	WHERE empresa_id = 1 AND modulo = 'propinas' AND accion = 'ajuste_manual'`).Scan(&auditoriaCount); err != nil {
		t.Fatalf("query auditoria events: %v", err)
	}
	if auditoriaCount == 0 {
		t.Fatal("expected auditoria event for ajuste_manual")
	}

	cierres, err := dbpkg.ListEmpresaCierresCaja(dbEmp, 1, dbpkg.EmpresaCierreCajaFilter{Limit: 10})
	if err != nil {
		t.Fatalf("list cierres caja: %v", err)
	}
	if len(cierres) == 0 {
		t.Fatal("expected at least one cierre caja")
	}
	if cierres[0].PropinasMovimientos == 0 {
		t.Fatalf("expected propinas_movimientos updated in cierre, got %+v", cierres[0])
	}
}
