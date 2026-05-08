package db

import (
	"strings"
	"testing"
)

func TestDefaultAlquileresLicenciaPlans(t *testing.T) {
	plans := DefaultAlquileresLicenciaPlans()
	if len(plans) != 4 {
		t.Fatalf("planes alquileres = %d, want 4", len(plans))
	}
	expectedValues := []float64{0, 60000, 100000, 150000}
	expectedDocs := []int{500, 1000, 2000, 5000}
	for i, plan := range plans {
		if plan.Valor != expectedValues[i] {
			t.Fatalf("valor plan %d = %.0f, want %.0f", i, plan.Valor, expectedValues[i])
		}
		if plan.MaxDocumentosMensuales != expectedDocs[i] {
			t.Fatalf("documentos plan %d = %d, want %d", i, plan.MaxDocumentosMensuales, expectedDocs[i])
		}
		if plan.DuracionDias != 15 && plan.DuracionDias != 30 {
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
