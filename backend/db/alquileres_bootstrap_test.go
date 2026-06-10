package db

import (
	"strings"
	"testing"
)

func TestDefaultAlquileresLicenciaPlans(t *testing.T) {
	plans := DefaultAlquileresLicenciaPlans()
	if len(plans) != 7 {
		t.Fatalf("planes alquileres = %d, want 7", len(plans))
	}
	expectedValues := []float64{0, 60000, 110000, 200000, 600000, 1100000, 2200000}
	expectedDocs := []int{250, 1000, 2000, 4000, 12000, 24000, 36000}
	for i, plan := range plans {
		if plan.Valor != expectedValues[i] {
			t.Fatalf("valor plan %d = %.0f, want %.0f", i, plan.Valor, expectedValues[i])
		}
		if plan.MaxDocumentosMensuales != expectedDocs[i] {
			t.Fatalf("documentos plan %d = %d, want %d", i, plan.MaxDocumentosMensuales, expectedDocs[i])
		}
		if plan.DuracionDias != 15 && plan.DuracionDias != 30 && plan.DuracionDias != 365 {
			t.Fatalf("duracion plan %d inesperada: %d", i, plan.DuracionDias)
		}
		if !strings.Contains(plan.ModulosHabilitados, "alquileres") {
			t.Fatalf("plan %d no habilita alquileres: %q", i, plan.ModulosHabilitados)
		}
		if !strings.Contains(plan.ModulosHabilitados, "contratos_obligaciones") {
			t.Fatalf("plan %d no habilita contratos_obligaciones: %q", i, plan.ModulosHabilitados)
		}
	}
}
