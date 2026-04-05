package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
