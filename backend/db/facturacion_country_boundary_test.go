package db

import (
	"context"
	"testing"
)

func TestFacturacionFiscalDefaultsNeverCrossCountries(t *testing.T) {
	for _, source := range []string{"CO", "EC", "PA", "CR", "AR", "VE", "", "ZZ"} {
		for _, target := range []string{"CO", "EC", "PA", "CR", "AR", "VE", "", "ZZ"} {
			want := source == target && source != "" && source != "ZZ"
			if got := facturacionCanHydrateCountry(target, source); got != want {
				t.Errorf("defaults %q -> %q: got %t want %t", source, target, got, want)
			}
		}
	}
	if !facturacionCanHydrateCountry(" co ", "CO") {
		t.Fatal("canonical country comparison rejected")
	}
}

func TestFacturacionForeignNumberingBlockedBeforeDatabaseIO(t *testing.T) {
	for _, country := range []string{"EC", "PA", "CR", "AR", "VE", "ZZ"} {
		got, err := PrepareFacturacionDocumentoLegalContext(context.Background(), nil, 12, country, "QA-NOT-ISSUED", 1, "USD")
		if err == nil || got != nil {
			t.Fatalf("%s used Colombian reservation: %#v %v", country, got, err)
		}
	}
}

func TestFacturacionNewCountryDoesNotFallbackToColombia(t *testing.T) {
	pais := paisFacturacionByCodigo("ZZ")
	if pais.Codigo != "ZZ" || pais.Codigo == "CO" {
		t.Fatalf("new country fell back to Colombia: %#v", pais)
	}
	cfg := defaultFacturacionConfig(12, "ZZ")
	if cfg.PaisCodigo != "ZZ" || cfg.Estado != "inactivo" || cfg.Ambiente != "sandbox" || cfg.Proveedor != "manual" {
		t.Fatalf("new country must be an isolated, blocked profile: %#v", cfg)
	}

	for _, country := range []string{"COL", "CO-invalid", "P1"} {
		if _, err := GetFacturacionElectronicaPaisConfigContext(context.Background(), nil, 12, country); err == nil {
			t.Fatalf("invalid country %q allowed for read", country)
		}
		if _, err := UpsertFacturacionElectronicaPaisConfigContext(context.Background(), nil, FacturacionElectronicaPaisConfig{EmpresaID: 12, PaisCodigo: country}); err == nil {
			t.Fatalf("invalid country %q allowed for write", country)
		}
	}
}

func TestFacturacionCountryChecklistDoesNotEnableFiscalEmission(t *testing.T) {
	if got := BuildFacturacionEcuadorChecklist(nil); got.EmisionHabilitada || len(got.Advertencias) == 0 {
		t.Fatalf("Ecuador checklist lacks explicit fiscal boundary: %#v", got)
	}
	if got := BuildFacturacionPanamaChecklist(nil); got.EmisionHabilitada || len(got.Advertencias) == 0 {
		t.Fatalf("Panama checklist lacks explicit fiscal boundary: %#v", got)
	}
}
