package handlers

import (
	"testing"
	"time"

	dbpkg "github.com/you/pos-backend/db"
)

func TestAsesorComercialCommissionRateForPaymentByYearStage(t *testing.T) {
	advisor := &dbpkg.AsesorComercial{
		PorcentajePrimerAnio:      40,
		PorcentajeRenovacionAnual: 30,
		MesesRenovacion:           24,
	}
	start := time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

	pct, until, stage, ok := asesorComercialCommissionRateForPayment(advisor, start, start.AddDate(0, 6, 0))
	if !ok || pct != 40 || stage != "primer_anio" {
		t.Fatalf("first year commission = pct %.2f stage %q ok %v, want 40 primer_anio true", pct, stage, ok)
	}
	if want := start.AddDate(3, 0, 0); !until.Equal(want) {
		t.Fatalf("association until = %s, want %s", until.Format("2006-01-02"), want.Format("2006-01-02"))
	}

	pct, _, stage, ok = asesorComercialCommissionRateForPayment(advisor, start, start.AddDate(1, 6, 0))
	if !ok || pct != 30 || stage != "renovacion_anual" {
		t.Fatalf("renewal commission = pct %.2f stage %q ok %v, want 30 renovacion_anual true", pct, stage, ok)
	}

	pct, _, stage, ok = asesorComercialCommissionRateForPayment(advisor, start, start.AddDate(3, 0, 1))
	if ok || pct != 0 || stage != "fuera_de_plazo" {
		t.Fatalf("expired commission = pct %.2f stage %q ok %v, want 0 fuera_de_plazo false", pct, stage, ok)
	}
}
