package handlers

import (
	"encoding/json"
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
