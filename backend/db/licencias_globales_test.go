package db

import "testing"

func TestDefaultGlobalLicenciaPlans(t *testing.T) {
	plans := DefaultGlobalLicenciaPlans()
	if len(plans) != 4 {
		t.Fatalf("planes globales = %d, want 4", len(plans))
	}

	seenCodes := map[string]bool{}
	expectedDocs := []int{250, 1000, 2000, 4000}
	expectedCajas := []int{2, 2, 3, 4}
	expectedValues := []float64{0, 60000, 100000, 150000}
	for i, plan := range plans {
		if plan.Codigo == "" {
			t.Fatalf("plan %d sin codigo_funcion", i)
		}
		if seenCodes[plan.Codigo] {
			t.Fatalf("codigo_funcion repetido: %s", plan.Codigo)
		}
		seenCodes[plan.Codigo] = true
		if plan.MaxDocumentosMensuales != expectedDocs[i] {
			t.Fatalf("documentos plan %d = %d, want %d", i, plan.MaxDocumentosMensuales, expectedDocs[i])
		}
		if plan.MaxCajasSimultaneas != expectedCajas[i] {
			t.Fatalf("cajas plan %d = %d, want %d", i, plan.MaxCajasSimultaneas, expectedCajas[i])
		}
		if plan.Valor != expectedValues[i] {
			t.Fatalf("valor plan %d = %.2f, want %.2f", i, plan.Valor, expectedValues[i])
		}
	}
	if plans[0].DuracionDias != 15 {
		t.Fatalf("duracion prueba gratis = %d, want 15", plans[0].DuracionDias)
	}
	if plans[0].Valor != 0 {
		t.Fatalf("valor prueba gratis = %.2f, want 0", plans[0].Valor)
	}
}
