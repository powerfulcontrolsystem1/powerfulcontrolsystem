package db

import (
	"strings"
	"testing"
)

func TestDefaultConstructoraLicenciaPlans(t *testing.T) {
	plans := DefaultConstructoraLicenciaPlans()
	if len(plans) != 4 {
		t.Fatalf("planes constructora = %d, want 4", len(plans))
	}
	expectedDocs := []int{250, 1000, 2000, 4000}
	expectedValues := []float64{0, 60000, 100000, 150000}
	for i, plan := range plans {
		if plan.DuracionDias != map[bool]int{true: 15, false: 30}[i == 0] {
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
