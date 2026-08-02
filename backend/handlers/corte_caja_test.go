package handlers

import (
	"net/http/httptest"
	"testing"

	dbpkg "github.com/you/pos-backend/db"
)

func TestCorteCajaSoloUsuarioCajaActualRecognizesReporteMiTurno(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/empresa/corte_caja?empresa_id=1&action=reporte_mi_turno", nil)
	if !corteCajaSoloUsuarioCajaActual(req) {
		t.Fatalf("expected reporte_mi_turno to use current-user cash close report")
	}

	plainReq := httptest.NewRequest("GET", "/api/empresa/corte_caja?empresa_id=1", nil)
	if corteCajaSoloUsuarioCajaActual(plainReq) {
		t.Fatalf("plain corte_caja request should not be forced to current-user cash report")
	}
}

func TestApplyCorteCajaPersistedCashSummaryIncludesMixedCash(t *testing.T) {
	resumen := corteCajaResumen{
		AperturaEfectivo:     0,
		EfectivoVentas:       0,
		IngresosEfectivo:     0,
		EgresosEfectivo:      0,
		OtrosMediosVentas:    100,
		EfectivoEsperadoCaja: 0,
	}
	cierre := dbpkg.EmpresaCierreCaja{
		AperturaMonto:    0,
		IngresosEfectivo: 50,
		EgresosEfectivo:  0,
		RetirosEfectivo:  0,
	}

	applyCorteCajaPersistedCashSummary(&resumen, &cierre)

	if resumen.EfectivoVentas != 50 {
		t.Fatalf("expected mixed cash portion 50, got %.2f", resumen.EfectivoVentas)
	}
	if resumen.EfectivoEsperadoCaja != 50 {
		t.Fatalf("expected cash drawer 50, got %.2f", resumen.EfectivoEsperadoCaja)
	}
	if resumen.OtrosMediosVentas != 50 {
		t.Fatalf("expected remaining non-cash mixed portion 50, got %.2f", resumen.OtrosMediosVentas)
	}
}
