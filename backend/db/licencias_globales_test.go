package db

import "testing"

func TestDefaultGlobalLicenciaPlans(t *testing.T) {
	plans := DefaultGlobalLicenciaPlans()
	if len(plans) != 7 {
		t.Fatalf("planes globales = %d, want 7", len(plans))
	}

	seenCodes := map[string]bool{}
	expectedDocs := []int{250, 1000, 2000, 4000, 12000, 24000, 36000}
	expectedCajas := []int{2, 2, 3, 4, 4, 5, 6}
	expectedValues := []float64{0, 60000, 110000, 200000, 600000, 1100000, 2200000}
	expectedNames := []string{
		"Prueba gratis 15 dias",
		"Plan mensual COP 60000",
		"Plan mensual COP 110000",
		"Plan mensual COP 200000",
		"Plan anual COP 600000",
		"Plan anual COP 1100000",
		"Plan anual COP 2200000",
	}
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
		if plan.Nombre != expectedNames[i] {
			t.Fatalf("nombre plan %d = %q, want %q", i, plan.Nombre, expectedNames[i])
		}
	}
	if plans[0].DuracionDias != 15 {
		t.Fatalf("duracion prueba gratis = %d, want 15", plans[0].DuracionDias)
	}
	if plans[0].Valor != 0 {
		t.Fatalf("valor prueba gratis = %.2f, want 0", plans[0].Valor)
	}
	for _, idx := range []int{4, 5, 6} {
		if plans[idx].DuracionDias != 365 {
			t.Fatalf("duracion plan anual %d = %d, want 365", idx, plans[idx].DuracionDias)
		}
	}
}

func TestIsGlobalLicenciaCatalogItem(t *testing.T) {
	item := Licencia{
		TipoID:        0,
		PaisCodigo:    "GLOBAL",
		CodigoFuncion: LicenciaCodigoBasicoGlobal,
		Activo:        1,
	}
	if !IsGlobalLicenciaCatalogItem(item) {
		t.Fatal("plan canonico global no fue reconocido como catalogo global")
	}
	item.EsAdicional = 1
	if IsGlobalLicenciaCatalogItem(item) {
		t.Fatal("addon no debe contar como plan canonico global")
	}
	item.EsAdicional = 0
	item.EmpresaID = 22
	if IsGlobalLicenciaCatalogItem(item) {
		t.Fatal("licencia asignada a empresa no debe contar como catalogo global")
	}
	item.EmpresaID = 0
	item.CodigoFuncion = "PLAN_ANTIGUO"
	if IsGlobalLicenciaCatalogItem(item) {
		t.Fatal("codigo antiguo no debe contar como catalogo global")
	}
}

func TestIsPowerfulSystemEmpresaNameRecognizesExistingCompanyAndLegacyTypo(t *testing.T) {
	valid := []string{
		"Powerful Control System",
		"powerful control system",
		"  Powerful   Control   System  ",
		"Powerful Control Systen",
	}
	for _, name := range valid {
		if !IsPowerfulSystemEmpresaName(name) {
			t.Fatalf("expected %q to be recognized as system company", name)
		}
	}
	if IsPowerfulSystemEmpresaName("Powerful Control System Demo") {
		t.Fatal("similar demo company must not be treated as the internal system company")
	}
}

func TestPowerfulSystemEmpresaInternalLicenseIsNotCommercialPlan(t *testing.T) {
	if IsGlobalLicenciaPlanCode(PowerfulSystemEmpresaLicenseCode) {
		t.Fatal("la licencia interna retirada no debe ser parte del catalogo comercial global")
	}
}
