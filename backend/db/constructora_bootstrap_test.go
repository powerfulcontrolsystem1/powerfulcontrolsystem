package db

import (
	"strings"
	"testing"
)

func TestDefaultConstructoraLicenciaPlans(t *testing.T) {
	plans := DefaultConstructoraLicenciaPlans()
	if len(plans) != 7 {
		t.Fatalf("planes constructora = %d, want 7", len(plans))
	}
	expectedDocs := []int{250, 1000, 2000, 4000, 12000, 24000, 36000}
	expectedValues := []float64{0, 60000, 110000, 200000, 600000, 1100000, 2200000}
	for i, plan := range plans {
		if plan.DuracionDias != 15 && plan.DuracionDias != 30 && plan.DuracionDias != 365 {
			t.Fatalf("duracion plan %d = %d", i, plan.DuracionDias)
		}
		if plan.MaxDocumentosMensuales != expectedDocs[i] {
			t.Fatalf("documentos plan %d = %d, want %d", i, plan.MaxDocumentosMensuales, expectedDocs[i])
		}
		if plan.Valor != expectedValues[i] {
			t.Fatalf("valor plan %d = %.2f, want %.2f", i, plan.Valor, expectedValues[i])
		}
		if !strings.Contains(plan.ModulosHabilitados, "aiu_construccion") || !strings.Contains(plan.ModulosHabilitados, "centros_costo") || !strings.Contains(plan.ModulosHabilitados, "contratos_obligaciones") {
			t.Fatalf("plan constructora sin modulos clave: %q", plan.ModulosHabilitados)
		}
	}
}

func TestDefaultConstructoraLicenciaModulesNoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, raw := range strings.Split(DefaultConstructoraLicenciaModules(), ",") {
		module := strings.TrimSpace(raw)
		if module == "" {
			t.Fatalf("modulo vacio en lista constructora")
		}
		if seen[module] {
			t.Fatalf("modulo duplicado: %s", module)
		}
		seen[module] = true
	}
}
