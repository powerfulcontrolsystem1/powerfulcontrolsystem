package db

import "testing"

func TestBuildPlan107FixtureManifestIsDeterministicAndScoped(t *testing.T) {
	first, err := BuildPlan107FixtureManifest(12, "p107-qa-contador")
	if err != nil {
		t.Fatalf("primer manifiesto: %v", err)
	}
	second, err := BuildPlan107FixtureManifest(12, "P107-QA-CONTADOR")
	if err != nil {
		t.Fatalf("segundo manifiesto: %v", err)
	}
	if first.ManifestSHA256 == "" || first.ManifestSHA256 != second.ManifestSHA256 {
		t.Fatalf("huella no determinista: %q / %q", first.ManifestSHA256, second.ManifestSHA256)
	}
	if !first.StagingOnly || first.EmpresaID != 12 || len(first.Scenarios) < 8 {
		t.Fatalf("manifiesto incompleto: %#v", first)
	}
	for _, scenario := range first.Scenarios {
		if scenario.IdempotencyKey == "" || scenario.IdempotencyKey[:8] != "p107-qa-" {
			t.Fatalf("clave de idempotencia invalida: %#v", scenario)
		}
	}
}

func TestBuildPlan107FixtureManifestRejectsUnsafeScope(t *testing.T) {
	if _, err := BuildPlan107FixtureManifest(0, "P107-QA"); err == nil {
		t.Fatal("empresa invalida debe rechazarse")
	}
	if _, err := BuildPlan107FixtureManifest(12, "PRODUCCION"); err == nil {
		t.Fatal("run sin prefijo P107-QA debe rechazarse")
	}
}
