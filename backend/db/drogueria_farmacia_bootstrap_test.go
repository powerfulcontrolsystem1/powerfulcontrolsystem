package db

import (
	"strings"
	"testing"
)

func TestDefaultDrogueriaFarmaciaLicenciaPlans(t *testing.T) {
	plans := DefaultDrogueriaFarmaciaLicenciaPlans()
	if len(plans) != 8 {
		t.Fatalf("planes drogueria farmacia = %d, want 8", len(plans))
	}
	expectedValues := []float64{0, 1000, 60000, 110000, 200000, 600000, 1100000, 2200000}
	expectedDocs := []int{250, 250, 1000, 2000, 4000, 12000, 24000, 36000}
	for i, plan := range plans {
		if plan.Valor != expectedValues[i] {
			t.Fatalf("valor plan %d = %.0f, want %.0f", i, plan.Valor, expectedValues[i])
		}
		if plan.MaxDocumentosMensuales != expectedDocs[i] {
			t.Fatalf("documentos plan %d = %d, want %d", i, plan.MaxDocumentosMensuales, expectedDocs[i])
		}
		if plan.DuracionDias != 1 && plan.DuracionDias != 15 && plan.DuracionDias != 30 && plan.DuracionDias != 365 {
			t.Fatalf("duracion plan %d inesperada: %d", i, plan.DuracionDias)
		}
		if !strings.Contains(plan.ModulosHabilitados, "drogueria_farmacia") {
			t.Fatalf("plan %d no habilita drogueria_farmacia: %q", i, plan.ModulosHabilitados)
		}
	}
}
