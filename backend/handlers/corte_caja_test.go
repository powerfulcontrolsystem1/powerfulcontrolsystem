package handlers

import (
	"net/http/httptest"
	"testing"
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
