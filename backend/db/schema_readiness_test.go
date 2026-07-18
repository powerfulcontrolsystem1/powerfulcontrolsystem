package db

import "testing"

func TestEmpresaModulosFaltantesSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaModulosFaltantesSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaModulosFaltantesRequiredTablesAreAllowed(t *testing.T) {
	if len(empresaModulosFaltantesRequiredTables) == 0 {
		t.Fatal("required ERP table catalog must not be empty")
	}
	seen := make(map[string]struct{}, len(empresaModulosFaltantesRequiredTables))
	for _, table := range empresaModulosFaltantesRequiredTables {
		if _, ok := empresaGenericAllowedTables[table]; !ok {
			t.Fatalf("required table %q is not allowed by the ERP repository", table)
		}
		if _, duplicate := seen[table]; duplicate {
			t.Fatalf("required table %q appears more than once", table)
		}
		seen[table] = struct{}{}
	}
}

func TestEmpresaDocumentosTransaccionalesSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaDocumentosTransaccionalesSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaFacturacionElectronicaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaFacturacionElectronicaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaConfiguracionAvanzadaSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaConfiguracionAvanzadaSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaCarritosSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaCarritosSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}

func TestEmpresaFinanzasSchemaReadyRejectsNilDatabase(t *testing.T) {
	if err := EmpresaFinanzasSchemaReady(nil); err == nil {
		t.Fatal("expected nil database to be rejected")
	}
}
