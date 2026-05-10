package db

import (
	"strings"
	"testing"
)

func TestDefaultDrogueriaFarmaciaLicenciaPlans(t *testing.T) {
	plans := DefaultDrogueriaFarmaciaLicenciaPlans()
	if len(plans) != 4 {
		t.Fatalf("planes drogueria farmacia = %d, want 4", len(plans))
	}
	expectedValues := []float64{0, 60000, 100000, 150000}
	expectedDocs := []int{250, 1000, 2000, 4000}
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
		if !strings.Contains(plan.ModulosHabilitados, "drogueria_farmacia") {
			t.Fatalf("plan %d no habilita drogueria_farmacia: %q", i, plan.ModulosHabilitados)
		}
	}
}
